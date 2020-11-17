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
	"github.com/spf13/pflag"
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

	var flagsPassed []string

	crosdev := new(admin.ChromeOsDevice)
	flagValueMap := map[string]interface{}{}

	// Collect names of command flags passed in
	cmd.Flags().Visit(func(f *pflag.Flag) {
		if f.Name != flgnm.FLG_SILENT {
			flagsPassed = append(flagsPassed, f.Name)
		}
	})

	// Populate flag value map
	for _, flg := range flagsPassed {
		val, err := cdevs.GetFlagVal(cmd, flg)
		if err != nil {
			return err
		}
		flagValueMap[flg] = val
	}

	// Process command flags
	err := processUpdCrOSDevFlags(cmd, crosdev, flagValueMap)
	if err != nil {
		lg.Error(err)
		return err
	}

	srv, err := cmn.CreateService(cmn.SRVTYPEADMIN, admin.AdminDirectoryDeviceChromeosScope)
	if err != nil {
		return err
	}
	ds := srv.(*admin.Service)

	customerID, err := cfg.ReadConfigString(cfg.CONFIGCUSTID)
	if err != nil {
		return err
	}

	cduc := ds.Chromeosdevices.Update(customerID, args[0], crosdev)

	projFlgVal, projFlgExists := flagValueMap[flgnm.FLG_PROJECTION]
	if projFlgExists {
		ucdProjectionFlag(cduc, "--"+flgnm.FLG_PROJECTION, fmt.Sprintf("%v", projFlgVal))
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
	updateCrOSDevCmd.Flags().StringP(flgnm.FLG_USER, "u", "", "device user")
}

func processUpdCrOSDevFlags(cmd *cobra.Command, crosdev *admin.ChromeOsDevice, flagValueMap map[string]interface{}) error {
	lg.Debug("starting processUpdCrOSDevFlags()")
	defer lg.Debug("finished processUpdCrOSDevFlags()")

	cdevUpdOneStrFuncMap := map[string]func(*admin.ChromeOsDevice, string){
		flgnm.FLG_LOCATION: ucdLocationFlag,
		flgnm.FLG_NOTES:    ucdNotesFlag,
		flgnm.FLG_USER:     ucdUserFlag,
	}

	cdevUpdTwoStrFuncMap := map[string]func(*admin.ChromeOsDevice, string, string) error{
		flgnm.FLG_ASSETID:     ucdAssetIDFlag,
		flgnm.FLG_ORGUNITPATH: ucdOrgUnitPathFlag,
	}

	for flName, flgVal := range flagValueMap {
		strOneFunc, sf1Exists := cdevUpdOneStrFuncMap[flName]
		if sf1Exists {
			strOneFunc(crosdev, fmt.Sprintf("%v", flgVal))
			continue
		}

		strTwoFunc, sf2Exists := cdevUpdTwoStrFuncMap[flName]
		if sf2Exists {
			err := strTwoFunc(crosdev, "--"+flName, fmt.Sprintf("%v", flgVal))
			if err != nil {
				return err
			}
			continue
		}
	}
	return nil
}

func ucdAssetIDFlag(crosdev *admin.ChromeOsDevice, flgName string, flgVal string) error {
	lg.Debugw("starting ucdAssetIDFlag()",
		"flgVal", flgVal)
	defer lg.Debug("finished ucdAssetIDFlag()")

	if flgVal == "" {
		err := fmt.Errorf(gmess.ERR_EMPTYSTRING, flgName)
		lg.Error(err)
		return err
	}
	crosdev.AnnotatedAssetId = flgVal

	return nil
}

func ucdLocationFlag(crosdev *admin.ChromeOsDevice, flgVal string) {
	lg.Debugw("starting ucdLocationFlag()",
		"flgVal", flgVal)
	defer lg.Debug("finished ucdLocationFlag()")

	crosdev.AnnotatedLocation = flgVal
}

func ucdNotesFlag(crosdev *admin.ChromeOsDevice, flgVal string) {
	lg.Debugw("starting ucdNotesFlag()",
		"flgVal", flgVal)
	defer lg.Debug("finished ucdNotesFlag()")

	crosdev.Notes = flgVal
}

func ucdOrgUnitPathFlag(crosdev *admin.ChromeOsDevice, flgName string, flgVal string) error {
	lg.Debugw("starting ucdOrgUnitPathFlag()",
		"flgVal", flgVal)
	defer lg.Debug("finished ucdOrgUnitPathFlag()")

	if flgVal == "" {
		err := fmt.Errorf(gmess.ERR_EMPTYSTRING, flgName)
		lg.Error(err)
		return err
	}
	crosdev.OrgUnitPath = flgVal

	return nil
}

func ucdProjectionFlag(cduc *admin.ChromeosdevicesUpdateCall, flgName string, flgVal string) error {
	lg.Debugw("starting ucdProjectionFlag()",
		"flgVal", flgVal)
	defer lg.Debug("finished ucdProjectionFlag()")

	if flgVal == "" {
		err := fmt.Errorf(gmess.ERR_EMPTYSTRING, flgName)
		lg.Error(err)
		return err
	}

	proj := strings.ToLower(flgVal)
	ok := cmn.SliceContainsStr(cdevs.ValidProjections, proj)
	if !ok {
		err := fmt.Errorf(gmess.ERR_INVALIDPROJECTIONTYPE, flgVal)
		lg.Error(err)
		return err
	}

	updCall := cdevs.AddProjection(cduc, proj)
	cduc = updCall.(*admin.ChromeosdevicesUpdateCall)

	return nil
}

func ucdUserFlag(crosdev *admin.ChromeOsDevice, flgVal string) {
	lg.Debugw("starting ucdUserFlag()",
		"flgVal", flgVal)
	defer lg.Debug("finished ucdUserFlag()")

	crosdev.AnnotatedUser = flgVal
}
