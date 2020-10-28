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

	cmn "github.com/plusworx/gmin/utils/common"
	flgnm "github.com/plusworx/gmin/utils/flagnames"
	gpars "github.com/plusworx/gmin/utils/gminparsers"
	grpset "github.com/plusworx/gmin/utils/groupsettings"
	lg "github.com/plusworx/gmin/utils/logging"
	"github.com/spf13/cobra"
	gset "google.golang.org/api/groupssettings/v1"
)

var getGroupSettingsCmd = &cobra.Command{
	Use:     "group-settings <group email address>",
	Aliases: []string{"grp-settings", "grp-set", "gsettings", "gset"},
	Args:    cobra.ExactArgs(1),
	Example: `gmin get group-settings agroup@mydomain.org
gmin get gset 042yioqz3p5ulpk -a email`,
	Short: "Outputs information about group settings",
	Long:  `Outputs information about group settings.`,
	RunE:  doGetGroupSettings,
}

func doGetGroupSettings(cmd *cobra.Command, args []string) error {
	lg.Debugw("starting doGetGroupSettings()",
		"args", args)
	defer lg.Debug("finished doGetGroupSettings()")

	var (
		jsonData []byte
		group    *gset.Groups
	)

	srv, err := cmn.CreateService(cmn.SRVTYPEGRPSETTING, gset.AppsGroupsSettingsScope)
	if err != nil {
		return err
	}
	gss := srv.(*gset.Service)

	gsgc := gss.Groups.Get(args[0])

	flgAttrsVal, err := cmd.Flags().GetString(flgnm.FLG_ATTRIBUTES)
	if err != nil {
		lg.Error(err)
		return err
	}
	if flgAttrsVal != "" {
		formattedAttrs, err := gpars.ParseOutputAttrs(flgAttrsVal, grpset.GroupSettingsAttrMap)
		if err != nil {
			return err
		}

		getCall := grpset.AddFields(gsgc, formattedAttrs)
		gsgc = getCall.(*gset.GroupsGetCall)
	}

	group, err = grpset.DoGet(gsgc)
	if err != nil {
		return err
	}

	jsonData, err = json.MarshalIndent(group, "", "    ")
	if err != nil {
		lg.Error(err)
		return err
	}

	fmt.Println(string(jsonData))

	return nil
}

func init() {
	getCmd.AddCommand(getGroupSettingsCmd)

	getGroupSettingsCmd.Flags().StringVarP(&attrs, flgnm.FLG_ATTRIBUTES, "a", "", "required group attributes (separated by ~)")
}
