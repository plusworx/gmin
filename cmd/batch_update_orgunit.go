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
	Short:   "Updates a batch of orgunits",
	Long: `Updates a batch of orgunits where orgunit details are provided in a Google Sheet, CSV/JSON input file or piped JSON.
	
	Examples:	gmin batch-update orgunits -i inputfile.json
			gmin bupd ous -i inputfile.csv -f csv
			gmin bupd ou -i 1odyAIp3jGspd3M4xeepxWD6aeQIUuHBgrZB2OHSu8MI -s 'Sheet1!A1:K25' -f gsheet
			  
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
		err := btchUpdOUProcessCSV(ds, inputFile)
		if err != nil {
			logger.Error(err)
			return err
		}
	case lwrFmt == "json":
		err := btchUpdOUProcessJSON(ds, inputFile, scanner)
		if err != nil {
			logger.Error(err)
			return err
		}
	case lwrFmt == "gsheet":
		err := btchUpdOUProcessSheet(ds, inputFile)
		if err != nil {
			logger.Error(err)
			return err
		}
	}

	return nil
}

func btchUpdJSONOrgUnit(ds *admin.Service, jsonData string) (*admin.OrgUnit, string, error) {
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

	return orgunit, ouKey.OUKey, nil
}

func btchUpdateOrgUnits(ds *admin.Service, orgunits []*admin.OrgUnit, ouKeys []string) error {
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

		go btchOUUpdateProcess(ou, wg, ouuc, ouKeys[idx])
	}

	wg.Wait()

	return nil
}

func btchOUUpdateProcess(orgunit *admin.OrgUnit, wg *sync.WaitGroup, ouuc *admin.OrgunitsUpdateCall, ouKey string) {
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
		logger.Errorw(err.Error(),
			"retrying", b.Clock.Now().String(),
			"orgunit", ouKey)
		return fmt.Errorf(cmn.ErrBatchOU, err.Error(), ouKey)
	}, b)
	if err != nil {
		// Log final error
		logger.Error(err)
		fmt.Println(cmn.GminMessage(err.Error()))
	}
}

func btchUpdOUProcessCSV(ds *admin.Service, filePath string) error {
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

		ouVar, ouKey, err := btchUpdProcessOrgUnit(hdrMap, iSlice)
		if err != nil {
			logger.Error(err)
			return err
		}

		orgunits = append(orgunits, ouVar)
		ouKeys = append(ouKeys, ouKey)

		count = count + 1
	}

	err = btchUpdateOrgUnits(ds, orgunits, ouKeys)
	if err != nil {
		logger.Error(err)
		return err
	}
	return nil
}

func btchUpdOUProcessJSON(ds *admin.Service, filePath string, scanner *bufio.Scanner) error {
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

		ouVar, ouKey, err := btchUpdJSONOrgUnit(ds, jsonData)
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

	err = btchUpdateOrgUnits(ds, orgunits, ouKeys)
	if err != nil {
		logger.Error(err)
		return err
	}

	return nil
}

func btchUpdOUProcessSheet(ds *admin.Service, sheetID string) error {
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

		ouVar, ouKey, err := btchUpdProcessOrgUnit(hdrMap, row)
		if err != nil {
			logger.Error(err)
			return err
		}

		ouKeys = append(ouKeys, ouKey)
		orgunits = append(orgunits, ouVar)
	}

	err = btchUpdateOrgUnits(ds, orgunits, ouKeys)
	if err != nil {
		logger.Error(err)
		return err
	}

	return nil
}

func btchUpdProcessOrgUnit(hdrMap map[int]string, ouData []interface{}) (*admin.OrgUnit, string, error) {
	var (
		orgunit *admin.OrgUnit
		ouKey   string
	)

	orgunit = new(admin.OrgUnit)

	for idx, attr := range ouData {
		attrName := hdrMap[idx]

		switch {
		case attrName == "blockInheritance":
			lwrAttr := strings.ToLower(fmt.Sprintf("%v", attr))
			if lwrAttr == "true" {
				orgunit.BlockInheritance = true
			}
			if lwrAttr == "false" {
				orgunit.BlockInheritance = false
				orgunit.ForceSendFields = append(orgunit.ForceSendFields, "BlockInheritance")
			}
		case attrName == "description":
			desc := fmt.Sprintf("%v", attr)
			orgunit.Description = desc
			if desc == "" {
				orgunit.ForceSendFields = append(orgunit.ForceSendFields, "Description")
			}
		case attrName == "name":
			orgunit.Name = fmt.Sprintf("%v", attr)
		case attrName == "parentOrgUnitPath":
			orgunit.ParentOrgUnitPath = fmt.Sprintf("%v", attr)
		case attrName == "ouKey":
			ouKey = fmt.Sprintf("%v", attr)
		}
	}

	return orgunit, ouKey, nil
}

func init() {
	batchUpdateCmd.AddCommand(batchUpdOUCmd)

	batchUpdOUCmd.Flags().StringVarP(&inputFile, "inputfile", "i", "", "filepath to orgunit data file or sheet id")
	batchUpdOUCmd.Flags().StringVarP(&format, "format", "f", "json", "user data file format")
	batchUpdOUCmd.Flags().StringVarP(&sheetRange, "sheetrange", "s", "", "user data gsheet range")
}
