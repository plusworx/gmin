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

var batchCrtOrgUnitCmd = &cobra.Command{
	Use:     "orgunits -i <input file path or google sheet id>",
	Aliases: []string{"orgunit", "ous", "ou"},
	Short:   "Creates a batch of orgunits",
	Long: `Creates a batch of orgunits where orgunit details are provided in a Google Sheet or CSV/JSON input file.
	
	Examples:	gmin batch-create orgunits -i inputfile.json
			gmin bcrt ous -i inputfile.csv -f csv
			gmin bcrt ou -i 1odyAIp3jGspd3M4xeepxWD6aeQIUuHBgrZB2OHSu8MI -s 'Sheet1!A1:K25' -f gsheet
		
	The contents of the JSON file should look something like this:
	
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
	ds, err := cmn.CreateDirectoryService(admin.AdminDirectoryOrgunitScope)
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
		err := btchOUProcessCSV(ds, inputFile)
		if err != nil {
			return err
		}
	case lwrFmt == "json":
		err := btchOUProcessJSON(ds, inputFile)
		if err != nil {
			return err
		}
	case lwrFmt == "gsheet":
		err := btchOUProcessSheet(ds, inputFile)
		if err != nil {
			return err
		}
	}

	return nil
}

func btchCreateJSONOrgUnit(ds *admin.Service, jsonData string) (*admin.OrgUnit, error) {
	var (
		emptyVals = cmn.EmptyValues{}
		orgunit   *admin.OrgUnit
	)

	orgunit = new(admin.OrgUnit)
	jsonBytes := []byte(jsonData)

	if !json.Valid(jsonBytes) {
		return nil, errors.New("gmin: error - attribute string is not valid JSON")
	}

	outStr, err := cmn.ParseInputAttrs(jsonBytes)
	if err != nil {
		return nil, err
	}

	err = cmn.ValidateInputAttrs(outStr, ous.OrgUnitAttrMap)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(jsonBytes, &orgunit)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(jsonBytes, &emptyVals)
	if err != nil {
		return nil, err
	}
	if len(emptyVals.ForceSendFields) > 0 {
		orgunit.ForceSendFields = emptyVals.ForceSendFields
	}

	return orgunit, nil
}

func btchInsertNewOrgUnits(ds *admin.Service, orgunits []*admin.OrgUnit) error {
	customerID, err := cfg.ReadConfigString("customerid")
	if err != nil {
		return err
	}

	wg := new(sync.WaitGroup)

	for _, ou := range orgunits {
		if ou.Name == "" || ou.ParentOrgUnitPath == "" {
			return errors.New("gmin: error - name and parentOrgUnitPath must be provided")
		}

		ouic := ds.Orgunits.Insert(customerID, ou)

		// Sleep for 2 seconds because only 1 orgunit can be created per second but 1 second interval
		// still results in rate limit errors
		time.Sleep(2 * time.Second)

		wg.Add(1)

		go btchOUInsertProcess(ou, wg, ouic)
	}

	wg.Wait()

	return nil
}

func btchOUInsertProcess(orgunit *admin.OrgUnit, wg *sync.WaitGroup, ouic *admin.OrgunitsInsertCall) {
	defer wg.Done()

	b := backoff.NewExponentialBackOff()
	b.MaxElapsedTime = 32 * time.Second

	err := backoff.Retry(func() error {
		var err error
		newOrgUnit, err := ouic.Do()
		if err == nil {
			fmt.Println(cmn.GminMessage("**** gmin: orgunit " + newOrgUnit.OrgUnitPath + " created ****"))
			return err
		}
		if !cmn.IsErrRetryable(err) {
			return backoff.Permanent(errors.New(cmn.GminMessage("gmin: error - " + err.Error() + " " + orgunit.Name)))
		}
		return errors.New(cmn.GminMessage("gmin: error - " + err.Error() + " " + orgunit.Name))
	}, b)
	if err != nil {
		fmt.Println(err)
	}
}

func btchOUProcessCSV(ds *admin.Service, filePath string) error {
	var (
		iSlice   []interface{}
		hdrMap   = map[int]string{}
		orgunits []*admin.OrgUnit
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
			err = cmn.ValidateHeader(hdrMap, ous.OrgUnitAttrMap)
			if err != nil {
				return err
			}
			count = count + 1
			continue
		}

		for idx, value := range record {
			iSlice[idx] = value
		}

		ouVar, err := btchCrtProcessOrgUnit(hdrMap, iSlice)
		if err != nil {
			fmt.Println(err.Error())
		}

		orgunits = append(orgunits, ouVar)

		count = count + 1
	}

	err = btchInsertNewOrgUnits(ds, orgunits)
	if err != nil {
		return err
	}
	return nil
}

func btchOUProcessJSON(ds *admin.Service, filePath string) error {
	var orgunits []*admin.OrgUnit

	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		jsonData := scanner.Text()

		ouVar, err := btchCreateJSONOrgUnit(ds, jsonData)
		if err != nil {
			return err
		}

		orgunits = append(orgunits, ouVar)
	}
	err = scanner.Err()
	if err != nil {
		return err
	}

	err = btchInsertNewOrgUnits(ds, orgunits)
	if err != nil {
		return err
	}

	return nil
}

func btchOUProcessSheet(ds *admin.Service, sheetID string) error {
	var orgunits []*admin.OrgUnit

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
	err = cmn.ValidateHeader(hdrMap, ous.OrgUnitAttrMap)
	if err != nil {
		return err
	}

	for idx, row := range sValRange.Values {
		if idx == 0 {
			continue
		}

		ouVar, err := btchCrtProcessOrgUnit(hdrMap, row)
		if err != nil {
			return err
		}

		orgunits = append(orgunits, ouVar)
	}

	err = btchInsertNewOrgUnits(ds, orgunits)
	if err != nil {
		return err
	}

	return nil
}

func btchCrtProcessOrgUnit(hdrMap map[int]string, ouData []interface{}) (*admin.OrgUnit, error) {
	var orgunit *admin.OrgUnit

	orgunit = new(admin.OrgUnit)

	for idx, attr := range ouData {
		attrName := hdrMap[idx]

		switch {
		case attrName == "blockInheritance":
			lwrAttr := strings.ToLower(fmt.Sprintf("%v", attr))
			if lwrAttr == "true" {
				orgunit.BlockInheritance = true
			}
		case attrName == "description":
			orgunit.Description = fmt.Sprintf("%v", attr)
		case attrName == "name":
			orgunit.Name = fmt.Sprintf("%v", attr)
		case attrName == "parentOrgUnitPath":
			orgunit.ParentOrgUnitPath = fmt.Sprintf("%v", attr)
		}
	}

	return orgunit, nil
}

func init() {
	batchCreateCmd.AddCommand(batchCrtOrgUnitCmd)

	batchCrtOrgUnitCmd.Flags().StringVarP(&inputFile, "inputfile", "i", "", "filepath to orgunit data file or sheet id")
	batchCrtOrgUnitCmd.Flags().StringVarP(&format, "format", "f", "json", "user data file format")
	batchCrtOrgUnitCmd.Flags().StringVarP(&sheetRange, "sheetrange", "s", "", "user data gsheet range")
}
