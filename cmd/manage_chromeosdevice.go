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
	"errors"
	"fmt"
	"strings"

	cdevs "github.com/plusworx/gmin/utils/chromeosdevices"
	cmn "github.com/plusworx/gmin/utils/common"
	cfg "github.com/plusworx/gmin/utils/config"
	"github.com/spf13/cobra"

	admin "google.golang.org/api/admin/directory/v1"
)

var manageCrOSDevCmd = &cobra.Command{
	Use:     "chromeosdevice <device id> <action>",
	Aliases: []string{"crosdevice", "crosdev", "cdev"},
	Args:    cobra.ExactArgs(2),
	Short:   "Performs an action on a ChromeOS device",
	Long: `Performs an action on a ChromeOS device.
	
	Examples:	gmin manage chromeosdevice 4cx07eba348f09b3 disable
			gmin mng cdev 4cx07eba348f09b3 reenable`,
	RunE: doManageCrOSDev,
}

func doManageCrOSDev(cmd *cobra.Command, args []string) error {
	var devAction = admin.ChromeOsDeviceAction{}

	customerID, err := cfg.ReadConfigString("customerid")
	if err != nil {
		return err
	}

	action := strings.ToLower(args[1])
	ok := cmn.SliceContainsStr(cdevs.ValidActions, action)
	if !ok {
		return fmt.Errorf("gmin: error - %v is not a valid action type", args[1])
	}

	devAction.Action = action
	if action == "deprovision" {
		if reason == "" {
			return errors.New("gmin: error - must provide a reason for deprovision")
		}
		devAction.DeprovisionReason = reason
	}

	ds, err := cmn.CreateDirectoryService(admin.AdminDirectoryDeviceChromeosScope)
	if err != nil {
		return err
	}

	cdac := ds.Chromeosdevices.Action(customerID, args[0], &devAction)

	if attrs != "" {
		manageAttrs, err := cmn.ParseOutputAttrs(attrs, cdevs.CrOSDevAttrMap)
		if err != nil {
			return err
		}
		formattedAttrs := cdevs.StartChromeDevicesField + manageAttrs + cdevs.EndField
		actionCall := cdevs.AddFields(cdac, formattedAttrs)
		cdac = actionCall.(*admin.ChromeosdevicesActionCall)
	}

	err = cdac.Do()
	if err != nil {
		return err
	}

	fmt.Println("**** gmin : " + args[1] + " successfully performed on ChromeOS device " + args[0] + " ****")

	return nil
}

func init() {
	manageCmd.AddCommand(manageCrOSDevCmd)

	manageCrOSDevCmd.Flags().StringVarP(&attrs, "attributes", "a", "", "device's attributes to display (separated by ~)")
	manageCrOSDevCmd.Flags().StringVarP(&reason, "reason", "r", "", "device deprovision reason")
}
