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

	"github.com/jinzhu/copier"
	cmn "github.com/plusworx/gmin/utils/common"
	mems "github.com/plusworx/gmin/utils/members"
	"github.com/spf13/cobra"
	admin "google.golang.org/api/admin/directory/v1"
)

var getMemberCmd = &cobra.Command{
	Use:     "group-member <member email address or id> <group email address or id>",
	Aliases: []string{"grp-member", "grp-mem", "gmember", "gmem"},
	Args:    cobra.ExactArgs(2),
	Short:   "Outputs information about a member of a group",
	Long: `Outputs information about a member of a group.
	
	Examples:	gmin get group-member 127987192327764327416 mygroup@mydomain.org -a email
			gmin get gmem jack.black@mydomain.org mygroup@mydomain.org -a email`,
	RunE: doGetMember,
}

func doGetMember(cmd *cobra.Command, args []string) error {
	var jsonData []byte

	jsonData, err := processGroupMember(args[0], attrs, args[1])
	if err != nil {
		return err
	}

	fmt.Println(string(jsonData))

	return nil
}

func init() {
	getCmd.AddCommand(getMemberCmd)

	getMemberCmd.Flags().StringVarP(&attrs, "attributes", "a", "", "required group attributes (separated by ~)")
}

func processGroupMember(memID string, attrs string, groupEmail string) ([]byte, error) {
	var (
		jsonData  []byte
		member    *admin.Member
		newMember = mems.GminMember{}
	)

	ds, err := cmn.CreateDirectoryService(admin.AdminDirectoryGroupMemberReadonlyScope)
	if err != nil {
		return nil, err
	}

	mgc := ds.Members.Get(groupEmail, memID)

	if attrs != "" {
		formattedAttrs, err := cmn.ParseOutputAttrs(attrs, mems.MemberAttrMap)
		if err != nil {
			return nil, err
		}

		getCall := mems.AddFields(mgc, formattedAttrs)
		mgc = getCall.(*admin.MembersGetCall)
	}

	member, err = mems.DoGet(mgc)
	if err != nil {
		return nil, err
	}

	if attrs == "" {
		copier.Copy(&newMember, member)

		jsonData, err = json.MarshalIndent(newMember, "", "    ")
		if err != nil {
			return nil, err
		}
	} else {
		jsonData, err = json.MarshalIndent(member, "", "    ")
		if err != nil {
			return nil, err
		}
	}

	return jsonData, nil
}
