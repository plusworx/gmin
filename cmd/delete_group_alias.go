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

var deleteGroupAliasCmd = &cobra.Command{
	Use:     "group-alias <alias email address> <group email address or id>",
	Aliases: []string{"grp-alias", "galias", "ga"},
	Args:    cobra.ExactArgs(2),
	Example: `gmin delete group-alias my.alias@mycompany.com mygroup@mycompany.com
gmin del ga my.alias@mycompany.com mygroup@mycompany.com`,
	Short: "Deletes group alias",
	Long:  `Deletes group alias.`,
	RunE:  doDeleteGroupAlias,
}

func doDeleteGroupAlias(cmd *cobra.Command, args []string) error {
	lg.Debugw("starting doDeleteGroupAlias()",
		"args", args)
	defer lg.Debug("finished doDeleteGroupAlias()")

	srv, err := cmn.CreateService(cmn.SRVTYPEADMIN, admin.AdminDirectoryGroupScope)
	if err != nil {
		return err
	}
	ds := srv.(*admin.Service)

	gadc := ds.Groups.Aliases.Delete(args[1], args[0])

	err = gadc.Do()
	if err != nil {
		lg.Error(err)
		return err
	}

	fmt.Println(cmn.GminMessage(fmt.Sprintf(gmess.INFO_GROUPALIASDELETED, args[0], args[1])))
	lg.Infof(gmess.INFO_GROUPALIASDELETED, args[0], args[1])

	return nil
}

func init() {
	deleteCmd.AddCommand(deleteGroupAliasCmd)
}
