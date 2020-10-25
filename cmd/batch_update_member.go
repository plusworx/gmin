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

var batchUpdMemberCmd = &cobra.Command{
	Use:     "group-members <group email address, alias or id> -i <input file path>",
	Aliases: []string{"group-member", "grp-members", "grp-member", "gmembers", "gmember", "gmems", "gmem"},
	Args:    cobra.ExactArgs(1),
	Example: `gmin batch-update group-members sales@mycompany.com -i inputfile.json
gmin bupd gmems sales@mycompany.com -i inputfile.csv -f csv
gmin bupd gmem finance@mycompany.com -i 1odyAIp3jGspd3M4xeepxWD6aeQIUuHBgrZB2OHSu8MI -s 'Sheet1!A1:K25' -f gsheet`,
	Short: "Updates a batch of group members",
	Long: `Updates a batch of group members where group member details are provided in a Google Sheet, CSV/JSON input file or piped JSON.
			  
The contents of the JSON file or piped input should look something like this:

{"memberKey":"rudolph.brown@mycompany.com","delivery_settings":"DIGEST","role":"MANAGER"}
{"memberKey":"emily.bronte@mycompany.com","delivery_settings":"DAILY","role":"MEMBER"}
{"memberKey":"charles.dickens@mycompany.com","delivery_settings":"NONE","role":"MEMBER"}

N.B. memberKey (member email address, alias or id) must be provided.

CSV and Google sheets must have a header row with the following column names being the only ones that are valid:

delivery_settings
memberKey [required]
role

The column names are case insensitive and can be in any order.`,
	RunE: doBatchUpdMember,
}

func doBatchUpdMember(cmd *cobra.Command, args []string) error {
	lg.Debugw("starting doBatchUpdMember()",
		"args", args)
	defer lg.Debug("finished doBatchUpdMember()")

	var (
		memKeys []string
		members []*admin.Member
	)

	srv, err := cmn.CreateService(cmn.SRVTYPEADMIN, admin.AdminDirectoryGroupMemberScope)
	if err != nil {
		return err
	}
	ds := srv.(*admin.Service)

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
		memKeys, members, err = bumProcessCSVFile(ds, inputFlgVal)
		if err != nil {
			return err
		}
	case lwrFmt == "json":
		memKeys, members, err = bumProcessJSON(ds, inputFlgVal, scanner)
		if err != nil {
			return err
		}
	case lwrFmt == "gsheet":
		rangeFlgVal, err := cmd.Flags().GetString(flgnm.FLG_SHEETRANGE)
		if err != nil {
			lg.Error(err)
			return err
		}

		memKeys, members, err = bumProcessGSheet(ds, inputFlgVal, rangeFlgVal)
		if err != nil {
			return err
		}
	default:
		err = fmt.Errorf(gmess.ERR_INVALIDFILEFORMAT, formatFlgVal)
		lg.Error(err)
		return err
	}

	err = bumProcessObjects(ds, groupKey, members, memKeys)
	if err != nil {
		return err
	}

	return nil
}

func bumFromFileFactory(hdrMap map[int]string, memData []interface{}) (*admin.Member, string, error) {
	lg.Debugw("starting bumFromFileFactory()",
		"hdrMap", hdrMap)
	defer lg.Debug("finished bumFromFileFactory()")

	var (
		member *admin.Member
		memKey string
	)

	member = new(admin.Member)

	for idx, attr := range memData {
		attrName := hdrMap[idx]
		attrVal := fmt.Sprintf("%v", attr)

		switch {
		case attrName == "delivery_settings":
			validDS, err := mems.ValidateDeliverySetting(attrVal)
			if err != nil {
				return nil, "", err
			}
			member.DeliverySettings = validDS
		case attrName == "role":
			validRole, err := mems.ValidateRole(attrVal)
			if err != nil {
				return nil, "", err
			}
			member.Role = validRole
		case attrName == "memberKey":
			if attrVal == "" {
				err := fmt.Errorf(gmess.ERR_EMPTYSTRING, attrName)
				lg.Error(err)
				return nil, "", err
			}
			memKey = attrVal
		}
	}
	return member, memKey, nil
}

func bumFromJSONFactory(ds *admin.Service, jsonData string) (*admin.Member, string, error) {
	lg.Debugw("starting bumFromJSONFactory()",
		"jsonData", jsonData)
	defer lg.Debug("finished bumFromJSONFactory()")

	var (
		member *admin.Member
		memKey = mems.Key{}
	)

	member = new(admin.Member)
	jsonBytes := []byte(jsonData)

	if !json.Valid(jsonBytes) {
		err := errors.New(gmess.ERR_INVALIDJSONATTR)
		lg.Error(err)
		return nil, "", err
	}

	outStr, err := cmn.ParseInputAttrs(jsonBytes)
	if err != nil {
		return nil, "", err
	}

	err = cmn.ValidateInputAttrs(outStr, mems.MemberAttrMap)
	if err != nil {
		return nil, "", err
	}

	err = json.Unmarshal(jsonBytes, &memKey)
	if err != nil {
		lg.Error(err)
		return nil, "", err
	}

	if memKey.MemberKey == "" {
		err = errors.New(gmess.ERR_NOJSONMEMBERKEY)
		lg.Error(err)
		return nil, "", err
	}

	err = json.Unmarshal(jsonBytes, &member)
	if err != nil {
		lg.Error(err)
		return nil, "", err
	}
	return member, memKey.MemberKey, nil
}

func bumProcessCSVFile(ds *admin.Service, filePath string) ([]string, []*admin.Member, error) {
	lg.Debugw("starting bumProcessCSVFile()",
		"filePath", filePath)
	defer lg.Debug("finished bumProcessCSVFile()")

	var (
		iSlice  []interface{}
		hdrMap  = map[int]string{}
		memKeys []string
		members []*admin.Member
	)

	csvfile, err := os.Open(filePath)
	if err != nil {
		lg.Error(err)
		return nil, nil, err
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
			return nil, nil, err
		}

		if count == 0 {
			iSlice = make([]interface{}, len(record))
			for idx, value := range record {
				iSlice[idx] = value
			}
			hdrMap = cmn.ProcessHeader(iSlice)
			err = cmn.ValidateHeader(hdrMap, mems.MemberAttrMap)
			if err != nil {
				return nil, nil, err
			}
			count = count + 1
			continue
		}

		for idx, value := range record {
			iSlice[idx] = value
		}

		memVar, memKey, err := bumFromFileFactory(hdrMap, iSlice)
		if err != nil {
			return nil, nil, err
		}

		members = append(members, memVar)
		memKeys = append(memKeys, memKey)

		count = count + 1
	}

	return memKeys, members, nil
}

func bumProcessGSheet(ds *admin.Service, sheetID string, sheetrange string) ([]string, []*admin.Member, error) {
	lg.Debugw("starting bumProcessGSheet()",
		"sheetID", sheetID,
		"sheetrange", sheetrange)
	defer lg.Debug("finished bumProcessGSheet()")

	var (
		memKeys []string
		members []*admin.Member
	)

	if sheetrange == "" {
		err := errors.New(gmess.ERR_NOSHEETRANGE)
		lg.Error(err)
		return nil, nil, err
	}

	srv, err := cmn.CreateService(cmn.SRVTYPESHEET, sheet.DriveReadonlyScope)
	if err != nil {
		return nil, nil, err
	}
	ss := srv.(*sheet.Service)

	ssvgc := ss.Spreadsheets.Values.Get(sheetID, sheetrange)
	sValRange, err := ssvgc.Do()
	if err != nil {
		lg.Error(err)
		return nil, nil, err
	}

	if len(sValRange.Values) == 0 {
		err = fmt.Errorf(gmess.ERR_NOSHEETDATAFOUND, sheetID, sheetrange)
		lg.Error(err)
		return nil, nil, err
	}

	hdrMap := cmn.ProcessHeader(sValRange.Values[0])
	err = cmn.ValidateHeader(hdrMap, mems.MemberAttrMap)
	if err != nil {
		return nil, nil, err
	}

	for idx, row := range sValRange.Values {
		if idx == 0 {
			continue
		}

		memVar, memKey, err := bumFromFileFactory(hdrMap, row)
		if err != nil {
			return nil, nil, err
		}

		memKeys = append(memKeys, memKey)
		members = append(members, memVar)
	}

	return memKeys, members, nil
}

func bumProcessJSON(ds *admin.Service, filePath string, scanner *bufio.Scanner) ([]string, []*admin.Member, error) {
	lg.Debugw("starting bumProcessJSON()",
		"filePath", filePath)
	defer lg.Debug("finished bumProcessJSON()")

	var (
		memKeys []string
		members []*admin.Member
	)

	if filePath != "" {
		file, err := os.Open(filePath)
		if err != nil {
			lg.Error(err)
			return nil, nil, err
		}
		defer file.Close()

		scanner = bufio.NewScanner(file)
	}

	for scanner.Scan() {
		jsonData := scanner.Text()

		memVar, memKey, err := bumFromJSONFactory(ds, jsonData)
		if err != nil {
			return nil, nil, err
		}

		memKeys = append(memKeys, memKey)
		members = append(members, memVar)
	}
	err := scanner.Err()
	if err != nil {
		lg.Error(err)
		return nil, nil, err
	}

	return memKeys, members, nil
}

func bumProcessObjects(ds *admin.Service, groupKey string, members []*admin.Member, memKeys []string) error {
	lg.Debugw("starting bumProcessObjects()",
		"groupKey", groupKey,
		"memKeys", memKeys)
	defer lg.Debug("finished bumProcessObjects()")

	wg := new(sync.WaitGroup)

	for idx, m := range members {
		muc := ds.Members.Update(groupKey, memKeys[idx], m)

		wg.Add(1)

		go bumUpdate(m, groupKey, wg, muc, memKeys[idx])
	}

	wg.Wait()

	return nil
}

func bumUpdate(member *admin.Member, groupKey string, wg *sync.WaitGroup, muc *admin.MembersUpdateCall, memKey string) {
	lg.Debugw("starting bumUpdate()",
		"groupKey", groupKey,
		"memKey", memKey)
	defer lg.Debug("finished bumUpdate()")

	defer wg.Done()

	b := backoff.NewExponentialBackOff()
	b.MaxElapsedTime = 32 * time.Second

	err := backoff.Retry(func() error {
		var err error
		_, err = muc.Do()
		if err == nil {
			fmt.Println(cmn.GminMessage(fmt.Sprintf(gmess.INFO_MEMBERUPDATED, memKey, groupKey)))
			lg.Infof(gmess.INFO_MEMBERUPDATED, memKey, groupKey)
			return err
		}
		if !cmn.IsErrRetryable(err) {
			return backoff.Permanent(fmt.Errorf(gmess.ERR_BATCHMEMBER, err.Error(), memKey, groupKey))
		}
		// Log the retries
		lg.Warnw(err.Error(),
			"retrying", b.GetElapsedTime().String(),
			"group", groupKey,
			"member", memKey)
		return fmt.Errorf(gmess.ERR_BATCHMEMBER, err.Error(), memKey, groupKey)
	}, b)
	if err != nil {
		// Log final error
		lg.Error(err)
		fmt.Println(cmn.GminMessage(err.Error()))
	}
}

func init() {
	batchUpdateCmd.AddCommand(batchUpdMemberCmd)

	batchUpdMemberCmd.Flags().StringVarP(&inputFile, flgnm.FLG_INPUTFILE, "i", "", "filepath to group member data file or sheet id")
	batchUpdMemberCmd.Flags().StringVarP(&format, flgnm.FLG_FORMAT, "f", "json", "user data file format")
	batchUpdMemberCmd.Flags().StringVarP(&sheetRange, flgnm.FLG_SHEETRANGE, "s", "", "user data gsheet range")
}
