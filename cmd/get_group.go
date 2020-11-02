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
	grps "github.com/plusworx/gmin/utils/groups"
	lg "github.com/plusworx/gmin/utils/logging"
	"github.com/spf13/cobra"
	admin "google.golang.org/api/admin/directory/v1"
)

var getGroupCmd = &cobra.Command{
	Use:     "group <email address or id>",
	Aliases: []string{"grp"},
	Args:    cobra.ExactArgs(1),
	Example: `gmin get group agroup@mydomain.org
gmin get grp 042yioqz3p5ulpk -a email`,
	Short: "Outputs information about a group",
	Long:  `Outputs information about a group.`,
	RunE:  doGetGroup,
}

func doGetGroup(cmd *cobra.Command, args []string) error {
	lg.Debugw("starting doGetGroup()",
		"args", args)
	defer lg.Debug("finished doGetGroup()")

	var (
		jsonData []byte
		group    *admin.Group
	)

	srv, err := cmn.CreateService(cmn.SRVTYPEADMIN, admin.AdminDirectoryGroupReadonlyScope)
	if err != nil {
		return err
	}
	ds := srv.(*admin.Service)

	ggc := ds.Groups.Get(args[0])

	flgAttrsVal, err := cmd.Flags().GetString(flgnm.FLG_ATTRIBUTES)
	if err != nil {
		lg.Error(err)
		return err
	}
	if flgAttrsVal != "" {
		formattedAttrs, err := gpars.ParseOutputAttrs(flgAttrsVal, grps.GroupAttrMap)
		if err != nil {
			return err
		}

		getCall := grps.AddFields(ggc, formattedAttrs)
		ggc = getCall.(*admin.GroupsGetCall)
	}

	group, err = grps.DoGet(ggc)
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
	getCmd.AddCommand(getGroupCmd)

	getGroupCmd.Flags().StringP(flgnm.FLG_ATTRIBUTES, "a", "", "required group attributes (separated by ~)")
}
