/*
Copyright © 2020 Chris Duncan <chris.duncan@plusworx.uk>

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/

package cmd

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/cenkalti/backoff/v4"
	cmn "github.com/plusworx/gmin/utils/common"
	flgnm "github.com/plusworx/gmin/utils/flagnames"
	gmess "github.com/plusworx/gmin/utils/gminmessages"
	grps "github.com/plusworx/gmin/utils/groups"
	lg "github.com/plusworx/gmin/utils/logging"
	"github.com/spf13/cobra"
	admin "google.golang.org/api/admin/directory/v1"
	sheet "google.golang.org/api/sheets/v4"
)

var batchDelGroupCmd = &cobra.Command{
	Use:     "groups [-i input file path]",
	Aliases: []string{"group", "grps", "grp"},
	Example: `gmin batch-delete groups -i inputfile.txt
gmin bdel grps -i inputfile.txt
gmin ls grp -q name:Test1* -a email | jq '.groups[] | .email' -r | gmin bdel grp`,
	Short: "Deletes a batch of groups",
	Long: `Deletes a batch of groups where group details are provided in a text input file or through a pipe.
			
The input file or piped in data should provide the group email addresses, aliases or ids to be deleted on separate lines like this:

oldsales@company.com
oldaccounts@company.com
unused_group@company.com

An input Google sheet must have a header row with the following column names being the only ones that are valid:

groupKey [required]

The column name is case insensitive.`,
	RunE: doBatchDelGroup,
}

func doBatchDelGroup(cmd *cobra.Command, args []string) error {
	lg.Debugw("starting doBatchDelGroup()",
		"args", args)
	defer lg.Debug("finished doBatchDelGroup()")

	var groups []string

	ds, err := cmn.CreateDirectoryService(admin.AdminDirectoryGroupScope)
	if err != nil {
		return err
	}

	inputFlgVal, err := cmd.Flags().GetString(flgnm.FLG_INPUTFILE)
	if err != nil {
		lg.Error(err)
		return err
	}

	scanner, err := cmn.InputFromStdIn(inputFlgVal)
	if err != nil {
		return err
	}

	if inputFlgVal == "" && scanner == nil {
		err := errors.New(gmess.ERR_NOINPUTFILE)
		lg.Error(err)
		return err
	}

	formatFlgVal, err := cmd.Flags().GetString(flgnm.FLG_FORMAT)
	if err != nil {
		lg.Error(err)
		return err
	}
	lwrFmt := strings.ToLower(formatFlgVal)

	ok := cmn.SliceContainsStr(cmn.ValidFileFormats, lwrFmt)
	if !ok {
		err = fmt.Errorf(gmess.ERR_INVALIDFILEFORMAT, formatFlgVal)
		lg.Error(err)
		return err
	}

	switch {
	case lwrFmt == "text":
		groups, err = bdgProcessTextFile(ds, inputFlgVal, scanner)
		if err != nil {
			return err
		}
	case lwrFmt == "gsheet":
		rangeFlgVal, err := cmd.Flags().GetString(flgnm.FLG_SHEETRANGE)
		if err != nil {
			return err
		}

		groups, err = bdgProcessGSheet(ds, inputFlgVal, rangeFlgVal)
		if err != nil {
			return err
		}
	default:
		err = fmt.Errorf(gmess.ERR_INVALIDFILEFORMAT, formatFlgVal)
		lg.Error(err)
		return err
	}

	err = bdgProcessDeletion(ds, groups)
	if err != nil {
		return err
	}

	return nil
}

func bdgDelete(wg *sync.WaitGroup, gdc *admin.GroupsDeleteCall, group string) {
	lg.Debugw("starting bdgDelete()",
		"group", group)
	defer lg.Debug("finished bdgDelete()")

	defer wg.Done()

	b := backoff.NewExponentialBackOff()
	b.MaxElapsedTime = 32 * time.Second

	err := backoff.Retry(func() error {
		var err error

		err = gdc.Do()
		if err == nil {
			fmt.Println(cmn.GminMessage(fmt.Sprintf(gmess.INFO_GROUPDELETED, group)))
			lg.Infof(gmess.INFO_GROUPDELETED, group)
			return err
		}
		if !cmn.IsErrRetryable(err) {
			return backoff.Permanent(fmt.Errorf(gmess.ERR_BATCHGROUP, err.Error(), group))
		}
		// Log the retries
		lg.Warnw(err.Error(),
			"retrying", b.GetElapsedTime().String(),
			"group", group)
		return fmt.Errorf(gmess.ERR_BATCHGROUP, err.Error(), group)
	}, b)
	if err != nil {
		// Log final error
		lg.Error(err)
		fmt.Println(cmn.GminMessage(err.Error()))
	}
}

func bdgFromFileFactory(hdrMap map[int]string, groupData []interface{}) (string, error) {
	lg.Debugw("starting bdgFromFileFactory()",
		"hdrMap", hdrMap)
	defer lg.Debug("finished bdgFromFileFactory()")

	var group string

	for idx, val := range groupData {
		attrName := hdrMap[idx]
		attrVal := fmt.Sprintf("%v", val)

		if attrName == "groupKey" {
			group = attrVal
		}
	}
	return group, nil
}

func bdgProcessDeletion(ds *admin.Service, groups []string) error {
	lg.Debug("starting bdgProcessDeletion()")
	defer lg.Debug("finished bdgProcessDeletion()")

	wg := new(sync.WaitGroup)

	for _, group := range groups {
		gdc := ds.Groups.Delete(group)

		wg.Add(1)

		go bdgDelete(wg, gdc, group)
	}

	wg.Wait()

	return nil
}

func bdgProcessGSheet(ds *admin.Service, sheetID string, sheetrange string) ([]string, error) {
	lg.Debugw("starting bdgProcessGSheet()",
		"sheetID", sheetID,
		"sheetrange", sheetrange)
	defer lg.Debug("finished bdgProcessGSheet()")

	var groups []string

	if sheetrange == "" {
		err := errors.New(gmess.ERR_NOSHEETRANGE)
		lg.Error(err)
		return nil, err
	}

	ss, err := cmn.CreateSheetService(sheet.DriveReadonlyScope)
	if err != nil {
		return nil, err
	}

	ssvgc := ss.Spreadsheets.Values.Get(sheetID, sheetrange)
	sValRange, err := ssvgc.Do()
	if err != nil {
		lg.Error(err)
		return nil, err
	}

	if len(sValRange.Values) == 0 {
		err = fmt.Errorf(gmess.ERR_NOSHEETDATAFOUND, sheetID, sheetrange)
		lg.Error(err)
		return nil, err
	}

	hdrMap := cmn.ProcessHeader(sValRange.Values[0])
	err = cmn.ValidateHeader(hdrMap, grps.GroupAttrMap)
	if err != nil {
		return nil, err
	}

	for idx, row := range sValRange.Values {
		if idx == 0 {
			continue
		}

		grpVar, err := bdgFromFileFactory(hdrMap, row)
		if err != nil {
			return nil, err
		}

		groups = append(groups, grpVar)
	}

	return groups, nil
}

func bdgProcessTextFile(ds *admin.Service, filePath string, scanner *bufio.Scanner) ([]string, error) {
	lg.Debugw("starting bdgProcessTextFile()",
		"filePath", filePath)
	defer lg.Debug("finished bdgProcessTextFile()")

	var groups []string

	if filePath != "" {
		file, err := os.Open(filePath)
		if err != nil {
			lg.Error(err)
			return nil, err
		}
		defer file.Close()
		scanner = bufio.NewScanner(file)
	}

	for scanner.Scan() {
		group := scanner.Text()
		groups = append(groups, group)
	}

	return groups, nil
}

func init() {
	batchDelCmd.AddCommand(batchDelGroupCmd)

	batchDelGroupCmd.Flags().StringVarP(&inputFile, flgnm.FLG_INPUTFILE, "i", "", "filepath to group data text file")
	batchDelGroupCmd.Flags().StringVarP(&delFormat, flgnm.FLG_FORMAT, "f", "text", "group data file format (text or gsheet)")
	batchDelGroupCmd.Flags().StringVarP(&sheetRange, flgnm.FLG_SHEETRANGE, "s", "", "group data gsheet range")
}
