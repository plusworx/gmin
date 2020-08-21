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
	"strings"

	cmn "github.com/plusworx/gmin/utils/common"
	cfg "github.com/plusworx/gmin/utils/config"
	mdevs "github.com/plusworx/gmin/utils/mobiledevices"
	"github.com/spf13/cobra"

	admin "google.golang.org/api/admin/directory/v1"
)

var manageMobDevCmd = &cobra.Command{
	Use:     "mobiledevice <resource id> <action>",
	Aliases: []string{"mobdevice", "mobdev", "mdev"},
	Args:    cobra.ExactArgs(2),
	Short:   "Performs an action on a mobile device",
	Long: `Performs an action on a mobile device.
	
	Examples:	gmin manage mobiledevice 4cx07eba348f09b3 block
			gmin mng mdev 4cx07eba348f09b3 admin_remote_wipe`,
	RunE: doManageMobDev,
}

func doManageMobDev(cmd *cobra.Command, args []string) error {
	var devAction = admin.MobileDeviceAction{}

	customerID, err := cfg.ReadConfigString("customerid")
	if err != nil {
		return err
	}

	action := strings.ToLower(args[1])
	ok := cmn.SliceContainsStr(mdevs.ValidActions, action)
	if !ok {
		return fmt.Errorf("gmin: error - %v is not a valid action type", args[1])
	}

	devAction.Action = action

	ds, err := cmn.CreateDirectoryService(admin.AdminDirectoryDeviceMobileActionScope)
	if err != nil {
		return err
	}

	mdac := ds.Mobiledevices.Action(customerID, args[0], &devAction)

	if attrs != "" {
		manageAttrs, err := cmn.ParseOutputAttrs(attrs, mdevs.MobDevAttrMap)
		if err != nil {
			return err
		}
		formattedAttrs := mdevs.StartMobDevicesField + manageAttrs + mdevs.EndField
		actionCall := mdevs.AddFields(mdac, formattedAttrs)
		mdac = actionCall.(*admin.MobiledevicesActionCall)
	}

	err = mdac.Do()
	if err != nil {
		return err
	}

	fmt.Println("**** gmin : " + args[1] + " successfully performed on mobile device " + args[0] + " ****")

	return nil
}

func init() {
	manageCmd.AddCommand(manageMobDevCmd)

	manageMobDevCmd.Flags().StringVarP(&attrs, "attributes", "a", "", "device's attributes to display (separated by ~)")
}
