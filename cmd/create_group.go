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
	flgnm "github.com/plusworx/gmin/utils/flagnames"
	gmess "github.com/plusworx/gmin/utils/gminmessages"
	lg "github.com/plusworx/gmin/utils/logging"
	"github.com/spf13/cobra"
	admin "google.golang.org/api/admin/directory/v1"
)

var createGroupCmd = &cobra.Command{
	Use:     "group <group email address>",
	Aliases: []string{"grp"},
	Args:    cobra.ExactArgs(1),
	Example: `gmin create group office@mycompany.com
gmin crt grp finance@mycompany.com -n Finance -d "Finance Department Group"`,
	Short: "Creates a group",
	Long:  `Creates a group.`,
	RunE:  doCreateGroup,
}

func doCreateGroup(cmd *cobra.Command, args []string) error {
	lg.Debugw("starting doCreateGroup()",
		"args", args)

	var group *admin.Group

	group = new(admin.Group)

	group.Email = args[0]

	flgDescVal, err := cmd.Flags().GetString(flgnm.FLG_DESCRIPTION)
	if err != nil {
		lg.Error(err)
		return err
	}

	if flgDescVal != "" {
		group.Description = flgDescVal
	}

	flgNameVal, err := cmd.Flags().GetString(flgnm.FLG_NAME)
	if err != nil {
		lg.Error(err)
		return err
	}

	if flgNameVal != "" {
		group.Name = flgNameVal
	}

	ds, err := cmn.CreateDirectoryService(admin.AdminDirectoryGroupScope)
	if err != nil {
		lg.Error(err)
		return err
	}

	gic := ds.Groups.Insert(group)
	newGroup, err := gic.Do()
	if err != nil {
		lg.Error(err)
		return err
	}

	lg.Infof(gmess.INFO_GROUPCREATED, newGroup.Email)
	fmt.Println(cmn.GminMessage(fmt.Sprintf(gmess.INFO_GROUPCREATED, newGroup.Email)))

	lg.Debug("finished doCreateGroup()")
	return nil
}

func init() {
	createCmd.AddCommand(createGroupCmd)

	createGroupCmd.Flags().StringVarP(&groupDesc, flgnm.FLG_DESCRIPTION, "d", "", "group description")
	createGroupCmd.Flags().StringVarP(&groupName, flgnm.FLG_NAME, "n", "", "group name")
}
