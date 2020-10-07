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
	Use:     "chromeos-devices -i <input file>",
	Aliases: []string{"chromeos-device", "cros-devices", "cros-device", "cros-devs", "cros-dev", "cdevs", "cdev"},
	Example: `gmin batch-manage chromeos-devices -i inputfile.json
gmin bmng cdevs -i inputfile.csv -f csv
gmin bmng cdev -i 1odyAIp3jGspd3M4xeepxWD6aeQIUuHBgrZB2OHSu8MI -s 'Sheet1!A1:C25' -f gsheet`,
	Short: "Manages a batch of ChromeOS devices",
	Long: `Manages a batch of ChromeOS devices where device details are provided in a Google Sheet, CSV/JSON input file or piped JSON.
			  
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
	logger.Debugw("starting doBatchMngCrOSDev()",
		"args", args)

	ds, err := cmn.CreateDirectoryService(admin.AdminDirectoryDeviceChromeosScope)
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
		err := bmngcProcessCSVFile(ds, inputFile)
		if err != nil {
			logger.Error(err)
			return err
		}
	case lwrFmt == "json":
		err := bmngcProcessJSON(ds, inputFile, scanner)
		if err != nil {
			logger.Error(err)
			return err
		}
	case lwrFmt == "gsheet":
		err := bmngcProcessGSheet(ds, inputFile)
		if err != nil {
			logger.Error(err)
			return err
		}
	}
	logger.Debug("finished doBatchMngCrOSDev()")
	return nil
}

func bmngcFromFileFactory(hdrMap map[int]string, cdevData []interface{}) (cdevs.ManagedDevice, error) {
	logger.Debugw("starting bmngcFromFileFactory()",
		"hdrMap", hdrMap)

	managedDev := cdevs.ManagedDevice{}

	for idx, attr := range cdevData {
		attrName := hdrMap[idx]
		attrVal := fmt.Sprintf("%v", attr)
		lowerAttrVal := strings.ToLower(fmt.Sprintf("%v", attr))

		switch {
		case attrName == "action":
			ok := cmn.SliceContainsStr(cdevs.ValidActions, lowerAttrVal)
			if !ok {
				err := fmt.Errorf(cmn.ErrInvalidActionType, attrVal)
				logger.Error(err)
				return managedDev, err
			}
			managedDev.Action = lowerAttrVal
		case attrName == "deviceId":
			managedDev.DeviceId = attrVal
		case attrName == "deprovisionReason":
			if lowerAttrVal != "" {
				ok := cmn.SliceContainsStr(cdevs.ValidDeprovisionReasons, lowerAttrVal)
				if !ok {
					err := fmt.Errorf(cmn.ErrInvalidDeprovisionReason, attrVal)
					logger.Error(err)
					return managedDev, err
				}
				managedDev.DeprovisionReason = lowerAttrVal
			}
		}
	}

	if managedDev.Action == "deprovision" && managedDev.DeprovisionReason == "" {
		err := errors.New(cmn.ErrNoDeprovisionReason)
		logger.Error(err)
		return managedDev, err
	}
	logger.Debug("finished bmngcFromFileFactory()")
	return managedDev, nil
}

func bmngcFromJSONFactory(ds *admin.Service, jsonData string) (cdevs.ManagedDevice, error) {
	logger.Debugw("starting bmngcFromJSONFactory()",
		"jsonData", jsonData)

	managedDev := cdevs.ManagedDevice{}
	jsonBytes := []byte(jsonData)

	if !json.Valid(jsonBytes) {
		logger.Error(cmn.ErrInvalidJSONAttr)
		return managedDev, errors.New(cmn.ErrInvalidJSONAttr)
	}

	outStr, err := cmn.ParseInputAttrs(jsonBytes)
	if err != nil {
		logger.Error(err)
		return managedDev, err
	}

	err = cmn.ValidateInputAttrs(outStr, cdevs.CrOSDevAttrMap)
	if err != nil {
		logger.Error(err)
		return managedDev, err
	}

	err = json.Unmarshal(jsonBytes, &managedDev)
	if err != nil {
		logger.Error(err)
		return managedDev, err
	}

	logger.Debug("finished bmngcFromJSONFactory()")
	return managedDev, nil
}

func bmngcPerformAction(deviceID string, action string, wg *sync.WaitGroup, cdac *admin.ChromeosdevicesActionCall) {
	logger.Debugw("starting bmngcPerformAction()",
		"action", action,
		"deviceID", deviceID)

	defer wg.Done()

	b := backoff.NewExponentialBackOff()
	b.MaxElapsedTime = 32 * time.Second

	err := backoff.Retry(func() error {
		var err error
		err = cdac.Do()
		if err == nil {
			logger.Infof(cmn.InfoCDevActionPerformed, action, deviceID)
			fmt.Println(cmn.GminMessage(fmt.Sprintf(cmn.InfoCDevActionPerformed, action, deviceID)))
			return err
		}
		if !cmn.IsErrRetryable(err) {
			return backoff.Permanent(fmt.Errorf(cmn.ErrBatchChromeOSDevice, err.Error(), deviceID))
		}
		// Log the retries
		logger.Warnw(err.Error(),
			"retrying", b.GetElapsedTime().String(),
			"ChromeOS device", deviceID)
		return fmt.Errorf(cmn.ErrBatchChromeOSDevice, err.Error(), deviceID)
	}, b)
	if err != nil {
		// Log final error
		logger.Error(err)
		fmt.Println(cmn.GminMessage(err.Error()))
	}
	logger.Debug("finished bmngcPerformAction()")
}

func bmngcProcessCSVFile(ds *admin.Service, filePath string) error {
	logger.Debugw("starting bmngcProcessCSVFile()",
		"filePath", filePath)

	var (
		iSlice      []interface{}
		hdrMap      = map[int]string{}
		managedDevs []cdevs.ManagedDevice
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
			err = cmn.ValidateHeader(hdrMap, cdevs.CrOSDevAttrMap)
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

		mngCdevVar, err := bmngcFromFileFactory(hdrMap, iSlice)
		if err != nil {
			logger.Error(err)
			return err
		}

		managedDevs = append(managedDevs, mngCdevVar)

		count = count + 1
	}

	err = bmngcProcessObjects(ds, managedDevs)
	if err != nil {
		logger.Error(err)
		return err
	}
	logger.Debug("finished bmngcProcessCSVFile()")
	return nil
}

func bmngcProcessGSheet(ds *admin.Service, sheetID string) error {
	logger.Debugw("starting bmngcProcessGSheet()",
		"sheetID", sheetID)

	var managedDevs []cdevs.ManagedDevice

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
	err = cmn.ValidateHeader(hdrMap, cdevs.CrOSDevAttrMap)
	if err != nil {
		logger.Error(err)
		return err
	}

	for idx, row := range sValRange.Values {
		if idx == 0 {
			continue
		}

		mngCdevVar, err := bmngcFromFileFactory(hdrMap, row)
		if err != nil {
			logger.Error(err)
			return err
		}

		managedDevs = append(managedDevs, mngCdevVar)
	}

	err = bmngcProcessObjects(ds, managedDevs)
	if err != nil {
		logger.Error(err)
		return err
	}
	logger.Debug("finished bmngcProcessGSheet()")
	return nil
}

func bmngcProcessJSON(ds *admin.Service, filePath string, scanner *bufio.Scanner) error {
	logger.Debugw("starting bmngcProcessJSON()",
		"filePath", filePath)

	var managedDevs []cdevs.ManagedDevice

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

		mngCdevVar, err := bmngcFromJSONFactory(ds, jsonData)
		if err != nil {
			logger.Error(err)
			return err
		}

		managedDevs = append(managedDevs, mngCdevVar)
	}
	err := scanner.Err()
	if err != nil {
		logger.Error(err)
		return err
	}

	err = bmngcProcessObjects(ds, managedDevs)
	if err != nil {
		logger.Error(err)
		return err
	}
	logger.Debug("finished bmngcProcessJSON()")
	return nil
}

func bmngcProcessObjects(ds *admin.Service, managedDevs []cdevs.ManagedDevice) error {
	logger.Debug("starting bmngcProcessObjects()")

	customerID, err := cfg.ReadConfigString("customerid")
	if err != nil {
		logger.Error(err)
		return err
	}

	wg := new(sync.WaitGroup)

	for _, md := range managedDevs {
		devAction := admin.ChromeOsDeviceAction{}

		devAction.Action = md.Action
		devAction.DeprovisionReason = md.DeprovisionReason

		cdac := ds.Chromeosdevices.Action(customerID, md.DeviceId, &devAction)

		wg.Add(1)

		go bmngcPerformAction(md.DeviceId, md.Action, wg, cdac)
	}

	wg.Wait()

	logger.Debug("finished bmngcProcessObjects()")
	return nil
}

func init() {
	batchManageCmd.AddCommand(batchMngCrOSDevCmd)

	batchMngCrOSDevCmd.Flags().StringVarP(&inputFile, "input-file", "i", "", "filepath to device data file")
	batchMngCrOSDevCmd.Flags().StringVarP(&format, "format", "f", "json", "user data file format")
	batchMngCrOSDevCmd.Flags().StringVarP(&sheetRange, "sheet-range", "s", "", "user data gsheet range")
}
