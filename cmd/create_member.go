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
	mems "github.com/plusworx/gmin/utils/members"
	"github.com/spf13/cobra"
	admin "google.golang.org/api/admin/directory/v1"
)

var createMemberCmd = &cobra.Command{
	Use:     "group-member <user/group email address> <group email address or id>",
	Aliases: []string{"grp-member", "grp-mem", "gmember", "gmem"},
	Args:    cobra.ExactArgs(2),
	Example: `gmin create group-member another.user@mycompany.com  office@mycompany.com -d NONE
gmin crt gmem finance.person@mycompany.com finance@mycompany.com -r MEMBER`,
	Short: "Makes a user a group member",
	Long:  `Makes a user a group member.`,
	RunE:  doCreateMember,
}

func doCreateMember(cmd *cobra.Command, args []string) error {
	lg.Debugw("starting doCreateMember()",
		"args", args)
	defer lg.Debug("finished doCreateMember()")

	var member *admin.Member

	member = new(admin.Member)

	member.Email = args[0]

	flgDelSetVal, err := cmd.Flags().GetString(flgnm.FLG_DELIVERYSETTING)
	if err != nil {
		lg.Error(err)
		return err
	}

	if flgDelSetVal != "" {
		validDS, err := mems.ValidateDeliverySetting(flgDelSetVal)
		if err != nil {
			return err
		}
		member.DeliverySettings = validDS
	}

	flgRoleVal, err := cmd.Flags().GetString(flgnm.FLG_ROLE)
	if err != nil {
		lg.Error(err)
		return err
	}

	if flgRoleVal != "" {
		validRole, err := mems.ValidateRole(flgRoleVal)
		if err != nil {
			return err
		}
		member.Role = validRole
	}

	ds, err := cmn.CreateDirectoryService(admin.AdminDirectoryGroupMemberScope)
	if err != nil {
		return err
	}

	mic := ds.Members.Insert(args[1], member)
	newMember, err := mic.Do()
	if err != nil {
		lg.Error(err)
		return err
	}

	lg.Infof(gmess.INFO_MEMBERCREATED, newMember.Email, args[1])
	fmt.Println(cmn.GminMessage(fmt.Sprintf(gmess.INFO_MEMBERCREATED, newMember.Email, args[1])))

	return nil
}

func init() {
	createCmd.AddCommand(createMemberCmd)

	createMemberCmd.Flags().StringVarP(&deliverySetting, flgnm.FLG_DELIVERYSETTING, "d", "", "member delivery setting")
	createMemberCmd.Flags().StringVarP(&role, flgnm.FLG_ROLE, "r", "", "member role")
}
