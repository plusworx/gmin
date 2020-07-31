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
	"strings"

	"github.com/jinzhu/copier"
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
		groups       *admin.Groups
		jsonData     []byte
		newGroups    = grps.GminGroups{}
		validOrderBy string
	)

	ds, err := cmn.CreateDirectoryService(admin.AdminDirectoryGroupReadonlyScope)
	if err != nil {
		return err
	}

	glc := ds.Groups.List()

	if attrs != "" {
		listAttrs, err := cmn.ParseOutputAttrs(attrs, grps.GroupAttrMap)
		if err != nil {
			return err
		}
		formattedAttrs := grps.StartGroupsField + listAttrs + grps.EndField

		listCall := grps.AddFields(glc, formattedAttrs)
		glc = listCall.(*admin.GroupsListCall)
	}

	if domain != "" {
		glc = grps.AddDomain(glc, domain)
	} else {
		customerID, err := cfg.ReadConfigString("customerid")
		if err != nil {
			return err
		}
		glc = grps.AddCustomer(glc, customerID)
	}

	if query != "" {
		formattedQuery, err := cmn.ParseQuery(query, grps.QueryAttrMap)
		if err != nil {
			return err
		}

		glc = grps.AddQuery(glc, formattedQuery)
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

		glc = grps.AddOrderBy(glc, validOrderBy)

		if sortOrder != "" {
			so := strings.ToLower(sortOrder)
			validSortOrder, err := cmn.IsValidAttr(so, cmn.ValidSortOrders)
			if err != nil {
				return err
			}

			glc = grps.AddSortOrder(glc, validSortOrder)
		}
	}

	if userKey != "" {
		if domain != "" {
			glc = grps.AddUserKey(glc, userKey)
		} else {
			return errors.New("gmin: error - you must provide a domain in addition to userkey")
		}
	}

	glc = grps.AddMaxResults(glc, maxResults)

	groups, err = grps.DoList(glc)
	if err != nil {
		return err
	}

	if pages != "" {
		err = doGrpPages(glc, groups, pages)
		if err != nil {
			return err
		}
	}

	if attrs == "" {
		copier.Copy(&newGroups, groups)

		jsonData, err = json.MarshalIndent(newGroups, "", "    ")
		if err != nil {
			return err
		}
	} else {
		jsonData, err = json.MarshalIndent(groups, "", "    ")
		if err != nil {
			return err
		}
	}

	if count {
		fmt.Println(len(groups.Groups))
	} else {
		fmt.Println(string(jsonData))
	}

	return nil
}

func doGrpAllPages(glc *admin.GroupsListCall, groups *admin.Groups) error {
	if groups.NextPageToken != "" {
		glc = grps.AddPageToken(glc, groups.NextPageToken)
		nxtGroups, err := grps.DoList(glc)
		if err != nil {
			return err
		}
		groups.Groups = append(groups.Groups, nxtGroups.Groups...)
		groups.Etag = nxtGroups.Etag
		groups.NextPageToken = nxtGroups.NextPageToken

		if nxtGroups.NextPageToken != "" {
			doGrpAllPages(glc, groups)
		}
	}

	return nil
}

func doGrpNumPages(glc *admin.GroupsListCall, groups *admin.Groups, numPages int) error {
	if groups.NextPageToken != "" && numPages > 0 {
		glc = grps.AddPageToken(glc, groups.NextPageToken)
		nxtGroups, err := grps.DoList(glc)
		if err != nil {
			return err
		}
		groups.Groups = append(groups.Groups, nxtGroups.Groups...)
		groups.Etag = nxtGroups.Etag
		groups.NextPageToken = nxtGroups.NextPageToken

		if nxtGroups.NextPageToken != "" {
			doGrpNumPages(glc, groups, numPages-1)
		}
	}

	return nil
}

func doGrpPages(glc *admin.GroupsListCall, groups *admin.Groups, pages string) error {
	if pages == "all" {
		err := doGrpAllPages(glc, groups)
		if err != nil {
			return err
		}
	} else {
		numPages, err := strconv.Atoi(pages)
		if err != nil {
			return errors.New("gmin: error - pages must be 'all' or a number")
		}

		if numPages > 1 {
			err = doGrpNumPages(glc, groups, numPages-1)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func init() {
	listCmd.AddCommand(listGroupsCmd)

	listGroupsCmd.Flags().StringVarP(&attrs, "attributes", "a", "", "required group attributes (separated by ~)")
	listGroupsCmd.Flags().BoolVarP(&count, "count", "", false, "count number of entities returned")
	listGroupsCmd.Flags().StringVarP(&domain, "domain", "d", "", "domain from which to get groups")
	listGroupsCmd.Flags().Int64VarP(&maxResults, "maxresults", "m", 200, "maximum number of results to return per page")
	listGroupsCmd.Flags().StringVarP(&orderBy, "orderby", "o", "", "field by which results will be ordered")
	listGroupsCmd.Flags().StringVarP(&pages, "pages", "p", "", "number of pages of results to be returned")
	listGroupsCmd.Flags().StringVarP(&query, "query", "q", "", "selection criteria to get groups (separated by ~)")
	listGroupsCmd.Flags().StringVarP(&sortOrder, "sortorder", "s", "", "sort order of returned results")
	listGroupsCmd.Flags().StringVarP(&userKey, "userkey", "u", "", "email address or id of user who belongs to returned groups")
}
