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
	"fmt"

	cdevs "github.com/plusworx/gmin/utils/chromeosdevices"
	cmn "github.com/plusworx/gmin/utils/common"
	cfg "github.com/plusworx/gmin/utils/config"
	"github.com/spf13/cobra"

	admin "google.golang.org/api/admin/directory/v1"
)

var moveCrOSDevCmd = &cobra.Command{
	Use:     "chromeosdevice <device id> <orgunitpath>",
	Aliases: []string{"crosdevice", "cdev"},
	Args:    cobra.ExactArgs(2),
	Short:   "Moves a ChromeOS device to another orgunit",
	Long: `Moves a ChromeOS device to another orgunit.
	
	Example: gmin move chromeosdevice 4cx07eba348f09b3 /Sales`,
	RunE: doMoveCrOSDev,
}

func doMoveCrOSDev(cmd *cobra.Command, args []string) error {
	var move = admin.ChromeOsMoveDevicesToOu{}

	customerID, err := cfg.ReadConfigString("customerid")
	if err != nil {
		return err
	}

	move.DeviceIds = append(move.DeviceIds, args[0])

	ds, err := cmn.CreateDirectoryService(admin.AdminDirectoryDeviceChromeosScope)
	if err != nil {
		return err
	}

	cdmc := ds.Chromeosdevices.MoveDevicesToOu(customerID, args[1], &move)

	if attrs != "" {
		moveAttrs, err := cmn.ParseOutputAttrs(attrs, cdevs.CrOSDevAttrMap)
		if err != nil {
			return err
		}
		formattedAttrs := cdevs.StartChromeDevicesField + moveAttrs + cdevs.EndField
		moveCall := cdevs.AddFields(cdmc, formattedAttrs)
		cdmc = moveCall.(*admin.ChromeosdevicesMoveDevicesToOuCall)
	}

	err = cdmc.Do()
	if err != nil {
		return err
	}

	fmt.Println("**** gmin: ChromeOS device " + args[0] + " moved to " + args[1] + "****")

	return nil
}

func init() {
	moveCmd.AddCommand(moveCrOSDevCmd)

	moveCrOSDevCmd.Flags().StringVarP(&attrs, "attributes", "a", "", "device's attributes to display (separated by ~)")
}
