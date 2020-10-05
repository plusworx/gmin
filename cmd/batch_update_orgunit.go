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
	ous "github.com/plusworx/gmin/utils/orgunits"
	"github.com/spf13/cobra"
	admin "google.golang.org/api/admin/directory/v1"
	sheet "google.golang.org/api/sheets/v4"
)

var batchUpdOUCmd = &cobra.Command{
	Use:     "orgunits -i <input file path>",
	Aliases: []string{"orgunit", "ous", "ou"},
	Example: `gmin batch-update orgunits -i inputfile.json
gmin bupd ous -i inputfile.csv -f csv
gmin bupd ou -i 1odyAIp3jGspd3M4xeepxWD6aeQIUuHBgrZB2OHSu8MI -s 'Sheet1!A1:K25' -f gsheet`,
	Short: "Updates a batch of orgunits",
	Long: `Updates a batch of orgunits where orgunit details are provided in a Google Sheet, CSV/JSON input file or piped JSON.
			  
The contents of the JSON file or piped input should look something like this:

{"ouKey":"Credit","parentOrgUnitPath":"/Finance","name":"Credit_Control"}
{"ouKey":"Audit","parentOrgUnitPath":"/Finance","name":"Audit_Governance"}
{"ouKey":"Planning","parentOrgUnitPath":"/Finance","name":"Planning_Reporting"}

N.B. ouKey (full orgunit path or id) must be provided.

CSV and Google sheets must have a header row with the following column names being the only ones that are valid:

blockInheritance [value true or false]
description
name
ouKey [required]
parentOrgUnitPath

The column names are case insensitive and can be in any order.`,
	RunE: doBatchUpdOU,
}

func doBatchUpdOU(cmd *cobra.Command, args []string) error {
	logger.Debugw("starting doBatchUpdOU()",
		"args", args)

	ds, err := cmn.CreateDirectoryService(admin.AdminDirectoryOrgunitScope)
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

	switch {
	case lwrFmt == "csv":
		err := buoProcessCSVFile(ds, inputFile)
		if err != nil {
			logger.Error(err)
			return err
		}
	case lwrFmt == "json":
		err := buoProcessJSON(ds, inputFile, scanner)
		if err != nil {
			logger.Error(err)
			return err
		}
	case lwrFmt == "gsheet":
		err := buoProcessGSheet(ds, inputFile)
		if err != nil {
			logger.Error(err)
			return err
		}
	}
	logger.Debug("finished doBatchUpdOU()")
	return nil
}

func buoFromFileFactory(hdrMap map[int]string, ouData []interface{}) (*admin.OrgUnit, string, error) {
	logger.Debugw("starting buoFromFileFactory()",
		"hdrMap", hdrMap)

	var (
		orgunit *admin.OrgUnit
		ouKey   string
	)

	orgunit = new(admin.OrgUnit)

	for idx, attr := range ouData {
		attrName := hdrMap[idx]
		attrVal := fmt.Sprintf("%v", attr)
		lowerAttrVal := strings.ToLower(fmt.Sprintf("%v", attr))

		switch {
		case attrName == "blockInheritance":
			if lowerAttrVal == "true" {
				orgunit.BlockInheritance = true
				break
			}
			orgunit.BlockInheritance = false
			orgunit.ForceSendFields = append(orgunit.ForceSendFields, "BlockInheritance")
		case attrName == "description":
			orgunit.Description = attrVal
			if attrVal == "" {
				orgunit.ForceSendFields = append(orgunit.ForceSendFields, "Description")
			}
		case attrName == "name":
			if attrVal == "" {
				err := fmt.Errorf(cmn.ErrEmptyString, attrName)
				return nil, "", err
			}
			orgunit.Name = attrVal
		case attrName == "parentOrgUnitPath":
			if attrVal == "" {
				err := fmt.Errorf(cmn.ErrEmptyString, attrName)
				return nil, "", err
			}
			orgunit.ParentOrgUnitPath = attrVal
		case attrName == "ouKey":
			if attrVal == "" {
				err := fmt.Errorf(cmn.ErrEmptyString, attrName)
				return nil, "", err
			}
			ouKey = attrVal
		}
	}
	logger.Debug("finished buoFromFileFactory()")
	return orgunit, ouKey, nil
}

func buoFromJSONFactory(ds *admin.Service, jsonData string) (*admin.OrgUnit, string, error) {
	logger.Debugw("starting buoFromJSONFactory()",
		"jsonData", jsonData)

	var (
		orgunit   *admin.OrgUnit
		ouKey     = ous.Key{}
		emptyVals = cmn.EmptyValues{}
	)

	orgunit = new(admin.OrgUnit)
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

	err = cmn.ValidateInputAttrs(outStr, ous.OrgUnitAttrMap)
	if err != nil {
		logger.Error(err)
		return nil, "", err
	}

	err = json.Unmarshal(jsonBytes, &ouKey)
	if err != nil {
		logger.Error(err)
		return nil, "", err
	}

	if ouKey.OUKey == "" {
		err = errors.New(cmn.ErrNoJSONOUKey)
		logger.Error(err)
		return nil, "", err
	}

	err = json.Unmarshal(jsonBytes, &orgunit)
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
		orgunit.ForceSendFields = emptyVals.ForceSendFields
	}
	logger.Debug("finished buoFromJSONFactory()")
	return orgunit, ouKey.OUKey, nil
}

func buoProcessCSVFile(ds *admin.Service, filePath string) error {
	logger.Debugw("starting buoProcessCSVFile()",
		"filePath", filePath)

	var (
		iSlice   []interface{}
		hdrMap   = map[int]string{}
		ouKeys   []string
		orgunits []*admin.OrgUnit
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
			err = cmn.ValidateHeader(hdrMap, ous.OrgUnitAttrMap)
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

		ouVar, ouKey, err := buoFromFileFactory(hdrMap, iSlice)
		if err != nil {
			logger.Error(err)
			return err
		}

		orgunits = append(orgunits, ouVar)
		ouKeys = append(ouKeys, ouKey)

		count = count + 1
	}

	err = buoProcessObjects(ds, orgunits, ouKeys)
	if err != nil {
		logger.Error(err)
		return err
	}
	logger.Debug("finished buoProcessCSVFile()")
	return nil
}

func buoProcessGSheet(ds *admin.Service, sheetID string) error {
	logger.Debugw("starting buoProcessGSheet()",
		"sheetID", sheetID)

	var (
		ouKeys   []string
		orgunits []*admin.OrgUnit
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
	err = cmn.ValidateHeader(hdrMap, ous.OrgUnitAttrMap)
	if err != nil {
		logger.Error(err)
		return err
	}

	for idx, row := range sValRange.Values {
		if idx == 0 {
			continue
		}

		ouVar, ouKey, err := buoFromFileFactory(hdrMap, row)
		if err != nil {
			logger.Error(err)
			return err
		}

		ouKeys = append(ouKeys, ouKey)
		orgunits = append(orgunits, ouVar)
	}

	err = buoProcessObjects(ds, orgunits, ouKeys)
	if err != nil {
		logger.Error(err)
		return err
	}
	logger.Debug("finished buoProcessGSheet()")
	return nil
}

func buoProcessJSON(ds *admin.Service, filePath string, scanner *bufio.Scanner) error {
	logger.Debugw("starting buoProcessJSON()",
		"filePath", filePath)

	var (
		ouKeys   []string
		orgunits []*admin.OrgUnit
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

		ouVar, ouKey, err := buoFromJSONFactory(ds, jsonData)
		if err != nil {
			logger.Error(err)
			return err
		}

		ouKeys = append(ouKeys, ouKey)
		orgunits = append(orgunits, ouVar)
	}
	err := scanner.Err()
	if err != nil {
		logger.Error(err)
		return err
	}

	err = buoProcessObjects(ds, orgunits, ouKeys)
	if err != nil {
		logger.Error(err)
		return err
	}
	logger.Debug("finished buoProcessJSON()")
	return nil
}

func buoProcessObjects(ds *admin.Service, orgunits []*admin.OrgUnit, ouKeys []string) error {
	logger.Debugw("starting buoProcessObjects()",
		"ouKeys", ouKeys)

	wg := new(sync.WaitGroup)

	customerID, err := cfg.ReadConfigString("customerid")
	if err != nil {
		logger.Error(err)
		return err
	}

	for idx, ou := range orgunits {
		ouuc := ds.Orgunits.Update(customerID, ouKeys[idx], ou)

		wg.Add(1)

		// Sleep for 2 seconds because only 1 orgunit can be updated per second but 1 second interval
		// still can result in rate limit errors
		time.Sleep(2 * time.Second)

		go buoUpdate(ou, wg, ouuc, ouKeys[idx])
	}

	wg.Wait()

	logger.Debug("finished buoProcessObjects()")
	return nil
}

func buoUpdate(orgunit *admin.OrgUnit, wg *sync.WaitGroup, ouuc *admin.OrgunitsUpdateCall, ouKey string) {
	logger.Debugw("starting buoUpdate()",
		"ouKey", ouKey)

	defer wg.Done()

	b := backoff.NewExponentialBackOff()
	b.MaxElapsedTime = 32 * time.Second

	err := backoff.Retry(func() error {
		var err error
		_, err = ouuc.Do()
		if err == nil {
			logger.Infof(cmn.InfoOUUpdated, ouKey)
			fmt.Println(cmn.GminMessage(fmt.Sprintf(cmn.InfoOUUpdated, ouKey)))
			return err
		}
		if !cmn.IsErrRetryable(err) {
			return backoff.Permanent(fmt.Errorf(cmn.ErrBatchOU, err.Error(), ouKey))
		}
		// Log the retries
		logger.Warnw(err.Error(),
			"retrying", b.GetElapsedTime().String(),
			"orgunit", ouKey)
		return fmt.Errorf(cmn.ErrBatchOU, err.Error(), ouKey)
	}, b)
	if err != nil {
		// Log final error
		logger.Error(err)
		fmt.Println(cmn.GminMessage(err.Error()))
	}
	logger.Debug("finished buoUpdate()")
}

func init() {
	batchUpdateCmd.AddCommand(batchUpdOUCmd)

	batchUpdOUCmd.Flags().StringVarP(&inputFile, "inputfile", "i", "", "filepath to orgunit data file or sheet id")
	batchUpdOUCmd.Flags().StringVarP(&format, "format", "f", "json", "user data file format")
	batchUpdOUCmd.Flags().StringVarP(&sheetRange, "sheetrange", "s", "", "user data gsheet range")
}
