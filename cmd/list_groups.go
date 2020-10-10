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

	cmn "github.com/plusworx/gmin/utils/common"
	cfg "github.com/plusworx/gmin/utils/config"
	gmess "github.com/plusworx/gmin/utils/gminmessages"
	grps "github.com/plusworx/gmin/utils/groups"
	"github.com/spf13/cobra"
	admin "google.golang.org/api/admin/directory/v1"
)

var listGroupsCmd = &cobra.Command{
	Use:     "groups",
	Aliases: []string{"group", "grp", "grps"},
	Args:    cobra.NoArgs,
	Example: `gmin list groups -a email~description~id
gmin ls grp -q email=mygroup@domain.com`,
	Short: "Outputs a list of groups",
	Long:  `Outputs a list of groups.`,
	RunE:  doListGroups,
}

func doListGroups(cmd *cobra.Command, args []string) error {
	logger.Debugw("starting doListGroups()",
		"args", args)

	var (
		groups       *admin.Groups
		jsonData     []byte
		validOrderBy string
	)

	ds, err := cmn.CreateDirectoryService(admin.AdminDirectoryGroupReadonlyScope)
	if err != nil {
		logger.Error(err)
		return err
	}

	glc := ds.Groups.List()

	if attrs != "" {
		listAttrs, err := cmn.ParseOutputAttrs(attrs, grps.GroupAttrMap)
		if err != nil {
			logger.Error(err)
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
			logger.Error(err)
			return err
		}
		glc = grps.AddCustomer(glc, customerID)
	}

	if query != "" {
		formattedQuery, err := cmn.ParseQuery(query, grps.QueryAttrMap)
		if err != nil {
			logger.Error(err)
			return err
		}

		glc = grps.AddQuery(glc, formattedQuery)
	}

	if orderBy != "" {
		ob := strings.ToLower(orderBy)
		ok := cmn.SliceContainsStr(grps.ValidOrderByStrs, ob)
		if !ok {
			err = fmt.Errorf(gmess.ErrInvalidOrderBy, orderBy)
			logger.Error(err)
			return err
		}

		validOrderBy, err = cmn.IsValidAttr(ob, grps.GroupAttrMap)
		if err != nil {
			logger.Error(err)
			return err
		}

		glc = grps.AddOrderBy(glc, validOrderBy)

		if sortOrder != "" {
			so := strings.ToLower(sortOrder)
			validSortOrder, err := cmn.IsValidAttr(so, cmn.ValidSortOrders)
			if err != nil {
				logger.Error(err)
				return err
			}

			glc = grps.AddSortOrder(glc, validSortOrder)
		}
	}

	if userKey != "" {
		if domain != "" {
			glc = grps.AddUserKey(glc, userKey)
		} else {
			err = errors.New(gmess.ErrNoDomainWithUserKey)
			logger.Error(err)
			return err
		}
	}

	glc = grps.AddMaxResults(glc, maxResults)

	groups, err = grps.DoList(glc)
	if err != nil {
		logger.Error(err)
		return err
	}

	if pages != "" {
		err = doGrpPages(glc, groups, pages)
		if err != nil {
			logger.Error(err)
			return err
		}
	}

	jsonData, err = json.MarshalIndent(groups, "", "    ")
	if err != nil {
		logger.Error(err)
		return err
	}

	if count {
		fmt.Println(len(groups.Groups))
	} else {
		fmt.Println(string(jsonData))
	}

	logger.Debug("finished doListGroups()")
	return nil
}

func doGrpAllPages(glc *admin.GroupsListCall, groups *admin.Groups) error {
	logger.Debug("starting doGrpAllPages()")

	if groups.NextPageToken != "" {
		glc = grps.AddPageToken(glc, groups.NextPageToken)
		nxtGroups, err := grps.DoList(glc)
		if err != nil {
			logger.Error(err)
			return err
		}
		groups.Groups = append(groups.Groups, nxtGroups.Groups...)
		groups.Etag = nxtGroups.Etag
		groups.NextPageToken = nxtGroups.NextPageToken

		if nxtGroups.NextPageToken != "" {
			doGrpAllPages(glc, groups)
		}
	}

	logger.Debug("finished doGrpAllPages()")
	return nil
}

func doGrpNumPages(glc *admin.GroupsListCall, groups *admin.Groups, numPages int) error {
	logger.Debugw("starting doGrpNumPages()",
		"numPages", numPages)

	if groups.NextPageToken != "" && numPages > 0 {
		glc = grps.AddPageToken(glc, groups.NextPageToken)
		nxtGroups, err := grps.DoList(glc)
		if err != nil {
			logger.Error(err)
			return err
		}
		groups.Groups = append(groups.Groups, nxtGroups.Groups...)
		groups.Etag = nxtGroups.Etag
		groups.NextPageToken = nxtGroups.NextPageToken

		if nxtGroups.NextPageToken != "" {
			doGrpNumPages(glc, groups, numPages-1)
		}
	}

	logger.Debug("finished doGrpNumPages()")
	return nil
}

func doGrpPages(glc *admin.GroupsListCall, groups *admin.Groups, pages string) error {
	logger.Debugw("starting doGrpPages()",
		"pages", pages)

	if pages == "all" {
		err := doGrpAllPages(glc, groups)
		if err != nil {
			logger.Error(err)
			return err
		}
	} else {
		numPages, err := strconv.Atoi(pages)
		if err != nil {
			err = errors.New(gmess.ErrInvalidPagesArgument)
			logger.Error(err)
			return err
		}

		if numPages > 1 {
			err = doGrpNumPages(glc, groups, numPages-1)
			if err != nil {
				logger.Error(err)
				return err
			}
		}
	}

	logger.Debug("finished doGrpPages()")
	return nil
}

func init() {
	listCmd.AddCommand(listGroupsCmd)

	listGroupsCmd.Flags().StringVarP(&attrs, "attributes", "a", "", "required group attributes (separated by ~)")
	listGroupsCmd.Flags().BoolVarP(&count, "count", "", false, "count number of entities returned")
	listGroupsCmd.Flags().StringVarP(&domain, "domain", "d", "", "domain from which to get groups")
	listGroupsCmd.Flags().Int64VarP(&maxResults, "max-results", "m", 200, "maximum number of results to return per page")
	listGroupsCmd.Flags().StringVarP(&orderBy, "order-by", "o", "", "field by which results will be ordered")
	listGroupsCmd.Flags().StringVarP(&pages, "pages", "p", "", "number of pages of results to be returned ('all' or a number)")
	listGroupsCmd.Flags().StringVarP(&query, "query", "q", "", "selection criteria to get groups (separated by ~)")
	listGroupsCmd.Flags().StringVarP(&sortOrder, "sort-order", "s", "", "sort order of returned results")
	listGroupsCmd.Flags().StringVarP(&userKey, "user-key", "u", "", "email address or id of user who belongs to returned groups")
}
