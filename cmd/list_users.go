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
		validOrderBy   string
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
		validAttrs, err = cmn.ValidateArgs(attrs, usrs.UserAttrMap, cmn.AttrStr)
		if err != nil {
			return err
		}

		formattedAttrs = usrs.FormatAttrs(validAttrs, false)
		listCall := usrs.AddFields(ulc, formattedAttrs)
		ulc = listCall.(*admin.UsersListCall)
	}

	if deleted {
		ulc = usrs.AddShowDeleted(ulc)
	}

	if domain != "" {
		ulc = usrs.AddDomain(ulc, domain)
	} else {
		customerID, err := cfg.ReadConfigString("customerid")
		if err != nil {
			return err
		}
		ulc = usrs.AddCustomer(ulc, customerID)
	}

	if projection != "" {
		proj := strings.ToLower(projection)
		ok := cmn.SliceContainsStr(usrs.ValidProjections, proj)
		if !ok {
			return fmt.Errorf("gmin: error - %v is not a valid projection type", projection)
		}

		listCall := usrs.AddProjection(ulc, proj)
		ulc = listCall.(*admin.UsersListCall)
	}

	if query != "" {
		formattedQuery, err := usrProcessQuery(query)
		if err != nil {
			return err
		}

		ulc = usrs.AddQuery(ulc, formattedQuery)
	}

	if orderBy != "" {
		ob := strings.ToLower(orderBy)
		ok := cmn.SliceContainsStr(usrs.ValidOrderByStrs, ob)
		if !ok {
			err = fmt.Errorf("gmin: error - %v is not a valid order by field", orderBy)
			return err
		}

		validOrderBy = ob

		if ob != "email" {
			validOrderBy, err = cmn.IsValidAttr(ob, usrs.UserAttrMap)
			if err != nil {
				return err
			}
		}

		ulc = usrs.AddOrderBy(ulc, validOrderBy)

		if sortOrder != "" {
			so := strings.ToLower(sortOrder)
			validSortOrder, err := cmn.IsValidAttr(so, cmn.ValidSortOrders)
			if err != nil {
				return err
			}

			ulc = usrs.AddSortOrder(ulc, validSortOrder)
		}
	}

	if viewType != "" {
		vt := strings.ToLower(viewType)
		ok := cmn.SliceContainsStr(usrs.ValidViewTypes, vt)
		if !ok {
			return fmt.Errorf("gmin: error - %v is not a valid view type", viewType)
		}

		listCall := usrs.AddViewType(ulc, vt)
		ulc = listCall.(*admin.UsersListCall)
	}

	ulc = usrs.AddMaxResults(ulc, maxResults)

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
	listUsersCmd.Flags().StringVarP(&orderBy, "orderby", "o", "", "field by which results will be ordered")
	listUsersCmd.Flags().StringVarP(&projection, "projection", "j", "", "type of projection")
	listUsersCmd.Flags().StringVarP(&query, "query", "q", "", "selection criteria to get users (separated by ~)")
	listUsersCmd.Flags().StringVarP(&sortOrder, "sortorder", "s", "", "sort order of returned results")
	listUsersCmd.Flags().StringVarP(&viewType, "viewtype", "v", "", "data view type")
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
