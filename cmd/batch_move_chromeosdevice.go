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

var batchMoveCrOSDevCmd = &cobra.Command{
	Use:     "chromeosdevices <orgunitpath> -i <input file>",
	Aliases: []string{"chromeosdevice", "crosdevices", "crosdevice", "crosdevs", "crosdev", "cdevs", "cdev"},
	Args:    cobra.ExactArgs(1),
	Short:   "Moves a batch of ChromeOS devices to another orgunit",
	Long: `Moves a batch of ChromeOS devices to another orgunit where device details are provided in a text input file.
	
	Examples:	gmin batch-move chromeosdevices /Sales -i inputfile.txt
			gmin bmove cdev /IT -i inputfile.txt
			
	The input file should contain a list of device ids like this:
	
	5ac7be73-5996-394e-9c30-62d41a8f10e8
	6ac9bd33-7095-453e-6c39-22d48a8f13e8
	6bc4be13-9916-494e-9c39-62d45c8f40e9`,
	RunE: doBatchMoveCrOSDev,
}

func doBatchMoveCrOSDev(cmd *cobra.Command, args []string) error {
	var deviceIDs = []string{}

	ds, err := cmn.CreateDirectoryService(admin.AdminDirectoryDeviceChromeosScope)
	if err != nil {
		return err
	}

	customerID, err := cfg.ReadConfigString("customerid")
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

	if scanner == nil {
		file, err := os.Open(inputFile)
		if err != nil {
			return err
		}
		defer file.Close()

		scanner = bufio.NewScanner(file)
	}

	for scanner.Scan() {
		deviceIDs = []string{}
		deviceIDs = append(deviceIDs, scanner.Text())

		b := backoff.NewExponentialBackOff()
		b.MaxElapsedTime = 30 * time.Second

		err = backoff.Retry(func() error {
			var err error
			err = moveCrOSDev(ds, customerID, deviceIDs, args[0])
			if err == nil {
				return err
			}

			if strings.Contains(err.Error(), "Resource Not Found") ||
				strings.Contains(err.Error(), "Bad Request") {
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

func moveCrOSDev(ds *admin.Service, customerID string, deviceIDs []string, oupath string) error {
	var move = admin.ChromeOsMoveDevicesToOu{}

	move.DeviceIds = deviceIDs

	cdmc := ds.Chromeosdevices.MoveDevicesToOu(customerID, oupath, &move)

	err := cdmc.Do()
	if err != nil {
		return err
	}

	fmt.Println("**** gmin: ChromeOS device " + deviceIDs[0] + " moved to " + oupath + " ****")

	return nil
}

func init() {
	batchMoveCmd.AddCommand(batchMoveCrOSDevCmd)

	batchMoveCrOSDevCmd.Flags().StringVarP(&inputFile, "inputfile", "i", "", "filepath to device data file")
	batchMoveCrOSDevCmd.MarkFlagRequired("inputfile")
}
