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

var batchCrtGroupCmd = &cobra.Command{
	Use:     "groups -i <input file path or google sheet id>",
	Aliases: []string{"group", "grps", "grp"},
	Example: `gmin batch-create groups -i inputfile.json
gmin bcrt grps -i inputfile.csv -f csv
gmin bcrt grp -i 1odyAIp3jGspd3M4xeepxWD6aeQIUuHBgrZB2OHSu8MI -s 'Sheet1!A1:K25' -f gsheet`,
	Short: "Creates a batch of groups",
	Long: `Creates a batch of groups where group details are provided in a Google Sheet, CSV/JSON input file or piped JSON.
				  
The contents of the JSON file or piped input should look something like this:
	
{"description":"Finance group","email":"finance@mycompany.com","name":"Finance"}
{"description":"Marketing group","email":"marketing@mycompany.com","name":"Marketing"}
{"description":"Sales group","email":"sales@mycompany.com","name":"Sales"}
	
CSV and Google sheets must have a header row with the following column names being the only ones that are valid:
	
description
email [required]
name

The column names are case insensitive and can be in any order.`,
	RunE: doBatchCrtGroup,
}

func doBatchCrtGroup(cmd *cobra.Command, args []string) error {
	logger.Debugw("starting doBatchCrtGroup()",
		"args", args)

	ds, err := cmn.CreateDirectoryService(admin.AdminDirectoryGroupScope)
	if err != nil {
		return err
	}

	scanner, err := cmn.InputFromStdIn(inputFile)
	if err != nil {
		logger.Error(err)
		return err
	}

	if inputFile == "" && scanner == nil {
		err := errors.New(gmess.ErrNoInputFile)
		logger.Error(err)
		return err
	}

	lwrFmt := strings.ToLower(format)

	ok := cmn.SliceContainsStr(cmn.ValidFileFormats, lwrFmt)
	if !ok {
		err = fmt.Errorf(gmess.ErrInvalidFileFormat, format)
		logger.Error(err)
		return err
	}

	switch {
	case lwrFmt == "csv":
		err := bcgProcessCSVFile(ds, inputFile)
		if err != nil {
			logger.Error(err)
			return err
		}
	case lwrFmt == "json":
		err := bcgProcessJSON(ds, inputFile, scanner)
		if err != nil {
			logger.Error(err)
			return err
		}
	case lwrFmt == "gsheet":
		err := bcgProcessGSheet(ds, inputFile)
		if err != nil {
			logger.Error(err)
			return err
		}
	default:
		return fmt.Errorf(gmess.ErrInvalidFileFormat, format)
	}
	logger.Debug("finished doBatchCrtGroup()")
	return nil
}

func bcgCreate(group *admin.Group, wg *sync.WaitGroup, gic *admin.GroupsInsertCall) {
	logger.Debug("starting bcgCreate()")

	defer wg.Done()

	b := backoff.NewExponentialBackOff()
	b.MaxElapsedTime = 32 * time.Second

	err := backoff.Retry(func() error {
		var err error
		newGroup, err := gic.Do()
		if err == nil {
			logger.Infof(gmess.InfoGroupCreated, newGroup.Email)
			fmt.Println(cmn.GminMessage(fmt.Sprintf(gmess.InfoGroupCreated, newGroup.Email)))
			return err
		}
		if !cmn.IsErrRetryable(err) {
			return backoff.Permanent(fmt.Errorf(gmess.ErrBatchGroup, err.Error(), group.Email))
		}
		// Log the retries
		logger.Warnw(err.Error(),
			"retrying", b.GetElapsedTime().String(),
			"group", group.Email)
		return fmt.Errorf(gmess.ErrBatchGroup, err.Error(), group.Email)
	}, b)
	if err != nil {
		// Log final error
		logger.Error(err)
		fmt.Println(cmn.GminMessage(err.Error()))
	}
	logger.Debug("finished bcgCreate()")
}

func bcgFromFileFactory(hdrMap map[int]string, grpData []interface{}) (*admin.Group, error) {
	logger.Debugw("starting bcgFromFileFactory()",
		"hdrMap", hdrMap)

	var group *admin.Group

	group = new(admin.Group)

	for idx, attr := range grpData {
		attrName := hdrMap[idx]
		attrVal := fmt.Sprintf("%v", attr)

		switch {
		case attrName == "description":
			group.Description = attrVal
		case attrName == "email":
			if attrVal == "" {
				err := fmt.Errorf(gmess.ErrEmptyString, attrName)
				return nil, err
			}
			group.Email = attrVal
		case attrName == "name":
			if attrVal == "" {
				err := fmt.Errorf(gmess.ErrEmptyString, attrName)
				return nil, err
			}
			group.Name = attrVal
		}
	}
	logger.Debug("finished bcgFromFileFactory()")
	return group, nil
}

func bcgFromJSONFactory(ds *admin.Service, jsonData string) (*admin.Group, error) {
	logger.Debugw("starting bcgFromJSONFactory()",
		"jsonData", jsonData)

	var (
		emptyVals = cmn.EmptyValues{}
		group     *admin.Group
	)

	group = new(admin.Group)
	jsonBytes := []byte(jsonData)

	if !json.Valid(jsonBytes) {
		logger.Error(gmess.ErrInvalidJSONAttr)
		return nil, errors.New(gmess.ErrInvalidJSONAttr)
	}

	outStr, err := cmn.ParseInputAttrs(jsonBytes)
	if err != nil {
		logger.Error(err)
		return nil, err
	}

	err = cmn.ValidateInputAttrs(outStr, grps.GroupAttrMap)
	if err != nil {
		logger.Error(err)
		return nil, err
	}

	err = json.Unmarshal(jsonBytes, &group)
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
		group.ForceSendFields = emptyVals.ForceSendFields
	}
	logger.Debug("finished bcgFromJSONFactory()")
	return group, nil
}

func bcgProcessCSVFile(ds *admin.Service, filePath string) error {
	logger.Debugw("starting bcgProcessCSVFile()",
		"filePath", filePath)

	var (
		iSlice []interface{}
		hdrMap = map[int]string{}
		groups []*admin.Group
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
			err = cmn.ValidateHeader(hdrMap, grps.GroupAttrMap)
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

		grpVar, err := bcgFromFileFactory(hdrMap, iSlice)
		if err != nil {
			logger.Error(err)
			return err
		}

		groups = append(groups, grpVar)

		count = count + 1
	}

	err = bcgProcessObjects(ds, groups)
	if err != nil {
		logger.Error(err)
		return err
	}
	logger.Debug("finished bcgProcessCSVFile()")
	return nil
}

func bcgProcessGSheet(ds *admin.Service, sheetID string) error {
	logger.Debugw("starting bcgProcessGSheet()",
		"sheetID", sheetID)

	var groups []*admin.Group

	if sheetRange == "" {
		err := errors.New(gmess.ErrNoSheetRange)
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
		err = fmt.Errorf(gmess.ErrNoSheetDataFound, sheetID, sheetRange)
		logger.Error(err)
		return err
	}

	hdrMap := cmn.ProcessHeader(sValRange.Values[0])
	err = cmn.ValidateHeader(hdrMap, grps.GroupAttrMap)
	if err != nil {
		logger.Error(err)
		return err
	}

	for idx, row := range sValRange.Values {
		if idx == 0 {
			continue
		}

		grpVar, err := bcgFromFileFactory(hdrMap, row)
		if err != nil {
			logger.Error(err)
			return err
		}

		groups = append(groups, grpVar)
	}

	err = bcgProcessObjects(ds, groups)
	if err != nil {
		logger.Error(err)
		return err
	}
	logger.Debug("finished bcgProcessGSheet()")
	return nil
}

func bcgProcessJSON(ds *admin.Service, filePath string, scanner *bufio.Scanner) error {
	logger.Debugw("starting bcgProcessJSON()",
		"filePath", filePath)

	var groups []*admin.Group

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

		grpVar, err := bcgFromJSONFactory(ds, jsonData)
		if err != nil {
			logger.Error(err)
			return err
		}

		groups = append(groups, grpVar)
	}
	err := scanner.Err()
	if err != nil {
		logger.Error(err)
		return err
	}

	err = bcgProcessObjects(ds, groups)
	if err != nil {
		logger.Error(err)
		return err
	}
	logger.Debug("finished bcgProcessJSON()")
	return nil
}

func bcgProcessObjects(ds *admin.Service, groups []*admin.Group) error {
	logger.Debug("starting bcgProcessObjects()")

	wg := new(sync.WaitGroup)

	for _, g := range groups {
		if g.Email == "" {
			err := errors.New(gmess.ErrNoGroupEmailAddress)
			logger.Error(err)
			return err
		}

		gic := ds.Groups.Insert(g)

		wg.Add(1)

		go bcgCreate(g, wg, gic)
	}

	wg.Wait()

	logger.Debug("finished bcgProcessObjects()")
	return nil
}

func init() {
	batchCreateCmd.AddCommand(batchCrtGroupCmd)

	batchCrtGroupCmd.Flags().StringVarP(&inputFile, "input-file", "i", "", "filepath to group data file or sheet id")
	batchCrtGroupCmd.Flags().StringVarP(&format, "format", "f", "json", "user data file format")
	batchCrtGroupCmd.Flags().StringVarP(&sheetRange, "sheet-range", "s", "", "user data gsheet range")
}
