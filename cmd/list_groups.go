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
	flgnm "github.com/plusworx/gmin/utils/flagnames"
	gmess "github.com/plusworx/gmin/utils/gminmessages"
	gpars "github.com/plusworx/gmin/utils/gminparsers"
	grps "github.com/plusworx/gmin/utils/groups"
	lg "github.com/plusworx/gmin/utils/logging"
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
	lg.Debugw("starting doListGroups()",
		"args", args)

	var (
		groups       *admin.Groups
		jsonData     []byte
		validOrderBy string
	)

	ds, err := cmn.CreateDirectoryService(admin.AdminDirectoryGroupReadonlyScope)
	if err != nil {
		lg.Error(err)
		return err
	}

	glc := ds.Groups.List()

	flgAttrsVal, err := cmd.Flags().GetString(flgnm.FLG_ATTRIBUTES)
	if err != nil {
		lg.Error(err)
		return err
	}
	if flgAttrsVal != "" {
		listAttrs, err := gpars.ParseOutputAttrs(flgAttrsVal, grps.GroupAttrMap)
		if err != nil {
			lg.Error(err)
			return err
		}
		formattedAttrs := grps.STARTGROUPSFIELD + listAttrs + grps.ENDFIELD

		listCall := grps.AddFields(glc, formattedAttrs)
		glc = listCall.(*admin.GroupsListCall)
	}

	flgDomainVal, err := cmd.Flags().GetString(flgnm.FLG_DOMAIN)
	if err != nil {
		lg.Error(err)
		return err
	}
	if flgDomainVal != "" {
		glc = grps.AddDomain(glc, flgDomainVal)
	} else {
		customerID, err := cfg.ReadConfigString(cfg.CONFIGCUSTID)
		if err != nil {
			lg.Error(err)
			return err
		}
		glc = grps.AddCustomer(glc, customerID)
	}

	flgQueryVal, err := cmd.Flags().GetString(flgnm.FLG_QUERY)
	if err != nil {
		lg.Error(err)
		return err
	}
	if flgQueryVal != "" {
		formattedQuery, err := gpars.ParseQuery(flgQueryVal, grps.QueryAttrMap)
		if err != nil {
			lg.Error(err)
			return err
		}

		glc = grps.AddQuery(glc, formattedQuery)
	}

	flgOrderByVal, err := cmd.Flags().GetString(flgnm.FLG_ORDERBY)
	if err != nil {
		lg.Error(err)
		return err
	}
	if flgOrderByVal != "" {
		ob := strings.ToLower(flgOrderByVal)
		ok := cmn.SliceContainsStr(grps.ValidOrderByStrs, ob)
		if !ok {
			err = fmt.Errorf(gmess.ERR_INVALIDORDERBY, flgOrderByVal)
			lg.Error(err)
			return err
		}

		validOrderBy, err = cmn.IsValidAttr(ob, grps.GroupAttrMap)
		if err != nil {
			lg.Error(err)
			return err
		}

		glc = grps.AddOrderBy(glc, validOrderBy)

		flgSrtOrdVal, err := cmd.Flags().GetString(flgnm.FLG_SORTORDER)
		if err != nil {
			lg.Error(err)
			return err
		}
		if flgSrtOrdVal != "" {
			so := strings.ToLower(flgSrtOrdVal)
			validSortOrder, err := cmn.IsValidAttr(so, cmn.ValidSortOrders)
			if err != nil {
				lg.Error(err)
				return err
			}

			glc = grps.AddSortOrder(glc, validSortOrder)
		}
	}

	flgUserKeyVal, err := cmd.Flags().GetString(flgnm.FLG_USERKEY)
	if err != nil {
		lg.Error(err)
		return err
	}
	if flgUserKeyVal != "" {
		if flgDomainVal != "" {
			glc = grps.AddUserKey(glc, flgUserKeyVal)
		} else {
			err = errors.New(gmess.ERR_NODOMAINWITHUSERKEY)
			lg.Error(err)
			return err
		}
	}

	flgMaxResultsVal, err := cmd.Flags().GetInt64(flgnm.FLG_MAXRESULTS)
	if err != nil {
		lg.Error(err)
		return err
	}
	glc = grps.AddMaxResults(glc, flgMaxResultsVal)

	groups, err = grps.DoList(glc)
	if err != nil {
		lg.Error(err)
		return err
	}

	flgPagesVal, err := cmd.Flags().GetString(flgnm.FLG_PAGES)
	if err != nil {
		lg.Error(err)
		return err
	}
	if flgPagesVal != "" {
		err = doGrpPages(glc, groups, flgPagesVal)
		if err != nil {
			lg.Error(err)
			return err
		}
	}

	jsonData, err = json.MarshalIndent(groups, "", "    ")
	if err != nil {
		lg.Error(err)
		return err
	}

	flgCountVal, err := cmd.Flags().GetBool(flgnm.FLG_COUNT)
	if err != nil {
		lg.Error(err)
		return err
	}
	if flgCountVal {
		fmt.Println(len(groups.Groups))
	} else {
		fmt.Println(string(jsonData))
	}

	lg.Debug("finished doListGroups()")
	return nil
}

func doGrpAllPages(glc *admin.GroupsListCall, groups *admin.Groups) error {
	lg.Debug("starting doGrpAllPages()")

	if groups.NextPageToken != "" {
		glc = grps.AddPageToken(glc, groups.NextPageToken)
		nxtGroups, err := grps.DoList(glc)
		if err != nil {
			lg.Error(err)
			return err
		}
		groups.Groups = append(groups.Groups, nxtGroups.Groups...)
		groups.Etag = nxtGroups.Etag
		groups.NextPageToken = nxtGroups.NextPageToken

		if nxtGroups.NextPageToken != "" {
			doGrpAllPages(glc, groups)
		}
	}

	lg.Debug("finished doGrpAllPages()")
	return nil
}

func doGrpNumPages(glc *admin.GroupsListCall, groups *admin.Groups, numPages int) error {
	lg.Debugw("starting doGrpNumPages()",
		"numPages", numPages)

	if groups.NextPageToken != "" && numPages > 0 {
		glc = grps.AddPageToken(glc, groups.NextPageToken)
		nxtGroups, err := grps.DoList(glc)
		if err != nil {
			lg.Error(err)
			return err
		}
		groups.Groups = append(groups.Groups, nxtGroups.Groups...)
		groups.Etag = nxtGroups.Etag
		groups.NextPageToken = nxtGroups.NextPageToken

		if nxtGroups.NextPageToken != "" {
			doGrpNumPages(glc, groups, numPages-1)
		}
	}

	lg.Debug("finished doGrpNumPages()")
	return nil
}

func doGrpPages(glc *admin.GroupsListCall, groups *admin.Groups, pages string) error {
	lg.Debugw("starting doGrpPages()",
		"pages", pages)

	if pages == "all" {
		err := doGrpAllPages(glc, groups)
		if err != nil {
			lg.Error(err)
			return err
		}
	} else {
		numPages, err := strconv.Atoi(pages)
		if err != nil {
			err = errors.New(gmess.ERR_INVALIDPAGESARGUMENT)
			lg.Error(err)
			return err
		}

		if numPages > 1 {
			err = doGrpNumPages(glc, groups, numPages-1)
			if err != nil {
				lg.Error(err)
				return err
			}
		}
	}

	lg.Debug("finished doGrpPages()")
	return nil
}

func init() {
	listCmd.AddCommand(listGroupsCmd)

	listGroupsCmd.Flags().StringVarP(&attrs, flgnm.FLG_ATTRIBUTES, "a", "", "required group attributes (separated by ~)")
	listGroupsCmd.Flags().BoolVarP(&count, flgnm.FLG_COUNT, "", false, "count number of entities returned")
	listGroupsCmd.Flags().StringVarP(&domain, flgnm.FLG_DOMAIN, "d", "", "domain from which to get groups")
	listGroupsCmd.Flags().Int64VarP(&maxResults, flgnm.FLG_MAXRESULTS, "m", 200, "maximum number of results to return per page")
	listGroupsCmd.Flags().StringVarP(&orderBy, flgnm.FLG_ORDERBY, "o", "", "field by which results will be ordered")
	listGroupsCmd.Flags().StringVarP(&pages, flgnm.FLG_PAGES, "p", "", "number of pages of results to be returned ('all' or a number)")
	listGroupsCmd.Flags().StringVarP(&query, flgnm.FLG_QUERY, "q", "", "selection criteria to get groups (separated by ~)")
	listGroupsCmd.Flags().StringVarP(&sortOrder, flgnm.FLG_SORTORDER, "s", "", "sort order of returned results")
	listGroupsCmd.Flags().StringVarP(&userKey, flgnm.FLG_USERKEY, "u", "", "email address or id of user who belongs to returned groups")
}
