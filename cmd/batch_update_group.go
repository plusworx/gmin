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
	grps "github.com/plusworx/gmin/utils/groups"
	"github.com/spf13/cobra"
	admin "google.golang.org/api/admin/directory/v1"
	sheet "google.golang.org/api/sheets/v4"
)

var batchUpdGrpCmd = &cobra.Command{
	Use:     "groups -i <input file path>",
	Aliases: []string{"group", "grps", "grp"},
	Short:   "Updates a batch of groups",
	Long: `Updates a batch of groups where group details are provided in a Google Sheet or CSV/JSON input file.
	
	Examples:	gmin batch-update groups -i inputfile.json
			gmin bupd grps -i inputfile.csv -f csv
			gmin bupd grp -i 1odyAIp3jGspd3M4xeepxWD6aeQIUuHBgrZB2OHSu8MI -s 'Sheet1!A1:K25' -f gsheet
			  
	The contents of the JSON file should look something like this:
	
	{"groupKey":"034gixby5n7pqal","email":"testgroup@mycompany.com","name":"Testing","description":"This is a testing group for all your testing needs."}
	{"groupKey":"032hioqz3p4ulyk","email":"info@mycompany.com","name":"Information","description":"Group for responding to general queries."}
	{"groupKey":"045fijmz6w8nkqc","email":"webmaster@mycompany.com","name":"Webmaster","description":"Group for responding to website queries."}

	N.B. groupKey (group email address, alias or id) must be provided.
	
	CSV and Google sheets must have a header row with the following column names being the only ones that are valid:
	
	description
	email
	groupKey [required]
	name
	
	The column names are case insensitive and can be in any order.`,
	RunE: doBatchUpdGrp,
}

func doBatchUpdGrp(cmd *cobra.Command, args []string) error {
	ds, err := cmn.CreateDirectoryService(admin.AdminDirectoryGroupScope)
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

	switch {
	case lwrFmt == "csv":
		err := btchUpdGrpProcessCSV(ds, inputFile)
		if err != nil {
			return err
		}
	case lwrFmt == "json":
		err := btchUpdGrpProcessJSON(ds, inputFile)
		if err != nil {
			return err
		}
	case lwrFmt == "gsheet":
		err := btchUpdGrpProcessSheet(ds, inputFile)
		if err != nil {
			return err
		}
	}

	return nil
}

func btchUpdJSONGroup(ds *admin.Service, jsonData string) (*admin.Group, string, error) {
	var (
		emptyVals = cmn.EmptyValues{}
		group     *admin.Group
		grpKey    = grps.Key{}
	)

	group = new(admin.Group)
	jsonBytes := []byte(jsonData)

	if !json.Valid(jsonBytes) {
		return nil, "", errors.New("gmin: error - attribute string is not valid JSON")
	}

	outStr, err := cmn.ParseInputAttrs(jsonBytes)
	if err != nil {
		return nil, "", err
	}

	err = cmn.ValidateInputAttrs(outStr, grps.GroupAttrMap)
	if err != nil {
		return nil, "", err
	}

	err = json.Unmarshal(jsonBytes, &grpKey)
	if err != nil {
		return nil, "", err
	}

	if grpKey.GroupKey == "" {
		return nil, "", errors.New("gmin: error - groupKey must be included in the JSON input string")
	}

	err = json.Unmarshal(jsonBytes, &group)
	if err != nil {
		return nil, "", err
	}

	err = json.Unmarshal(jsonBytes, &emptyVals)
	if err != nil {
		return nil, "", err
	}
	if len(emptyVals.ForceSendFields) > 0 {
		group.ForceSendFields = emptyVals.ForceSendFields
	}

	return group, grpKey.GroupKey, nil
}

func btchUpdateGroups(ds *admin.Service, groups []*admin.Group, groupKeys []string) error {
	wg := new(sync.WaitGroup)

	for idx, g := range groups {
		guc := ds.Groups.Update(groupKeys[idx], g)

		wg.Add(1)

		go btchGrpUpdateProcess(g, wg, guc, groupKeys[idx])
	}

	wg.Wait()

	return nil
}

func btchGrpUpdateProcess(group *admin.Group, wg *sync.WaitGroup, guc *admin.GroupsUpdateCall, groupKey string) {
	defer wg.Done()

	b := backoff.NewExponentialBackOff()
	b.MaxElapsedTime = 32 * time.Second

	err := backoff.Retry(func() error {
		var err error
		_, err = guc.Do()
		if err == nil {
			fmt.Println(cmn.GminMessage("**** gmin: group " + groupKey + " updated ****"))
			return err
		}
		if !cmn.IsErrRetryable(err) {
			return backoff.Permanent(errors.New(cmn.GminMessage("gmin: error - " + err.Error() + " " + groupKey)))
		}
		return errors.New(cmn.GminMessage("gmin: error - " + err.Error() + " " + groupKey))
	}, b)
	if err != nil {
		fmt.Println(err)
	}
}

func btchUpdGrpProcessCSV(ds *admin.Service, filePath string) error {
	var (
		iSlice    []interface{}
		hdrMap    = map[int]string{}
		groupKeys []string
		groups    []*admin.Group
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
			err = cmn.ValidateHeader(hdrMap, grps.GroupAttrMap)
			if err != nil {
				return err
			}
			count = count + 1
			continue
		}

		for idx, value := range record {
			iSlice[idx] = value
		}

		grpVar, groupKey, err := btchUpdProcessGroup(hdrMap, iSlice)
		if err != nil {
			fmt.Println(err.Error())
		}

		groups = append(groups, grpVar)
		groupKeys = append(groupKeys, groupKey)

		count = count + 1
	}

	err = btchUpdateGroups(ds, groups, groupKeys)
	if err != nil {
		return err
	}
	return nil
}

func btchUpdGrpProcessJSON(ds *admin.Service, filePath string) error {
	var (
		groupKeys []string
		groups    []*admin.Group
	)

	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		jsonData := scanner.Text()

		grpVar, groupKey, err := btchUpdJSONGroup(ds, jsonData)
		if err != nil {
			return err
		}

		groupKeys = append(groupKeys, groupKey)
		groups = append(groups, grpVar)
	}
	err = scanner.Err()
	if err != nil {
		return err
	}

	err = btchUpdateGroups(ds, groups, groupKeys)
	if err != nil {
		return err
	}

	return nil
}

func btchUpdGrpProcessSheet(ds *admin.Service, sheetID string) error {
	var (
		groupKeys []string
		groups    []*admin.Group
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
	err = cmn.ValidateHeader(hdrMap, grps.GroupAttrMap)
	if err != nil {
		return err
	}

	for idx, row := range sValRange.Values {
		if idx == 0 {
			continue
		}

		grpVar, groupKey, err := btchUpdProcessGroup(hdrMap, row)
		if err != nil {
			return err
		}

		groupKeys = append(groupKeys, groupKey)
		groups = append(groups, grpVar)
	}

	err = btchUpdateGroups(ds, groups, groupKeys)
	if err != nil {
		return err
	}

	return nil
}

func btchUpdProcessGroup(hdrMap map[int]string, grpData []interface{}) (*admin.Group, string, error) {
	var (
		group    *admin.Group
		groupKey string
	)

	group = new(admin.Group)

	for idx, attr := range grpData {
		attrName := hdrMap[idx]

		switch {
		case attrName == "description":
			desc := fmt.Sprintf("%v", attr)
			group.Description = desc
			if desc == "" {
				group.ForceSendFields = append(group.ForceSendFields, "Description")
			}
		case attrName == "email":
			group.Email = fmt.Sprintf("%v", attr)
		case attrName == "name":
			group.Name = fmt.Sprintf("%v", attr)
		case attrName == "groupKey":
			groupKey = fmt.Sprintf("%v", attr)
		}
	}

	return group, groupKey, nil
}

func init() {
	batchUpdateCmd.AddCommand(batchUpdGrpCmd)

	batchUpdGrpCmd.Flags().StringVarP(&inputFile, "inputfile", "i", "", "filepath to group data file or sheet id")
	batchUpdGrpCmd.Flags().StringVarP(&format, "format", "f", "json", "user data file format")
	batchUpdGrpCmd.Flags().StringVarP(&sheetRange, "sheetrange", "s", "", "user data gsheet range")
}
