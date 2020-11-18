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
	gd "github.com/plusworx/gmin/utils/gendatastructs"
	gmess "github.com/plusworx/gmin/utils/gminmessages"
	gpars "github.com/plusworx/gmin/utils/gminparsers"
	grps "github.com/plusworx/gmin/utils/groups"
	lg "github.com/plusworx/gmin/utils/logging"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
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
	defer lg.Debug("finished doListGroups()")

	var (
		flagsPassed []string
		groups      *admin.Groups
		jsonData    []byte
	)

	flagValueMap := map[string]interface{}{}

	// Collect names of command flags passed in
	cmd.Flags().Visit(func(f *pflag.Flag) {
		flagsPassed = append(flagsPassed, f.Name)
	})

	// Populate flag value map
	for _, flg := range flagsPassed {
		val, err := grps.GetFlagVal(cmd, flg)
		if err != nil {
			return err
		}
		flagValueMap[flg] = val
	}

	srv, err := cmn.CreateService(cmn.SRVTYPEADMIN, admin.AdminDirectoryGroupReadonlyScope)
	if err != nil {
		return err
	}
	ds := srv.(*admin.Service)

	glc := ds.Groups.List()

	err = lstGrpProcessFlags(glc, flagValueMap)
	if err != nil {
		return err
	}

	// If maxresults flag wasn't passed in then use the default value
	_, maxResultsPresent := flagValueMap[flgnm.FLG_MAXRESULTS]
	if !maxResultsPresent {
		err := lstGrpMaxResults(glc, int64(200))
		if err != nil {
			return err
		}
	}

	groups, err = grps.DoList(glc)
	if err != nil {
		return err
	}

	err = lstGrpPages(glc, groups, flagValueMap)
	if err != nil {
		lg.Error(err)
		return err
	}

	flgCountVal, countPresent := flagValueMap[flgnm.FLG_COUNT]
	if countPresent {
		countVal := flgCountVal.(bool)
		if countVal {
			fmt.Println(len(groups.Groups))
			return nil
		}
	}

	jsonData, err = json.MarshalIndent(groups, "", "    ")
	if err != nil {
		lg.Error(err)
		return err
	}
	fmt.Println(string(jsonData))

	return nil
}

func doGrpAllPages(glc *admin.GroupsListCall, groups *admin.Groups) error {
	lg.Debug("starting doGrpAllPages()")
	defer lg.Debug("finished doGrpAllPages()")

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
	lg.Debugw("starting doGrpNumPages()",
		"numPages", numPages)
	defer lg.Debug("finished doGrpNumPages()")

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
	lg.Debugw("starting doGrpPages()",
		"pages", pages)
	defer lg.Debug("finished doGrpPages()")

	if pages == "all" {
		err := doGrpAllPages(glc, groups)
		if err != nil {
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
				return err
			}
		}
	}

	return nil
}

func init() {
	listCmd.AddCommand(listGroupsCmd)

	listGroupsCmd.Flags().StringP(flgnm.FLG_ATTRIBUTES, "a", "", "required group attributes (separated by ~)")
	listGroupsCmd.Flags().Bool(flgnm.FLG_COUNT, false, "count number of entities returned")
	listGroupsCmd.Flags().StringP(flgnm.FLG_DOMAIN, "d", "", "domain from which to get groups")
	listGroupsCmd.Flags().Int64P(flgnm.FLG_MAXRESULTS, "m", 200, "maximum number of results to return per page")
	listGroupsCmd.Flags().StringP(flgnm.FLG_ORDERBY, "o", "", "field by which results will be ordered")
	listGroupsCmd.Flags().StringP(flgnm.FLG_PAGES, "p", "", "number of pages of results to be returned ('all' or a number)")
	listGroupsCmd.Flags().StringP(flgnm.FLG_QUERY, "q", "", "selection criteria to get groups (separated by ~)")
	listGroupsCmd.Flags().StringP(flgnm.FLG_SORTORDER, "s", "", "sort order of returned results")
	listGroupsCmd.Flags().StringP(flgnm.FLG_USERKEY, "u", "", "email address or id of user who belongs to returned groups")
}

func lstGrpAttributes(glc *admin.GroupsListCall, flagVal interface{}) error {
	lg.Debug("starting lstGrpAttributes()")
	defer lg.Debug("finished lstGrpAttributes()")

	attrsVal := fmt.Sprintf("%v", flagVal)
	if attrsVal != "" {
		listAttrs, err := gpars.ParseOutputAttrs(attrsVal, grps.GroupAttrMap)
		if err != nil {
			lg.Error(err)
			return err
		}
		formattedAttrs := grps.STARTGROUPSFIELD + listAttrs + grps.ENDFIELD

		listCall := grps.AddFields(glc, formattedAttrs)
		glc = listCall.(*admin.GroupsListCall)
	}
	return nil
}

func lstGrpDomain(glc *admin.GroupsListCall, flagVal interface{}, userKey string) error {
	lg.Debug("starting lstGrpDomain()")
	defer lg.Debug("finished lstGrpDomain()")

	domVal := fmt.Sprintf("%v", flagVal)

	if domVal == "" {
		return fmt.Errorf(gmess.ERR_EMPTYSTRING, flgnm.FLG_DOMAIN)
	}

	glc = grps.AddDomain(glc, domVal)

	if userKey != "" {
		glc = grps.AddUserKey(glc, userKey)
	}

	return nil
}

func lstGrpDomainUserKey(flgValMap map[string]interface{}) error {
	lg.Debug("starting lstGrpDomainUserKey()")
	defer lg.Debug("finished lstGrpDomainUserKey()")

	_, keyPresent := flgValMap[flgnm.FLG_USERKEY]
	if !keyPresent {
		return nil
	}

	flgDomainVal, domainPresent := flgValMap[flgnm.FLG_DOMAIN]
	if domainPresent {
		domVal := fmt.Sprintf("%v", flgDomainVal)
		if domVal != "" {
			return nil
		}
	}

	err := errors.New(gmess.ERR_NODOMAINWITHUSERKEY)
	lg.Error(err)
	return err
}

func lstGrpMaxResults(glc *admin.GroupsListCall, flagVal interface{}) error {
	lg.Debug("starting lstGrpMaxResults()")
	defer lg.Debug("finished lstGrpMaxResults()")

	flgMaxResultsVal := flagVal.(int64)

	glc = grps.AddMaxResults(glc, flgMaxResultsVal)
	return nil
}

func lstGrpOrderBy(glc *admin.GroupsListCall, inData interface{}) error {
	lg.Debug("starting lstGrpOrderBy()")
	defer lg.Debug("finished lstGrpOrderBy()")

	var (
		err          error
		validOrderBy string
	)

	inStruct := inData.(gd.TwoStrStruct)

	orderVal := inStruct.Element1
	if orderVal != "" {
		ob := strings.ToLower(orderVal)
		ok := cmn.SliceContainsStr(grps.ValidOrderByStrs, ob)
		if !ok {
			err := fmt.Errorf(gmess.ERR_INVALIDORDERBY, orderVal)
			lg.Error(err)
			return err
		}

		validOrderBy, err = cmn.IsValidAttr(ob, grps.GroupAttrMap)
		if err != nil {
			lg.Error(err)
			return err
		}
		glc = grps.AddOrderBy(glc, validOrderBy)

		flgSrtOrdByVal := inStruct.Element2
		if flgSrtOrdByVal != "" {
			so := strings.ToLower(flgSrtOrdByVal)
			validSortOrder, err := cmn.IsValidAttr(so, cmn.ValidSortOrders)
			if err != nil {
				lg.Error(err)
				return err
			}

			glc = grps.AddSortOrder(glc, validSortOrder)
		}
	}
	return nil
}

func lstGrpPages(glc *admin.GroupsListCall, groups *admin.Groups, flgValMap map[string]interface{}) error {
	lg.Debug("starting lstGrpPages()")
	defer lg.Debug("finished lstGrpPages()")

	flgPagesVal, pagesPresent := flgValMap[flgnm.FLG_PAGES]
	if !pagesPresent {
		return nil
	}
	if flgPagesVal != "" {
		pagesVal := fmt.Sprintf("%v", flgPagesVal)
		err := doGrpPages(glc, groups, pagesVal)
		if err != nil {
			lg.Error(err)
			return err
		}
	}
	return nil
}

func lstGrpProcessFlags(glc *admin.GroupsListCall, flgValMap map[string]interface{}) error {
	lg.Debug("starting lstGrpProcessFlags()")
	defer lg.Debug("finished lstGrpProcessFlags()")

	var keyVal string

	lstGrpFuncMap := map[string]func(*admin.GroupsListCall, interface{}) error{
		flgnm.FLG_ATTRIBUTES: lstGrpAttributes,
		flgnm.FLG_MAXRESULTS: lstGrpMaxResults,
		flgnm.FLG_ORDERBY:    lstGrpOrderBy,
		flgnm.FLG_QUERY:      lstGrpQuery,
	}

	// Get customer id if the domain flag is not present
	_, domPresent := flgValMap[flgnm.FLG_DOMAIN]
	if !domPresent {
		customerID, err := cfg.ReadConfigString(cfg.CONFIGCUSTID)
		if err != nil {
			lg.Error(err)
			return err
		}
		glc = grps.AddCustomer(glc, customerID)
	}

	// Check that domain present if user key has been supplied
	err := lstGrpDomainUserKey(flgValMap)
	if err != nil {
		return err
	}

	// Cycle through flags that build the ulc excluding pages and count
	for key, val := range flgValMap {
		// Do domain flag specific processing
		if key == flgnm.FLG_DOMAIN {
			flgUserKeyVal, usrKeyPresent := flgValMap[flgnm.FLG_USERKEY]
			if usrKeyPresent {
				keyVal = fmt.Sprintf("%v", flgUserKeyVal)
			} else {
				keyVal = ""
			}
			err := lstGrpDomain(glc, val, keyVal)
			if err != nil {
				return err
			}
			continue
		}
		// Order by has dependent sort order so deal with that
		if key == flgnm.FLG_ORDERBY {
			retStruct, err := lstGrpSortOrderBy(flgValMap)
			if err != nil {
				return err
			}
			val = retStruct
		}

		lgf, ok := lstGrpFuncMap[key]
		if !ok {
			continue
		}
		err := lgf(glc, val)
		if err != nil {
			return err
		}
	}
	return nil
}

func lstGrpQuery(glc *admin.GroupsListCall, flagVal interface{}) error {
	lg.Debug("starting lstGrpQuery()")
	defer lg.Debug("finished lstGrpQuery()")

	qryVal := fmt.Sprintf("%v", flagVal)
	if qryVal != "" {
		formattedQuery, err := gpars.ParseQuery(qryVal, grps.QueryAttrMap)
		if err != nil {
			lg.Error(err)
			return err
		}

		glc = grps.AddQuery(glc, formattedQuery)
	}
	return nil
}

func lstGrpSortOrderBy(flgValMap map[string]interface{}) (gd.TwoStrStruct, error) {
	lg.Debug("starting lstGrpSortOrderBy()")
	defer lg.Debug("finished lstGrpSortOrderBy()")

	outData := gd.TwoStrStruct{}

	orderByVal := flgValMap[flgnm.FLG_ORDERBY]
	sortOrderVal, sortOrderPresent := flgValMap[flgnm.FLG_SORTORDER]

	outData.Element1 = fmt.Sprintf("%v", orderByVal)
	if sortOrderPresent {
		outData.Element2 = fmt.Sprintf("%v", sortOrderVal)
	} else {
		outData.Element2 = ""
	}

	return outData, nil
}
