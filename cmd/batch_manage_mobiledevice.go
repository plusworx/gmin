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
	Use:     "mobile-devices -i <input file>",
	Aliases: []string{"mobile-device", "mob-devices", "mob-device", "mob-devs", "mob-dev", "mdevs", "mdev"},
	Example: `gmin batch-manage mobile-devices -i inputfile.json
gmin bmng mdevs -i inputfile.csv -f csv
gmin bmng mdev -i 1odyAIp3jGspd3M4xeepxWD6aeQIUuHBgrZB2OHSu8MI -s 'Sheet1!A1:B25' -f gsheet`,
	Short: "Manages a batch of mobile devices",
	Long: `Manages a batch of mobile devices where device details are provided in a Google Sheet, CSV/JSON input file or piped JSON.
			  
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
	logger.Debugw("starting doBatchMngMobDev()",
		"args", args)

	ds, err := cmn.CreateDirectoryService(admin.AdminDirectoryDeviceMobileActionScope)
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
		err := bmngmProcessCSVFile(ds, inputFile)
		if err != nil {
			logger.Error(err)
			return err
		}
	case lwrFmt == "json":
		err := bmngmProcessJSON(ds, inputFile, scanner)
		if err != nil {
			logger.Error(err)
			return err
		}
	case lwrFmt == "gsheet":
		err := bmngmProcessGSheet(ds, inputFile)
		if err != nil {
			logger.Error(err)
			return err
		}
	default:
		return fmt.Errorf(cmn.ErrInvalidFileFormat, format)
	}
	logger.Debug("finished doBatchMngMobDev()")
	return nil
}

func bmngmFromFileFactory(hdrMap map[int]string, mdevData []interface{}) (mdevs.ManagedDevice, error) {
	logger.Debugw("starting bmngmFromFileFactory()",
		"hdrMap", hdrMap)

	managedDev := mdevs.ManagedDevice{}

	for idx, attr := range mdevData {
		attrName := hdrMap[idx]
		attrVal := fmt.Sprintf("%v", attr)
		lowerAttrVal := strings.ToLower(fmt.Sprintf("%v", attr))

		switch {
		case attrName == "action":
			ok := cmn.SliceContainsStr(mdevs.ValidActions, lowerAttrVal)
			if !ok {
				err := fmt.Errorf(cmn.ErrInvalidActionType, attrVal)
				logger.Error(err)
				return managedDev, err
			}
			managedDev.Action = lowerAttrVal
		case attrName == "resourceId":
			managedDev.ResourceId = attrVal
		}
	}
	logger.Debug("finished bmngmFromFileFactory()")
	return managedDev, nil
}

func bmngmFromJSONFactory(ds *admin.Service, jsonData string) (mdevs.ManagedDevice, error) {
	logger.Debugw("starting bmngmFromJSONFactory()",
		"jsonData", jsonData)

	managedDev := mdevs.ManagedDevice{}
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

	err = cmn.ValidateInputAttrs(outStr, mdevs.MobDevAttrMap)
	if err != nil {
		logger.Error(err)
		return managedDev, err
	}

	err = json.Unmarshal(jsonBytes, &managedDev)
	if err != nil {
		logger.Error(err)
		return managedDev, err
	}
	logger.Debug("finished bmngmFromJSONFactory()")
	return managedDev, nil
}

func bmngmPerformAction(resourceID string, action string, wg *sync.WaitGroup, mdac *admin.MobiledevicesActionCall) {
	logger.Debugw("starting bmngmPerformAction()",
		"action", action,
		"resourceID", resourceID)

	defer wg.Done()

	b := backoff.NewExponentialBackOff()
	b.MaxElapsedTime = 32 * time.Second

	err := backoff.Retry(func() error {
		var err error
		err = mdac.Do()
		if err == nil {
			logger.Infof(cmn.InfoMDevActionPerformed, action, resourceID)
			fmt.Println(cmn.GminMessage(fmt.Sprintf(cmn.InfoMDevActionPerformed, action, resourceID)))
			return err
		}
		if !cmn.IsErrRetryable(err) {
			return backoff.Permanent(fmt.Errorf(cmn.ErrBatchMobileDevice, err.Error(), resourceID))
		}
		// Log the retries
		logger.Warnw(err.Error(),
			"retrying", b.GetElapsedTime().String(),
			"mobile device", resourceID)
		return fmt.Errorf(cmn.ErrBatchMobileDevice, err.Error(), resourceID)
	}, b)
	if err != nil {
		// Log final error
		logger.Error(err)
		fmt.Println(cmn.GminMessage(err.Error()))
	}
	logger.Debug("finished bmngmPerformAction()")
}

func bmngmProcessCSVFile(ds *admin.Service, filePath string) error {
	logger.Debugw("starting bmngmProcessCSVFile()",
		"filePath", filePath)

	var (
		iSlice      []interface{}
		hdrMap      = map[int]string{}
		managedDevs []mdevs.ManagedDevice
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
			err = cmn.ValidateHeader(hdrMap, mdevs.MobDevAttrMap)
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

		mngMdevVar, err := bmngmFromFileFactory(hdrMap, iSlice)
		if err != nil {
			logger.Error(err)
			return err
		}

		managedDevs = append(managedDevs, mngMdevVar)

		count = count + 1
	}

	err = bmngmProcessObjects(ds, managedDevs)
	if err != nil {
		logger.Error(err)
		return err
	}
	logger.Debug("finished bmngmProcessCSVFile()")
	return nil
}

func bmngmProcessGSheet(ds *admin.Service, sheetID string) error {
	logger.Debugw("starting bmngmProcessGSheet()",
		"sheetID", sheetID)

	var managedDevs []mdevs.ManagedDevice

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
	err = cmn.ValidateHeader(hdrMap, mdevs.MobDevAttrMap)
	if err != nil {
		logger.Error(err)
		return err
	}

	for idx, row := range sValRange.Values {
		if idx == 0 {
			continue
		}

		mngMdevVar, err := bmngmFromFileFactory(hdrMap, row)
		if err != nil {
			logger.Error(err)
			return err
		}

		managedDevs = append(managedDevs, mngMdevVar)
	}

	err = bmngmProcessObjects(ds, managedDevs)
	if err != nil {
		logger.Error(err)
		return err
	}
	logger.Debug("finished bmngmProcessGSheet()")
	return nil
}

func bmngmProcessJSON(ds *admin.Service, filePath string, scanner *bufio.Scanner) error {
	logger.Debugw("starting bmngmProcessJSON()",
		"filePath", filePath)

	var managedDevs []mdevs.ManagedDevice

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

		mngMdevVar, err := bmngmFromJSONFactory(ds, jsonData)
		if err != nil {
			logger.Error(err)
			return err
		}

		managedDevs = append(managedDevs, mngMdevVar)
	}
	err := scanner.Err()
	if err != nil {
		logger.Error(err)
		return err
	}

	err = bmngmProcessObjects(ds, managedDevs)
	if err != nil {
		logger.Error(err)
		return err
	}
	logger.Debug("finished bmngmProcessJSON()")
	return nil
}

func bmngmProcessObjects(ds *admin.Service, managedDevs []mdevs.ManagedDevice) error {
	logger.Debug("starting bmngmProcessObjects()")

	customerID, err := cfg.ReadConfigString("customerid")
	if err != nil {
		logger.Error(err)
		return err
	}

	wg := new(sync.WaitGroup)

	for _, md := range managedDevs {
		devAction := admin.MobileDeviceAction{}

		devAction.Action = md.Action

		mdac := ds.Mobiledevices.Action(customerID, md.ResourceId, &devAction)

		wg.Add(1)

		go bmngmPerformAction(md.ResourceId, md.Action, wg, mdac)
	}

	wg.Wait()

	logger.Debug("finished bmngmProcessObjects()")
	return nil
}

func init() {
	batchManageCmd.AddCommand(batchMngMobDevCmd)

	batchMngMobDevCmd.Flags().StringVarP(&inputFile, "input-file", "i", "", "filepath to device data file")
	batchMngMobDevCmd.Flags().StringVarP(&format, "format", "f", "json", "user data file format")
	batchMngMobDevCmd.Flags().StringVarP(&sheetRange, "sheet-range", "s", "", "user data gsheet range")
}
