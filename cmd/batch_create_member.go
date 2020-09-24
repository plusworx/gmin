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
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"io"
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

var batchCrtMemberCmd = &cobra.Command{
	Use:     "group-members <group email address or id> -i <input file path or google sheet id>",
	Aliases: []string{"group-member", "grp-members", "grp-member", "gmembers", "gmember", "gmems", "gmem"},
	Args:    cobra.ExactArgs(1),
	Short:   "Creates a batch of group members",
	Long: `Creates a batch of group members where group member details are provided in a Google Sheet, CSV/JSON input file or piped JSON.
	
	Examples:	gmin batch-create group-members engineering@mycompany.com -i inputfile.json
			gmin bcrt gmems sales@mycompany.com -i inputfile.csv -f csv
			gmin bcrt gmem finance@mycompany.com -i 1odyAIp3jGspd3M4xeepxWD6aeQIUuHBgrZB2OHSu8MI -s 'Sheet1!A1:K25' -f gsheet
			
	The contents of the JSON file or piped input should look something like this:
	
	{"delivery_settings":"DIGEST","email":"kayden.yundt@mycompany.com","role":"MEMBER"}
	{"delivery_settings":"ALL_MAIL","email":"kenyatta.tillman@mycompany.com","role":"MANAGER"}
	{"delivery_settings":"DAILY","email":"keon.stroman@mycompany.com","role":"MEMBER"}
	
	CSV and Google sheets must have a header row with the following column names being the only ones that are valid:
	
	delivery_settings
	email [required]
	role

	The column names are case insensitive and can be in any order.`,
	RunE: doBatchCrtMember,
}

func doBatchCrtMember(cmd *cobra.Command, args []string) error {
	logger.Debugw("starting doBatchCrtMember()",
		"args", args)

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

	lwrFmt := strings.ToLower(format)

	ok := cmn.SliceContainsStr(cmn.ValidFileFormats, lwrFmt)
	if !ok {
		err = fmt.Errorf(cmn.ErrInvalidFileFormat, format)
		logger.Error(err)
		return err
	}

	groupKey := args[0]

	switch {
	case lwrFmt == "csv":
		err := btchCrtMemProcessCSV(ds, groupKey, inputFile)
		if err != nil {
			logger.Error(err)
			return err
		}
	case lwrFmt == "json":
		err := btchCrtMemProcessJSON(ds, groupKey, inputFile, scanner)
		if err != nil {
			logger.Error(err)
			return err
		}
	case lwrFmt == "gsheet":
		err := btchCrtMemProcessSheet(ds, groupKey, inputFile)
		if err != nil {
			logger.Error(err)
			return err
		}
	}
	logger.Debug("finished doBatchCrtMember()")
	return nil
}

func btchCreateJSONMember(ds *admin.Service, jsonData string) (*admin.Member, error) {
	logger.Debugw("starting btchCreateJSONMember()",
		"jsonData", jsonData)

	var (
		emptyVals = cmn.EmptyValues{}
		member    *admin.Member
	)

	member = new(admin.Member)
	jsonBytes := []byte(jsonData)

	if !json.Valid(jsonBytes) {
		logger.Error(cmn.ErrInvalidJSONAttr)
		return nil, errors.New(cmn.ErrInvalidJSONAttr)
	}

	outStr, err := cmn.ParseInputAttrs(jsonBytes)
	if err != nil {
		logger.Error(err)
		return nil, err
	}

	err = cmn.ValidateInputAttrs(outStr, mems.MemberAttrMap)
	if err != nil {
		logger.Error(err)
		return nil, err
	}

	err = json.Unmarshal(jsonBytes, &member)
	if err != nil {
		logger.Error(err)
		return nil, err
	}

	err = json.Unmarshal(jsonBytes, &emptyVals)
	if err != nil {
		logger.Error(err)
		return nil, err
	}
	if len(emptyVals.ForceSendFields) > 0 {
		member.ForceSendFields = emptyVals.ForceSendFields
	}
	logger.Debug("finished btchCreateJSONMember()")
	return member, nil
}

func btchInsertNewMembers(ds *admin.Service, groupKey string, members []*admin.Member) error {
	logger.Debugw("starting btchInsertNewMembers()",
		"groupKey", groupKey)

	wg := new(sync.WaitGroup)

	for _, m := range members {
		if m.Email == "" {
			err := errors.New(cmn.ErrNoMemberEmailAddress)
			logger.Error(err)
			return err
		}

		mic := ds.Members.Insert(groupKey, m)

		wg.Add(1)

		go btchMemInsertProcess(m, groupKey, wg, mic)
	}

	wg.Wait()

	logger.Debug("finished btchInsertNewMembers()")
	return nil
}

func btchMemInsertProcess(member *admin.Member, groupKey string, wg *sync.WaitGroup, mic *admin.MembersInsertCall) {
	logger.Debugw("starting btchMemInsertProcess()",
		"groupKey", groupKey)

	defer wg.Done()

	b := backoff.NewExponentialBackOff()
	b.MaxElapsedTime = 32 * time.Second

	err := backoff.Retry(func() error {
		var err error
		newMember, err := mic.Do()
		if err == nil {
			logger.Infof(cmn.InfoMemberCreated, newMember.Email, groupKey)
			fmt.Println(cmn.GminMessage(fmt.Sprintf(cmn.InfoMemberCreated, newMember.Email, groupKey)))
			return err
		}
		if !cmn.IsErrRetryable(err) {
			return backoff.Permanent(fmt.Errorf(cmn.ErrBatchMember, err.Error(), member.Email))
		}
		// Log the retries
		logger.Warnw(err.Error(),
			"retrying", b.GetElapsedTime().String(),
			"group", groupKey,
			"member", member.Email)
		return fmt.Errorf(cmn.ErrBatchMember, err.Error(), member.Email)
	}, b)
	if err != nil {
		// Log final error
		logger.Error(err)
		fmt.Println(cmn.GminMessage(err.Error()))
	}
	logger.Debug("finished btchMemInsertProcess()")
}

func btchCrtMemProcessCSV(ds *admin.Service, groupKey string, filePath string) error {
	logger.Debugw("starting btchCrtMemProcessCSV()",
		"filePath", filePath,
		"groupKey", groupKey)

	var (
		iSlice  []interface{}
		hdrMap  = map[int]string{}
		members []*admin.Member
	)

	csvfile, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer csvfile.Close()

	r := csv.NewReader(csvfile)

	count := 0
	for {
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			logger.Error(err)
			return err
		}

		if count == 0 {
			iSlice = make([]interface{}, len(record))
			for idx, value := range record {
				iSlice[idx] = value
			}
			hdrMap = cmn.ProcessHeader(iSlice)
			err = cmn.ValidateHeader(hdrMap, mems.MemberAttrMap)
			if err != nil {
				logger.Error(err)
				return err
			}
			count = count + 1
			continue
		}

		for idx, value := range record {
			iSlice[idx] = value
		}

		memVar, err := btchCrtProcessMember(hdrMap, iSlice)
		if err != nil {
			logger.Error(err)
			return err
		}

		members = append(members, memVar)

		count = count + 1
	}

	err = btchInsertNewMembers(ds, groupKey, members)
	if err != nil {
		logger.Error(err)
		return err
	}
	logger.Debug("finished btchCrtMemProcessCSV")
	return nil
}

func btchCrtMemProcessJSON(ds *admin.Service, groupKey, filePath string, scanner *bufio.Scanner) error {
	logger.Debugw("starting btchCrtMemProcessJSON()",
		"filePath", filePath,
		"groupKey", groupKey)

	var members []*admin.Member

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
		jsonData := scanner.Text()

		memVar, err := btchCreateJSONMember(ds, jsonData)
		if err != nil {
			logger.Error(err)
			return err
		}

		members = append(members, memVar)
	}
	err := scanner.Err()
	if err != nil {
		logger.Error(err)
		return err
	}

	err = btchInsertNewMembers(ds, groupKey, members)
	if err != nil {
		logger.Error(err)
		return err
	}
	logger.Debug("finished btchCrtMemProcessJSON()")
	return nil
}

func btchCrtMemProcessSheet(ds *admin.Service, groupKey string, sheetID string) error {
	logger.Debugw("starting btchCrtMemProcessSheet()",
		"groupKey", groupKey,
		"sheetID", sheetID)

	var members []*admin.Member

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

		memVar, err := btchCrtProcessMember(hdrMap, row)
		if err != nil {
			logger.Error(err)
			return err
		}

		members = append(members, memVar)
	}

	err = btchInsertNewMembers(ds, groupKey, members)
	if err != nil {
		logger.Error(err)
		return err
	}
	logger.Debug("finished btchCrtMemProcessSheet()")
	return nil
}

func btchCrtProcessMember(hdrMap map[int]string, grpData []interface{}) (*admin.Member, error) {
	logger.Debugw("starting btchCrtProcessMember()",
		"hdrMap", hdrMap)

	var member *admin.Member

	member = new(admin.Member)

	for idx, attr := range grpData {
		attrName := hdrMap[idx]

		switch {
		case attrName == "delivery_settings":
			sAttr := fmt.Sprintf("%v", attr)
			validDS, err := mems.ValidateDeliverySetting(sAttr)
			if err != nil {
				logger.Error(err)
				return nil, err
			}
			member.DeliverySettings = validDS
		case attrName == "email":
			member.Email = fmt.Sprintf("%v", attr)
		case attrName == "role":
			sAttr := fmt.Sprintf("%v", attr)
			validRole, err := mems.ValidateRole(sAttr)
			if err != nil {
				logger.Error(err)
				return nil, err
			}
			member.Role = validRole
		}
	}
	logger.Debug("finished btchCrtProcessMember()")
	return member, nil
}

func init() {
	batchCreateCmd.AddCommand(batchCrtMemberCmd)

	batchCrtMemberCmd.Flags().StringVarP(&inputFile, "inputfile", "i", "", "filepath to group member data file or sheet id")
	batchCrtMemberCmd.Flags().StringVarP(&format, "format", "f", "json", "user data file format")
	batchCrtMemberCmd.Flags().StringVarP(&sheetRange, "sheetrange", "s", "", "user data gsheet range")
}
