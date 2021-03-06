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
	gmess "github.com/plusworx/gmin/utils/gminmessages"
	lg "github.com/plusworx/gmin/utils/logging"
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
	lg.Debugw("starting doMoveCrOSDev()",
		"args", args)
	defer lg.Debug("finished doMoveCrOSDev()")

	var move = admin.ChromeOsMoveDevicesToOu{}

	customerID, err := cfg.ReadConfigString(cfg.CONFIGCUSTID)
	if err != nil {
		return err
	}

	move.DeviceIds = append(move.DeviceIds, args[0])

	srv, err := cmn.CreateService(cmn.SRVTYPEADMIN, admin.AdminDirectoryDeviceChromeosScope)
	if err != nil {
		return err
	}
	ds := srv.(*admin.Service)

	cdmc := ds.Chromeosdevices.MoveDevicesToOu(customerID, args[1], &move)

	err = cdmc.Do()
	if err != nil {
		lg.Error(err)
		return err
	}

	fmt.Println(cmn.GminMessage(fmt.Sprintf(gmess.INFO_CDEVMOVEPERFORMED, args[0], args[1])))
	lg.Infof(gmess.INFO_CDEVMOVEPERFORMED, args[0], args[1])

	return nil
}

func init() {
	moveCmd.AddCommand(moveCrOSDevCmd)
}
