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
	gmess "github.com/plusworx/gmin/utils/gminmessages"
	grps "github.com/plusworx/gmin/utils/groups"
	"github.com/spf13/cobra"
	admin "google.golang.org/api/admin/directory/v1"
	sheet "google.golang.org/api/sheets/v4"
)

var batchUpdGrpCmd = &cobra.Command{
	Use:     "groups -i <input file path>",
	Aliases: []string{"group", "grps", "grp"},
	Example: `gmin batch-update groups -i inputfile.json
gmin bupd grps -i inputfile.csv -f csv
gmin bupd grp -i 1odyAIp3jGspd3M4xeepxWD6aeQIUuHBgrZB2OHSu8MI -s 'Sheet1!A1:K25' -f gsheet`,
	Short: "Updates a batch of groups",
	Long: `Updates a batch of groups where group details are provided in a Google Sheet, CSV/JSON input file or piped JSON.
			  
The contents of the JSON file or piped input should look something like this:

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
	logger.Debugw("starting doBatchUpdGrp()",
		"args", args)

	var (
		groupKeys []string
		groups    []*admin.Group
	)

	ds, err := cmn.CreateDirectoryService(admin.AdminDirectoryGroupScope)
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
	case lwrFmt == "csv":
		groupKeys, groups, err = bugProcessCSVFile(ds, inputFlgVal)
		if err != nil {
			logger.Error(err)
			return err
		}
	case lwrFmt == "json":
		groupKeys, groups, err = bugProcessJSON(ds, inputFlgVal, scanner)
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

		groupKeys, groups, err = bugProcessGSheet(ds, inputFlgVal, rangeFlgVal)
		if err != nil {
			logger.Error(err)
			return err
		}
	default:
		return fmt.Errorf(gmess.ERRINVALIDFILEFORMAT, formatFlgVal)
	}

	err = bugProcessObjects(ds, groups, groupKeys)
	if err != nil {
		logger.Error(err)
		return err
	}

	logger.Debug("finished doBatchUpdGrp()")
	return nil
}

func bugFromFileFactory(hdrMap map[int]string, grpData []interface{}) (*admin.Group, string, error) {
	logger.Debugw("starting bugFromFileFactory()",
		"hdrMap", hdrMap)

	var (
		group    *admin.Group
		groupKey string
	)

	group = new(admin.Group)

	for idx, attr := range grpData {
		attrName := hdrMap[idx]
		attrVal := fmt.Sprintf("%v", attr)

		switch {
		case attrName == "description":
			group.Description = attrVal
			if attrVal == "" {
				group.ForceSendFields = append(group.ForceSendFields, "Description")
			}
		case attrName == "email":
			if attrVal == "" {
				err := fmt.Errorf(gmess.ERREMPTYSTRING, attrName)
				return nil, "", err
			}
			group.Email = attrVal
		case attrName == "name":
			if attrVal == "" {
				err := fmt.Errorf(gmess.ERREMPTYSTRING, attrName)
				return nil, "", err
			}
			group.Name = attrVal
		case attrName == "groupKey":
			if attrVal == "" {
				err := fmt.Errorf(gmess.ERREMPTYSTRING, attrName)
				return nil, "", err
			}
			groupKey = attrVal
		}
	}
	logger.Debug("finished bugFromFileFactory()")
	return group, groupKey, nil
}

func bugFromJSONFactory(ds *admin.Service, jsonData string) (*admin.Group, string, error) {
	logger.Debugw("starting bugFromJSONFactory()",
		"jsonData", jsonData)

	var (
		emptyVals = cmn.EmptyValues{}
		group     *admin.Group
		grpKey    = grps.Key{}
	)

	group = new(admin.Group)
	jsonBytes := []byte(jsonData)

	if !json.Valid(jsonBytes) {
		logger.Error(gmess.ERRINVALIDJSONATTR)
		return nil, "", errors.New(gmess.ERRINVALIDJSONATTR)
	}

	outStr, err := cmn.ParseInputAttrs(jsonBytes)
	if err != nil {
		logger.Error(err)
		return nil, "", err
	}

	err = cmn.ValidateInputAttrs(outStr, grps.GroupAttrMap)
	if err != nil {
		logger.Error(err)
		return nil, "", err
	}

	err = json.Unmarshal(jsonBytes, &grpKey)
	if err != nil {
		logger.Error(err)
		return nil, "", err
	}

	if grpKey.GroupKey == "" {
		err = errors.New(gmess.ERRNOJSONGROUPKEY)
		logger.Error(err)
		return nil, "", err
	}

	err = json.Unmarshal(jsonBytes, &group)
	if err != nil {
		logger.Error(err)
		return nil, "", err
	}

	err = json.Unmarshal(jsonBytes, &emptyVals)
	if err != nil {
		logger.Error(err)
		return nil, "", err
	}
	if len(emptyVals.ForceSendFields) > 0 {
		group.ForceSendFields = emptyVals.ForceSendFields
	}
	logger.Debug("finished bugFromJSONFactory()")
	return group, grpKey.GroupKey, nil
}

func bugProcessCSVFile(ds *admin.Service, filePath string) ([]string, []*admin.Group, error) {
	logger.Debugw("starting bugProcessCSVFile()",
		"filePath", filePath)

	var (
		iSlice    []interface{}
		hdrMap    = map[int]string{}
		groupKeys []string
		groups    []*admin.Group
	)

	csvfile, err := os.Open(filePath)
	if err != nil {
		logger.Error(err)
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
			logger.Error(err)
			return nil, nil, err
		}

		if count == 0 {
			iSlice = make([]interface{}, len(record))
			for idx, value := range record {
				iSlice[idx] = value
			}
			hdrMap = cmn.ProcessHeader(iSlice)
			err = cmn.ValidateHeader(hdrMap, grps.GroupAttrMap)
			if err != nil {
				logger.Error(err)
				return nil, nil, err
			}
			count = count + 1
			continue
		}

		for idx, value := range record {
			iSlice[idx] = value
		}

		grpVar, groupKey, err := bugFromFileFactory(hdrMap, iSlice)
		if err != nil {
			logger.Error(err)
			return nil, nil, err
		}

		groups = append(groups, grpVar)
		groupKeys = append(groupKeys, groupKey)

		count = count + 1
	}

	logger.Debug("finished bugProcessCSVFile()")
	return groupKeys, groups, nil
}

func bugProcessGSheet(ds *admin.Service, sheetID string, sheetrange string) ([]string, []*admin.Group, error) {
	logger.Debugw("starting bugProcessGSheet()",
		"sheetID", sheetID,
		"sheetrange", sheetrange)

	var (
		groupKeys []string
		groups    []*admin.Group
	)

	if sheetrange == "" {
		err := errors.New(gmess.ERRNOSHEETRANGE)
		logger.Error(err)
		return nil, nil, err
	}

	ss, err := cmn.CreateSheetService(sheet.DriveReadonlyScope)
	if err != nil {
		logger.Error(err)
		return nil, nil, err
	}

	ssvgc := ss.Spreadsheets.Values.Get(sheetID, sheetrange)
	sValRange, err := ssvgc.Do()
	if err != nil {
		logger.Error(err)
		return nil, nil, err
	}

	if len(sValRange.Values) == 0 {
		err = fmt.Errorf(gmess.ERRNOSHEETDATAFOUND, sheetID, sheetrange)
		logger.Error(err)
		return nil, nil, err
	}

	hdrMap := cmn.ProcessHeader(sValRange.Values[0])
	err = cmn.ValidateHeader(hdrMap, grps.GroupAttrMap)
	if err != nil {
		logger.Error(err)
		return nil, nil, err
	}

	for idx, row := range sValRange.Values {
		if idx == 0 {
			continue
		}

		grpVar, groupKey, err := bugFromFileFactory(hdrMap, row)
		if err != nil {
			logger.Error(err)
			return nil, nil, err
		}

		groupKeys = append(groupKeys, groupKey)
		groups = append(groups, grpVar)
	}

	logger.Debug("finished bugProcessGSheet()")
	return groupKeys, groups, nil
}

func bugProcessJSON(ds *admin.Service, filePath string, scanner *bufio.Scanner) ([]string, []*admin.Group, error) {
	logger.Debugw("starting bugProcessJSON()",
		"filePath", filePath)

	var (
		groupKeys []string
		groups    []*admin.Group
	)

	if filePath != "" {
		file, err := os.Open(filePath)
		if err != nil {
			logger.Error(err)
			return nil, nil, err
		}
		defer file.Close()

		scanner = bufio.NewScanner(file)
	}

	for scanner.Scan() {
		jsonData := scanner.Text()

		grpVar, groupKey, err := bugFromJSONFactory(ds, jsonData)
		if err != nil {
			logger.Error(err)
			return nil, nil, err
		}

		groupKeys = append(groupKeys, groupKey)
		groups = append(groups, grpVar)
	}
	err := scanner.Err()
	if err != nil {
		logger.Error(err)
		return nil, nil, err
	}

	logger.Debug("finished bugProcessJSON()")
	return groupKeys, groups, nil
}

func bugProcessObjects(ds *admin.Service, groups []*admin.Group, groupKeys []string) error {
	logger.Debugw("starting bugProcessObjects()",
		"groupKeys", groupKeys)

	wg := new(sync.WaitGroup)

	for idx, g := range groups {
		guc := ds.Groups.Update(groupKeys[idx], g)

		wg.Add(1)

		go bugUpdate(g, wg, guc, groupKeys[idx])
	}

	wg.Wait()

	logger.Debug("finished bugProcessObjects()")
	return nil
}

func bugUpdate(group *admin.Group, wg *sync.WaitGroup, guc *admin.GroupsUpdateCall, groupKey string) {
	logger.Debugw("starting bugUpdate()",
		"groupKey", groupKey)

	defer wg.Done()

	b := backoff.NewExponentialBackOff()
	b.MaxElapsedTime = 32 * time.Second

	err := backoff.Retry(func() error {
		var err error
		_, err = guc.Do()
		if err == nil {
			logger.Infof(gmess.INFOGROUPUPDATED, groupKey)
			fmt.Println(cmn.GminMessage(fmt.Sprintf(gmess.INFOGROUPUPDATED, groupKey)))
			return err
		}
		if !cmn.IsErrRetryable(err) {
			return backoff.Permanent(fmt.Errorf(gmess.ERRBATCHGROUP, err.Error(), groupKey))
		}
		// Log the retries
		logger.Warnw(err.Error(),
			"retrying", b.GetElapsedTime().String(),
			"group", groupKey)
		return fmt.Errorf(gmess.ERRBATCHGROUP, err.Error(), groupKey)
	}, b)
	if err != nil {
		// Log final error
		logger.Error(err)
		fmt.Println(cmn.GminMessage(err.Error()))
	}
	logger.Debug("finished bugUpdate()")
}

func init() {
	batchUpdateCmd.AddCommand(batchUpdGrpCmd)

	batchUpdGrpCmd.Flags().StringVarP(&inputFile, "input-file", "i", "", "filepath to group data file or sheet id")
	batchUpdGrpCmd.Flags().StringVarP(&format, "format", "f", "json", "user data file format")
	batchUpdGrpCmd.Flags().StringVarP(&sheetRange, "sheet-range", "s", "", "user data gsheet range")
}
