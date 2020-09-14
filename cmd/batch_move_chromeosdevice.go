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

var batchMoveCrOSDevCmd = &cobra.Command{
	Use:     "chromeosdevices -i <input file>",
	Aliases: []string{"chromeosdevice", "crosdevices", "crosdevice", "crosdevs", "crosdev", "cdevs", "cdev"},
	Short:   "Moves a batch of ChromeOS devices to another orgunit",
	Long: `Moves a batch of ChromeOS devices to another orgunit where device details are provided in a Google Sheet, CSV/JSON input file or piped JSON.
	
	Examples:	gmin batch-move chromeosdevices -i inputfile.txt
			gmin bmv cdevs -i inputfile.csv -f csv
			gmin bmv cdev -i 1odyAIp3jGspd3M4xeepxWD6aeQIUuHBgrZB2OHSu8MI -s 'Sheet1!A1:C25' -f gsheet
			
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
		err := btchMoveCrOSDevProcessCSV(ds, inputFile)
		if err != nil {
			return err
		}
	case lwrFmt == "json":
		err := btchMoveCrOSDevProcessJSON(ds, inputFile, scanner)
		if err != nil {
			return err
		}
	case lwrFmt == "gsheet":
		err := btchMoveCrOSDevProcessSheet(ds, inputFile)
		if err != nil {
			return err
		}
	}

	return nil
}

func btchMoveJSONCrOSDev(ds *admin.Service, jsonData string) (cdevs.MovedDevice, error) {
	movedDev := cdevs.MovedDevice{}
	jsonBytes := []byte(jsonData)

	if !json.Valid(jsonBytes) {
		return movedDev, errors.New("gmin: error - attribute string is not valid JSON")
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
		return movedDev, err
	}

	return movedDev, nil
}

func btchMoveCrOSDevs(ds *admin.Service, movedDevs []cdevs.MovedDevice) error {
	customerID, err := cfg.ReadConfigString("customerid")
	if err != nil {
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

		go btchMoveCrOSDevProcess(md.DeviceId, md.OrgUnitPath, wg, cdmc)
	}

	wg.Wait()

	return nil
}

func btchMoveCrOSDevProcess(deviceID string, ouPath string, wg *sync.WaitGroup, cdmc *admin.ChromeosdevicesMoveDevicesToOuCall) {
	defer wg.Done()

	b := backoff.NewExponentialBackOff()
	b.MaxElapsedTime = 32 * time.Second

	err := backoff.Retry(func() error {
		var err error
		err = cdmc.Do()
		if err == nil {
			fmt.Println(cmn.GminMessage("**** gmin: ChromeOS device " + deviceID + " moved to " + ouPath + " ****"))
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

func btchMoveCrOSDevProcessCSV(ds *admin.Service, filePath string) error {
	var (
		iSlice    []interface{}
		hdrMap    = map[int]string{}
		movedDevs []cdevs.MovedDevice
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

		moveCdevVar, err := btchMoveProcessCrOSDev(hdrMap, iSlice)
		if err != nil {
			return err
		}

		movedDevs = append(movedDevs, moveCdevVar)

		count = count + 1
	}

	err = btchMoveCrOSDevs(ds, movedDevs)
	if err != nil {
		return err
	}
	return nil
}

func btchMoveCrOSDevProcessJSON(ds *admin.Service, filePath string, scanner *bufio.Scanner) error {
	var movedDevs []cdevs.MovedDevice

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

		moveCdevVar, err := btchMoveJSONCrOSDev(ds, jsonData)
		if err != nil {
			return err
		}

		movedDevs = append(movedDevs, moveCdevVar)
	}
	err := scanner.Err()
	if err != nil {
		return err
	}

	err = btchMoveCrOSDevs(ds, movedDevs)
	if err != nil {
		return err
	}

	return nil
}

func btchMoveCrOSDevProcessSheet(ds *admin.Service, sheetID string) error {
	var movedDevs []cdevs.MovedDevice

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

		moveCdevVar, err := btchMoveProcessCrOSDev(hdrMap, row)
		if err != nil {
			return err
		}

		movedDevs = append(movedDevs, moveCdevVar)
	}

	err = btchMoveCrOSDevs(ds, movedDevs)
	if err != nil {
		return err
	}

	return nil
}

func btchMoveProcessCrOSDev(hdrMap map[int]string, cdevData []interface{}) (cdevs.MovedDevice, error) {
	movedDev := cdevs.MovedDevice{}

	for idx, attr := range cdevData {
		attrName := hdrMap[idx]

		switch {
		case attrName == "deviceId":
			movedDev.DeviceId = fmt.Sprintf("%v", attr)
		case attrName == "orgUnitPath":
			movedDev.OrgUnitPath = fmt.Sprintf("%v", attr)
		}
	}

	return movedDev, nil
}

func init() {
	batchMoveCmd.AddCommand(batchMoveCrOSDevCmd)

	batchMoveCrOSDevCmd.Flags().StringVarP(&inputFile, "inputfile", "i", "", "filepath to device data file")
	batchMoveCrOSDevCmd.Flags().StringVarP(&format, "format", "f", "json", "user data file format")
	batchMoveCrOSDevCmd.Flags().StringVarP(&sheetRange, "sheetrange", "s", "", "user data gsheet range")

	batchMoveCrOSDevCmd.MarkFlagRequired("inputfile")
}
