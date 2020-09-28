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
	logger.Debugw("starting doBatchUpdMember()",
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
		err := btchUpdMemProcessCSV(ds, groupKey, inputFile)
		if err != nil {
			logger.Error(err)
			return err
		}
	case lwrFmt == "json":
		err := btchUpdMemProcessJSON(ds, groupKey, inputFile, scanner)
		if err != nil {
			logger.Error(err)
			return err
		}
	case lwrFmt == "gsheet":
		err := btchUpdMemProcessSheet(ds, groupKey, inputFile)
		if err != nil {
			logger.Error(err)
			return err
		}
	}
	logger.Debug("finished doBatchUpdMember()")
	return nil
}

func btchUpdJSONMember(ds *admin.Service, jsonData string) (*admin.Member, string, error) {
	logger.Debugw("starting btchUpdJSONMember()",
		"jsonData", jsonData)

	var (
		member *admin.Member
		memKey = mems.Key{}
	)

	member = new(admin.Member)
	jsonBytes := []byte(jsonData)

	if !json.Valid(jsonBytes) {
		logger.Error(cmn.ErrInvalidJSONAttr)
		return nil, "", errors.New(cmn.ErrInvalidJSONAttr)
	}

	outStr, err := cmn.ParseInputAttrs(jsonBytes)
	if err != nil {
		logger.Error(err)
		return nil, "", err
	}

	err = cmn.ValidateInputAttrs(outStr, mems.MemberAttrMap)
	if err != nil {
		logger.Error(err)
		return nil, "", err
	}

	err = json.Unmarshal(jsonBytes, &memKey)
	if err != nil {
		logger.Error(err)
		return nil, "", err
	}

	if memKey.MemberKey == "" {
		err = errors.New(cmn.ErrNoJSONMemberKey)
		logger.Error(err)
		return nil, "", err
	}

	err = json.Unmarshal(jsonBytes, &member)
	if err != nil {
		logger.Error(err)
		return nil, "", err
	}
	logger.Debug("finished btchUpdJSONMember()")
	return member, memKey.MemberKey, nil
}

func btchUpdateMembers(ds *admin.Service, groupKey string, members []*admin.Member, memKeys []string) error {
	logger.Debugw("starting btchUpdateMembers()",
		"groupKey", groupKey,
		"memKeys", memKeys)

	wg := new(sync.WaitGroup)

	for idx, m := range members {
		muc := ds.Members.Update(groupKey, memKeys[idx], m)

		wg.Add(1)

		go btchMemUpdateProcess(m, groupKey, wg, muc, memKeys[idx])
	}

	wg.Wait()

	logger.Debug("finished btchUpdateMembers()")
	return nil
}

func btchMemUpdateProcess(member *admin.Member, groupKey string, wg *sync.WaitGroup, muc *admin.MembersUpdateCall, memKey string) {
	logger.Debugw("starting btchMemUpdateProcess()",
		"groupKey", groupKey,
		"memKey", memKey)

	defer wg.Done()

	b := backoff.NewExponentialBackOff()
	b.MaxElapsedTime = 32 * time.Second

	err := backoff.Retry(func() error {
		var err error
		_, err = muc.Do()
		if err == nil {
			logger.Infof(cmn.InfoMemberUpdated, memKey, groupKey)
			fmt.Println(cmn.GminMessage(fmt.Sprintf(cmn.InfoMemberUpdated, memKey, groupKey)))
			return err
		}
		if !cmn.IsErrRetryable(err) {
			return backoff.Permanent(fmt.Errorf(cmn.ErrBatchMember, err.Error(), memKey))
		}
		// Log the retries
		logger.Warnw(err.Error(),
			"retrying", b.GetElapsedTime().String(),
			"group", groupKey,
			"member", memKey)
		return fmt.Errorf(cmn.ErrBatchMember, err.Error(), memKey)
	}, b)
	if err != nil {
		// Log final error
		logger.Error(err)
		fmt.Println(cmn.GminMessage(err.Error()))
	}
	logger.Debug("finished btchMemUpdateProcess()")
}

func btchUpdMemProcessCSV(ds *admin.Service, groupKey string, filePath string) error {
	logger.Debugw("starting btchUpdMemProcessCSV()",
		"filePath", filePath,
		"groupKey", groupKey)

	var (
		iSlice  []interface{}
		hdrMap  = map[int]string{}
		memKeys []string
		members []*admin.Member
	)

	csvfile, err := os.Open(filePath)
	if err != nil {
		logger.Error(err)
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

		memVar, memKey, err := btchUpdProcessMember(hdrMap, iSlice)
		if err != nil {
			logger.Error(err)
			return err
		}

		members = append(members, memVar)
		memKeys = append(memKeys, memKey)

		count = count + 1
	}

	err = btchUpdateMembers(ds, groupKey, members, memKeys)
	if err != nil {
		logger.Error(err)
		return err
	}
	logger.Debug("finished btchUpdMemProcessCSV()")
	return nil
}

func btchUpdMemProcessJSON(ds *admin.Service, groupKey string, filePath string, scanner *bufio.Scanner) error {
	logger.Debugw("starting btchUpdMemProcessJSON()",
		"groupKey", groupKey,
		"filePath", filePath)

	var (
		memKeys []string
		members []*admin.Member
	)

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

		memVar, memKey, err := btchUpdJSONMember(ds, jsonData)
		if err != nil {
			logger.Error(err)
			return err
		}

		memKeys = append(memKeys, memKey)
		members = append(members, memVar)
	}
	err := scanner.Err()
	if err != nil {
		logger.Error(err)
		return err
	}

	err = btchUpdateMembers(ds, groupKey, members, memKeys)
	if err != nil {
		logger.Error(err)
		return err
	}
	logger.Debug("finished btchUpdMemProcessJSON()")
	return nil
}

func btchUpdMemProcessSheet(ds *admin.Service, groupKey string, sheetID string) error {
	logger.Debugw("starting btchUpdMemProcessSheet()",
		"groupKey", groupKey,
		"sheetID", sheetID)

	var (
		memKeys []string
		members []*admin.Member
	)

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

		memVar, memKey, err := btchUpdProcessMember(hdrMap, row)
		if err != nil {
			logger.Error(err)
			return err
		}

		memKeys = append(memKeys, memKey)
		members = append(members, memVar)
	}

	err = btchUpdateMembers(ds, groupKey, members, memKeys)
	if err != nil {
		logger.Error(err)
		return err
	}
	logger.Debug("finished btchUpdMemProcessSheet()")
	return nil
}

func btchUpdProcessMember(hdrMap map[int]string, memData []interface{}) (*admin.Member, string, error) {
	logger.Debugw("starting btchUpdProcessMember()",
		"hdrMap", hdrMap)

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
				logger.Error(err)
				return nil, "", err
			}
			member.DeliverySettings = validDS
		case attrName == "role":
			validRole, err := mems.ValidateRole(fmt.Sprintf("%v", attr))
			if err != nil {
				logger.Error(err)
				return nil, "", err
			}
			member.Role = validRole
		case attrName == "memberKey":
			memKey = fmt.Sprintf("%v", attr)
		}
	}
	logger.Debug("finished btchUpdProcessMember()")
	return member, memKey, nil
}

func init() {
	batchUpdateCmd.AddCommand(batchUpdMemberCmd)

	batchUpdMemberCmd.Flags().StringVarP(&inputFile, "inputfile", "i", "", "filepath to group member data file or sheet id")
	batchUpdMemberCmd.Flags().StringVarP(&format, "format", "f", "json", "user data file format")
	batchUpdMemberCmd.Flags().StringVarP(&sheetRange, "sheetrange", "s", "", "user data gsheet range")
}
