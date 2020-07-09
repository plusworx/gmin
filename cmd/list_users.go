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

	cmn "github.com/plusworx/gmin/utils/common"
	cfg "github.com/plusworx/gmin/utils/config"
	usrs "github.com/plusworx/gmin/utils/users"
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
		ulc = usrs.AddListFields(ulc, formattedAttrs)
	}

	if deleted {
		ulc = usrs.AddListShowDeleted(ulc)
	}

	if domain != "" {
		ulc = usrs.AddListDomain(ulc, domain)
	} else {
		customerID, err := cfg.ReadConfigString("customerid")
		if err != nil {
			return err
		}
		ulc = usrs.AddListCustomer(ulc, customerID)
	}

	if query != "" {
		formattedQuery, err := usrProcessQuery(query)
		if err != nil {
			return err
		}

		ulc = usrs.AddListQuery(ulc, formattedQuery)
	}

	ulc = usrs.AddListMaxResults(ulc, maxResults)

	users, err = usrs.DoList(ulc)
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

	listUsersCmd.Flags().StringVarP(&attrs, "attributes", "a", "", "required user attributes (separated by ~)")
	listUsersCmd.Flags().StringVarP(&domain, "domain", "d", "", "domain from which to get users")
	listUsersCmd.Flags().Int64VarP(&maxResults, "maxresults", "m", 500, "maximum number of results to return")
	listUsersCmd.Flags().StringVarP(&query, "query", "q", "", "selection criteria to get users (separated by ~)")
	listUsersCmd.Flags().BoolVarP(&deleted, "deleted", "x", false, "show deleted users")

}

func usrProcessQuery(query string) (string, error) {
	queryParts, err := cmn.ValidateQuery(query, usrs.QueryAttrMap)
	if err != nil {
		return "", err
	}

	formattedQuery := strings.Join(queryParts, " ")

	return formattedQuery, nil
}
