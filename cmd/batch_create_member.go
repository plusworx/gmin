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
	flgnm "github.com/plusworx/gmin/utils/flagnames"
	gmess "github.com/plusworx/gmin/utils/gminmessages"
	lg "github.com/plusworx/gmin/utils/logging"
	mems "github.com/plusworx/gmin/utils/members"
	"github.com/spf13/cobra"
	admin "google.golang.org/api/admin/directory/v1"
	sheet "google.golang.org/api/sheets/v4"
)

var batchCrtMemberCmd = &cobra.Command{
	Use:     "group-members <group email address or id> -i <input file path or google sheet id>",
	Aliases: []string{"group-member", "grp-members", "grp-member", "gmembers", "gmember", "gmems", "gmem"},
	Args:    cobra.ExactArgs(1),
	Example: `gmin batch-create group-members engineering@mycompany.com -i inputfile.json
gmin bcrt gmems sales@mycompany.com -i inputfile.csv -f csv
gmin bcrt gmem finance@mycompany.com -i 1odyAIp3jGspd3M4xeepxWD6aeQIUuHBgrZB2OHSu8MI -s 'Sheet1!A1:K25' -f gsheet`,
	Short: "Creates a batch of group members",
	Long: `Creates a batch of group members where group member details are provided in a Google Sheet, CSV/JSON input file or piped JSON.
			
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
	lg.Debugw("starting doBatchCrtMember()",
		"args", args)
	defer lg.Debug("finished doBatchCrtMember()")

	var members []*admin.Member

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

	groupKey := args[0]

	switch {
	case lwrFmt == "csv":
		members, err = bcmProcessCSVFile(ds, inputFlgVal)
		if err != nil {
			return err
		}
	case lwrFmt == "json":
		members, err = bcmProcessJSON(ds, inputFlgVal, scanner)
		if err != nil {
			return err
		}
	case lwrFmt == "gsheet":
		rangeFlgVal, err := cmd.Flags().GetString(flgnm.FLG_SHEETRANGE)
		if err != nil {
			return err
		}

		members, err = bcmProcessGSheet(ds, inputFlgVal, rangeFlgVal)
		if err != nil {
			return err
		}
	default:
		err = fmt.Errorf(gmess.ERR_INVALIDFILEFORMAT, formatFlgVal)
		lg.Error(err)
		return err
	}

	err = bcmProcessObjects(ds, groupKey, members)
	if err != nil {
		return err
	}

	return nil
}

func bcmCreate(member *admin.Member, groupKey string, wg *sync.WaitGroup, mic *admin.MembersInsertCall) {
	lg.Debugw("starting bcmCreate()",
		"groupKey", groupKey)
	defer lg.Debug("finished bcmCreate()")

	defer wg.Done()

	b := backoff.NewExponentialBackOff()
	b.MaxElapsedTime = 32 * time.Second

	err := backoff.Retry(func() error {
		var err error
		newMember, err := mic.Do()
		if err == nil {
			lg.Infof(gmess.INFO_MEMBERCREATED, newMember.Email, groupKey)
			fmt.Println(cmn.GminMessage(fmt.Sprintf(gmess.INFO_MEMBERCREATED, newMember.Email, groupKey)))
			return err
		}
		if !cmn.IsErrRetryable(err) {
			return backoff.Permanent(fmt.Errorf(gmess.ERR_BATCHMEMBER, err.Error(), member.Email, groupKey))
		}
		// Log the retries
		lg.Warnw(err.Error(),
			"retrying", b.GetElapsedTime().String(),
			"group", groupKey,
			"member", member.Email)
		return fmt.Errorf(gmess.ERR_BATCHMEMBER, err.Error(), member.Email, groupKey)
	}, b)
	if err != nil {
		// Log final error
		lg.Error(err)
		fmt.Println(cmn.GminMessage(err.Error()))
	}
}

func bcmFromFileFactory(hdrMap map[int]string, grpData []interface{}) (*admin.Member, error) {
	lg.Debugw("starting bcmFromFileFactory()",
		"hdrMap", hdrMap)
	defer lg.Debug("finished bcmFromFileFactory()")

	var member *admin.Member

	member = new(admin.Member)

	for idx, attr := range grpData {
		attrName := hdrMap[idx]
		attrVal := fmt.Sprintf("%v", attr)

		switch {
		case attrName == "delivery_settings":
			validDS, err := mems.ValidateDeliverySetting(attrVal)
			if err != nil {
				return nil, err
			}
			member.DeliverySettings = validDS
		case attrName == "email":
			if attrVal == "" {
				err := fmt.Errorf(gmess.ERR_EMPTYSTRING, attrName)
				lg.Error(err)
				return nil, err
			}
			member.Email = attrVal
		case attrName == "role":
			validRole, err := mems.ValidateRole(attrVal)
			if err != nil {
				return nil, err
			}
			member.Role = validRole
		}
	}
	return member, nil
}

func bcmFromJSONFactory(ds *admin.Service, jsonData string) (*admin.Member, error) {
	lg.Debugw("starting bcmFromJSONFactory()",
		"jsonData", jsonData)
	defer lg.Debug("finished bcmFromJSONFactory()")

	var (
		emptyVals = cmn.EmptyValues{}
		member    *admin.Member
	)

	member = new(admin.Member)
	jsonBytes := []byte(jsonData)

	if !json.Valid(jsonBytes) {
		lg.Error(gmess.ERR_INVALIDJSONATTR)
		return nil, errors.New(gmess.ERR_INVALIDJSONATTR)
	}

	outStr, err := cmn.ParseInputAttrs(jsonBytes)
	if err != nil {
		return nil, err
	}

	err = cmn.ValidateInputAttrs(outStr, mems.MemberAttrMap)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(jsonBytes, &member)
	if err != nil {
		lg.Error(err)
		return nil, err
	}

	err = json.Unmarshal(jsonBytes, &emptyVals)
	if err != nil {
		lg.Error(err)
		return nil, err
	}
	if len(emptyVals.ForceSendFields) > 0 {
		member.ForceSendFields = emptyVals.ForceSendFields
	}
	return member, nil
}

func bcmProcessCSVFile(ds *admin.Service, filePath string) ([]*admin.Member, error) {
	lg.Debugw("starting bcmProcessCSVFile()",
		"filePath", filePath)
	defer lg.Debug("finished bcmProcessCSVFile")

	var (
		iSlice  []interface{}
		hdrMap  = map[int]string{}
		members []*admin.Member
	)

	csvfile, err := os.Open(filePath)
	if err != nil {
		return nil, err
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
			lg.Error(err)
			return nil, err
		}

		if count == 0 {
			iSlice = make([]interface{}, len(record))
			for idx, value := range record {
				iSlice[idx] = value
			}
			hdrMap = cmn.ProcessHeader(iSlice)
			err = cmn.ValidateHeader(hdrMap, mems.MemberAttrMap)
			if err != nil {
				return nil, err
			}
			count = count + 1
			continue
		}

		for idx, value := range record {
			iSlice[idx] = value
		}

		memVar, err := bcmFromFileFactory(hdrMap, iSlice)
		if err != nil {
			return nil, err
		}

		members = append(members, memVar)

		count = count + 1
	}

	return members, nil
}

func bcmProcessGSheet(ds *admin.Service, sheetID string, sheetrange string) ([]*admin.Member, error) {
	lg.Debugw("starting bcmProcessGSheet()",
		"sheetID", sheetID,
		"sheetrange", sheetrange)
	defer lg.Debug("finished bcmProcessGSheet()")

	var members []*admin.Member

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

		memVar, err := bcmFromFileFactory(hdrMap, row)
		if err != nil {
			return nil, err
		}

		members = append(members, memVar)
	}

	return members, nil
}

func bcmProcessJSON(ds *admin.Service, filePath string, scanner *bufio.Scanner) ([]*admin.Member, error) {
	lg.Debugw("starting bcmProcessJSON()",
		"filePath", filePath)
	defer lg.Debug("finished bcmProcessJSON()")

	var members []*admin.Member

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
		jsonData := scanner.Text()

		memVar, err := bcmFromJSONFactory(ds, jsonData)
		if err != nil {
			return nil, err
		}

		members = append(members, memVar)
	}
	err := scanner.Err()
	if err != nil {
		lg.Error(err)
		return nil, err
	}

	return members, nil
}

func bcmProcessObjects(ds *admin.Service, groupKey string, members []*admin.Member) error {
	lg.Debugw("starting bcmProcessObjects()",
		"groupKey", groupKey)
	defer lg.Debug("finished bcmProcessObjects()")

	wg := new(sync.WaitGroup)

	for _, m := range members {
		if m.Email == "" {
			err := errors.New(gmess.ERR_NOMEMBEREMAILADDRESS)
			lg.Error(err)
			return err
		}

		mic := ds.Members.Insert(groupKey, m)

		wg.Add(1)

		go bcmCreate(m, groupKey, wg, mic)
	}

	wg.Wait()

	return nil
}

func init() {
	batchCreateCmd.AddCommand(batchCrtMemberCmd)

	batchCrtMemberCmd.Flags().StringVarP(&inputFile, flgnm.FLG_INPUTFILE, "i", "", "filepath to group member data file or sheet id")
	batchCrtMemberCmd.Flags().StringVarP(&format, flgnm.FLG_FORMAT, "f", "json", "user data file format")
	batchCrtMemberCmd.Flags().StringVarP(&sheetRange, flgnm.FLG_SHEETRANGE, "s", "", "user data gsheet range")
}
