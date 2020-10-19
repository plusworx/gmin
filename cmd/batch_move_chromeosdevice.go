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

var batchMoveCrOSDevCmd = &cobra.Command{
	Use:     "chromeos-devices -i <input file>",
	Aliases: []string{"chromeos-device", "cros-devices", "cros-device", "cros-devs", "cros-dev", "cdevs", "cdev"},
	Example: `gmin batch-move chromeos-devices -i inputfile.txt
gmin bmv cdevs -i inputfile.csv -f csv
gmin bmv cdev -i 1odyAIp3jGspd3M4xeepxWD6aeQIUuHBgrZB2OHSu8MI -s 'Sheet1!A1:C25' -f gsheet`,
	Short: "Moves a batch of ChromeOS devices to another orgunit",
	Long: `Moves a batch of ChromeOS devices to another orgunit where device details are provided in a Google Sheet, CSV/JSON input file or piped JSON.
			
The JSON file or piped input should contain ChromeOS device move details like this:

{"deviceId":"5ac7be73-5996-394e-9c30-62d41a8f10e8","orgUnitPath":"/IT"}
{"deviceId":"6ac9bd33-7095-453e-6c39-22d48a8f13e8","orgUnitPath":"/Engineering"}
{"deviceId":"6bc4be13-9916-494e-9c39-62d45c8f40e9","orgUnitPath":"/Sales"}

CSV and Google sheets must have a header row with the following column names being the only ones that are valid:

deviceId [required]
orgUnitPath [required]

The column names are case insensitive and can be in any order.`,
	RunE: doBatchMoveCrOSDev,
}

func doBatchMoveCrOSDev(cmd *cobra.Command, args []string) error {
	lg.Debugw("starting doBatchMoveCrOSDev()",
		"args", args)
	defer lg.Debug("finished doBatchMoveCrOSDev()")

	var movedDevs []cdevs.MovedDevice

	ds, err := cmn.CreateDirectoryService(admin.AdminDirectoryDeviceChromeosScope)
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
		movedDevs, err = bmvcProcessCSVFile(ds, inputFlgVal)
		if err != nil {
			return err
		}
	case lwrFmt == "json":
		movedDevs, err = bmvcProcessJSON(ds, inputFlgVal, scanner)
		if err != nil {
			return err
		}
	case lwrFmt == "gsheet":
		rangeFlgVal, err := cmd.Flags().GetString(flgnm.FLG_SHEETRANGE)
		if err != nil {
			lg.Error(err)
			return err
		}

		movedDevs, err = bmvcProcessGSheet(ds, inputFlgVal, rangeFlgVal)
		if err != nil {
			return err
		}
	default:
		return fmt.Errorf(gmess.ERR_INVALIDFILEFORMAT, formatFlgVal)
	}

	err = bmvcProcessObjects(ds, movedDevs)
	if err != nil {
		return err
	}

	return nil
}

func bmvcFromFileFactory(hdrMap map[int]string, cdevData []interface{}) (cdevs.MovedDevice, error) {
	lg.Debugw("starting bmvcFromFileFactory()",
		"hdrMap", hdrMap)
	defer lg.Debug("finished bmvcFromFileFactory()")

	movedDev := cdevs.MovedDevice{}

	for idx, attr := range cdevData {
		attrName := hdrMap[idx]
		attrVal := fmt.Sprintf("%v", attr)

		switch {
		case attrName == "deviceId":
			if attrVal == "" {
				err := fmt.Errorf(gmess.ERR_EMPTYSTRING, attrName)
				lg.Error(err)
				return movedDev, err
			}
			movedDev.DeviceId = attrVal
		case attrName == "orgUnitPath":
			if attrVal == "" {
				err := fmt.Errorf(gmess.ERR_EMPTYSTRING, attrName)
				lg.Error(err)
				return movedDev, err
			}
			movedDev.OrgUnitPath = attrVal
		}
	}
	return movedDev, nil
}

func bmvcFromJSONFactory(ds *admin.Service, jsonData string) (cdevs.MovedDevice, error) {
	lg.Debugw("starting bmvcFromJSONFactory()",
		"jsonData", jsonData)
	defer lg.Debug("finished bmvcFromJSONFactory()")

	movedDev := cdevs.MovedDevice{}
	jsonBytes := []byte(jsonData)

	if !json.Valid(jsonBytes) {
		err := errors.New(gmess.ERR_INVALIDJSONATTR)
		lg.Error(err)
		return movedDev, err
	}

	outStr, err := cmn.ParseInputAttrs(jsonBytes)
	if err != nil {
		return movedDev, err
	}

	err = cmn.ValidateInputAttrs(outStr, cdevs.CrOSDevAttrMap)
	if err != nil {
		return movedDev, err
	}

	err = json.Unmarshal(jsonBytes, &movedDev)
	if err != nil {
		lg.Error(err)
		return movedDev, err
	}
	return movedDev, nil
}

func bmvcPerformMove(deviceID string, ouPath string, wg *sync.WaitGroup, cdmc *admin.ChromeosdevicesMoveDevicesToOuCall) {
	lg.Debugw("starting bmvcPerformMove()",
		"deviceID", deviceID,
		"ouPath", ouPath)
	defer lg.Debug("finished bmvcPerformMove()")

	defer wg.Done()

	b := backoff.NewExponentialBackOff()
	b.MaxElapsedTime = 32 * time.Second

	err := backoff.Retry(func() error {
		var err error
		err = cdmc.Do()
		if err == nil {
			lg.Infof(gmess.INFO_CDEVMOVEPERFORMED, deviceID, ouPath)
			fmt.Println(cmn.GminMessage(fmt.Sprintf(gmess.INFO_CDEVMOVEPERFORMED, deviceID, ouPath)))
			return err
		}
		if !cmn.IsErrRetryable(err) {
			return backoff.Permanent(fmt.Errorf(gmess.ERR_BATCHCHROMEOSDEVICE, err.Error(), deviceID))
		}
		// Log the retries
		lg.Warnw(err.Error(),
			"retrying", b.GetElapsedTime().String(),
			"ChromeOS device", deviceID,
			"orgunit", ouPath)
		return fmt.Errorf(gmess.ERR_BATCHCHROMEOSDEVICE, err.Error(), deviceID)
	}, b)
	if err != nil {
		// Log final error
		lg.Error(err)
		fmt.Println(cmn.GminMessage(err.Error()))
	}
}

func bmvcProcessCSVFile(ds *admin.Service, filePath string) ([]cdevs.MovedDevice, error) {
	lg.Debugw("starting bmvcProcessCSVFile()",
		"filePath", filePath)
	defer lg.Debug("finished bmvcProcessCSVFile()")

	var (
		iSlice    []interface{}
		hdrMap    = map[int]string{}
		movedDevs []cdevs.MovedDevice
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

		moveCdevVar, err := bmvcFromFileFactory(hdrMap, iSlice)
		if err != nil {
			return nil, err
		}

		movedDevs = append(movedDevs, moveCdevVar)

		count = count + 1
	}

	return movedDevs, nil
}

func bmvcProcessGSheet(ds *admin.Service, sheetID string, sheetrange string) ([]cdevs.MovedDevice, error) {
	lg.Debugw("starting bmvcProcessGSheet()",
		"sheetID", sheetID,
		"sheetrange", sheetrange)
	defer lg.Debug("finished bmvcProcessGSheet()")

	var movedDevs []cdevs.MovedDevice

	if sheetrange == "" {
		err := errors.New(gmess.ERR_NOSHEETRANGE)
		lg.Error(err)
		return nil, err
	}

	ss, err := cmn.CreateSheetService(sheet.DriveReadonlyScope)
	if err != nil {
		return nil, err
	}

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

		moveCdevVar, err := bmvcFromFileFactory(hdrMap, row)
		if err != nil {
			return nil, err
		}

		movedDevs = append(movedDevs, moveCdevVar)
	}

	return movedDevs, nil
}

func bmvcProcessJSON(ds *admin.Service, filePath string, scanner *bufio.Scanner) ([]cdevs.MovedDevice, error) {
	lg.Debugw("starting bmvcProcessJSON()",
		"filePath", filePath)
	defer lg.Debug("finished bmvcProcessJSON()")

	var movedDevs []cdevs.MovedDevice

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

		moveCdevVar, err := bmvcFromJSONFactory(ds, jsonData)
		if err != nil {
			return nil, err
		}

		movedDevs = append(movedDevs, moveCdevVar)
	}
	err := scanner.Err()
	if err != nil {
		lg.Error(err)
		return nil, err
	}

	return movedDevs, nil
}

func bmvcProcessObjects(ds *admin.Service, movedDevs []cdevs.MovedDevice) error {
	lg.Debug("starting bmvcProcessObjects()")
	defer lg.Debug("finished bmvcProcessObjects()")

	customerID, err := cfg.ReadConfigString(cfg.CONFIGCUSTID)
	if err != nil {
		lg.Error(err)
		return err
	}

	wg := new(sync.WaitGroup)

	for _, md := range movedDevs {
		move := admin.ChromeOsMoveDevicesToOu{}
		deviceIDs := []string{}

		deviceIDs = append(deviceIDs, md.DeviceId)
		move.DeviceIds = deviceIDs

		cdmc := ds.Chromeosdevices.MoveDevicesToOu(customerID, md.OrgUnitPath, &move)

		wg.Add(1)

		go bmvcPerformMove(md.DeviceId, md.OrgUnitPath, wg, cdmc)
	}

	wg.Wait()

	return nil
}

func init() {
	batchMoveCmd.AddCommand(batchMoveCrOSDevCmd)

	batchMoveCrOSDevCmd.Flags().StringVarP(&inputFile, flgnm.FLG_INPUTFILE, "i", "", "filepath to device data file")
	batchMoveCrOSDevCmd.Flags().StringVarP(&format, flgnm.FLG_FORMAT, "f", "json", "user data file format")
	batchMoveCrOSDevCmd.Flags().StringVarP(&sheetRange, flgnm.FLG_SHEETRANGE, "s", "", "user data gsheet range")
}
