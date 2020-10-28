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
	lg "github.com/plusworx/gmin/utils/logging"
	uas "github.com/plusworx/gmin/utils/useraliases"
	"github.com/spf13/cobra"
	admin "google.golang.org/api/admin/directory/v1"
)

var listUserAliasesCmd = &cobra.Command{
	Use:     "user-aliases <user email address or id>",
	Aliases: []string{"user-alias", "ualiases", "ualias", "uas", "ua"},
	Args:    cobra.ExactArgs(1),
	Example: `gmin list user-aliases myuser@mycompany.com
gmin ls uas myuser@mycompany.com`,
	Short: "Outputs a list of user aliases",
	Long:  `Outputs a list of user aliases.`,
	RunE:  doListUserAliases,
}

func doListUserAliases(cmd *cobra.Command, args []string) error {
	lg.Debugw("starting doListUserAliases()",
		"args", args)
	defer lg.Debug("finished doListUserAliases()")

	var aliases *admin.Aliases

	srv, err := cmn.CreateService(cmn.SRVTYPEADMIN, admin.AdminDirectoryUserAliasReadonlyScope)
	if err != nil {
		return err
	}
	ds := srv.(*admin.Service)

	ualc := ds.Users.Aliases.List(args[0])

	flgAttrsVal, err := cmd.Flags().GetString(flgnm.FLG_ATTRIBUTES)
	if err != nil {
		lg.Error(err)
		return err
	}
	if flgAttrsVal != "" {
		listAttrs, err := gpars.ParseOutputAttrs(flgAttrsVal, uas.UserAliasAttrMap)
		if err != nil {
			return err
		}
		formattedAttrs := uas.STARTALIASESFIELD + listAttrs + uas.ENDFIELD

		ualc = uas.AddFields(ualc, formattedAttrs)
	}

	aliases, err = uas.DoList(ualc)
	if err != nil {
		return err
	}

	jsonData, err := json.MarshalIndent(aliases, "", "    ")
	if err != nil {
		lg.Error(err)
		return err
	}

	fmt.Println(string(jsonData))

	return nil
}

func init() {
	listCmd.AddCommand(listUserAliasesCmd)

	listUserAliasesCmd.Flags().StringVarP(&attrs, flgnm.FLG_ATTRIBUTES, "a", "", "required user alias attributes (separated by ~)")
}
