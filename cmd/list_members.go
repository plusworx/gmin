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
	"strconv"

	"github.com/jinzhu/copier"
	cmn "github.com/plusworx/gmin/utils/common"
	mems "github.com/plusworx/gmin/utils/members"
	"github.com/spf13/cobra"
	admin "google.golang.org/api/admin/directory/v1"
)

var listMembersCmd = &cobra.Command{
	Use:     "group-members <group email address or id>",
	Aliases: []string{"group-member", "grp-members", "grp-member", "grp-mems", "grp-mem", "gmembers", "gmember", "gmems", "gmem"},
	Args:    cobra.ExactArgs(1),
	Short:   "Outputs a list of group members",
	Long:    `Outputs a list of group members. Must specify a group email address or id.`,
	RunE:    doListMembers,
}

func doListMembers(cmd *cobra.Command, args []string) error {
	var (
		jsonData   []byte
		members    *admin.Members
		newMembers = mems.GminMembers{}
	)

	ds, err := cmn.CreateDirectoryService(admin.AdminDirectoryGroupMemberReadonlyScope)
	if err != nil {
		return err
	}

	mlc := ds.Members.List(args[0])

	if attrs != "" {
		listAttrs, err := cmn.ParseOutputAttrs(attrs, mems.MemberAttrMap)
		if err != nil {
			return err
		}
		formattedAttrs := mems.StartMembersField + listAttrs + mems.EndField

		listCall := mems.AddFields(mlc, formattedAttrs)
		mlc = listCall.(*admin.MembersListCall)
	}

	if role != "" {
		formattedRoles, err := cmn.ParseOutputAttrs(role, mems.RoleMap)
		if err != nil {
			return err
		}
		mlc = mems.AddRoles(mlc, formattedRoles)
	}

	mlc = mems.AddMaxResults(mlc, maxResults)

	members, err = mems.DoList(mlc)
	if err != nil {
		return err
	}

	if pages != "" {
		err = doMemPages(mlc, members, pages)
		if err != nil {
			return err
		}
	}

	if attrs == "" {
		copier.Copy(&newMembers, members)

		jsonData, err = json.MarshalIndent(newMembers, "", "    ")
		if err != nil {
			return err
		}
	} else {
		jsonData, err = json.MarshalIndent(members, "", "    ")
		if err != nil {
			return err
		}
	}

	fmt.Println(string(jsonData))

	return nil
}

func doMemAllPages(mlc *admin.MembersListCall, members *admin.Members) error {
	if members.NextPageToken != "" {
		mlc = mems.AddPageToken(mlc, members.NextPageToken)
		nxtMems, err := mems.DoList(mlc)
		if err != nil {
			return err
		}
		members.Members = append(members.Members, nxtMems.Members...)
		members.Etag = nxtMems.Etag
		members.NextPageToken = nxtMems.NextPageToken

		if nxtMems.NextPageToken != "" {
			doMemAllPages(mlc, members)
		}
	}

	return nil
}

func doMemNumPages(mlc *admin.MembersListCall, members *admin.Members, numPages int) error {
	if members.NextPageToken != "" && numPages > 0 {
		mlc = mems.AddPageToken(mlc, members.NextPageToken)
		nxtMems, err := mems.DoList(mlc)
		if err != nil {
			return err
		}
		members.Members = append(members.Members, nxtMems.Members...)
		members.Etag = nxtMems.Etag
		members.NextPageToken = nxtMems.NextPageToken

		if nxtMems.NextPageToken != "" {
			doMemNumPages(mlc, members, numPages-1)
		}
	}

	return nil
}

func doMemPages(mlc *admin.MembersListCall, members *admin.Members, pages string) error {
	if pages == "all" {
		err := doMemAllPages(mlc, members)
		if err != nil {
			return err
		}
	} else {
		numPages, err := strconv.Atoi(pages)
		if err != nil {
			return errors.New("gmin: error - pages must be 'all' or a number")
		}

		if numPages > 1 {
			err = doMemNumPages(mlc, members, numPages-1)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func init() {
	listCmd.AddCommand(listMembersCmd)

	listMembersCmd.Flags().StringVarP(&attrs, "attributes", "a", "", "required member attributes (separated by ~)")
	listMembersCmd.Flags().Int64VarP(&maxResults, "maxresults", "m", 200, "maximum number or results to return")
	listMembersCmd.Flags().StringVarP(&pages, "pages", "p", "", "number of pages of results to be returned")
	listMembersCmd.Flags().StringVarP(&role, "roles", "r", "", "roles to filter results by (separated by ~)")
}
