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

	cmn "github.com/plusworx/gmin/utils/common"
	cfg "github.com/plusworx/gmin/utils/config"
	"github.com/spf13/cobra"

	admin "google.golang.org/api/admin/directory/v1"
)

var moveCrOSDevCmd = &cobra.Command{
	Use:     "chromeos-device <device id> <orgunitpath>",
	Aliases: []string{"cros-device", "cros-dev", "cdev"},
	Args:    cobra.ExactArgs(2),
	Example: `gmin move chromeos-device 4cx07eba348f09b3 /Sales
gmin mv cdev 4cx07eba348f09b3 /IT`,
	Short: "Moves a ChromeOS device to another orgunit",
	Long:  `Moves a ChromeOS device to another orgunit.`,
	RunE:  doMoveCrOSDev,
}

func doMoveCrOSDev(cmd *cobra.Command, args []string) error {
	logger.Debugw("starting doMoveCrOSDev()",
		"args", args)

	var move = admin.ChromeOsMoveDevicesToOu{}

	customerID, err := cfg.ReadConfigString("customerid")
	if err != nil {
		logger.Error(err)
		return err
	}

	move.DeviceIds = append(move.DeviceIds, args[0])

	ds, err := cmn.CreateDirectoryService(admin.AdminDirectoryDeviceChromeosScope)
	if err != nil {
		logger.Error(err)
		return err
	}

	cdmc := ds.Chromeosdevices.MoveDevicesToOu(customerID, args[1], &move)

	err = cdmc.Do()
	if err != nil {
		logger.Error(err)
		return err
	}

	logger.Infof(cmn.InfoCDevMovePerformed, args[0], args[1])
	fmt.Println(cmn.GminMessage(fmt.Sprintf(cmn.InfoCDevMovePerformed, args[0], args[1])))

	logger.Debug("finished doMoveCrOSDev()")
	return nil
}

func init() {
	moveCmd.AddCommand(moveCrOSDevCmd)
}
