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

var batchUpdMemberCmd = &cobra.Command{
	Use:     "group-members <group email address, alias or id> -i <input file path>",
	Aliases: []string{"group-member", "grp-members", "grp-member", "gmembers", "gmember", "gmems", "gmem"},
	Args:    cobra.ExactArgs(1),
	Short:   "Updates a batch of group members",
	Long: `Updates a batch of group members where group member details are provided in a Google Sheet or CSV/JSON input file.
	
	Examples:	gmin batch-update group-members sales@mycompany.com -i inputfile.json
			gmin bupd gmems sales@mycompany.com -i inputfile.csv -f csv
			gmin bupd gmem finance@mycompany.com -i 1odyAIp3jGspd3M4xeepxWD6aeQIUuHBgrZB2OHSu8MI -s 'Sheet1!A1:K25' -f gsheet
			  
	The contents of the JSON file should look something like this:
	
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
	ds, err := cmn.CreateDirectoryService(admin.AdminDirectoryGroupMemberScope)
	if err != nil {
		return err
	}

	if inputFile == "" {
		err := errors.New("gmin: error - must provide inputfile")
		return err
	}

	lwrFmt := strings.ToLower(format)

	ok := cmn.SliceContainsStr(cmn.ValidFileFormats, lwrFmt)
	if !ok {
		return fmt.Errorf("gmin: error - %v is not a valid file format", format)
	}

	groupKey := args[0]

	switch {
	case lwrFmt == "csv":
		err := btchUpdMemProcessCSV(ds, groupKey, inputFile)
		if err != nil {
			return err
		}
	case lwrFmt == "json":
		err := btchUpdMemProcessJSON(ds, groupKey, inputFile)
		if err != nil {
			return err
		}
	case lwrFmt == "gsheet":
		err := btchUpdMemProcessSheet(ds, groupKey, inputFile)
		if err != nil {
			return err
		}
	}

	return nil
}

func btchUpdJSONMember(ds *admin.Service, jsonData string) (*admin.Member, string, error) {
	var (
		member *admin.Member
		memKey = mems.Key{}
	)

	member = new(admin.Member)
	jsonBytes := []byte(jsonData)

	if !json.Valid(jsonBytes) {
		return nil, "", errors.New("gmin: error - attribute string is not valid JSON")
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
		return nil, "", err
	}

	if memKey.MemberKey == "" {
		return nil, "", errors.New("gmin: error - memberKey must be included in the JSON input string")
	}

	err = json.Unmarshal(jsonBytes, &member)
	if err != nil {
		return nil, "", err
	}

	return member, memKey.MemberKey, nil
}

func btchUpdateMembers(ds *admin.Service, groupKey string, members []*admin.Member, memKeys []string) error {
	wg := new(sync.WaitGroup)

	for idx, m := range members {
		muc := ds.Members.Update(groupKey, memKeys[idx], m)

		wg.Add(1)

		go btchMemUpdateProcess(m, groupKey, wg, muc, memKeys[idx])
	}

	wg.Wait()

	return nil
}

func btchMemUpdateProcess(member *admin.Member, groupKey string, wg *sync.WaitGroup, muc *admin.MembersUpdateCall, memKey string) {
	defer wg.Done()

	b := backoff.NewExponentialBackOff()
	b.MaxElapsedTime = 32 * time.Second

	err := backoff.Retry(func() error {
		var err error
		_, err = muc.Do()
		if err == nil {
			fmt.Println(cmn.GminMessage("**** gmin: member " + member.Email + " updated in group " + groupKey + " ****"))
			return err
		}
		if !cmn.IsErrRetryable(err) {
			return backoff.Permanent(errors.New(cmn.GminMessage("gmin: error - " + err.Error() + " " + memKey)))
		}
		return errors.New(cmn.GminMessage("gmin: error - " + err.Error() + " " + memKey))
	}, b)
	if err != nil {
		fmt.Println(err)
	}
}

func btchUpdMemProcessCSV(ds *admin.Service, groupKey string, filePath string) error {
	var (
		iSlice  []interface{}
		hdrMap  = map[int]string{}
		memKeys []string
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
				return err
			}
			count = count + 1
			continue
		}

		for idx, value := range record {
			iSlice[idx] = value
		}

		memVar, memKey, err := btchUpdProcessMember(hdrMap, iSlice)
		if err != nil {
			fmt.Println(err.Error())
		}

		members = append(members, memVar)
		memKeys = append(memKeys, memKey)

		count = count + 1
	}

	err = btchUpdateMembers(ds, groupKey, members, memKeys)
	if err != nil {
		return err
	}
	return nil
}

func btchUpdMemProcessJSON(ds *admin.Service, groupKey string, filePath string) error {
	var (
		memKeys []string
		members []*admin.Member
	)

	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		jsonData := scanner.Text()

		memVar, memKey, err := btchUpdJSONMember(ds, jsonData)
		if err != nil {
			return err
		}

		memKeys = append(memKeys, memKey)
		members = append(members, memVar)
	}
	err = scanner.Err()
	if err != nil {
		return err
	}

	err = btchUpdateMembers(ds, groupKey, members, memKeys)
	if err != nil {
		return err
	}

	return nil
}

func btchUpdMemProcessSheet(ds *admin.Service, groupKey string, sheetID string) error {
	var (
		memKeys []string
		members []*admin.Member
	)

	if sheetRange == "" {
		return errors.New("gmin: error - sheetrange must be provided")
	}

	ss, err := cmn.CreateSheetService(sheet.DriveReadonlyScope)
	if err != nil {
		return err
	}

	ssvgc := ss.Spreadsheets.Values.Get(sheetID, sheetRange)
	sValRange, err := ssvgc.Do()
	if err != nil {
		return err
	}

	if len(sValRange.Values) == 0 {
		return errors.New("gmin: error - no data found in sheet " + sheetID + " range: " + sheetRange)
	}

	hdrMap := cmn.ProcessHeader(sValRange.Values[0])
	err = cmn.ValidateHeader(hdrMap, mems.MemberAttrMap)
	if err != nil {
		return err
	}

	for idx, row := range sValRange.Values {
		if idx == 0 {
			continue
		}

		memVar, memKey, err := btchUpdProcessMember(hdrMap, row)
		if err != nil {
			return err
		}

		memKeys = append(memKeys, memKey)
		members = append(members, memVar)
	}

	err = btchUpdateMembers(ds, groupKey, members, memKeys)
	if err != nil {
		return err
	}

	return nil
}

func btchUpdProcessMember(hdrMap map[int]string, memData []interface{}) (*admin.Member, string, error) {
	var (
		member *admin.Member
		memKey string
	)

	member = new(admin.Member)

	for idx, attr := range memData {
		attrName := hdrMap[idx]

		switch {
		case attrName == "delivery_settings":
			validDS, err := mems.ValidateDeliverySetting(fmt.Sprintf("%v", attr))
			if err != nil {
				return nil, "", err
			}
			member.DeliverySettings = validDS
		case attrName == "role":
			validRole, err := mems.ValidateRole(fmt.Sprintf("%v", attr))
			if err != nil {
				return nil, "", err
			}
			member.Role = validRole
		case attrName == "memberKey":
			memKey = fmt.Sprintf("%v", attr)
		}
	}

	return member, memKey, nil
}

func init() {
	batchUpdateCmd.AddCommand(batchUpdMemberCmd)

	batchUpdMemberCmd.Flags().StringVarP(&inputFile, "inputfile", "i", "", "filepath to group member data file or sheet id")
	batchUpdMemberCmd.Flags().StringVarP(&format, "format", "f", "json", "user data file format")
	batchUpdMemberCmd.Flags().StringVarP(&sheetRange, "sheetrange", "s", "", "user data gsheet range")
}
