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
	"strings"

	cmn "github.com/plusworx/gmin/common"
	usrs "github.com/plusworx/gmin/users"
	"github.com/spf13/cobra"
	admin "google.golang.org/api/admin/directory/v1"
)

var listUsersCmd = &cobra.Command{
	Use:     "users",
	Aliases: []string{"user"},
	Short:   "Outputs a list of users",
	Long:    `Outputs a list of users.`,
	RunE:    doListUsers,
}

func doListUsers(cmd *cobra.Command, args []string) error {
	var (
		formattedAttrs string
		users          *admin.Users
		validAttrs     []string
	)

	if query != "" && deleted {
		err := errors.New("gmin: error - cannot provide both --query and --deleted flags")
		return err
	}

	ds, err := cmn.CreateDirectoryService(admin.AdminDirectoryUserReadonlyScope)
	if err != nil {
		return err
	}

	ulc := ds.Users.List()

	if attrs != "" {
		validAttrs, err = cmn.ValidateAttrs(attrs, usrs.UserAttrMap)
		if err != nil {
			return err
		}

		formattedAttrs = usrs.FormatAttrs(validAttrs, false)
	}

	switch true {
	case domain != "" && !deleted:
		users, err = userDomainCall(domain, ulc, formattedAttrs)
	case domain != "" && deleted:
		users, err = userDelDomainCall(domain, ulc, formattedAttrs)
	case domain == "" && !deleted:
		users, err = userAllDomainCall(ulc, formattedAttrs)
	case domain == "" && deleted:
		users, err = userDelAllDomainCall(ulc, formattedAttrs)
	}

	if err != nil {
		return err
	}

	jsonData, err := json.MarshalIndent(users, "", "    ")
	if err != nil {
		return err
	}

	fmt.Println(string(jsonData))

	return nil
}

func init() {
	listCmd.AddCommand(listUsersCmd)

	listUsersCmd.Flags().StringVarP(&domain, "domain", "d", "", "domain from which to get users")
	listUsersCmd.Flags().StringVarP(&attrs, "attrs", "a", "", "required user attributes (separated by ~)")
	listUsersCmd.Flags().StringVarP(&query, "query", "q", "", "selection criteria to get users (separated by ~)")
	listUsersCmd.Flags().BoolVarP(&deleted, "deleted", "x", false, "show deleted users")

}

func userAllDomainCall(ulc *admin.UsersListCall, fmtAttrs string) (*admin.Users, error) {
	var (
		err   error
		users *admin.Users
	)

	formattedQuery, err := usrProcessQuery(query)
	if err != nil {
		return nil, err
	}

	switch true {
	case formattedQuery == "" && attrs == "":
		users, err = usrs.AllDomain(ulc)
	case formattedQuery != "" && attrs == "":
		users, err = usrs.AllDomainQuery(ulc, formattedQuery)
	case formattedQuery == "" && attrs != "":
		users, err = usrs.AllDomainAttrs(ulc, fmtAttrs)
	case formattedQuery != "" && attrs != "":
		users, err = usrs.AllDomainQueryAttrs(ulc, formattedQuery, fmtAttrs)
	}

	if err != nil {
		return nil, err
	}

	return users, nil
}

func userDelDomainCall(domain string, ulc *admin.UsersListCall, fmtAttrs string) (*admin.Users, error) {
	var (
		err   error
		users *admin.Users
	)

	if attrs == "" {
		users, err = usrs.DelDomain(domain, ulc)
	} else {
		users, err = usrs.DelDomainAttrs(domain, ulc, fmtAttrs)
	}

	if err != nil {
		return nil, err
	}

	return users, nil
}

func userDelAllDomainCall(ulc *admin.UsersListCall, fmtAttrs string) (*admin.Users, error) {
	var (
		err   error
		users *admin.Users
	)

	if attrs == "" {
		users, err = usrs.DelAllDomain(ulc)
	} else {
		users, err = usrs.DelAllDomainAttrs(ulc, fmtAttrs)
	}

	if err != nil {
		return nil, err
	}

	return users, nil
}

func userDomainCall(domain string, ulc *admin.UsersListCall, fmtAttrs string) (*admin.Users, error) {
	var (
		err   error
		users *admin.Users
	)

	formattedQuery, err := usrProcessQuery(query)
	if err != nil {
		return nil, err
	}

	switch true {
	case formattedQuery == "" && attrs == "":
		users, err = usrs.Domain(domain, ulc)
	case formattedQuery != "" && attrs == "":
		users, err = usrs.DomainQuery(domain, ulc, formattedQuery)
	case formattedQuery == "" && attrs != "":
		users, err = usrs.DomainAttrs(domain, ulc, fmtAttrs)
	case formattedQuery != "" && attrs != "":
		users, err = usrs.DomainQueryAttrs(domain, ulc, formattedQuery, fmtAttrs)
	}

	if err != nil {
		return nil, err
	}

	return users, nil
}

func usrProcessQuery(query string) (string, error) {
	var formattedQuery string

	if query != "" {
		queryParts, err := cmn.ValidateQuery(query, usrs.QueryAttrMap)
		if err != nil {
			return "", err
		}

		formattedQuery = strings.Join(queryParts, " ")
	} else {
		formattedQuery = ""
	}

	return formattedQuery, nil
}
