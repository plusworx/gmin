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
	mdevs "github.com/plusworx/gmin/utils/mobiledevices"
	"github.com/spf13/cobra"
	admin "google.golang.org/api/admin/directory/v1"
	sheet "google.golang.org/api/sheets/v4"
)

var batchMngMobDevCmd = &cobra.Command{
	Use:     "mobiledevices -i <input file>",
	Aliases: []string{"mobiledevice", "mobdevices", "mobdevice", "mobdevs", "mobdev", "mdevs", "mdev"},
	Short:   "Manages a batch of mobile devices",
	Long: `Manages a batch of mobile devices where device details are provided in a Google Sheet, CSV/JSON input file or piped JSON.
	
	Examples:	gmin batch-manage mobiledevices -i inputfile.json
			gmin bmng mdevs -i inputfile.csv -f csv
			gmin bmng mdev -i 1odyAIp3jGspd3M4xeepxWD6aeQIUuHBgrZB2OHSu8MI -s 'Sheet1!A1:B25' -f gsheet
			  
	The JSON file or piped input should contain mobile device management details like this:
	
	{"resourceId":"4cx07eba348f09b3Yjklj93xjsol0kE30lkl","action":"admin_account_wipe"}
	{"resourceId":"Hkj98764yKK4jw8yyoyq9987js07q1hs7y98","action":"approve"}
	{"resourceId":"lkalkju9027ja98na65wqHaTBOOUgarTQKk9","action":"admin_remote_wipe"}

	N.B. resourceId must be used NOT deviceId.
	
	CSV and Google sheets must have a header row with the following column names being the only ones that are valid:
	
	action [required]
	resourceId [required]
	
	The column names are case insensitive and can be in any order.
	
	Valid actions are:
	admin_account_wipe
	admin_remote_wipe
	approve
	block
	cancel_remote_wipe_then_activate
	cancel_remote_wipe_then_block`,
	RunE: doBatchMngMobDev,
}

func doBatchMngMobDev(cmd *cobra.Command, args []string) error {
	ds, err := cmn.CreateDirectoryService(admin.AdminDirectoryDeviceMobileActionScope)
	if err != nil {
		return err
	}

	scanner, err := cmn.InputFromStdIn(inputFile)
	if err != nil {
		return err
	}

	if inputFile == "" && scanner == nil {
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
		err := btchMngMobDevProcessCSV(ds, inputFile)
		if err != nil {
			return err
		}
	case lwrFmt == "json":
		err := btchMngMobDevProcessJSON(ds, inputFile, scanner)
		if err != nil {
			return err
		}
	case lwrFmt == "gsheet":
		err := btchMngMobDevProcessSheet(ds, inputFile)
		if err != nil {
			return err
		}
	}

	return nil
}

func btchMngJSONMobDev(ds *admin.Service, jsonData string) (mdevs.ManagedDevice, error) {
	managedDev := mdevs.ManagedDevice{}
	jsonBytes := []byte(jsonData)

	if !json.Valid(jsonBytes) {
		return managedDev, errors.New("gmin: error - attribute string is not valid JSON")
	}

	outStr, err := cmn.ParseInputAttrs(jsonBytes)
	if err != nil {
		return managedDev, err
	}

	err = cmn.ValidateInputAttrs(outStr, mdevs.MobDevAttrMap)
	if err != nil {
		return managedDev, err
	}

	err = json.Unmarshal(jsonBytes, &managedDev)
	if err != nil {
		return managedDev, err
	}

	return managedDev, nil
}

func btchMngMobDevs(ds *admin.Service, managedDevs []mdevs.ManagedDevice) error {
	customerID, err := cfg.ReadConfigString("customerid")
	if err != nil {
		return err
	}

	wg := new(sync.WaitGroup)

	for _, md := range managedDevs {
		devAction := admin.MobileDeviceAction{}

		devAction.Action = md.Action

		mdac := ds.Mobiledevices.Action(customerID, md.ResourceId, &devAction)

		wg.Add(1)

		go btchMngMobDevProcess(md.ResourceId, md.Action, wg, mdac)
	}

	wg.Wait()

	return nil
}

func btchMngMobDevProcess(resourceID string, action string, wg *sync.WaitGroup, mdac *admin.MobiledevicesActionCall) {
	defer wg.Done()

	b := backoff.NewExponentialBackOff()
	b.MaxElapsedTime = 32 * time.Second

	err := backoff.Retry(func() error {
		var err error
		err = mdac.Do()
		if err == nil {
			fmt.Println(cmn.GminMessage("**** gmin: " + action + " successfully performed on mobile device " + resourceID + " ****"))
			return err
		}
		if !cmn.IsErrRetryable(err) {
			return backoff.Permanent(errors.New(cmn.GminMessage("gmin: error - " + err.Error() + " " + resourceID)))
		}
		return errors.New(cmn.GminMessage("gmin: error - " + err.Error() + " " + resourceID))
	}, b)
	if err != nil {
		fmt.Println(err)
	}
}

func btchMngMobDevProcessCSV(ds *admin.Service, filePath string) error {
	var (
		iSlice      []interface{}
		hdrMap      = map[int]string{}
		managedDevs []mdevs.ManagedDevice
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
			err = cmn.ValidateHeader(hdrMap, mdevs.MobDevAttrMap)
			if err != nil {
				return err
			}
			count = count + 1
			continue
		}

		for idx, value := range record {
			iSlice[idx] = value
		}

		mngMdevVar, err := btchMngProcessMobDev(hdrMap, iSlice)
		if err != nil {
			return err
		}

		managedDevs = append(managedDevs, mngMdevVar)

		count = count + 1
	}

	err = btchMngMobDevs(ds, managedDevs)
	if err != nil {
		return err
	}
	return nil
}

func btchMngMobDevProcessJSON(ds *admin.Service, filePath string, scanner *bufio.Scanner) error {
	var managedDevs []mdevs.ManagedDevice

	if filePath != "" {
		file, err := os.Open(filePath)
		if err != nil {
			return err
		}
		defer file.Close()

		scanner = bufio.NewScanner(file)
	}

	for scanner.Scan() {
		jsonData := scanner.Text()

		mngMdevVar, err := btchMngJSONMobDev(ds, jsonData)
		if err != nil {
			return err
		}

		managedDevs = append(managedDevs, mngMdevVar)
	}
	err := scanner.Err()
	if err != nil {
		return err
	}

	err = btchMngMobDevs(ds, managedDevs)
	if err != nil {
		return err
	}

	return nil
}

func btchMngMobDevProcessSheet(ds *admin.Service, sheetID string) error {
	var managedDevs []mdevs.ManagedDevice

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
	err = cmn.ValidateHeader(hdrMap, mdevs.MobDevAttrMap)
	if err != nil {
		return err
	}

	for idx, row := range sValRange.Values {
		if idx == 0 {
			continue
		}

		mngMdevVar, err := btchMngProcessMobDev(hdrMap, row)
		if err != nil {
			return err
		}

		managedDevs = append(managedDevs, mngMdevVar)
	}

	err = btchMngMobDevs(ds, managedDevs)
	if err != nil {
		return err
	}

	return nil
}

func btchMngProcessMobDev(hdrMap map[int]string, mdevData []interface{}) (mdevs.ManagedDevice, error) {
	managedDev := mdevs.ManagedDevice{}

	for idx, attr := range mdevData {
		attrName := hdrMap[idx]

		switch {
		case attrName == "action":
			lwrAction := strings.ToLower(fmt.Sprintf("%v", attr))
			ok := cmn.SliceContainsStr(mdevs.ValidActions, lwrAction)
			if !ok {
				return managedDev, fmt.Errorf("gmin: error - %v is not a valid action type", fmt.Sprintf("%v", attr))
			}
			managedDev.Action = lwrAction
		case attrName == "resourceId":
			managedDev.ResourceId = fmt.Sprintf("%v", attr)
		}
	}

	return managedDev, nil
}

func init() {
	batchManageCmd.AddCommand(batchMngMobDevCmd)

	batchMngMobDevCmd.Flags().StringVarP(&inputFile, "inputfile", "i", "", "filepath to device data file")
	batchMngMobDevCmd.Flags().StringVarP(&format, "format", "f", "json", "user data file format")
	batchMngMobDevCmd.Flags().StringVarP(&sheetRange, "sheetrange", "s", "", "user data gsheet range")

	batchMngMobDevCmd.MarkFlagRequired("inputfile")
}
