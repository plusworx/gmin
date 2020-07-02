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
	mems "github.com/plusworx/gmin/members"
	"github.com/spf13/cobra"
	admin "google.golang.org/api/admin/directory/v1"
)

var createMemberCmd = &cobra.Command{
	Use:     "group-member <user/group email address> -g <group email address or id>",
	Aliases: []string{"grp-member", "grp-mem", "gmember", "gmem"},
	Args:    cobra.ExactArgs(1),
	Short:   "Makes a user a group member",
	Long: `Makes a user a group member.
	
	Examples: gmin create group-member another.user@mycompany.com  -g office@mycompany.com -d NONE
	          gmin crt gmem finance.person@mycompany.com -g finance@mycompany.com -r MEMBER`,
	RunE: doCreateMember,
}

func doCreateMember(cmd *cobra.Command, args []string) error {
	var member *admin.Member

	member = new(admin.Member)

	member.Email = args[0]

	if deliverySetting != "" {
		validDS, err := mems.ValidateDeliverySetting(deliverySetting)
		if err != nil {
			return err
		}
		member.DeliverySettings = validDS
	}

	if group == "" {
		err := errors.New("gmin: error - group must be provided")
		return err
	}

	if role != "" {
		validRole, err := mems.ValidateRole(role)
		if err != nil {
			return err
		}
		member.Role = validRole
	}

	ds, err := cmn.CreateDirectoryService(admin.AdminDirectoryGroupMemberScope)
	if err != nil {
		return err
	}

	mic := ds.Members.Insert(group, member)
	newMember, err := mic.Do()
	if err != nil {
		return err
	}

	fmt.Println("**** member " + newMember.Email + " created in group " + group + " ****")

	return nil
}

func init() {
	createCmd.AddCommand(createMemberCmd)

	createMemberCmd.Flags().StringVarP(&deliverySetting, "deliverysetting", "d", "", "member delivery setting")
	createMemberCmd.Flags().StringVarP(&group, "group", "g", "", "email address or id of group")
	createMemberCmd.Flags().StringVarP(&role, "role", "r", "", "member role")
}
