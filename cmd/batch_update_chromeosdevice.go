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

var batchUpdCrOSDevCmd = &cobra.Command{
	Use:     "chromeos-devices -i <input file>",
	Aliases: []string{"chromeos-device", "cros-devices", "cros-device", "cros-devs", "cros-dev", "cdevs", "cdev"},
	Example: `gmin batch-update chromeos-devices -i inputfile.json
gmin bupd cdevs -i inputfile.csv -f csv
gmin bupd cdev -i 1odyAIp3jGspd3M4xeepxWD6aeQIUuHBgrZB2OHSu8MI -s 'Sheet1!A1:K25' -f gsheet`,
	Short: "Updates a batch of ChromeOS devices",
	Long: `Updates a batch of ChromeOS devices with device details provided in a Google Sheet, CSV/JSON input file or piped JSON.
			
The JSON file or piped input should contain device update details like this:

{"deviceId":"5ac7be43-5906-394e-7c39-62d45a8f10e8","annotatedAssetId":"CB1","annotatedLocation":"Batcave","annotatedUser":"Bruce Wayne","notes":"Test machine","orgUnitPath":"/Anticrime"}
{"deviceId":"4ac7be43-5906-394e-7c39-62d45a8f10e8","annotatedAssetId":"CB2","annotatedLocation":"Wayne Manor","annotatedUser":"Alfred Pennyworth","notes":"Another test machine","orgUnitPath":"/Anticorruption"}
{"deviceId":"3ac7be43-5906-394e-7c39-62d45a8f10e8","annotatedAssetId":"CB3","annotatedLocation":"Wayne Towers","annotatedUser":"The Big Enchilada","notes":"Yet another test machine","orgUnitPath":"/Legal"}

N.B. deviceId must be provided.

CSV and Google sheets must have a header row with the following column names being the only ones that are valid:

annotatedAssetId
annotatedLocation
annotatedUser
deviceId [required]
notes
orgUnitPath

The column names are case insensitive and can be in any order.`,
	RunE: doBatchUpdCrOSDev,
}

func doBatchUpdCrOSDev(cmd *cobra.Command, args []string) error {
	lg.Debugw("starting doBatchUpdCrOSDev()",
		"args", args)
	defer lg.Debug("finished doBatchUpdCrOSDev()")

	var crosdevs []*admin.ChromeOsDevice

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
		crosdevs, err = bucProcessCSVFile(ds, inputFlgVal)
		if err != nil {
			return err
		}
	case lwrFmt == "json":
		crosdevs, err = bucProcessJSON(ds, inputFlgVal, scanner)
		if err != nil {
			return err
		}
	case lwrFmt == "gsheet":
		rangeFlgVal, err := cmd.Flags().GetString(flgnm.FLG_SHEETRANGE)
		if err != nil {
			lg.Error(err)
			return err
		}

		crosdevs, err = bucProcessGSheet(ds, inputFlgVal, rangeFlgVal)
		if err != nil {
			return err
		}
	default:
		err = fmt.Errorf(gmess.ERR_INVALIDFILEFORMAT, formatFlgVal)
		lg.Error(err)
		return err
	}

	err = bucProcessObjects(ds, crosdevs)
	if err != nil {
		return err
	}

	return nil
}

func bucFromFileFactory(hdrMap map[int]string, cdevData []interface{}) (*admin.ChromeOsDevice, error) {
	lg.Debugw("starting bucFromFileFactory()",
		"hdrMap", hdrMap)
	defer lg.Debug("finished bucFromFileFactory()")

	var crosdev *admin.ChromeOsDevice

	crosdev = new(admin.ChromeOsDevice)

	for idx, attr := range cdevData {
		attrName := hdrMap[idx]
		attrVal := fmt.Sprintf("%v", attr)

		switch {
		case attrName == "annotatedAssetId":
			crosdev.AnnotatedAssetId = attrVal
			if assetID == "" {
				crosdev.ForceSendFields = append(crosdev.ForceSendFields, "AnnotatedAssetId")
			}
		case attrName == "annotatedLocation":
			crosdev.AnnotatedLocation = attrVal
			if location == "" {
				crosdev.ForceSendFields = append(crosdev.ForceSendFields, "AnnotatedLocation")
			}
		case attrName == "annotatedUser":
			crosdev.AnnotatedUser = attrVal
			if attrVal == "" {
				crosdev.ForceSendFields = append(crosdev.ForceSendFields, "AnnotatedUser")
			}
		case attrName == "notes":
			crosdev.Notes = attrVal
			if notes == "" {
				crosdev.ForceSendFields = append(crosdev.ForceSendFields, "Notes")
			}
		case attrName == "orgUnitPath":
			crosdev.OrgUnitPath = attrVal
		}
	}
	return crosdev, nil
}

func bucFromJSONFactory(ds *admin.Service, jsonData string) (*admin.ChromeOsDevice, error) {
	lg.Debugw("starting bucFromJSONFactory()",
		"jsonData", jsonData)
	defer lg.Debug("finished bucFromJSONFactory()")

	var (
		crosdev   *admin.ChromeOsDevice
		emptyVals = cmn.EmptyValues{}
	)

	crosdev = new(admin.ChromeOsDevice)
	jsonBytes := []byte(jsonData)

	if !json.Valid(jsonBytes) {
		err := errors.New(gmess.ERR_INVALIDJSONATTR)
		lg.Error(err)
		return nil, err
	}

	outStr, err := cmn.ParseInputAttrs(jsonBytes)
	if err != nil {
		return nil, err
	}

	err = cmn.ValidateInputAttrs(outStr, cdevs.CrOSDevAttrMap)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(jsonBytes, &crosdev)
	if err != nil {
		lg.Error(err)
		return nil, err
	}

	if crosdev.DeviceId == "" {
		err = errors.New(gmess.ERR_NOJSONDEVICEID)
		lg.Error(err)
		return nil, err
	}

	err = json.Unmarshal(jsonBytes, &emptyVals)
	if err != nil {
		lg.Error(err)
		return nil, err
	}
	if len(emptyVals.ForceSendFields) > 0 {
		crosdev.ForceSendFields = emptyVals.ForceSendFields
	}
	return crosdev, nil
}

func bucProcessCSVFile(ds *admin.Service, filePath string) ([]*admin.ChromeOsDevice, error) {
	lg.Debugw("starting bucProcessCSVFile()",
		"filePath", filePath)
	defer lg.Debug("finished bucProcessCSVFile()")

	var (
		iSlice   []interface{}
		hdrMap   = map[int]string{}
		crosdevs []*admin.ChromeOsDevice
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

		cdevVar, err := bucFromFileFactory(hdrMap, iSlice)
		if err != nil {
			return nil, err
		}

		crosdevs = append(crosdevs, cdevVar)

		count = count + 1
	}

	return crosdevs, nil
}
func bucProcessGSheet(ds *admin.Service, sheetID string, sheetrange string) ([]*admin.ChromeOsDevice, error) {
	lg.Debugw("starting bucProcessGSheet()",
		"sheetID", sheetID,
		"sheetrange", sheetrange)
	defer lg.Debug("finished bucProcessGSheet()")

	var crosdevs []*admin.ChromeOsDevice

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

		cdevVar, err := bucFromFileFactory(hdrMap, row)
		if err != nil {
			return nil, err
		}

		crosdevs = append(crosdevs, cdevVar)
	}

	return crosdevs, nil
}

func bucProcessJSON(ds *admin.Service, filePath string, scanner *bufio.Scanner) ([]*admin.ChromeOsDevice, error) {
	lg.Debugw("starting bucProcessJSON()",
		"filePath", filePath)
	defer lg.Debug("finished bucProcessJSON()")

	var crosdevs []*admin.ChromeOsDevice

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

		cdevVar, err := bucFromJSONFactory(ds, jsonData)
		if err != nil {
			return nil, err
		}

		crosdevs = append(crosdevs, cdevVar)
	}
	err := scanner.Err()
	if err != nil {
		lg.Error(err)
		return nil, err
	}

	return crosdevs, nil
}

func bucProcessObjects(ds *admin.Service, crosdevs []*admin.ChromeOsDevice) error {
	lg.Debug("starting bucProcessObjects()")
	wg := new(sync.WaitGroup)
	defer lg.Debug("finished bucProcessObjects()")

	customerID, err := cfg.ReadConfigString(cfg.CONFIGCUSTID)
	if err != nil {
		lg.Error(err)
		return err
	}

	for _, c := range crosdevs {
		cduc := ds.Chromeosdevices.Update(customerID, c.DeviceId, c)

		wg.Add(1)

		go bucUpdate(c, wg, cduc)
	}

	wg.Wait()

	return nil
}

func bucUpdate(crosdev *admin.ChromeOsDevice, wg *sync.WaitGroup, cduc *admin.ChromeosdevicesUpdateCall) {
	lg.Debug("starting bucUpdate()")
	defer lg.Debug("finished bucUpdate()")

	defer wg.Done()

	b := backoff.NewExponentialBackOff()
	b.MaxElapsedTime = 32 * time.Second

	err := backoff.Retry(func() error {
		var err error
		_, err = cduc.Do()
		if err == nil {
			fmt.Println(cmn.GminMessage(fmt.Sprintf(gmess.INFO_CDEVUPDATED, crosdev.DeviceId)))
			lg.Infof(gmess.INFO_CDEVUPDATED, crosdev.DeviceId)
			return err
		}
		if !cmn.IsErrRetryable(err) {
			return backoff.Permanent(fmt.Errorf(gmess.ERR_BATCHCHROMEOSDEVICE, err.Error(), crosdev.DeviceId))
		}
		// Log the retries
		lg.Warnw(err.Error(),
			"retrying", b.GetElapsedTime().String(),
			"ChromeOS device", crosdev.DeviceId)
		return fmt.Errorf(gmess.ERR_BATCHCHROMEOSDEVICE, err.Error(), crosdev.DeviceId)
	}, b)
	if err != nil {
		// Log final error
		lg.Error(err)
		fmt.Println(cmn.GminMessage(err.Error()))
	}
}

func init() {
	batchUpdateCmd.AddCommand(batchUpdCrOSDevCmd)

	batchUpdCrOSDevCmd.Flags().StringVarP(&inputFile, flgnm.FLG_INPUTFILE, "i", "", "filepath to device data file or sheet id")
	batchUpdCrOSDevCmd.Flags().StringVarP(&format, flgnm.FLG_FORMAT, "f", "json", "user data file format")
	batchUpdCrOSDevCmd.Flags().StringVarP(&sheetRange, flgnm.FLG_SHEETRANGE, "s", "", "user data gsheet range")
}
