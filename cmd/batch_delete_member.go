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
	lg "github.com/plusworx/gmin/utils/logging"
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
	lg.Debugw("starting doBatchDelMember()",
		"args", args)
	defer lg.Debug("finished doBatchDelMember()")

	var members []string

	group := args[0]

	ds, err := cmn.CreateDirectoryService(admin.AdminDirectoryGroupMemberScope)
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
		members, err = bdmProcessTextFile(ds, inputFlgVal, scanner)
		if err != nil {
			return err
		}
	case lwrFmt == "gsheet":
		rangeFlgVal, err := cmd.Flags().GetString(flgnm.FLG_SHEETRANGE)
		if err != nil {
			return err
		}

		members, err = bdmProcessGSheet(ds, inputFlgVal, rangeFlgVal)
		if err != nil {
			return err
		}
	default:
		err = fmt.Errorf(gmess.ERR_INVALIDFILEFORMAT, formatFlgVal)
		lg.Error(err)
		return err
	}

	err = bdmProcessDeletion(ds, group, members)
	if err != nil {
		return err
	}

	return nil
}

func bdmDelete(wg *sync.WaitGroup, mdc *admin.MembersDeleteCall, member string, group string) {
	lg.Debugw("starting bdmDelete()",
		"group", group,
		"member", member)
	defer lg.Debug("finished bdmDelete()")

	defer wg.Done()

	b := backoff.NewExponentialBackOff()
	b.MaxElapsedTime = 32 * time.Second

	err := backoff.Retry(func() error {
		var err error
		err = mdc.Do()
		if err == nil {
			fmt.Println(cmn.GminMessage(fmt.Sprintf(gmess.INFO_MEMBERDELETED, member, group)))
			lg.Infof(gmess.INFO_MEMBERDELETED, member, group)
			return err
		}
		if !cmn.IsErrRetryable(err) {
			return backoff.Permanent(fmt.Errorf(gmess.ERR_BATCHMEMBER, err.Error(), member, group))
		}
		lg.Warnw(err.Error(),
			"retrying", b.GetElapsedTime().String(),
			"group", group,
			"member", member)
		return fmt.Errorf(gmess.ERR_BATCHMEMBER, err.Error(), member, group)
	}, b)
	if err != nil {
		// Log final error
		lg.Error(err)
		fmt.Println(cmn.GminMessage(err.Error()))
	}
}

func bdmFromFileFactory(hdrMap map[int]string, memberData []interface{}) (string, error) {
	lg.Debugw("starting bdmFromFileFactory()",
		"hdrMap", hdrMap)
	defer lg.Debug("finished bdmFromFileFactory()")

	var member string

	for idx, val := range memberData {
		attrName := hdrMap[idx]
		attrVal := fmt.Sprintf("%v", val)

		if attrName == "memberKey" {
			member = attrVal
		}
	}
	return member, nil
}

func bdmProcessDeletion(ds *admin.Service, group string, members []string) error {
	lg.Debug("starting bdmProcessDeletion()")
	defer lg.Debug("finished bdmProcessDeletion()")

	wg := new(sync.WaitGroup)

	for _, member := range members {
		mdc := ds.Members.Delete(group, member)

		wg.Add(1)

		go bdmDelete(wg, mdc, member, group)
	}

	wg.Wait()

	return nil
}

func bdmProcessGSheet(ds *admin.Service, sheetID string, sheetrange string) ([]string, error) {
	lg.Debugw("starting bdmProcessGSheet()",
		"sheetID", sheetID,
		"sheetrange", sheetrange)
	defer lg.Debug("finished bdmProcessGSheet()")

	var members []string

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
	err = cmn.ValidateHeader(hdrMap, mems.MemberAttrMap)
	if err != nil {
		return nil, err
	}

	for idx, row := range sValRange.Values {
		if idx == 0 {
			continue
		}

		memVar, err := bdgFromFileFactory(hdrMap, row)
		if err != nil {
			return nil, err
		}

		members = append(members, memVar)
	}

	return members, nil
}

func bdmProcessTextFile(ds *admin.Service, filePath string, scanner *bufio.Scanner) ([]string, error) {
	lg.Debugw("starting bdmProcessTextFile()",
		"filePath", filePath)
	defer lg.Debug("finished bdmProcessTextFile()")

	var members []string

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
		member := scanner.Text()
		members = append(members, member)
	}

	return members, nil
}

func init() {
	batchDelCmd.AddCommand(batchDelMemberCmd)

	batchDelMemberCmd.Flags().StringVarP(&inputFile, flgnm.FLG_INPUTFILE, "i", "", "filepath to member data text file")
	batchDelMemberCmd.Flags().StringVarP(&delFormat, flgnm.FLG_FORMAT, "f", "text", "member data file format (text or gsheet)")
	batchDelMemberCmd.Flags().StringVarP(&sheetRange, flgnm.FLG_SHEETRANGE, "s", "", "member data gsheet range")
}
