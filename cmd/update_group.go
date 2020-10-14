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
	"github.com/spf13/cobra"
	admin "google.golang.org/api/admin/directory/v1"
)

var updateGroupCmd = &cobra.Command{
	Use:     "group <group email address, alias or id>",
	Aliases: []string{"grp"},

	Args: cobra.ExactArgs(1),
	Example: `gmin update group office@mycompany.com
gmin upd grp 02502m921to3a9m -e newfinance@mycompany.com -n "New Finance" -d "New Finance Department"`,
	Short: "Updates a group",
	Long:  `Updates a group .`,
	RunE:  doUpdateGroup,
}

func doUpdateGroup(cmd *cobra.Command, args []string) error {
	logger.Debugw("starting doUpdateGroup()",
		"args", args)

	var (
		group    *admin.Group
		groupKey string
	)

	groupKey = args[0]
	group = new(admin.Group)

	if groupEmail != "" {
		group.Email = groupEmail
	}

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

	guc := ds.Groups.Update(groupKey, group)
	_, err = guc.Do()
	if err != nil {
		logger.Error(err)
		return err
	}

	logger.Infof(gmess.INFO_GROUPUPDATED, groupKey)
	fmt.Println(cmn.GminMessage(fmt.Sprintf(gmess.INFO_GROUPUPDATED, groupKey)))

	logger.Debug("finished doUpdateGroup()")
	return nil
}

func init() {
	updateCmd.AddCommand(updateGroupCmd)

	updateGroupCmd.Flags().StringVarP(&groupDesc, "description", "d", "", "group description")
	updateGroupCmd.Flags().StringVarP(&groupEmail, "email", "e", "", "group email")
	updateGroupCmd.Flags().StringVarP(&groupName, "name", "n", "", "group name")
}
