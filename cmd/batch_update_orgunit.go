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
	flgnm "github.com/plusworx/gmin/utils/flagnames"
	gmess "github.com/plusworx/gmin/utils/gminmessages"
	lg "github.com/plusworx/gmin/utils/logging"
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
	lg.Debugw("starting doBatchUpdOU()",
		"args", args)
	defer lg.Debug("finished doBatchUpdOU()")

	var (
		ouKeys   []string
		orgunits []*admin.OrgUnit
	)

	ds, err := cmn.CreateDirectoryService(admin.AdminDirectoryOrgunitScope)
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

	switch {
	case lwrFmt == "csv":
		ouKeys, orgunits, err = buoProcessCSVFile(ds, inputFlgVal)
		if err != nil {
			return err
		}
	case lwrFmt == "json":
		ouKeys, orgunits, err = buoProcessJSON(ds, inputFlgVal, scanner)
		if err != nil {
			return err
		}
	case lwrFmt == "gsheet":
		rangeFlgVal, err := cmd.Flags().GetString(flgnm.FLG_SHEETRANGE)
		if err != nil {
			lg.Error(err)
			return err
		}

		ouKeys, orgunits, err = buoProcessGSheet(ds, inputFlgVal, rangeFlgVal)
		if err != nil {
			return err
		}
	default:
		err = fmt.Errorf(gmess.ERR_INVALIDFILEFORMAT, formatFlgVal)
		lg.Error(err)
		return err
	}

	err = buoProcessObjects(ds, orgunits, ouKeys)
	if err != nil {
		return err
	}

	return nil
}

func buoFromFileFactory(hdrMap map[int]string, ouData []interface{}) (*admin.OrgUnit, string, error) {
	lg.Debugw("starting buoFromFileFactory()",
		"hdrMap", hdrMap)
	defer lg.Debug("finished buoFromFileFactory()")

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
				err := fmt.Errorf(gmess.ERR_EMPTYSTRING, attrName)
				lg.Error(err)
				return nil, "", err
			}
			orgunit.Name = attrVal
		case attrName == "parentOrgUnitPath":
			if attrVal == "" {
				err := fmt.Errorf(gmess.ERR_EMPTYSTRING, attrName)
				lg.Error(err)
				return nil, "", err
			}
			orgunit.ParentOrgUnitPath = attrVal
		case attrName == "ouKey":
			if attrVal == "" {
				err := fmt.Errorf(gmess.ERR_EMPTYSTRING, attrName)
				lg.Error(err)
				return nil, "", err
			}
			ouKey = attrVal
		}
	}
	return orgunit, ouKey, nil
}

func buoFromJSONFactory(ds *admin.Service, jsonData string) (*admin.OrgUnit, string, error) {
	lg.Debugw("starting buoFromJSONFactory()",
		"jsonData", jsonData)
	defer lg.Debug("finished buoFromJSONFactory()")

	var (
		orgunit   *admin.OrgUnit
		ouKey     = ous.Key{}
		emptyVals = cmn.EmptyValues{}
	)

	orgunit = new(admin.OrgUnit)
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

	err = cmn.ValidateInputAttrs(outStr, ous.OrgUnitAttrMap)
	if err != nil {
		return nil, "", err
	}

	err = json.Unmarshal(jsonBytes, &ouKey)
	if err != nil {
		lg.Error(err)
		return nil, "", err
	}

	if ouKey.OUKey == "" {
		err = errors.New(gmess.ERR_NOJSONOUKEY)
		lg.Error(err)
		return nil, "", err
	}

	err = json.Unmarshal(jsonBytes, &orgunit)
	if err != nil {
		lg.Error(err)
		return nil, "", err
	}

	err = json.Unmarshal(jsonBytes, &emptyVals)
	if err != nil {
		lg.Error(err)
		return nil, "", err
	}
	if len(emptyVals.ForceSendFields) > 0 {
		orgunit.ForceSendFields = emptyVals.ForceSendFields
	}
	return orgunit, ouKey.OUKey, nil
}

func buoProcessCSVFile(ds *admin.Service, filePath string) ([]string, []*admin.OrgUnit, error) {
	lg.Debugw("starting buoProcessCSVFile()",
		"filePath", filePath)
	defer lg.Debug("finished buoProcessCSVFile()")

	var (
		iSlice   []interface{}
		hdrMap   = map[int]string{}
		ouKeys   []string
		orgunits []*admin.OrgUnit
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
			err = cmn.ValidateHeader(hdrMap, ous.OrgUnitAttrMap)
			if err != nil {
				return nil, nil, err
			}
			count = count + 1
			continue
		}

		for idx, value := range record {
			iSlice[idx] = value
		}

		ouVar, ouKey, err := buoFromFileFactory(hdrMap, iSlice)
		if err != nil {
			return nil, nil, err
		}

		orgunits = append(orgunits, ouVar)
		ouKeys = append(ouKeys, ouKey)

		count = count + 1
	}

	return ouKeys, orgunits, nil
}

func buoProcessGSheet(ds *admin.Service, sheetID string, sheetrange string) ([]string, []*admin.OrgUnit, error) {
	lg.Debugw("starting buoProcessGSheet()",
		"sheetID", sheetID,
		"sheetrange", sheetrange)
	defer lg.Debug("finished buoProcessGSheet()")

	var (
		ouKeys   []string
		orgunits []*admin.OrgUnit
	)

	if sheetrange == "" {
		err := errors.New(gmess.ERR_NOSHEETRANGE)
		lg.Error(err)
		return nil, nil, err
	}

	ss, err := cmn.CreateSheetService(sheet.DriveReadonlyScope)
	if err != nil {
		return nil, nil, err
	}

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
	err = cmn.ValidateHeader(hdrMap, ous.OrgUnitAttrMap)
	if err != nil {
		return nil, nil, err
	}

	for idx, row := range sValRange.Values {
		if idx == 0 {
			continue
		}

		ouVar, ouKey, err := buoFromFileFactory(hdrMap, row)
		if err != nil {
			return nil, nil, err
		}

		ouKeys = append(ouKeys, ouKey)
		orgunits = append(orgunits, ouVar)
	}

	return ouKeys, orgunits, nil
}

func buoProcessJSON(ds *admin.Service, filePath string, scanner *bufio.Scanner) ([]string, []*admin.OrgUnit, error) {
	lg.Debugw("starting buoProcessJSON()",
		"filePath", filePath)
	defer lg.Debug("finished buoProcessJSON()")

	var (
		ouKeys   []string
		orgunits []*admin.OrgUnit
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

		ouVar, ouKey, err := buoFromJSONFactory(ds, jsonData)
		if err != nil {
			return nil, nil, err
		}

		ouKeys = append(ouKeys, ouKey)
		orgunits = append(orgunits, ouVar)
	}
	err := scanner.Err()
	if err != nil {
		lg.Error(err)
		return nil, nil, err
	}

	return ouKeys, orgunits, nil
}

func buoProcessObjects(ds *admin.Service, orgunits []*admin.OrgUnit, ouKeys []string) error {
	lg.Debugw("starting buoProcessObjects()",
		"ouKeys", ouKeys)
	defer lg.Debug("finished buoProcessObjects()")

	wg := new(sync.WaitGroup)

	customerID, err := cfg.ReadConfigString(cfg.CONFIGCUSTID)
	if err != nil {
		lg.Error(err)
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

	return nil
}

func buoUpdate(orgunit *admin.OrgUnit, wg *sync.WaitGroup, ouuc *admin.OrgunitsUpdateCall, ouKey string) {
	lg.Debugw("starting buoUpdate()",
		"ouKey", ouKey)
	defer lg.Debug("finished buoUpdate()")

	defer wg.Done()

	b := backoff.NewExponentialBackOff()
	b.MaxElapsedTime = 32 * time.Second

	err := backoff.Retry(func() error {
		var err error
		_, err = ouuc.Do()
		if err == nil {
			fmt.Println(cmn.GminMessage(fmt.Sprintf(gmess.INFO_OUUPDATED, ouKey)))
			lg.Infof(gmess.INFO_OUUPDATED, ouKey)
			return err
		}
		if !cmn.IsErrRetryable(err) {
			return backoff.Permanent(fmt.Errorf(gmess.ERR_BATCHOU, err.Error(), ouKey))
		}
		// Log the retries
		lg.Warnw(err.Error(),
			"retrying", b.GetElapsedTime().String(),
			"orgunit", ouKey)
		return fmt.Errorf(gmess.ERR_BATCHOU, err.Error(), ouKey)
	}, b)
	if err != nil {
		// Log final error
		lg.Error(err)
		fmt.Println(cmn.GminMessage(err.Error()))
	}
}

func init() {
	batchUpdateCmd.AddCommand(batchUpdOUCmd)

	batchUpdOUCmd.Flags().StringVarP(&inputFile, flgnm.FLG_INPUTFILE, "i", "", "filepath to orgunit data file or sheet id")
	batchUpdOUCmd.Flags().StringVarP(&format, flgnm.FLG_FORMAT, "f", "json", "user data file format")
	batchUpdOUCmd.Flags().StringVarP(&sheetRange, flgnm.FLG_SHEETRANGE, "s", "", "user data gsheet range")
}
