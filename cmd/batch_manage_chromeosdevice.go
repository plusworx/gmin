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
	cdevs "github.com/plusworx/gmin/utils/chromeosdevices"
	cmn "github.com/plusworx/gmin/utils/common"
	cfg "github.com/plusworx/gmin/utils/config"
	"github.com/spf13/cobra"
	admin "google.golang.org/api/admin/directory/v1"
)

var batchMngCrOSDevCmd = &cobra.Command{
	Use:     "chromeosdevices <action> -i <input file>",
	Aliases: []string{"chromeosdevice", "crosdevices", "crosdevice", "cdevs", "cdev"},
	Args:    cobra.ExactArgs(1),
	Short:   "Manages a batch of ChromeOS devices",
	Long: `Manages a batch of ChromeOS devices.
	
	Examples: gmin batch-manage chromeosdevices disable -i inputfile.txt
	          gmin bmng cdev deprovision -i inputfile.txt -r retiring_device`,
	RunE: doBatchMngCrOSDev,
}

func doBatchMngCrOSDev(cmd *cobra.Command, args []string) error {
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
			err = manageCrOSDev(ds, customerID, crosDevID, args[0])
			if err == nil {
				return err
			}

			if strings.Contains(err.Error(), "Missing required field") {
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

func manageCrOSDev(ds *admin.Service, customerID string, deviceID string, action string) error {
	var devAction = admin.ChromeOsDeviceAction{}

	lwrAction := strings.ToLower(action)
	ok := cmn.SliceContainsStr(cdevs.ValidActions, lwrAction)
	if !ok {
		return fmt.Errorf("gmin: error - %v is not a valid action type", action)
	}

	devAction.Action = lwrAction
	if lwrAction == "deprovision" {
		if reason == "" {
			return errors.New("gmin: error - must provide a reason for deprovision")
		}
		devAction.DeprovisionReason = reason
	}

	cdac := ds.Chromeosdevices.Action(customerID, deviceID, &devAction)

	err := cdac.Do()
	if err != nil {
		return err
	}

	fmt.Println("**** gmin : " + action + " successfully performed on ChromeOS device " + deviceID + " ****")

	return nil
}

func init() {
	batchManageCmd.AddCommand(batchMngCrOSDevCmd)

	batchMngCrOSDevCmd.Flags().StringVarP(&inputFile, "inputfile", "i", "", "filepath to device data file")
	batchMngCrOSDevCmd.Flags().StringVarP(&reason, "reason", "r", "", "device deprovision reason")
}
