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
	"errors"
	"fmt"

	cmn "github.com/plusworx/gmin/common"
	"github.com/spf13/cobra"
	admin "google.golang.org/api/admin/directory/v1"
)

var deleteMemberCmd = &cobra.Command{
	Use:     "group-member <member email address or id> -g <group email address or id>",
	Aliases: []string{"grp-member", "grp-mem", "gmember", "gmem"},
	Args:    cobra.ExactArgs(1),
	Short:   "Deletes member of a group",
	Long:    `Deletes member of a group.`,
	RunE:    doDeleteMember,
}

func doDeleteMember(cmd *cobra.Command, args []string) error {
	if group == "" {
		err := errors.New("gmin: error - group email address or id must be provided")
		return err
	}

	ds, err := cmn.CreateDirectoryService(admin.AdminDirectoryGroupMemberScope)
	if err != nil {
		return err
	}

	mdc := ds.Members.Delete(group, args[0])

	err = mdc.Do()
	if err != nil {
		return err
	}

	fmt.Printf("**** gmin: member %s of group %s deleted ****\n", args[0], group)

	return nil
}

func init() {
	deleteCmd.AddCommand(deleteMemberCmd)

	deleteMemberCmd.Flags().StringVarP(&group, "group", "g", "", "email address or id of group")
}
