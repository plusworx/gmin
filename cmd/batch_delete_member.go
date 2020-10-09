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

	group := args[0]

	ds, err := cmn.CreateDirectoryService(admin.AdminDirectoryGroupMemberScope)
	if err != nil {
		logger.Error(err)
		return err
	}

	scanner, err := cmn.InputFromStdIn(inputFile)
	if err != nil {
		logger.Error(err)
		return err
	}

	if inputFile == "" && scanner == nil {
		err := errors.New(cmn.ErrNoInputFile)
		logger.Error(err)
		return err
	}

	lwrFmt := strings.ToLower(delFormat)

	ok := cmn.SliceContainsStr(cmn.ValidFileFormats, lwrFmt)
	if !ok {
		err = fmt.Errorf(cmn.ErrInvalidFileFormat, delFormat)
		logger.Error(err)
		return err
	}

	switch {
	case lwrFmt == "text":
		err := bdmProcessTextFile(ds, inputFile, group, scanner)
		if err != nil {
			logger.Error(err)
			return err
		}
	case lwrFmt == "gsheet":
		err := bdmProcessGSheet(ds, group, inputFile)
		if err != nil {
			logger.Error(err)
			return err
		}
	default:
		return fmt.Errorf(cmn.ErrInvalidFileFormat, format)
	}

	logger.Debug("finished doBatchDelMember()")
	return nil
}

func bdmDeleteObject(wg *sync.WaitGroup, mdc *admin.MembersDeleteCall, member string, group string) {
	logger.Debugw("starting bdmDeleteObject()",
		"group", group,
		"member", member)

	defer wg.Done()

	b := backoff.NewExponentialBackOff()
	b.MaxElapsedTime = 32 * time.Second

	err := backoff.Retry(func() error {
		var err error
		err = mdc.Do()
		if err == nil {
			logger.Infof(cmn.InfoMemberDeleted, member, group)
			fmt.Println(cmn.GminMessage(fmt.Sprintf(cmn.InfoMemberDeleted, member, group)))
			return err
		}
		if !cmn.IsErrRetryable(err) {
			return backoff.Permanent(fmt.Errorf(cmn.ErrBatchMember, err.Error(), member))
		}
		logger.Warnw(err.Error(),
			"retrying", b.GetElapsedTime().String(),
			"group", group,
			"member", member)
		return fmt.Errorf(cmn.ErrBatchMember, err.Error(), member)
	}, b)
	if err != nil {
		// Log final error
		logger.Error(err)
		fmt.Println(cmn.GminMessage(err.Error()))
	}
	logger.Debug("finished bdmDeleteObject()")
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

		go bdmDeleteObject(wg, mdc, member, group)
	}

	wg.Wait()

	logger.Debug("finished bdmProcessDeletion()")
	return nil
}

func bdmProcessGSheet(ds *admin.Service, group string, sheetID string) error {
	logger.Debugw("starting bdmProcessGSheet()",
		"sheetID", sheetID)

	var members []string

	if sheetRange == "" {
		err := errors.New(cmn.ErrNoSheetRange)
		logger.Error(err)
		return err
	}

	ss, err := cmn.CreateSheetService(sheet.DriveReadonlyScope)
	if err != nil {
		logger.Error(err)
		return err
	}

	ssvgc := ss.Spreadsheets.Values.Get(sheetID, sheetRange)
	sValRange, err := ssvgc.Do()
	if err != nil {
		logger.Error(err)
		return err
	}

	if len(sValRange.Values) == 0 {
		err = fmt.Errorf(cmn.ErrNoSheetDataFound, sheetID, sheetRange)
		logger.Error(err)
		return err
	}

	hdrMap := cmn.ProcessHeader(sValRange.Values[0])
	err = cmn.ValidateHeader(hdrMap, mems.MemberAttrMap)
	if err != nil {
		logger.Error(err)
		return err
	}

	for idx, row := range sValRange.Values {
		if idx == 0 {
			continue
		}

		memVar, err := bdgFromFileFactory(hdrMap, row)
		if err != nil {
			logger.Error(err)
			return err
		}

		members = append(members, memVar)
	}

	err = bdmProcessDeletion(ds, group, members)
	if err != nil {
		logger.Error(err)
		return err
	}
	logger.Debug("finished bdmProcessGSheet()")
	return nil
}

func bdmProcessTextFile(ds *admin.Service, filePath string, group string, scanner *bufio.Scanner) error {
	logger.Debugw("starting bdmProcessTextFile()",
		"filePath", filePath)

	var members []string

	if filePath != "" {
		file, err := os.Open(filePath)
		if err != nil {
			logger.Error(err)
			return err
		}
		defer file.Close()
		scanner = bufio.NewScanner(file)
	}

	for scanner.Scan() {
		member := scanner.Text()
		members = append(members, member)
	}

	err := bdmProcessDeletion(ds, group, members)
	if err != nil {
		logger.Error(err)
		return err
	}

	logger.Debug("finished bdmProcessTextFile()")
	return nil
}

func init() {
	batchDelCmd.AddCommand(batchDelMemberCmd)

	batchDelMemberCmd.Flags().StringVarP(&inputFile, "input-file", "i", "", "filepath to member data text file")
	batchDelMemberCmd.Flags().StringVarP(&delFormat, "format", "f", "text", "member data file format (text or gsheet)")
	batchDelMemberCmd.Flags().StringVarP(&sheetRange, "sheet-range", "s", "", "member data gsheet range")
}
