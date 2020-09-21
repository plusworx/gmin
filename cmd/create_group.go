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
	"github.com/spf13/cobra"
	admin "google.golang.org/api/admin/directory/v1"
)

var createGroupCmd = &cobra.Command{
	Use:     "group <group email address>",
	Aliases: []string{"grp"},
	Args:    cobra.ExactArgs(1),
	Short:   "Creates a group",
	Long: `Creates a group.
	
	Examples:	gmin create group office@mycompany.com
			gmin crt grp finance@mycompany.com -n Finance -d "Finance Department Group"`,
	RunE: doCreateGroup,
}

func doCreateGroup(cmd *cobra.Command, args []string) error {
	var group *admin.Group

	group = new(admin.Group)

	group.Email = args[0]

	if groupDesc != "" {
		group.Description = groupDesc
	}

	if groupName != "" {
		group.Name = groupName
	}

	ds, err := cmn.CreateDirectoryService(admin.AdminDirectoryGroupScope)
	if err != nil {
		logger.Error(err)
		return err
	}

	gic := ds.Groups.Insert(group)
	newGroup, err := gic.Do()
	if err != nil {
		logger.Error(err)
		return err
	}

	logger.Infof(cmn.InfoGroupCreated, newGroup.Email)
	fmt.Println(cmn.GminMessage(fmt.Sprintf(cmn.InfoGroupCreated, newGroup.Email)))

	return nil
}

func init() {
	createCmd.AddCommand(createGroupCmd)

	createGroupCmd.Flags().StringVarP(&groupDesc, "description", "d", "", "group description")
	createGroupCmd.Flags().StringVarP(&groupName, "name", "n", "", "group name")
}
