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

var batchUpdCrOSDevCmd = &cobra.Command{
	Use:     "chromeosdevices -i <input file>",
	Aliases: []string{"chromeosdevice", "crosdevices", "crosdevice", "crosdevs", "crosdev", "cdevs", "cdev"},
	Short:   "Updates a batch of ChromeOS devices",
	Long: `Updates a batch of ChromeOS devices with device details provided in a Google Sheet, CSV/JSON input file or piped JSON.
	
	Examples:	gmin batch-update chromeosdevices -i inputfile.json
			gmin bupd cdevs -i inputfile.csv -f csv
			gmin bupd cdev -i 1odyAIp3jGspd3M4xeepxWD6aeQIUuHBgrZB2OHSu8MI -s 'Sheet1!A1:K25' -f gsheet
			
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
		err := btchUpdCDevProcessCSV(ds, inputFile)
		if err != nil {
			logger.Error(err)
			return err
		}
	case lwrFmt == "json":
		err := btchUpdCDevProcessJSON(ds, inputFile, scanner)
		if err != nil {
			logger.Error(err)
			return err
		}
	case lwrFmt == "gsheet":
		err := btchUpdCDevProcessSheet(ds, inputFile)
		if err != nil {
			logger.Error(err)
			return err
		}
	}

	return nil
}

func btchUpdJSONCDev(ds *admin.Service, jsonData string) (*admin.ChromeOsDevice, error) {
	var (
		crosdev   *admin.ChromeOsDevice
		emptyVals = cmn.EmptyValues{}
	)

	crosdev = new(admin.ChromeOsDevice)
	jsonBytes := []byte(jsonData)

	if !json.Valid(jsonBytes) {
		logger.Error(cmn.ErrInvalidJSONAttr)
		return nil, errors.New(cmn.ErrInvalidJSONAttr)
	}

	outStr, err := cmn.ParseInputAttrs(jsonBytes)
	if err != nil {
		logger.Error(err)
		return nil, err
	}

	err = cmn.ValidateInputAttrs(outStr, cdevs.CrOSDevAttrMap)
	if err != nil {
		logger.Error(err)
		return nil, err
	}

	err = json.Unmarshal(jsonBytes, &crosdev)
	if err != nil {
		logger.Error(err)
		return nil, err
	}

	if crosdev.DeviceId == "" {
		err = errors.New(cmn.ErrNoJSONDeviceID)
		logger.Error(err)
		return nil, err
	}

	err = json.Unmarshal(jsonBytes, &emptyVals)
	if err != nil {
		logger.Error(err)
		return nil, err
	}
	if len(emptyVals.ForceSendFields) > 0 {
		crosdev.ForceSendFields = emptyVals.ForceSendFields
	}

	return crosdev, nil
}

func btchUpdateCDevs(ds *admin.Service, crosdevs []*admin.ChromeOsDevice) error {
	wg := new(sync.WaitGroup)

	customerID, err := cfg.ReadConfigString("customerid")
	if err != nil {
		logger.Error(err)
		return err
	}

	for _, c := range crosdevs {
		cduc := ds.Chromeosdevices.Update(customerID, c.DeviceId, c)

		wg.Add(1)

		go btchCDevUpdateProcess(c, wg, cduc)
	}

	wg.Wait()

	return nil
}

func btchCDevUpdateProcess(crosdev *admin.ChromeOsDevice, wg *sync.WaitGroup, cduc *admin.ChromeosdevicesUpdateCall) {
	defer wg.Done()

	b := backoff.NewExponentialBackOff()
	b.MaxElapsedTime = 32 * time.Second

	err := backoff.Retry(func() error {
		var err error
		_, err = cduc.Do()
		if err == nil {
			logger.Infof(cmn.InfoCDevUpdated, crosdev.DeviceId)
			fmt.Println(cmn.GminMessage(fmt.Sprintf(cmn.InfoCDevUpdated, crosdev.DeviceId)))
			return err
		}
		if !cmn.IsErrRetryable(err) {
			return backoff.Permanent(fmt.Errorf(cmn.ErrBatchChromeOSDevice, err.Error(), crosdev.DeviceId))
		}
		// Log the retries
		logger.Errorw(err.Error(),
			"retrying", b.Clock.Now().String(),
			"ChromeOS device", crosdev.DeviceId)
		return fmt.Errorf(cmn.ErrBatchChromeOSDevice, err.Error(), crosdev.DeviceId)
	}, b)
	if err != nil {
		// Log final error
		logger.Error(err)
		fmt.Println(cmn.GminMessage(err.Error()))
	}
}

func btchUpdCDevProcessCSV(ds *admin.Service, filePath string) error {
	var (
		iSlice   []interface{}
		hdrMap   = map[int]string{}
		crosdevs []*admin.ChromeOsDevice
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

		cdevVar, err := btchUpdProcessCDev(hdrMap, iSlice)
		if err != nil {
			logger.Error(err)
			return err
		}

		crosdevs = append(crosdevs, cdevVar)

		count = count + 1
	}

	err = btchUpdateCDevs(ds, crosdevs)
	if err != nil {
		logger.Error(err)
		return err
	}
	return nil
}

func btchUpdCDevProcessJSON(ds *admin.Service, filePath string, scanner *bufio.Scanner) error {
	var crosdevs []*admin.ChromeOsDevice

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

		cdevVar, err := btchUpdJSONCDev(ds, jsonData)
		if err != nil {
			logger.Error(err)
			return err
		}

		crosdevs = append(crosdevs, cdevVar)
	}
	err := scanner.Err()
	if err != nil {
		logger.Error(err)
		return err
	}

	err = btchUpdateCDevs(ds, crosdevs)
	if err != nil {
		logger.Error(err)
		return err
	}

	return nil
}

func btchUpdCDevProcessSheet(ds *admin.Service, sheetID string) error {
	var crosdevs []*admin.ChromeOsDevice

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

		cdevVar, err := btchUpdProcessCDev(hdrMap, row)
		if err != nil {
			logger.Error(err)
			return err
		}

		crosdevs = append(crosdevs, cdevVar)
	}

	err = btchUpdateCDevs(ds, crosdevs)
	if err != nil {
		logger.Error(err)
		return err
	}

	return nil
}

func btchUpdProcessCDev(hdrMap map[int]string, cdevData []interface{}) (*admin.ChromeOsDevice, error) {
	var crosdev *admin.ChromeOsDevice

	crosdev = new(admin.ChromeOsDevice)

	for idx, attr := range cdevData {
		attrName := hdrMap[idx]

		switch {
		case attrName == "annotatedAssetId":
			assetID := fmt.Sprintf("%v", attr)
			crosdev.AnnotatedAssetId = assetID
			if assetID == "" {
				crosdev.ForceSendFields = append(crosdev.ForceSendFields, "AnnotatedAssetId")
			}
		case attrName == "annotatedLocation":
			location := fmt.Sprintf("%v", attr)
			crosdev.AnnotatedLocation = location
			if location == "" {
				crosdev.ForceSendFields = append(crosdev.ForceSendFields, "AnnotatedLocation")
			}
		case attrName == "annotatedUser":
			user := fmt.Sprintf("%v", attr)
			crosdev.AnnotatedUser = user
			if user == "" {
				crosdev.ForceSendFields = append(crosdev.ForceSendFields, "AnnotatedUser")
			}
		case attrName == "notes":
			notes := fmt.Sprintf("%v", attr)
			crosdev.Notes = notes
			if notes == "" {
				crosdev.ForceSendFields = append(crosdev.ForceSendFields, "Notes")
			}
		case attrName == "orgUnitPath":
			crosdev.OrgUnitPath = fmt.Sprintf("%v", attr)
		}
	}

	return crosdev, nil
}

func init() {
	batchUpdateCmd.AddCommand(batchUpdCrOSDevCmd)

	batchUpdCrOSDevCmd.Flags().StringVarP(&inputFile, "inputfile", "i", "", "filepath to device data file or sheet id")
	batchUpdCrOSDevCmd.Flags().StringVarP(&format, "format", "f", "json", "user data file format")
	batchUpdCrOSDevCmd.Flags().StringVarP(&sheetRange, "sheetrange", "s", "", "user data gsheet range")
}
