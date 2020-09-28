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
	Example: `gmin manage chromeosdevice 4cx07eba348f09b3 disable
gmin mng cdev 4cx07eba348f09b3 reenable`,
	Short: "Performs an action on a ChromeOS device",
	Long:  `Performs an action on a ChromeOS device.`,
	RunE:  doManageCrOSDev,
}

func doManageCrOSDev(cmd *cobra.Command, args []string) error {
	logger.Debugw("starting doManageCrOSDev()",
		"args", args)

	var devAction = admin.ChromeOsDeviceAction{}

	customerID, err := cfg.ReadConfigString("customerid")
	if err != nil {
		logger.Error(err)
		return err
	}

	action := strings.ToLower(args[1])
	ok := cmn.SliceContainsStr(cdevs.ValidActions, action)
	if !ok {
		err = fmt.Errorf(cmn.ErrInvalidActionType, args[1])
		logger.Error(err)
		return err
	}

	devAction.Action = action
	if action == "deprovision" {
		if reason == "" {
			err = errors.New(cmn.ErrNoDeprovisionReason)
			logger.Error(err)
			return err
		}
		devAction.DeprovisionReason = reason
	}

	ds, err := cmn.CreateDirectoryService(admin.AdminDirectoryDeviceChromeosScope)
	if err != nil {
		logger.Error(err)
		return err
	}

	cdac := ds.Chromeosdevices.Action(customerID, args[0], &devAction)

	err = cdac.Do()
	if err != nil {
		logger.Error(err)
		return err
	}

	logger.Infof(cmn.InfoCDevActionPerformed, args[1], args[0])
	fmt.Println(cmn.GminMessage(fmt.Sprintf(cmn.InfoCDevActionPerformed, args[1], args[0])))

	logger.Debug("finished doManageCrOSDev()")
	return nil
}

func init() {
	manageCmd.AddCommand(manageCrOSDevCmd)

	manageCrOSDevCmd.Flags().StringVarP(&reason, "reason", "r", "", "device deprovision reason")
}
