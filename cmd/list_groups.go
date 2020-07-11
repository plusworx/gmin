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
	"strings"

	cmn "github.com/plusworx/gmin/utils/common"
	cfg "github.com/plusworx/gmin/utils/config"
	grps "github.com/plusworx/gmin/utils/groups"
	"github.com/spf13/cobra"
	admin "google.golang.org/api/admin/directory/v1"
)

var listGroupsCmd = &cobra.Command{
	Use:     "groups",
	Aliases: []string{"group", "grp", "grps"},
	Short:   "Outputs a list of groups",
	Long:    `Outputs a list of groups.`,
	RunE:    doListGroups,
}

func doListGroups(cmd *cobra.Command, args []string) error {
	var (
		formattedAttrs string
		groups         *admin.Groups
		validAttrs     []string
		validOrderBy   string
	)

	ds, err := cmn.CreateDirectoryService(admin.AdminDirectoryGroupReadonlyScope)
	if err != nil {
		return err
	}

	glc := ds.Groups.List()

	if attrs != "" {
		validAttrs, err = cmn.ValidateArgs(attrs, grps.GroupAttrMap, cmn.AttrStr)
		if err != nil {
			return err
		}

		formattedAttrs = grps.FormatAttrs(validAttrs, false)
		glc = grps.AddListFields(glc, formattedAttrs)
	}

	if domain != "" {
		glc = grps.AddListDomain(glc, domain)
	} else {
		customerID, err := cfg.ReadConfigString("customerid")
		if err != nil {
			return err
		}
		glc = grps.AddListCustomer(glc, customerID)
	}

	if query != "" {
		formattedQuery, err := processQuery(query)
		if err != nil {
			return err
		}

		glc = grps.AddListQuery(glc, formattedQuery)
	}

	if orderBy != "" {
		ob := strings.ToLower(orderBy)
		ok := cmn.SliceContainsStr(grps.ValidOrderByStrs, ob)
		if !ok {
			err = fmt.Errorf("gmin: error - %v is not a valid order by field", orderBy)
			return err
		}

		validOrderBy, err = cmn.IsValidAttr(ob, grps.GroupAttrMap)
		if err != nil {
			return err
		}

		glc = grps.AddListOrderBy(glc, validOrderBy)

		if sortOrder != "" {
			so := strings.ToLower(sortOrder)
			validSortOrder, err := cmn.IsValidAttr(so, cmn.ValidSortOrders)
			if err != nil {
				return err
			}

			glc = grps.AddListSortOrder(glc, validSortOrder)
		}
	}

	if userKey != "" {
		glc = grps.AddListUserKey(glc, userKey)
	}

	glc = grps.AddListMaxResults(glc, maxResults)

	groups, err = grps.DoList(glc)
	if err != nil {
		return err
	}

	jsonData, err := json.MarshalIndent(groups, "", "    ")
	if err != nil {
		return err
	}

	fmt.Println(string(jsonData))

	return nil
}

func init() {
	listCmd.AddCommand(listGroupsCmd)

	listGroupsCmd.Flags().StringVarP(&attrs, "attributes", "a", "", "required group attributes (separated by ~)")
	listGroupsCmd.Flags().StringVarP(&domain, "domain", "d", "", "domain from which to get groups")
	listGroupsCmd.Flags().Int64VarP(&maxResults, "maxresults", "m", 200, "maximum number of results to return")
	listGroupsCmd.Flags().StringVarP(&orderBy, "orderby", "o", "", "field by which results will be ordered")
	listGroupsCmd.Flags().StringVarP(&query, "query", "q", "", "selection criteria to get groups (separated by ~)")
	listGroupsCmd.Flags().StringVarP(&sortOrder, "sortorder", "s", "", "sort order of returned results")
	listGroupsCmd.Flags().StringVarP(&userKey, "userkey", "u", "", "email address or id of user who belongs to returned groups")
}

func processQuery(query string) (string, error) {
	var formattedQuery string

	validQuery, err := cmn.ValidateQuery(query, grps.QueryAttrMap)
	if err != nil {
		return "", err
	}

	formattedQuery, err = grps.FormatQuery(validQuery)
	if err != nil {
		return "", err
	}

	return formattedQuery, nil
}
