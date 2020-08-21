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
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/cenkalti/backoff/v4"
	cdevs "github.com/plusworx/gmin/utils/chromeosdevices"
	cmn "github.com/plusworx/gmin/utils/common"
	cfg "github.com/plusworx/gmin/utils/config"
	"github.com/spf13/cobra"
	admin "google.golang.org/api/admin/directory/v1"
)

var batchUpdCrOSDevCmd = &cobra.Command{
	Use:     "chromeosdevices -i <input file>",
	Aliases: []string{"chromeosdevice", "crosdevices", "crosdevice", "crosdevs", "crosdev", "cdevs", "cdev"},
	Short:   "Updates a batch of ChromeOS devices",
	Long: `Updates a batch of ChromeOS devices with device details provided by a JSON input file.
	
	Examples:	gmin batch-update chromeosdevices -i inputfile.json
			gmin bupd cdev -i inputfile.json
			
	The JSON file should contain device update details like this:
	
	{"deviceId":"5ac7be43-5906-394e-7c39-62d45a8f10e8","annotatedAssetId":"CB1","annotatedLocation":"Batcave","annotatedUser":"Bruce Wayne","notes":"Test machine","orgUnitPath":"/Anticrime"}
	{"deviceId":"4ac7be43-5906-394e-7c39-62d45a8f10e8","annotatedAssetId":"CB2","annotatedLocation":"Wayne Manor","annotatedUser":"Alfred Pennyworth","notes":"Another test machine","orgUnitPath":"/Anticorruption"}
	{"deviceId":"3ac7be43-5906-394e-7c39-62d45a8f10e8","annotatedAssetId":"CB3","annotatedLocation":"Wayne Towers","annotatedUser":"The Big Enchilada","notes":"Yet another test machine","orgUnitPath":"/Legal"}`,
	RunE: doBatchUpdCrOSDev,
}

func doBatchUpdCrOSDev(cmd *cobra.Command, args []string) error {
	ds, err := cmn.CreateDirectoryService(admin.AdminDirectoryDeviceChromeosScope)
	if err != nil {
		return err
	}

	customerID, err := cfg.ReadConfigString("customerid")
	if err != nil {
		return err
	}

	if inputFile == "" {
		err := errors.New("gmin: error - must provide inputfile")
		return err
	}

	file, err := os.Open(inputFile)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		jsonData := scanner.Text()

		b := backoff.NewExponentialBackOff()
		b.MaxElapsedTime = 30 * time.Second

		err = backoff.Retry(func() error {
			var err error
			err = updateCrOSDev(ds, customerID, jsonData)
			if err == nil {
				return err
			}

			if strings.Contains(err.Error(), "Missing required field") ||
				strings.Contains(err.Error(), "not valid") ||
				strings.Contains(err.Error(), "unrecognized") ||
				strings.Contains(err.Error(), "should be") ||
				strings.Contains(err.Error(), "Resource Not Found") ||
				strings.Contains(err.Error(), "must be included") ||
				strings.Contains(err.Error(), "INVALID_OU_ID") {
				return backoff.Permanent(err)
			}

			return err
		}, b)
		if err != nil {
			return err
		}
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	return nil
}

func updateCrOSDev(ds *admin.Service, customerID string, jsonData string) error {
	var crosdev = admin.ChromeOsDevice{}

	jsonBytes := []byte(jsonData)

	if !json.Valid(jsonBytes) {
		return errors.New("gmin: error - attribute string is not valid JSON")
	}

	outStr, err := cmn.ParseInputAttrs(jsonBytes)
	if err != nil {
		return err
	}

	err = cmn.ValidateInputAttrs(outStr, cdevs.CrOSDevAttrMap)
	if err != nil {
		return err
	}

	err = json.Unmarshal(jsonBytes, &crosdev)
	if err != nil {
		return err
	}

	if crosdev.DeviceId == "" {
		return errors.New("gmin: error - deviceId must be included in the JSON input string")
	}

	cduc := ds.Chromeosdevices.Update(customerID, crosdev.DeviceId, &crosdev)

	_, err = cduc.Do()
	if err != nil {
		return err
	}

	fmt.Println("**** gmin: ChromeOS device " + crosdev.DeviceId + " updated ****")

	return nil
}

func init() {
	batchUpdateCmd.AddCommand(batchUpdCrOSDevCmd)

	batchUpdCrOSDevCmd.Flags().StringVarP(&inputFile, "inputfile", "i", "", "filepath to device data file")
}
