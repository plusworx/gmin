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

	cdevs "github.com/plusworx/gmin/utils/chromeosdevices"
	cmn "github.com/plusworx/gmin/utils/common"
	cfg "github.com/plusworx/gmin/utils/config"
	flgnm "github.com/plusworx/gmin/utils/flagnames"
	gmess "github.com/plusworx/gmin/utils/gminmessages"
	lg "github.com/plusworx/gmin/utils/logging"
	"github.com/spf13/cobra"
	admin "google.golang.org/api/admin/directory/v1"
)

var updateCrOSDevCmd = &cobra.Command{
	Use:     "chromeos-device <device id>",
	Aliases: []string{"cros-device", "cros-dev", "cdev"},
	Args:    cobra.ExactArgs(1),
	Example: `gmin update chromeos-device 4cx07eba348f09b3 --location "Head Office"
gmin upd cdev 4cx07eba348f09b3 -u "Mark Zuckerberg"`,
	Short: "Updates a ChromeOS device",
	Long:  `Updates a ChromeOS device.`,
	RunE:  doUpdateCrOSDev,
}

func doUpdateCrOSDev(cmd *cobra.Command, args []string) error {
	lg.Debugw("starting doUpdateCrOSDev()",
		"args", args)
	defer lg.Debug("finished doUpdateCrOSDev()")

	var crosdev = admin.ChromeOsDevice{}

	srv, err := cmn.CreateService(cmn.SRVTYPEADMIN, admin.AdminDirectoryDeviceChromeosScope)
	if err != nil {
		return err
	}
	ds := srv.(*admin.Service)

	customerID, err := cfg.ReadConfigString(cfg.CONFIGCUSTID)
	if err != nil {
		return err
	}

	flgAssetIDVal, err := cmd.Flags().GetString(flgnm.FLG_ASSETID)
	if err != nil {
		lg.Error(err)
		return err
	}
	if flgAssetIDVal != "" {
		crosdev.AnnotatedAssetId = flgAssetIDVal
	}

	flgLocationVal, err := cmd.Flags().GetString(flgnm.FLG_LOCATION)
	if err != nil {
		lg.Error(err)
		return err
	}
	if flgLocationVal != "" {
		crosdev.AnnotatedLocation = flgLocationVal
	}

	flgNotesVal, err := cmd.Flags().GetString(flgnm.FLG_NOTES)
	if err != nil {
		lg.Error(err)
		return err
	}
	if flgNotesVal != "" {
		crosdev.Notes = flgNotesVal
	}

	flgOUPathVal, err := cmd.Flags().GetString(flgnm.FLG_ORGUNITPATH)
	if err != nil {
		lg.Error(err)
		return err
	}
	if flgOUPathVal != "" {
		crosdev.OrgUnitPath = flgOUPathVal
	}

	flgUserKeyVal, err := cmd.Flags().GetString(flgnm.FLG_USERKEY)
	if err != nil {
		lg.Error(err)
		return err
	}
	if flgUserKeyVal != "" {
		crosdev.AnnotatedUser = flgUserKeyVal
	}

	cduc := ds.Chromeosdevices.Update(customerID, args[0], &crosdev)

	flgProjectionVal, err := cmd.Flags().GetString(flgnm.FLG_PROJECTION)
	if err != nil {
		lg.Error(err)
		return err
	}
	if flgProjectionVal != "" {
		proj := strings.ToLower(flgProjectionVal)
		ok := cmn.SliceContainsStr(cdevs.ValidProjections, proj)
		if !ok {
			err = fmt.Errorf(gmess.ERR_INVALIDPROJECTIONTYPE, flgProjectionVal)
			lg.Error(err)
			return err
		}

		updCall := cdevs.AddProjection(cduc, proj)
		cduc = updCall.(*admin.ChromeosdevicesUpdateCall)
	}

	updCrOSDev, err := cduc.Do()
	if err != nil {
		lg.Error(err)
		return err
	}

	fmt.Println(cmn.GminMessage(fmt.Sprintf(gmess.INFO_CDEVUPDATED, updCrOSDev.DeviceId)))
	lg.Infof(gmess.INFO_CDEVUPDATED, updCrOSDev.DeviceId)

	return nil
}

func init() {
	updateCmd.AddCommand(updateCrOSDevCmd)

	updateCrOSDevCmd.Flags().StringP(flgnm.FLG_ASSETID, "d", "", "device asset id")
	updateCrOSDevCmd.Flags().StringP(flgnm.FLG_PROJECTION, "j", "", "type of projection")
	updateCrOSDevCmd.Flags().StringP(flgnm.FLG_LOCATION, "l", "", "device location")
	updateCrOSDevCmd.Flags().StringP(flgnm.FLG_NOTES, "n", "", "notes about device")
	updateCrOSDevCmd.Flags().StringP(flgnm.FLG_ORGUNITPATH, "t", "", "orgunit device belongs to")
	updateCrOSDevCmd.Flags().StringP(flgnm.FLG_USERKEY, "u", "", "device user")
}
