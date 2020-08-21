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

var deleteMobDevCmd = &cobra.Command{
	Use:     "mobiledevice <resource id>",
	Aliases: []string{"mobdevice", "mobdev", "mdev"},
	Args:    cobra.ExactArgs(1),
	Short:   "Deletes a mobile device",
	Long: `Deletes a mobile device.
	
	Examples:	gmin delete mobiledevice AFiQxQ83IZT4llDfTWPZt69JvwSJU0YECe1TVyVZC4x
			gmin del mdev AFiQxQ83IZT4llDfTWPZt69JvwSJU0YECe1TVyVZC4x`,
	RunE: doDeleteMobDev,
}

func doDeleteMobDev(cmd *cobra.Command, args []string) error {
	ds, err := cmn.CreateDirectoryService(admin.AdminDirectoryDeviceMobileScope)
	if err != nil {
		return err
	}

	customerID, err := cfg.ReadConfigString("customerid")
	if err != nil {
		return err
	}
	mdc := ds.Mobiledevices.Delete(customerID, args[0])

	err = mdc.Do()
	if err != nil {
		return err
	}

	fmt.Printf("**** gmin: mobile device %s deleted ****\n", args[0])

	return nil
}

func init() {
	deleteCmd.AddCommand(deleteMobDevCmd)
}