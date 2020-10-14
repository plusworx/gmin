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
	gmess "github.com/plusworx/gmin/utils/gminmessages"
	mems "github.com/plusworx/gmin/utils/members"
	"github.com/spf13/cobra"
	admin "google.golang.org/api/admin/directory/v1"
	sheet "google.golang.org/api/sheets/v4"
)

var batchDelMemberCmd = &cobra.Command{
	Use:     "group-members <group email address or id> [-i input file path]",
	Aliases: []string{"group-member", "grp-members", "grp-member", "gmembers", "gmember", "gmems", "gmem"},
	Args:    cobra.ExactArgs(1),
	Example: `gmin batch-delete group-members somegroup@mycompany.com -i inputfile.txt
gmin bdel gmems somegroup@mycompany.com -i inputfile.txt
gmin ls gmem mygroup@mycompany.co.uk -a email | jq '.members[] | .email' -r | ./gmin bdel gmem mygroup@mycompany.co.uk`,
	Short: "Deletes a batch of group members",
	Long: `Deletes a batch of group members where group member details are provided in a text input file or through a pipe.
			
The input file or piped in data should provide the group member email addresses, aliases or ids to be deleted on separate lines like this:

frank.castle@mycompany.com
bruce.wayne@mycompany.com
peter.parker@mycompany.com

An input Google sheet must have a header row with the following column names being the only ones that are valid:

memberKey [required]

The column name is case insensitive.`,
	RunE: doBatchDelMember,
}

func doBatchDelMember(cmd *cobra.Command, args []string) error {
	logger.Debugw("starting doBatchDelMember()",
		"args", args)

	var members []string

	group := args[0]

	ds, err := cmn.CreateDirectoryService(admin.AdminDirectoryGroupMemberScope)
	if err != nil {
		logger.Error(err)
		return err
	}

	inputFlgVal, err := cmd.Flags().GetString("input-file")
	if err != nil {
		logger.Error(err)
		return err
	}

	scanner, err := cmn.InputFromStdIn(inputFlgVal)
	if err != nil {
		logger.Error(err)
		return err
	}

	if inputFlgVal == "" && scanner == nil {
		err := errors.New(gmess.ERRNOINPUTFILE)
		logger.Error(err)
		return err
	}

	formatFlgVal, err := cmd.Flags().GetString("format")
	if err != nil {
		logger.Error(err)
		return err
	}
	lwrFmt := strings.ToLower(formatFlgVal)

	ok := cmn.SliceContainsStr(cmn.ValidFileFormats, lwrFmt)
	if !ok {
		err = fmt.Errorf(gmess.ERRINVALIDFILEFORMAT, formatFlgVal)
		logger.Error(err)
		return err
	}

	switch {
	case lwrFmt == "text":
		members, err = bdmProcessTextFile(ds, inputFlgVal, scanner)
		if err != nil {
			logger.Error(err)
			return err
		}
	case lwrFmt == "gsheet":
		rangeFlgVal, err := cmd.Flags().GetString("sheet-range")
		if err != nil {
			logger.Error(err)
			return err
		}

		members, err = bdmProcessGSheet(ds, inputFlgVal, rangeFlgVal)
		if err != nil {
			logger.Error(err)
			return err
		}
	default:
		return fmt.Errorf(gmess.ERRINVALIDFILEFORMAT, formatFlgVal)
	}

	err = bdmProcessDeletion(ds, group, members)
	if err != nil {
		logger.Error(err)
		return err
	}

	logger.Debug("finished doBatchDelMember()")
	return nil
}

func bdmDelete(wg *sync.WaitGroup, mdc *admin.MembersDeleteCall, member string, group string) {
	logger.Debugw("starting bdmDelete()",
		"group", group,
		"member", member)

	defer wg.Done()

	b := backoff.NewExponentialBackOff()
	b.MaxElapsedTime = 32 * time.Second

	err := backoff.Retry(func() error {
		var err error
		err = mdc.Do()
		if err == nil {
			logger.Infof(gmess.INFOMEMBERDELETED, member, group)
			fmt.Println(cmn.GminMessage(fmt.Sprintf(gmess.INFOMEMBERDELETED, member, group)))
			return err
		}
		if !cmn.IsErrRetryable(err) {
			return backoff.Permanent(fmt.Errorf(gmess.ERRBATCHMEMBER, err.Error(), member, group))
		}
		logger.Warnw(err.Error(),
			"retrying", b.GetElapsedTime().String(),
			"group", group,
			"member", member)
		return fmt.Errorf(gmess.ERRBATCHMEMBER, err.Error(), member, group)
	}, b)
	if err != nil {
		// Log final error
		logger.Error(err)
		fmt.Println(cmn.GminMessage(err.Error()))
	}
	logger.Debug("finished bdmDelete()")
}

func bdmFromFileFactory(hdrMap map[int]string, memberData []interface{}) (string, error) {
	logger.Debugw("starting bdmFromFileFactory()",
		"hdrMap", hdrMap)

	var member string

	for idx, val := range memberData {
		attrName := hdrMap[idx]
		attrVal := fmt.Sprintf("%v", val)

		if attrName == "memberKey" {
			member = attrVal
		}
	}
	logger.Debug("finished bdmFromFileFactory()")
	return member, nil
}

func bdmProcessDeletion(ds *admin.Service, group string, members []string) error {
	logger.Debug("starting bdmProcessDeletion()")

	wg := new(sync.WaitGroup)

	for _, member := range members {
		mdc := ds.Members.Delete(group, member)

		wg.Add(1)

		go bdmDelete(wg, mdc, member, group)
	}

	wg.Wait()

	logger.Debug("finished bdmProcessDeletion()")
	return nil
}

func bdmProcessGSheet(ds *admin.Service, sheetID string, sheetrange string) ([]string, error) {
	logger.Debugw("starting bdmProcessGSheet()",
		"sheetID", sheetID,
		"sheetrange", sheetrange)

	var members []string

	if sheetrange == "" {
		err := errors.New(gmess.ERRNOSHEETRANGE)
		logger.Error(err)
		return nil, err
	}

	ss, err := cmn.CreateSheetService(sheet.DriveReadonlyScope)
	if err != nil {
		logger.Error(err)
		return nil, err
	}

	ssvgc := ss.Spreadsheets.Values.Get(sheetID, sheetrange)
	sValRange, err := ssvgc.Do()
	if err != nil {
		logger.Error(err)
		return nil, err
	}

	if len(sValRange.Values) == 0 {
		err = fmt.Errorf(gmess.ERRNOSHEETDATAFOUND, sheetID, sheetrange)
		logger.Error(err)
		return nil, err
	}

	hdrMap := cmn.ProcessHeader(sValRange.Values[0])
	err = cmn.ValidateHeader(hdrMap, mems.MemberAttrMap)
	if err != nil {
		logger.Error(err)
		return nil, err
	}

	for idx, row := range sValRange.Values {
		if idx == 0 {
			continue
		}

		memVar, err := bdgFromFileFactory(hdrMap, row)
		if err != nil {
			logger.Error(err)
			return nil, err
		}

		members = append(members, memVar)
	}

	logger.Debug("finished bdmProcessGSheet()")
	return members, nil
}

func bdmProcessTextFile(ds *admin.Service, filePath string, scanner *bufio.Scanner) ([]string, error) {
	logger.Debugw("starting bdmProcessTextFile()",
		"filePath", filePath)

	var members []string

	if filePath != "" {
		file, err := os.Open(filePath)
		if err != nil {
			logger.Error(err)
			return nil, err
		}
		defer file.Close()
		scanner = bufio.NewScanner(file)
	}

	for scanner.Scan() {
		member := scanner.Text()
		members = append(members, member)
	}

	logger.Debug("finished bdmProcessTextFile()")
	return members, nil
}

func init() {
	batchDelCmd.AddCommand(batchDelMemberCmd)

	batchDelMemberCmd.Flags().StringVarP(&inputFile, "input-file", "i", "", "filepath to member data text file")
	batchDelMemberCmd.Flags().StringVarP(&delFormat, "format", "f", "text", "member data file format (text or gsheet)")
	batchDelMemberCmd.Flags().StringVarP(&sheetRange, "sheet-range", "s", "", "member data gsheet range")
}
