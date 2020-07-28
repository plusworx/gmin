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

	cmn "github.com/plusworx/gmin/utils/common"
	cfg "github.com/plusworx/gmin/utils/config"
	"github.com/spf13/cobra"
	admin "google.golang.org/api/admin/directory/v1"
)

var batchMoveCrOSDevCmd = &cobra.Command{
	Use:     "chromeosdevices <orgunitpath> -i <input file>",
	Aliases: []string{"chromeosdevice", "crosdevices", "crosdevice", "cdevs", "cdev"},
	Args:    cobra.ExactArgs(1),
	Short:   "Moves a batch of ChromeOS devices to another orgunit",
	Long: `Moves a batch of ChromeOS devices to another orgunit.
	
	Examples: gmin batch-move chromeosdevices /Sales -i inputfile.txt
	          gmin bmng cdev /IT -i inputfile.txt`,
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
		deviceIDs = append(deviceIDs, crosDevID)
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	err = moveCrOSDev(ds, customerID, deviceIDs, args[0])
	if err == nil {
		return err
	}

	return nil
}

func moveCrOSDev(ds *admin.Service, customerID string, deviceIDs []string, oupath string) error {
	var move = admin.ChromeOsMoveDevicesToOu{}

	customerID, err := cfg.ReadConfigString("customerid")
	if err != nil {
		return err
	}

	move.DeviceIds = deviceIDs

	cdmc := ds.Chromeosdevices.MoveDevicesToOu(customerID, oupath, &move)

	err = cdmc.Do()
	if err != nil {
		return err
	}

	fmt.Println("**** gmin: ChromeOS devices moved to " + oupath + " ****")

	return nil
}

func init() {
	batchMoveCmd.AddCommand(batchMoveCrOSDevCmd)

	batchMoveCrOSDevCmd.Flags().StringVarP(&inputFile, "inputfile", "i", "", "filepath to device data file")
}
