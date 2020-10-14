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
	mems "github.com/plusworx/gmin/utils/members"
	"github.com/spf13/cobra"
	admin "google.golang.org/api/admin/directory/v1"
)

var updateMemberCmd = &cobra.Command{
	Use:     "group-member <member email address, alias or id> <group email address, alias or id>",
	Aliases: []string{"grp-member", "gmember", "gmem"},
	Args:    cobra.ExactArgs(2),
	Example: `gmin update group-member another.user@mycompany.com office@mycompany.com -d DAILY
gmin upd gmem finance.person@mycompany.com finance@mycompany.com -r MEMBER`,
	Short: "Updates a group member",
	Long:  `Updates a group member.`,
	RunE:  doUpdateMember,
}

func doUpdateMember(cmd *cobra.Command, args []string) error {
	logger.Debugw("starting doUpdateMember()",
		"args", args)

	var (
		member    *admin.Member
		memberKey string
	)

	memberKey = args[0]
	member = new(admin.Member)

	if deliverySetting != "" {
		validDS, err := mems.ValidateDeliverySetting(deliverySetting)
		if err != nil {
			logger.Error(err)
			return err
		}
		member.DeliverySettings = validDS
	}

	if role != "" {
		validRole, err := mems.ValidateRole(role)
		if err != nil {
			logger.Error(err)
			return err
		}
		member.Role = validRole
	}

	ds, err := cmn.CreateDirectoryService(admin.AdminDirectoryGroupMemberScope, admin.AdminDirectoryGroupScope)
	if err != nil {
		logger.Error(err)
		return err
	}

	muc := ds.Members.Update(args[1], memberKey, member)
	_, err = muc.Do()
	if err != nil {
		logger.Error(err)
		return err
	}

	logger.Infof(gmess.INFOMEMBERUPDATED, memberKey, args[1])
	fmt.Println(cmn.GminMessage(fmt.Sprintf(gmess.INFOMEMBERUPDATED, memberKey, args[1])))

	logger.Debug("finished doUpdateMember()")
	return nil
}

func init() {
	updateCmd.AddCommand(updateMemberCmd)

	updateMemberCmd.Flags().StringVarP(&deliverySetting, "delivery-setting", "d", "", "member delivery setting")
	updateMemberCmd.Flags().StringVarP(&role, "role", "r", "", "member role")
}
