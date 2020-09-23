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

	cmn "github.com/plusworx/gmin/utils/common"
	cfg "github.com/plusworx/gmin/utils/config"
	mdevs "github.com/plusworx/gmin/utils/mobiledevices"
	"github.com/spf13/cobra"
	admin "google.golang.org/api/admin/directory/v1"
)

var getMobDevCmd = &cobra.Command{
	Use:     "mobiledevice <resource id>",
	Aliases: []string{"mobdevice", "mobdev", "mdev"},
	Args:    cobra.ExactArgs(1),
	Short:   "Outputs information about a mobile device",
	Long: `Outputs information about a mobile device.
	
	Examples:	gmin get mobiledevice AFiQxQ83IZT4llDfTWPZt69JvwSJU0YECe1TVyVZC4x
			gmin get mdev AFiQxQ83IZT4llDfTWPZt69JvwSJU0YECe1TVyVZC4x -a serialnumber`,
	RunE: doGetMobDev,
}

func doGetMobDev(cmd *cobra.Command, args []string) error {
	var (
		jsonData []byte
		mobdev   *admin.MobileDevice
	)

	ds, err := cmn.CreateDirectoryService(admin.AdminDirectoryDeviceMobileReadonlyScope)
	if err != nil {
		logger.Error(err)
		return err
	}

	customerID, err := cfg.ReadConfigString("customerid")
	if err != nil {
		logger.Error(err)
		return err
	}
	mdgc := ds.Mobiledevices.Get(customerID, args[0])

	if attrs != "" {
		formattedAttrs, err := cmn.ParseOutputAttrs(attrs, mdevs.MobDevAttrMap)
		if err != nil {
			logger.Error(err)
			return err
		}

		getCall := mdevs.AddFields(mdgc, formattedAttrs)
		mdgc = getCall.(*admin.MobiledevicesGetCall)
	}

	if projection != "" {
		proj := strings.ToLower(projection)
		ok := cmn.SliceContainsStr(mdevs.ValidProjections, proj)
		if !ok {
			err = fmt.Errorf(cmn.ErrInvalidProjectionType, projection)
			logger.Error(err)
			return err
		}

		getCall := mdevs.AddProjection(mdgc, proj)
		mdgc = getCall.(*admin.MobiledevicesGetCall)
	}

	mobdev, err = mdevs.DoGet(mdgc)
	if err != nil {
		logger.Error(err)
		return err
	}

	jsonData, err = json.MarshalIndent(mobdev, "", "    ")
	if err != nil {
		logger.Error(err)
		return err
	}

	fmt.Println(string(jsonData))

	return nil
}

func init() {
	getCmd.AddCommand(getMobDevCmd)

	getMobDevCmd.Flags().StringVarP(&attrs, "attributes", "a", "", "required device attributes (separated by ~)")
	getMobDevCmd.Flags().StringVarP(&projection, "projection", "j", "", "type of projection")
}
