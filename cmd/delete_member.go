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

var deleteMemberCmd = &cobra.Command{
	Use:     "group-member <member email address or id> <group email address or id>",
	Aliases: []string{"grp-member", "grp-mem", "gmember", "gmem"},
	Args:    cobra.ExactArgs(2),
	Example: `gmin delete group-member mymember@mycompany.org mygroup@mycompany.org
gmin del gmem mymember@mycompany.org mygroup@mycompany.org`,
	Short: "Deletes member of a group",
	Long:  `Deletes member of a group.`,
	RunE:  doDeleteMember,
}

func doDeleteMember(cmd *cobra.Command, args []string) error {
	lg.Debugw("starting doDeleteMember()",
		"args", args)

	ds, err := cmn.CreateDirectoryService(admin.AdminDirectoryGroupMemberScope)
	if err != nil {
		lg.Error(err)
		return err
	}

	mdc := ds.Members.Delete(args[1], args[0])

	err = mdc.Do()
	if err != nil {
		lg.Error(err)
		return err
	}

	lg.Infof(gmess.INFO_MEMBERDELETED, args[0], args[1])
	fmt.Println(cmn.GminMessage(fmt.Sprintf(gmess.INFO_MEMBERDELETED, args[0], args[1])))

	lg.Debug("finished doDeleteMember()")
	return nil
}

func init() {
	deleteCmd.AddCommand(deleteMemberCmd)
}
