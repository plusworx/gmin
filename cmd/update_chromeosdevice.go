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
	"encoding/json"
	"fmt"
	"strings"

	cdevs "github.com/plusworx/gmin/utils/chromeosdevices"
	cmn "github.com/plusworx/gmin/utils/common"
	cfg "github.com/plusworx/gmin/utils/config"
	"github.com/spf13/cobra"
	admin "google.golang.org/api/admin/directory/v1"
)

var updateCrOSDevCmd = &cobra.Command{
	Use:     "chromeosdevice <device id>",
	Aliases: []string{"crosdevice", "cdev"},
	Args:    cobra.ExactArgs(1),
	Short:   "Updates a ChromeOS device",
	Long: `Updates a ChromeOS device.
	
	Examples:	gmin update chromeosdevice 4cx07eba348f09b3 --location "Head Office"
			gmin upd cdev 4cx07eba348f09b3 -u "Mark Zuckerberg"`,
	RunE: doUpdateCrOSDev,
}

func doUpdateCrOSDev(cmd *cobra.Command, args []string) error {
	var crosdev = admin.ChromeOsDevice{}

	ds, err := cmn.CreateDirectoryService(admin.AdminDirectoryDeviceChromeosScope)
	if err != nil {
		logger.Error(err)
		return err
	}

	customerID, err := cfg.ReadConfigString("customerid")
	if err != nil {
		logger.Error(err)
		return err
	}

	if assetID != "" {
		crosdev.AnnotatedAssetId = assetID
	}

	if location != "" {
		crosdev.AnnotatedLocation = location
	}

	if notes != "" {
		crosdev.Notes = notes
	}

	if orgUnit != "" {
		crosdev.OrgUnitPath = orgUnit
	}

	if userKey != "" {
		crosdev.AnnotatedUser = userKey
	}

	cduc := ds.Chromeosdevices.Update(customerID, args[0], &crosdev)

	if attrs != "" {
		updAttrs, err := cmn.ParseOutputAttrs(attrs, cdevs.CrOSDevAttrMap)
		if err != nil {
			logger.Error(err)
			return err
		}
		formattedAttrs := cdevs.StartChromeDevicesField + updAttrs + cdevs.EndField
		updCall := cdevs.AddFields(cduc, formattedAttrs)
		cduc = updCall.(*admin.ChromeosdevicesUpdateCall)
	}

	if projection != "" {
		proj := strings.ToLower(projection)
		ok := cmn.SliceContainsStr(cdevs.ValidProjections, proj)
		if !ok {
			err = fmt.Errorf(cmn.ErrInvalidProjectionType, projection)
			logger.Error(err)
			return err
		}

		updCall := cdevs.AddProjection(cduc, proj)
		cduc = updCall.(*admin.ChromeosdevicesUpdateCall)
	}

	updCrOSDev, err := cduc.Do()
	if err != nil {
		logger.Error(err)
		return err
	}

	logger.Infof(cmn.InfoCDevUpdated, updCrOSDev.DeviceId)
	fmt.Println(cmn.GminMessage(fmt.Sprintf(cmn.InfoCDevUpdated, updCrOSDev.DeviceId)))

	jsonData, err := json.MarshalIndent(updCrOSDev, "", "    ")
	if err != nil {
		logger.Error(err)
		return err
	}

	fmt.Println(string(jsonData))

	return nil
}

func init() {
	updateCmd.AddCommand(updateCrOSDevCmd)

	updateCrOSDevCmd.Flags().StringVarP(&attrs, "attributes", "a", "", "required device attributes (separated by ~)")
	updateCrOSDevCmd.Flags().StringVarP(&assetID, "assetid", "d", "", "device asset id")
	updateCrOSDevCmd.Flags().StringVarP(&projection, "projection", "j", "", "type of projection")
	updateCrOSDevCmd.Flags().StringVarP(&location, "location", "l", "", "device location")
	updateCrOSDevCmd.Flags().StringVarP(&notes, "notes", "n", "", "notes about device")
	updateCrOSDevCmd.Flags().StringVarP(&orgUnit, "orgunitpath", "t", "", "orgunit device belongs to")
	updateCrOSDevCmd.Flags().StringVarP(&userKey, "user", "u", "", "device user")
}
