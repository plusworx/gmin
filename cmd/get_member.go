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
	"errors"
	"fmt"
	"log"

	valid "github.com/asaskevich/govalidator"
	cmn "github.com/plusworx/gmin/common"
	mems "github.com/plusworx/gmin/members"
	usrs "github.com/plusworx/gmin/users"
	"github.com/spf13/cobra"
	admin "google.golang.org/api/admin/directory/v1"
)

var getMemberCmd = &cobra.Command{
	Use:     "member <email address or id>",
	Aliases: []string{"mem"},
	Args:    cobra.ExactArgs(1),
	Short:   "Outputs information about a member of a group or orgunit",
	Long: `Outputs information about a member of a group or orgunit.
	
	Examples: gmin get member auser@mydomain.org -o /IT
	          gmin get member 12345678 -g mygroup@mydomain.org -a email`,
	RunE: doGetMember,
}

func doGetMember(cmd *cobra.Command, args []string) error {
	var jsonData []byte

	err := mems.ValidateFlags(group, orgUnit)
	if err != nil {
		return err
	}

	if group != "" {
		jsonData, err = processGroupMember(args[0], attrs, group)
		if err != nil {
			return err
		}
	}

	if orgUnit != "" {
		ok := valid.IsEmail(args[0])
		if ok {
			jsonData = processOrgUnitMember(args[0], attrs, orgUnit)
		} else {
			err := errors.New("gmin: error - argument must be an email address for orgunit member")
			return err
		}
	}

	fmt.Println(string(jsonData))

	return nil
}

func init() {
	getCmd.AddCommand(getMemberCmd)

	getMemberCmd.Flags().StringVarP(&attrs, "attributes", "a", "", "required group attributes (separated by ~)")
	getMemberCmd.Flags().StringVarP(&group, "group", "g", "", "email address of group")
	getMemberCmd.Flags().StringVarP(&orgUnit, "orgunit", "o", "", "path of orgunit")
}

func processGroupMember(memID string, attrs string, groupEmail string) ([]byte, error) {
	var (
		member     *admin.Member
		validAttrs []string
	)

	ds, err := cmn.CreateDirectoryService(admin.AdminDirectoryGroupMemberReadonlyScope)
	if err != nil {
		return nil, err
	}

	mgc := ds.Members.Get(groupEmail, memID)

	if attrs != "" {
		validAttrs, err = cmn.ValidateAttrs(attrs, mems.MemberAttrMap)
		if err != nil {
			return nil, err
		}

		formattedAttrs := mems.FormatAttrs(validAttrs, true)
		member, err = mems.SingleAttrs(mgc, formattedAttrs)
		if err != nil {
			return nil, err
		}
	} else {
		member, err = mems.Single(mgc)
		if err != nil {
			return nil, err
		}
	}

	jsonData, err := json.MarshalIndent(member, "", "    ")
	if err != nil {
		return nil, err
	}

	return jsonData, nil
}

func processOrgUnitMember(email string, attrs string, orgUnit string) []byte {
	var (
		query      string
		users      *admin.Users
		validAttrs []string
	)

	ds, err := cmn.CreateDirectoryService(admin.AdminDirectoryUserReadonlyScope)
	if err != nil {
		log.Fatal(err)
	}

	ulc := ds.Users.List()

	query = "orgUnitPath=" + orgUnit + " " + "email=" + email

	if attrs != "" {
		validAttrs, err = cmn.ValidateAttrs(attrs, usrs.UserAttrMap)
		if err != nil {
			log.Fatal(err)
		}

		formattedAttrs := usrs.FormatAttrs(validAttrs, false)

		users, err = usrs.AllDomainQueryAttrs(ulc, query, formattedAttrs)
		if err != nil {
			log.Fatal(err)
		}
	} else {
		users, err = usrs.AllDomainQuery(ulc, query)
		if err != nil {
			log.Fatal(err)
		}
	}

	jsonData, err := json.MarshalIndent(users, "", "    ")
	if err != nil {
		log.Fatal(err)
	}

	return jsonData
}
