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
	"encoding/json"
	"fmt"

	cmn "github.com/plusworx/gmin/common"
	mems "github.com/plusworx/gmin/members"
	"github.com/spf13/cobra"
	admin "google.golang.org/api/admin/directory/v1"
)

var listMembersCmd = &cobra.Command{
	Use:     "group-members <group email address or id>",
	Aliases: []string{"group-member", "grp-members", "grp-member", "grp-mems", "grp-mem", "gmembers", "gmember", "gmems", "gmem"},
	Args:    cobra.ExactArgs(1),
	Short:   "Outputs a list of group members",
	Long:    `Outputs a list of group members. Must specify a group email address.`,
	RunE:    doListMembers,
}

func doListMembers(cmd *cobra.Command, args []string) error {
	var jsonData []byte

	jsonData, err := processGroupMembers(attrs, args[0])
	if err != nil {
		return err
	}

	fmt.Println(string(jsonData))

	return nil
}

func init() {
	listCmd.AddCommand(listMembersCmd)

	listMembersCmd.Flags().StringVarP(&attrs, "attributes", "a", "", "required member attributes (separated by ~)")
}

func processGroupMembers(attrs string, groupEmail string) ([]byte, error) {
	var (
		members    *admin.Members
		validAttrs []string
	)

	ds, err := cmn.CreateDirectoryService(admin.AdminDirectoryGroupMemberReadonlyScope)
	if err != nil {
		return nil, err
	}

	mlc := ds.Members.List(groupEmail)

	if attrs != "" {
		validAttrs, err = cmn.ValidateAttrs(attrs, mems.MemberAttrMap)
		if err != nil {
			return nil, err
		}

		formattedAttrs := mems.FormatAttrs(validAttrs, false)
		members, err = mems.Attrs(mlc, formattedAttrs)
		if err != nil {
			return nil, err
		}
	} else {
		members, err = mems.Members(mlc)
		if err != nil {
			return nil, err
		}
	}

	jsonData, err := json.MarshalIndent(members, "", "    ")
	if err != nil {
		return nil, err
	}

	return jsonData, nil
}
