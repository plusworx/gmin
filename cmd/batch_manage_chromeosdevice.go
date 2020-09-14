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
	cdevs "github.com/plusworx/gmin/utils/chromeosdevices"
	cmn "github.com/plusworx/gmin/utils/common"
	cfg "github.com/plusworx/gmin/utils/config"
	"github.com/spf13/cobra"
	admin "google.golang.org/api/admin/directory/v1"
	sheet "google.golang.org/api/sheets/v4"
)

var batchMngCrOSDevCmd = &cobra.Command{
	Use:     "chromeosdevices -i <input file>",
	Aliases: []string{"chromeosdevice", "crosdevices", "crosdevice", "crosdevs", "crosdev", "cdevs", "cdev"},
	Short:   "Manages a batch of ChromeOS devices",
	Long: `Manages a batch of ChromeOS devices where device details are provided in a Google Sheet, CSV/JSON input file or piped JSON.
	
	Examples:	gmin batch-manage chromeosdevices -i inputfile.json
			gmin bmng cdevs -i inputfile.csv -f csv
			gmin bmng cdev -i 1odyAIp3jGspd3M4xeepxWD6aeQIUuHBgrZB2OHSu8MI -s 'Sheet1!A1:C25' -f gsheet
			  
	The JSON file or piped input should contain ChromeOS device management details like this:
	
	{"deviceId":"5ac7be73-5996-394e-9c30-62d41a8f10e8","action":"disable"}
	{"deviceId":"6ac9bd33-7095-453e-6c39-22d48a8f13e8","action":"reenable"}
	{"deviceId":"6bc4be13-9916-494e-9c39-62d45c8f40e9","action":"deprovision","deprovisionReason":"retiring_device"}
	
	N.B. If you are deprovisioning devices then a deprovision reason must be provided.
	
	CSV and Google sheets must have a header row with the following column names being the only ones that are valid:
	
	action [required]
	deprovisionReason
	deviceId [required]
	
	The column names are case insensitive and can be in any order.
	
	Valid actions are:
	deprovision
	disable
	reenable
	
	Valid deprovision reasons are:
	different_model_replacement
	retiring_device
	same_model_replacement
	upgrade_transfer`,
	RunE: doBatchMngCrOSDev,
}

func doBatchMngCrOSDev(cmd *cobra.Command, args []string) error {
	ds, err := cmn.CreateDirectoryService(admin.AdminDirectoryDeviceChromeosScope)
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
		err := btchMngCrOSDevProcessCSV(ds, inputFile)
		if err != nil {
			return err
		}
	case lwrFmt == "json":
		err := btchMngCrOSDevProcessJSON(ds, inputFile, scanner)
		if err != nil {
			return err
		}
	case lwrFmt == "gsheet":
		err := btchMngCrOSDevProcessSheet(ds, inputFile)
		if err != nil {
			return err
		}
	}

	return nil
}

func btchMngJSONCrOSDev(ds *admin.Service, jsonData string) (cdevs.ManagedDevice, error) {
	managedDev := cdevs.ManagedDevice{}
	jsonBytes := []byte(jsonData)

	if !json.Valid(jsonBytes) {
		return managedDev, errors.New("gmin: error - attribute string is not valid JSON")
	}

	outStr, err := cmn.ParseInputAttrs(jsonBytes)
	if err != nil {
		return managedDev, err
	}

	err = cmn.ValidateInputAttrs(outStr, cdevs.CrOSDevAttrMap)
	if err != nil {
		return managedDev, err
	}

	err = json.Unmarshal(jsonBytes, &managedDev)
	if err != nil {
		return managedDev, err
	}

	return managedDev, nil
}

func btchMngCrOSDevs(ds *admin.Service, managedDevs []cdevs.ManagedDevice) error {
	customerID, err := cfg.ReadConfigString("customerid")
	if err != nil {
		return err
	}

	wg := new(sync.WaitGroup)

	for _, md := range managedDevs {
		devAction := admin.ChromeOsDeviceAction{}

		devAction.Action = md.Action
		devAction.DeprovisionReason = md.DeprovisionReason

		cdac := ds.Chromeosdevices.Action(customerID, md.DeviceId, &devAction)

		wg.Add(1)

		go btchMngCrOSDevProcess(md.DeviceId, md.Action, wg, cdac)
	}

	wg.Wait()

	return nil
}

func btchMngCrOSDevProcess(deviceID string, action string, wg *sync.WaitGroup, cdac *admin.ChromeosdevicesActionCall) {
	defer wg.Done()

	b := backoff.NewExponentialBackOff()
	b.MaxElapsedTime = 32 * time.Second

	err := backoff.Retry(func() error {
		var err error
		err = cdac.Do()
		if err == nil {
			fmt.Println(cmn.GminMessage("**** gmin: " + action + " successfully performed on ChromeOS device " + deviceID + " ****"))
			return err
		}
		if !cmn.IsErrRetryable(err) {
			return backoff.Permanent(errors.New(cmn.GminMessage("gmin: error - " + err.Error() + " " + deviceID)))
		}
		return errors.New(cmn.GminMessage("gmin: error - " + err.Error() + " " + deviceID))
	}, b)
	if err != nil {
		fmt.Println(err)
	}
}

func btchMngCrOSDevProcessCSV(ds *admin.Service, filePath string) error {
	var (
		iSlice      []interface{}
		hdrMap      = map[int]string{}
		managedDevs []cdevs.ManagedDevice
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
			err = cmn.ValidateHeader(hdrMap, cdevs.CrOSDevAttrMap)
			if err != nil {
				return err
			}
			count = count + 1
			continue
		}

		for idx, value := range record {
			iSlice[idx] = value
		}

		mngCdevVar, err := btchMngProcessCrOSDev(hdrMap, iSlice)
		if err != nil {
			return err
		}

		managedDevs = append(managedDevs, mngCdevVar)

		count = count + 1
	}

	err = btchMngCrOSDevs(ds, managedDevs)
	if err != nil {
		return err
	}
	return nil
}

func btchMngCrOSDevProcessJSON(ds *admin.Service, filePath string, scanner *bufio.Scanner) error {
	var managedDevs []cdevs.ManagedDevice

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

		mngCdevVar, err := btchMngJSONCrOSDev(ds, jsonData)
		if err != nil {
			return err
		}

		managedDevs = append(managedDevs, mngCdevVar)
	}
	err := scanner.Err()
	if err != nil {
		return err
	}

	err = btchMngCrOSDevs(ds, managedDevs)
	if err != nil {
		return err
	}

	return nil
}

func btchMngCrOSDevProcessSheet(ds *admin.Service, sheetID string) error {
	var managedDevs []cdevs.ManagedDevice

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
	err = cmn.ValidateHeader(hdrMap, cdevs.CrOSDevAttrMap)
	if err != nil {
		return err
	}

	for idx, row := range sValRange.Values {
		if idx == 0 {
			continue
		}

		mngCdevVar, err := btchMngProcessCrOSDev(hdrMap, row)
		if err != nil {
			return err
		}

		managedDevs = append(managedDevs, mngCdevVar)
	}

	err = btchMngCrOSDevs(ds, managedDevs)
	if err != nil {
		return err
	}

	return nil
}

func btchMngProcessCrOSDev(hdrMap map[int]string, cdevData []interface{}) (cdevs.ManagedDevice, error) {
	managedDev := cdevs.ManagedDevice{}

	for idx, attr := range cdevData {
		attrName := hdrMap[idx]

		switch {
		case attrName == "action":
			lwrAction := strings.ToLower(fmt.Sprintf("%v", attr))
			ok := cmn.SliceContainsStr(cdevs.ValidActions, lwrAction)
			if !ok {
				return managedDev, fmt.Errorf("gmin: error - %v is not a valid action type", fmt.Sprintf("%v", attr))
			}
			managedDev.Action = lwrAction
		case attrName == "deviceId":
			managedDev.DeviceId = fmt.Sprintf("%v", attr)
		case attrName == "deprovisionReason":
			lwrReason := strings.ToLower(fmt.Sprintf("%v", attr))
			if lwrReason != "" {
				ok := cmn.SliceContainsStr(cdevs.ValidDeprovisionReasons, lwrReason)
				if !ok {
					return managedDev, fmt.Errorf("gmin: error - %v is not a valid deprovision reason", fmt.Sprintf("%v", attr))
				}
			}
			managedDev.DeprovisionReason = fmt.Sprintf("%v", attr)
		}
	}

	if managedDev.Action == "deprovision" && managedDev.DeprovisionReason == "" {
		return managedDev, errors.New("gmin: error - must provide a deprovision reason")
	}

	return managedDev, nil
}

func init() {
	batchManageCmd.AddCommand(batchMngCrOSDevCmd)

	batchMngCrOSDevCmd.Flags().StringVarP(&inputFile, "inputfile", "i", "", "filepath to device data file")
	batchMngCrOSDevCmd.Flags().StringVarP(&format, "format", "f", "json", "user data file format")
	batchMngCrOSDevCmd.Flags().StringVarP(&sheetRange, "sheetrange", "s", "", "user data gsheet range")

	batchMngCrOSDevCmd.MarkFlagRequired("inputfile")
}
