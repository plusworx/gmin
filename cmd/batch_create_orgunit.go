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
	cfg "github.com/plusworx/gmin/utils/config"
	gmess "github.com/plusworx/gmin/utils/gminmessages"
	ous "github.com/plusworx/gmin/utils/orgunits"
	"github.com/spf13/cobra"
	admin "google.golang.org/api/admin/directory/v1"
	sheet "google.golang.org/api/sheets/v4"
)

var batchCrtOrgUnitCmd = &cobra.Command{
	Use:     "orgunits -i <input file path or google sheet id>",
	Aliases: []string{"orgunit", "ous", "ou"},
	Example: `gmin batch-create orgunits -i inputfile.json
gmin bcrt ous -i inputfile.csv -f csv
gmin bcrt ou -i 1odyAIp3jGspd3M4xeepxWD6aeQIUuHBgrZB2OHSu8MI -s 'Sheet1!A1:K25' -f gsheet`,
	Short: "Creates a batch of orgunits",
	Long: `Creates a batch of orgunits where orgunit details are provided in a Google Sheet, CSV/JSON input file or piped JSON.
			
The contents of the JSON file or piped input should look something like this:

{"description":"Fabrication department OU","name":"Fabrication","parentOrgUnitPath":"/Engineering"}
{"description":"Electrical department OU","name":"Electrical","parentOrgUnitPath":"/Engineering"}
{"description":"Paintwork department OU","name":"Paintwork","parentOrgUnitPath":"/Engineering"}

CSV and Google sheets must have a header row with the following column names being the only ones that are valid:

blockInheritance [value true or false]
description
name [required]
parentOrgUnitPath [required]

The column names are case insensitive and can be in any order.`,
	RunE: doBatchCrtOrgUnit,
}

func doBatchCrtOrgUnit(cmd *cobra.Command, args []string) error {
	logger.Debugw("starting doBatchCrtOrgUnit()",
		"args", args)

	var orgunits []*admin.OrgUnit

	ds, err := cmn.CreateDirectoryService(admin.AdminDirectoryOrgunitScope)
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
		err := errors.New(gmess.ErrNoInputFile)
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
		err = fmt.Errorf(gmess.ErrInvalidFileFormat, formatFlgVal)
		logger.Error(err)
		return err
	}

	switch {
	case lwrFmt == "csv":
		orgunits, err = bcoProcessCSVFile(ds, inputFlgVal)
		if err != nil {
			logger.Error(err)
			return err
		}
	case lwrFmt == "json":
		orgunits, err = bcoProcessJSON(ds, inputFlgVal, scanner)
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

		orgunits, err = bcoProcessGSheet(ds, inputFlgVal, rangeFlgVal)
		if err != nil {
			logger.Error(err)
			return err
		}
	default:
		return fmt.Errorf(gmess.ErrInvalidFileFormat, formatFlgVal)
	}

	err = bcoProcessObjects(ds, orgunits)
	if err != nil {
		logger.Error(err)
		return err
	}

	logger.Debug("finished doBatchCrtOrgUnit()")
	return nil
}

func bcoCreate(orgunit *admin.OrgUnit, wg *sync.WaitGroup, ouic *admin.OrgunitsInsertCall) {
	logger.Debug("starting bcoCreate()")

	defer wg.Done()

	b := backoff.NewExponentialBackOff()
	b.MaxElapsedTime = 32 * time.Second

	err := backoff.Retry(func() error {
		var err error
		newOrgUnit, err := ouic.Do()
		if err == nil {
			logger.Infof(gmess.InfoOUCreated, newOrgUnit.OrgUnitPath)
			fmt.Println(cmn.GminMessage(fmt.Sprintf(gmess.InfoOUCreated, newOrgUnit.OrgUnitPath)))
			return err
		}
		if !cmn.IsErrRetryable(err) {
			return backoff.Permanent(fmt.Errorf(gmess.ErrBatchOU, err.Error(), orgunit.Name))
		}
		// Log the retries
		logger.Warnw(err.Error(),
			"retrying", b.GetElapsedTime().String(),
			"orgunit", orgunit.Name)
		return fmt.Errorf(gmess.ErrBatchOU, err.Error(), orgunit.Name)
	}, b)
	if err != nil {
		// Log final error
		logger.Error(err)
		fmt.Println(cmn.GminMessage(err.Error()))
	}
	logger.Debug("finished bcoCreate()")
}

func bcoFromFileFactory(hdrMap map[int]string, ouData []interface{}) (*admin.OrgUnit, error) {
	logger.Debugw("starting bcoFromFileFactory()",
		"hdrMap", hdrMap)

	var orgunit *admin.OrgUnit

	orgunit = new(admin.OrgUnit)

	for idx, attr := range ouData {
		attrName := hdrMap[idx]
		attrVal := fmt.Sprintf("%v", attr)
		lowerAttrVal := strings.ToLower(fmt.Sprintf("%v", attr))

		switch {
		case attrName == "blockInheritance":
			if lowerAttrVal == "true" {
				orgunit.BlockInheritance = true
			}
		case attrName == "description":
			orgunit.Description = attrVal
		case attrName == "name":
			if attrVal == "" {
				err := fmt.Errorf(gmess.ErrEmptyString, attrName)
				return nil, err
			}
			orgunit.Name = attrVal
		case attrName == "parentOrgUnitPath":
			if attrVal == "" {
				err := fmt.Errorf(gmess.ErrEmptyString, attrName)
				return nil, err
			}
			orgunit.ParentOrgUnitPath = attrVal
		}
	}
	logger.Debug("finished bcoFromFileFactory()")
	return orgunit, nil
}

func bcoFromJSONFactory(ds *admin.Service, jsonData string) (*admin.OrgUnit, error) {
	logger.Debugw("starting bcoFromJSONFactory()",
		"jsonData", jsonData)

	var (
		emptyVals = cmn.EmptyValues{}
		orgunit   *admin.OrgUnit
	)

	orgunit = new(admin.OrgUnit)
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

	err = cmn.ValidateInputAttrs(outStr, ous.OrgUnitAttrMap)
	if err != nil {
		logger.Error(err)
		return nil, err
	}

	err = json.Unmarshal(jsonBytes, &orgunit)
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
		orgunit.ForceSendFields = emptyVals.ForceSendFields
	}
	logger.Debug("finished bcoFromJSONFactory()")
	return orgunit, nil
}

func bcoProcessCSVFile(ds *admin.Service, filePath string) ([]*admin.OrgUnit, error) {
	logger.Debugw("starting bcoProcessCSVFile()",
		"filePath", filePath)

	var (
		iSlice   []interface{}
		hdrMap   = map[int]string{}
		orgunits []*admin.OrgUnit
	)

	csvfile, err := os.Open(filePath)
	if err != nil {
		logger.Error(err)
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
			logger.Error(err)
			return nil, err
		}

		if count == 0 {
			iSlice = make([]interface{}, len(record))
			for idx, value := range record {
				iSlice[idx] = value
			}
			hdrMap = cmn.ProcessHeader(iSlice)
			err = cmn.ValidateHeader(hdrMap, ous.OrgUnitAttrMap)
			if err != nil {
				logger.Error(err)
				return nil, err
			}
			count = count + 1
			continue
		}

		for idx, value := range record {
			iSlice[idx] = value
		}

		ouVar, err := bcoFromFileFactory(hdrMap, iSlice)
		if err != nil {
			logger.Error(err)
			return nil, err
		}

		orgunits = append(orgunits, ouVar)

		count = count + 1
	}

	logger.Debug("finished bcoProcessCSVFile()")
	return orgunits, nil
}

func bcoProcessGSheet(ds *admin.Service, sheetID string, sheetrange string) ([]*admin.OrgUnit, error) {
	logger.Debugw("starting bcoProcessGSheet()",
		"sheetID", sheetID,
		"sheetrange", sheetrange)

	var orgunits []*admin.OrgUnit

	if sheetrange == "" {
		err := errors.New(gmess.ErrNoSheetRange)
		logger.Error(err)
		return nil, err
	}

	ss, err := cmn.CreateSheetService(sheet.DriveReadonlyScope)
	if err != nil {
		logger.Error(err)
		return nil, err
	}

	ssvgc := ss.Spreadsheets.Values.Get(sheetID, sheetrange)
	sValRange, err := ssvgc.Do()
	if err != nil {
		logger.Error(err)
		return nil, err
	}

	if len(sValRange.Values) == 0 {
		err = fmt.Errorf(gmess.ErrNoSheetDataFound, sheetID, sheetrange)
		logger.Error(err)
		return nil, err
	}

	hdrMap := cmn.ProcessHeader(sValRange.Values[0])
	err = cmn.ValidateHeader(hdrMap, ous.OrgUnitAttrMap)
	if err != nil {
		logger.Error(err)
		return nil, err
	}

	for idx, row := range sValRange.Values {
		if idx == 0 {
			continue
		}

		ouVar, err := bcoFromFileFactory(hdrMap, row)
		if err != nil {
			logger.Error(err)
			return nil, err
		}

		orgunits = append(orgunits, ouVar)
	}

	logger.Debug("finished bcoProcessGSheet()")
	return orgunits, nil
}

func bcoProcessJSON(ds *admin.Service, filePath string, scanner *bufio.Scanner) ([]*admin.OrgUnit, error) {
	logger.Debugw("starting bcoProcessJSON()",
		"filePath", filePath)

	var orgunits []*admin.OrgUnit

	if filePath != "" {
		file, err := os.Open(filePath)
		if err != nil {
			logger.Error(err)
			return nil, err
		}
		defer file.Close()

		scanner = bufio.NewScanner(file)
	}

	for scanner.Scan() {
		jsonData := scanner.Text()

		ouVar, err := bcoFromJSONFactory(ds, jsonData)
		if err != nil {
			logger.Error(err)
			return nil, err
		}

		orgunits = append(orgunits, ouVar)
	}
	err := scanner.Err()
	if err != nil {
		logger.Error(err)
		return nil, err
	}

	logger.Debug("finished bcoProcessJSON()")
	return orgunits, nil
}

func bcoProcessObjects(ds *admin.Service, orgunits []*admin.OrgUnit) error {
	logger.Debug("starting bcoProcessObjects()")

	customerID, err := cfg.ReadConfigString("customerid")
	if err != nil {
		logger.Error(err)
		return err
	}

	wg := new(sync.WaitGroup)

	for _, ou := range orgunits {
		if ou.Name == "" || ou.ParentOrgUnitPath == "" {
			err = errors.New(gmess.ErrNoNameOrOuPath)
			logger.Error(err)
			return err
		}

		ouic := ds.Orgunits.Insert(customerID, ou)

		// Sleep for 2 seconds because only 1 orgunit can be created per second but 1 second interval
		// still can result in rate limit errors
		time.Sleep(2 * time.Second)

		wg.Add(1)

		go bcoCreate(ou, wg, ouic)
	}

	wg.Wait()

	logger.Debug("finished bcoProcessObjects()")
	return nil
}

func init() {
	batchCreateCmd.AddCommand(batchCrtOrgUnitCmd)

	batchCrtOrgUnitCmd.Flags().StringVarP(&inputFile, "input-file", "i", "", "filepath to orgunit data file or sheet id")
	batchCrtOrgUnitCmd.Flags().StringVarP(&format, "format", "f", "json", "user data file format")
	batchCrtOrgUnitCmd.Flags().StringVarP(&sheetRange, "sheet-range", "s", "", "user data gsheet range")
}
