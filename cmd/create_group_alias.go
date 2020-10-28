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
	gmess "github.com/plusworx/gmin/utils/gminmessages"
	lg "github.com/plusworx/gmin/utils/logging"
	"github.com/spf13/cobra"
	admin "google.golang.org/api/admin/directory/v1"
)

var createGroupAliasCmd = &cobra.Command{
	Use:     "group-alias <alias email address> <group email address or id>",
	Aliases: []string{"galias", "ga"},
	Args:    cobra.ExactArgs(2),
	Example: `gmin create group-alias group.alias@mycompany.com finance@mycompany.com
gmin crt ga group.alias@mycompany.com sales@mycompany.com`,
	Short: "Creates a group alias",
	Long:  `Creates a group alias.`,
	RunE:  doCreateGroupAlias,
}

func doCreateGroupAlias(cmd *cobra.Command, args []string) error {
	lg.Debugw("starting doCreateGroupAlias()",
		"args", args)
	defer lg.Debug("finished doCreateGroupAlias()")

	var alias *admin.Alias

	alias = new(admin.Alias)

	alias.Alias = args[0]

	srv, err := cmn.CreateService(cmn.SRVTYPEADMIN, admin.AdminDirectoryGroupScope)
	if err != nil {
		return err
	}
	ds := srv.(*admin.Service)

	gaic := ds.Groups.Aliases.Insert(args[1], alias)
	newAlias, err := gaic.Do()
	if err != nil {
		lg.Error(err)
		return err
	}

	fmt.Println(cmn.GminMessage(fmt.Sprintf(gmess.INFO_GROUPALIASCREATED, newAlias.Alias, args[1])))
	lg.Infof(gmess.INFO_GROUPALIASCREATED, newAlias.Alias, args[1])

	return nil
}

func init() {
	createCmd.AddCommand(createGroupAliasCmd)
}
