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
	flgnm "github.com/plusworx/gmin/utils/flagnames"
	gmess "github.com/plusworx/gmin/utils/gminmessages"
	lg "github.com/plusworx/gmin/utils/logging"
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
	lg.Debugw("starting doBatchMngCrOSDev()",
		"args", args)
	defer lg.Debug("finished doBatchMngCrOSDev()")

	var managedDevs []cdevs.ManagedDevice

	srv, err := cmn.CreateService(cmn.SRVTYPEADMIN, admin.AdminDirectoryDeviceChromeosScope)
	if err != nil {
		return err
	}
	ds := srv.(*admin.Service)

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
		managedDevs, err = bmngcProcessCSVFile(ds, inputFlgVal)
		if err != nil {
			return err
		}
	case lwrFmt == "json":
		managedDevs, err = bmngcProcessJSON(ds, inputFlgVal, scanner)
		if err != nil {
			return err
		}
	case lwrFmt == "gsheet":
		rangeFlgVal, err := cmd.Flags().GetString(flgnm.FLG_SHEETRANGE)
		if err != nil {
			return err
		}

		managedDevs, err = bmngcProcessGSheet(ds, inputFlgVal, rangeFlgVal)
		if err != nil {
			return err
		}
	default:
		err = fmt.Errorf(gmess.ERR_INVALIDFILEFORMAT, formatFlgVal)
		lg.Error(err)
		return err
	}

	err = bmngcProcessObjects(ds, managedDevs)
	if err != nil {
		return err
	}

	return nil
}

func bmngcFromFileFactory(hdrMap map[int]string, cdevData []interface{}) (cdevs.ManagedDevice, error) {
	lg.Debugw("starting bmngcFromFileFactory()",
		"hdrMap", hdrMap)
	defer lg.Debug("finished bmngcFromFileFactory()")

	managedDev := cdevs.ManagedDevice{}

	for idx, attr := range cdevData {
		attrName := hdrMap[idx]
		attrVal := fmt.Sprintf("%v", attr)
		lowerAttrVal := strings.ToLower(fmt.Sprintf("%v", attr))

		switch {
		case attrName == "action":
			ok := cmn.SliceContainsStr(cdevs.ValidActions, lowerAttrVal)
			if !ok {
				err := fmt.Errorf(gmess.ERR_INVALIDACTIONTYPE, attrVal)
				lg.Error(err)
				return managedDev, err
			}
			managedDev.Action = lowerAttrVal
		case attrName == "deviceId":
			managedDev.DeviceId = attrVal
		case attrName == "deprovisionReason":
			if lowerAttrVal != "" {
				ok := cmn.SliceContainsStr(cdevs.ValidDeprovisionReasons, lowerAttrVal)
				if !ok {
					err := fmt.Errorf(gmess.ERR_INVALIDDEPROVISIONREASON, attrVal)
					lg.Error(err)
					return managedDev, err
				}
				managedDev.DeprovisionReason = lowerAttrVal
			}
		}
	}

	if managedDev.Action == "deprovision" && managedDev.DeprovisionReason == "" {
		err := errors.New(gmess.ERR_NODEPROVISIONREASON)
		lg.Error(err)
		return managedDev, err
	}
	return managedDev, nil
}

func bmngcFromJSONFactory(ds *admin.Service, jsonData string) (cdevs.ManagedDevice, error) {
	lg.Debugw("starting bmngcFromJSONFactory()",
		"jsonData", jsonData)
	defer lg.Debug("finished bmngcFromJSONFactory()")

	managedDev := cdevs.ManagedDevice{}
	jsonBytes := []byte(jsonData)

	if !json.Valid(jsonBytes) {
		lg.Error(gmess.ERR_INVALIDJSONATTR)
		return managedDev, errors.New(gmess.ERR_INVALIDJSONATTR)
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
		lg.Error(err)
		return managedDev, err
	}

	return managedDev, nil
}

func bmngcPerformAction(deviceID string, action string, wg *sync.WaitGroup, cdac *admin.ChromeosdevicesActionCall) {
	lg.Debugw("starting bmngcPerformAction()",
		"action", action,
		"deviceID", deviceID)
	defer lg.Debug("finished bmngcPerformAction()")

	defer wg.Done()

	b := backoff.NewExponentialBackOff()
	b.MaxElapsedTime = 32 * time.Second

	err := backoff.Retry(func() error {
		var err error
		err = cdac.Do()
		if err == nil {
			fmt.Println(cmn.GminMessage(fmt.Sprintf(gmess.INFO_CDEVACTIONPERFORMED, action, deviceID)))
			lg.Infof(gmess.INFO_CDEVACTIONPERFORMED, action, deviceID)
			return err
		}
		if !cmn.IsErrRetryable(err) {
			return backoff.Permanent(fmt.Errorf(gmess.ERR_BATCHCHROMEOSDEVICE, err.Error(), deviceID))
		}
		// Log the retries
		lg.Warnw(err.Error(),
			"retrying", b.GetElapsedTime().String(),
			"ChromeOS device", deviceID)
		return fmt.Errorf(gmess.ERR_BATCHCHROMEOSDEVICE, err.Error(), deviceID)
	}, b)
	if err != nil {
		// Log final error
		lg.Error(err)
		fmt.Println(cmn.GminMessage(err.Error()))
	}
}

func bmngcProcessCSVFile(ds *admin.Service, filePath string) ([]cdevs.ManagedDevice, error) {
	lg.Debugw("starting bmngcProcessCSVFile()",
		"filePath", filePath)
	defer lg.Debug("finished bmngcProcessCSVFile()")

	var (
		iSlice      []interface{}
		hdrMap      = map[int]string{}
		managedDevs []cdevs.ManagedDevice
	)

	csvfile, err := os.Open(filePath)
	if err != nil {
		lg.Error(err)
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
			lg.Error(err)
			return nil, err
		}

		if count == 0 {
			iSlice = make([]interface{}, len(record))
			for idx, value := range record {
				iSlice[idx] = value
			}
			hdrMap = cmn.ProcessHeader(iSlice)
			err = cmn.ValidateHeader(hdrMap, cdevs.CrOSDevAttrMap)
			if err != nil {
				return nil, err
			}
			count = count + 1
			continue
		}

		for idx, value := range record {
			iSlice[idx] = value
		}

		mngCdevVar, err := bmngcFromFileFactory(hdrMap, iSlice)
		if err != nil {
			return nil, err
		}

		managedDevs = append(managedDevs, mngCdevVar)

		count = count + 1
	}

	return managedDevs, nil
}

func bmngcProcessGSheet(ds *admin.Service, sheetID string, sheetrange string) ([]cdevs.ManagedDevice, error) {
	lg.Debugw("starting bmngcProcessGSheet()",
		"sheetID", sheetID,
		"sheetrange", sheetrange)
	defer lg.Debug("finished bmngcProcessGSheet()")

	var managedDevs []cdevs.ManagedDevice

	if sheetrange == "" {
		err := errors.New(gmess.ERR_NOSHEETRANGE)
		lg.Error(err)
		return nil, err
	}

	srv, err := cmn.CreateService(cmn.SRVTYPESHEET, sheet.DriveReadonlyScope)
	if err != nil {
		return nil, err
	}
	ss := srv.(*sheet.Service)

	ssvgc := ss.Spreadsheets.Values.Get(sheetID, sheetrange)
	sValRange, err := ssvgc.Do()
	if err != nil {
		lg.Error(err)
		return nil, err
	}

	if len(sValRange.Values) == 0 {
		err = fmt.Errorf(gmess.ERR_NOSHEETDATAFOUND, sheetID, sheetrange)
		lg.Error(err)
		return nil, err
	}

	hdrMap := cmn.ProcessHeader(sValRange.Values[0])
	err = cmn.ValidateHeader(hdrMap, cdevs.CrOSDevAttrMap)
	if err != nil {
		return nil, err
	}

	for idx, row := range sValRange.Values {
		if idx == 0 {
			continue
		}

		mngCdevVar, err := bmngcFromFileFactory(hdrMap, row)
		if err != nil {
			return nil, err
		}

		managedDevs = append(managedDevs, mngCdevVar)
	}

	return managedDevs, nil
}

func bmngcProcessJSON(ds *admin.Service, filePath string, scanner *bufio.Scanner) ([]cdevs.ManagedDevice, error) {
	lg.Debugw("starting bmngcProcessJSON()",
		"filePath", filePath)
	defer lg.Debug("finished bmngcProcessJSON()")

	var managedDevs []cdevs.ManagedDevice

	if filePath != "" {
		file, err := os.Open(filePath)
		if err != nil {
			lg.Error(err)
			return nil, err
		}
		defer file.Close()

		scanner = bufio.NewScanner(file)
	}

	for scanner.Scan() {
		jsonData := scanner.Text()

		mngCdevVar, err := bmngcFromJSONFactory(ds, jsonData)
		if err != nil {
			return nil, err
		}

		managedDevs = append(managedDevs, mngCdevVar)
	}
	err := scanner.Err()
	if err != nil {
		lg.Error(err)
		return nil, err
	}

	return managedDevs, nil
}

func bmngcProcessObjects(ds *admin.Service, managedDevs []cdevs.ManagedDevice) error {
	lg.Debug("starting bmngcProcessObjects()")
	defer lg.Debug("finished bmngcProcessObjects()")

	customerID, err := cfg.ReadConfigString(cfg.CONFIGCUSTID)
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

		go bmngcPerformAction(md.DeviceId, md.Action, wg, cdac)
	}

	wg.Wait()

	return nil
}

func init() {
	batchManageCmd.AddCommand(batchMngCrOSDevCmd)

	batchMngCrOSDevCmd.Flags().StringVarP(&inputFile, flgnm.FLG_INPUTFILE, "i", "", "filepath to device data file")
	batchMngCrOSDevCmd.Flags().StringVarP(&format, flgnm.FLG_FORMAT, "f", "json", "user data file format")
	batchMngCrOSDevCmd.Flags().StringVarP(&sheetRange, flgnm.FLG_SHEETRANGE, "s", "", "user data gsheet range")
}
