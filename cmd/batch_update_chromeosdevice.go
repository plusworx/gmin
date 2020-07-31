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
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/cenkalti/backoff/v4"
	cmn "github.com/plusworx/gmin/utils/common"
	cfg "github.com/plusworx/gmin/utils/config"
	"github.com/spf13/cobra"
	admin "google.golang.org/api/admin/directory/v1"
)

var batchUpdCrOSDevCmd = &cobra.Command{
	Use:     "chromeosdevices -i <input file>",
	Aliases: []string{"chromeosdevice", "crosdevices", "crosdevice", "cdevs", "cdev"},
	Short:   "Updates a batch of ChromeOS devices",
	Long: `Updates a batch of ChromeOS devices.
	
	Examples: gmin batch-update chromeosdevices -i inputfile.txt -l London
	          gmin bupd cdev -i inputfile.txt -n "New chromebook trial"`,
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
		crosDevID := scanner.Text()

		b := backoff.NewExponentialBackOff()
		b.MaxElapsedTime = 30 * time.Second

		err = backoff.Retry(func() error {
			var err error
			err = updateCrOSDev(ds, customerID, crosDevID)
			if err == nil {
				return err
			}

			if strings.Contains(err.Error(), "Missing required field") ||
				strings.Contains(err.Error(), "Resource Not Found") ||
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

func updateCrOSDev(ds *admin.Service, customerID string, deviceID string) error {
	var crosdev = admin.ChromeOsDevice{}

	if location != "" {
		crosdev.AnnotatedLocation = location
	}

	if notes != "" {
		crosdev.Notes = notes
	}

	if orgUnit != "" {
		crosdev.OrgUnitPath = orgUnit
	}

	if userKey != "" {
		crosdev.AnnotatedUser = userKey
	}

	cduc := ds.Chromeosdevices.Update(customerID, deviceID, &crosdev)

	updCrosDev, err := cduc.Do()
	if err != nil {
		return err
	}

	fmt.Println("**** gmin: ChromeOS device " + updCrosDev.DeviceId + " updated ****")

	return nil
}

func init() {
	batchUpdateCmd.AddCommand(batchUpdCrOSDevCmd)

	batchUpdCrOSDevCmd.Flags().StringVarP(&inputFile, "inputfile", "i", "", "filepath to device data file")
	batchUpdCrOSDevCmd.Flags().StringVarP(&location, "location", "l", "", "device location")
	batchUpdCrOSDevCmd.Flags().StringVarP(&notes, "notes", "n", "", "notes about device")
	batchUpdCrOSDevCmd.Flags().StringVarP(&orgUnit, "orgunitpath", "o", "", "orgunit device belongs to")
	batchUpdCrOSDevCmd.Flags().StringVarP(&userKey, "user", "u", "", "device user")
}
