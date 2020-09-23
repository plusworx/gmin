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

var getCrOSDevCmd = &cobra.Command{
	Use:     "chromeosdevice <device id>",
	Aliases: []string{"crosdevice", "crosdev", "cdev"},
	Args:    cobra.ExactArgs(1),
	Short:   "Outputs information about a ChromeOS device",
	Long: `Outputs information about a ChromeOS device.
	
	Examples:	gmin get chromeosdevice 5ad9ae43-5996-394e-9c39-12d45a8f10e8
			gmin get cdev 5ad9ae43-5996-394e-9c39-12d45a8f10e8 -a serialnumber`,
	RunE: doGetCrOSDev,
}

func doGetCrOSDev(cmd *cobra.Command, args []string) error {
	var (
		jsonData []byte
		crosdev  *admin.ChromeOsDevice
	)

	ds, err := cmn.CreateDirectoryService(admin.AdminDirectoryDeviceChromeosReadonlyScope)
	if err != nil {
		logger.Error(err)
		return err
	}

	customerID, err := cfg.ReadConfigString("customerid")
	if err != nil {
		logger.Error(err)
		return err
	}
	cdgc := ds.Chromeosdevices.Get(customerID, args[0])

	if attrs != "" {
		formattedAttrs, err := cmn.ParseOutputAttrs(attrs, cdevs.CrOSDevAttrMap)
		if err != nil {
			logger.Error(err)
			return err
		}

		getCall := cdevs.AddFields(cdgc, formattedAttrs)
		cdgc = getCall.(*admin.ChromeosdevicesGetCall)
	}

	if projection != "" {
		proj := strings.ToLower(projection)
		ok := cmn.SliceContainsStr(cdevs.ValidProjections, proj)
		if !ok {
			err = fmt.Errorf(cmn.ErrInvalidProjectionType, projection)
			logger.Error(err)
			return err
		}

		getCall := cdevs.AddProjection(cdgc, proj)
		cdgc = getCall.(*admin.ChromeosdevicesGetCall)
	}

	crosdev, err = cdevs.DoGet(cdgc)
	if err != nil {
		logger.Error(err)
		return err
	}

	jsonData, err = json.MarshalIndent(crosdev, "", "    ")
	if err != nil {
		logger.Error(err)
		return err
	}

	fmt.Println(string(jsonData))

	return nil
}

func init() {
	getCmd.AddCommand(getCrOSDevCmd)

	getCrOSDevCmd.Flags().StringVarP(&attrs, "attributes", "a", "", "required device attributes (separated by ~)")
	getCrOSDevCmd.Flags().StringVarP(&projection, "projection", "j", "", "type of projection")
}
