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
	lg "github.com/plusworx/gmin/utils/logging"
	usrs "github.com/plusworx/gmin/utils/users"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	admin "google.golang.org/api/admin/directory/v1"
)

var listUsersCmd = &cobra.Command{
	Use:     "users",
	Aliases: []string{"user", "usrs", "usr"},
	Args:    cobra.NoArgs,
	Example: `gmin list users -a primaryemail~addresses
gmin ls user -q name:Fred`,
	Short: "Outputs a list of users",
	Long:  `Outputs a list of users.`,
	RunE:  doListUsers,
}

func doListUsers(cmd *cobra.Command, args []string) error {
	lg.Debugw("starting doListUsers()",
		"args", args)
	defer lg.Debug("finished doListUsers()")

	var (
		flagsPassed []string
		jsonData    []byte
		users       *admin.Users
	)

	flagValueMap := map[string]interface{}{}

	// Collect names of command flags passed in
	cmd.Flags().Visit(func(f *pflag.Flag) {
		flagsPassed = append(flagsPassed, f.Name)
	})

	// Populate flag value map
	for _, flg := range flagsPassed {
		val, err := usrs.GetFlagVal(cmd, flg)
		if err != nil {
			return err
		}
		flagValueMap[flg] = val
	}

	srv, err := cmn.CreateService(cmn.SRVTYPEADMIN, admin.AdminDirectoryUserReadonlyScope)
	if err != nil {
		return err
	}
	ds := srv.(*admin.Service)

	ulc := ds.Users.List()

	err = lstUsrProcessFlags(ulc, flagValueMap)
	if err != nil {
		return err
	}

	// If maxresults flag wasn't passed in then use the default value
	_, maxResultsPresent := flagValueMap[flgnm.FLG_MAXRESULTS]
	if !maxResultsPresent {
		err := lstUsrMaxResults(ulc, int64(500))
		if err != nil {
			return err
		}
	}

	users, err = usrs.DoList(ulc)
	if err != nil {
		lg.Error(err)
		return err
	}

	err = lstUsrPages(ulc, users, flagValueMap)
	if err != nil {
		lg.Error(err)
		return err
	}

	flgCountVal, countPresent := flagValueMap[flgnm.FLG_COUNT]
	if countPresent {
		countVal := flgCountVal.(bool)
		if countVal {
			fmt.Println(len(users.Users))
			return nil
		}
	}

	jsonData, err = json.MarshalIndent(users, "", "    ")
	if err != nil {
		lg.Error(err)
		return err
	}
	fmt.Println(string(jsonData))

	return nil
}

func doUserAllPages(ulc *admin.UsersListCall, users *admin.Users) error {
	lg.Debug("starting doUserAllPages()")
	defer lg.Debug("finished doUserAllPages()")

	if users.NextPageToken != "" {
		ulc = usrs.AddPageToken(ulc, users.NextPageToken)
		nxtUsers, err := usrs.DoList(ulc)
		if err != nil {
			lg.Error(err)
			return err
		}
		users.Users = append(users.Users, nxtUsers.Users...)
		users.Etag = nxtUsers.Etag
		users.NextPageToken = nxtUsers.NextPageToken

		if nxtUsers.NextPageToken != "" {
			doUserAllPages(ulc, users)
		}
	}

	return nil
}

func doUserNumPages(ulc *admin.UsersListCall, users *admin.Users, numPages int) error {
	lg.Debugw("starting doUserNumPages()",
		"numPages", numPages)
	defer lg.Debug("finished doUserNumPages()")

	if users.NextPageToken != "" && numPages > 0 {
		ulc = usrs.AddPageToken(ulc, users.NextPageToken)
		nxtUsers, err := usrs.DoList(ulc)
		if err != nil {
			lg.Error(err)
			return err
		}
		users.Users = append(users.Users, nxtUsers.Users...)
		users.Etag = nxtUsers.Etag
		users.NextPageToken = nxtUsers.NextPageToken

		if nxtUsers.NextPageToken != "" {
			doUserNumPages(ulc, users, numPages-1)
		}
	}

	return nil
}

func doUserPages(ulc *admin.UsersListCall, users *admin.Users, pages string) error {
	lg.Debugw("starting doUserPages()",
		"pages", pages)
	defer lg.Debug("finished doUserPages()")

	if pages == "all" {
		err := doUserAllPages(ulc, users)
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
			err = doUserNumPages(ulc, users, numPages-1)
			if err != nil {
				lg.Error(err)
				return err
			}
		}
	}

	return nil
}

func init() {
	listCmd.AddCommand(listUsersCmd)

	listUsersCmd.Flags().StringP(flgnm.FLG_ATTRIBUTES, "a", "", "required user attributes (separated by ~)")
	listUsersCmd.Flags().Bool(flgnm.FLG_COUNT, false, "count number of entities returned")
	listUsersCmd.Flags().StringP(flgnm.FLG_CUSTFLDMASK, "c", "", "custom field mask schemas (separated by ~)")
	listUsersCmd.Flags().StringP(flgnm.FLG_DOMAIN, "d", "", "domain from which to get users")
	listUsersCmd.Flags().Int64P(flgnm.FLG_MAXRESULTS, "m", 500, "maximum number of results to return per page")
	listUsersCmd.Flags().StringP(flgnm.FLG_ORDERBY, "o", "", "field by which results will be ordered")
	listUsersCmd.Flags().StringP(flgnm.FLG_PAGES, "p", "", "number of pages of results to be returned ('all' or a number)")
	listUsersCmd.Flags().StringP(flgnm.FLG_PROJECTION, "j", "", "type of projection")
	listUsersCmd.Flags().StringP(flgnm.FLG_QUERY, "q", "", "selection criteria to get users (separated by ~)")
	listUsersCmd.Flags().StringP(flgnm.FLG_SORTORDER, "s", "", "sort order of returned results")
	listUsersCmd.Flags().StringP(flgnm.FLG_VIEWTYPE, "v", "", "data view type")
	listUsersCmd.Flags().BoolP(flgnm.FLG_DELETED, "x", false, "show deleted users")

}

func lstUsrAttributes(ulc *admin.UsersListCall, flagVal interface{}) error {
	lg.Debug("starting lstUsrAttributes()")
	defer lg.Debug("finished lstUsrAttributes()")

	attrsVal := fmt.Sprintf("%v", flagVal)
	if attrsVal != "" {
		listAttrs, err := gpars.ParseOutputAttrs(attrsVal, usrs.UserAttrMap)
		if err != nil {
			lg.Error(err)
			return err
		}
		formattedAttrs := usrs.STARTUSERSFIELD + listAttrs + usrs.ENDFIELD

		listCall := usrs.AddFields(ulc, formattedAttrs)
		ulc = listCall.(*admin.UsersListCall)
	}
	return nil
}

func lstUsrCustomProjection(flgValMap map[string]interface{}) (gd.TwoStrStruct, error) {
	lg.Debug("starting lstUsrCustomProjection()")
	defer lg.Debug("finished lstUsrCustomProjection()")

	var retMaskVal string

	projVal := flgValMap[flgnm.FLG_PROJECTION]
	custFldMaskVal, maskPresent := flgValMap[flgnm.FLG_CUSTFLDMASK]

	lowerProjVal := strings.ToLower(fmt.Sprintf("%v", projVal))

	if lowerProjVal == "custom" && !maskPresent {
		err := errors.New(gmess.ERR_NOCUSTOMFIELDMASK)
		lg.Error(err)
		return gd.TwoStrStruct{}, err
	}

	if maskPresent && lowerProjVal != "custom" {
		err := errors.New(gmess.ERR_PROJECTIONFLAGNOTCUSTOM)
		lg.Error(err)
		return gd.TwoStrStruct{}, err
	}

	if maskPresent {
		retMaskVal = fmt.Sprintf("%v", custFldMaskVal)
	} else {
		retMaskVal = ""
	}

	retStruct := gd.TwoStrStruct{Element1: lowerProjVal, Element2: retMaskVal}
	return retStruct, nil
}

func lstUsrDeleted(ulc *admin.UsersListCall, flagVal interface{}) error {
	lg.Debug("starting lstUsrDeleted()")
	defer lg.Debug("finished lstUsrDeleted()")

	delVal := flagVal.(bool)
	if delVal {
		ulc = usrs.AddShowDeleted(ulc)
	}
	return nil
}

func lstUsrDeletedQuery(flgValMap map[string]interface{}) error {
	lg.Debug("starting lstUsrDeletedQuery()")
	defer lg.Debug("finished lstUsrDeletedQuery()")

	flgDeletedVal, deletedPresent := flgValMap[flgnm.FLG_DELETED]
	flgQueryVal, queryPresent := flgValMap[flgnm.FLG_QUERY]

	if !deletedPresent || !queryPresent {
		return nil
	}

	dVal := flgDeletedVal.(bool)
	qVal := fmt.Sprintf("%v", flgQueryVal)

	if qVal != "" && dVal {
		err := errors.New(gmess.ERR_QUERYANDDELETEDFLAGS)
		lg.Error(err)
		return err
	}

	return nil
}

func lstUsrDomain(ulc *admin.UsersListCall, flagVal interface{}) error {
	lg.Debug("starting lstUsrDomain()")
	defer lg.Debug("finished lstUsrDomain()")

	domVal := fmt.Sprintf("%v", flagVal)

	if domVal == "" {
		return fmt.Errorf(gmess.ERR_EMPTYSTRING, flgnm.FLG_DOMAIN)
	}

	ulc = usrs.AddDomain(ulc, domVal)
	return nil
}

func lstUsrMaxResults(ulc *admin.UsersListCall, flagVal interface{}) error {
	lg.Debug("starting lstUsrMaxResults()")
	defer lg.Debug("finished lstUsrMaxResults()")

	flgMaxResultsVal := flagVal.(int64)

	ulc = usrs.AddMaxResults(ulc, flgMaxResultsVal)
	return nil
}

func lstUsrOrderBy(ulc *admin.UsersListCall, inData interface{}) error {
	lg.Debug("starting lstUsrOrderBy()")
	defer lg.Debug("finished lstUsrOrderBy()")

	var (
		err          error
		validOrderBy string
	)

	inStruct := inData.(gd.TwoStrStruct)

	orderVal := inStruct.Element1
	if orderVal != "" {
		ob := strings.ToLower(orderVal)
		ok := cmn.SliceContainsStr(usrs.ValidOrderByStrs, ob)
		if !ok {
			err := fmt.Errorf(gmess.ERR_INVALIDORDERBY, orderVal)
			lg.Error(err)
			return err
		}

		if ob == "email" {
			validOrderBy = ob
		} else {
			validOrderBy, err = cmn.IsValidAttr(ob, usrs.UserAttrMap)
			if err != nil {
				lg.Error(err)
				return err
			}
		}

		ulc = usrs.AddOrderBy(ulc, validOrderBy)

		flgSrtOrdByVal := inStruct.Element2
		if flgSrtOrdByVal != "" {
			so := strings.ToLower(flgSrtOrdByVal)
			validSortOrder, err := cmn.IsValidAttr(so, cmn.ValidSortOrders)
			if err != nil {
				lg.Error(err)
				return err
			}

			ulc = usrs.AddSortOrder(ulc, validSortOrder)
		}
	}
	return nil
}

func lstUsrPages(ulc *admin.UsersListCall, users *admin.Users, flgValMap map[string]interface{}) error {
	lg.Debug("starting lstUsrPages()")
	defer lg.Debug("finished lstUsrPages()")

	flgPagesVal, pagesPresent := flgValMap[flgnm.FLG_PAGES]
	if !pagesPresent {
		return nil
	}
	if flgPagesVal != "" {
		pagesVal := fmt.Sprintf("%v", flgPagesVal)
		err := doUserPages(ulc, users, pagesVal)
		if err != nil {
			lg.Error(err)
			return err
		}
	}
	return nil
}

func lstUsrProcessFlags(ulc *admin.UsersListCall, flgValMap map[string]interface{}) error {
	lg.Debug("starting lstUsrProcessFlags()")
	defer lg.Debug("finished lstUsrProcessFlags()")

	lstUsrFuncMap := map[string]func(*admin.UsersListCall, interface{}) error{
		flgnm.FLG_ATTRIBUTES: lstUsrAttributes,
		flgnm.FLG_DELETED:    lstUsrDeleted,
		flgnm.FLG_DOMAIN:     lstUsrDomain,
		flgnm.FLG_MAXRESULTS: lstUsrMaxResults,
		flgnm.FLG_ORDERBY:    lstUsrOrderBy,
		flgnm.FLG_PROJECTION: lstUsrProjection,
		flgnm.FLG_QUERY:      lstUsrQuery,
		flgnm.FLG_VIEWTYPE:   lstUsrViewType,
	}

	// Get customer id if the domain flag is not present
	_, domPresent := flgValMap[flgnm.FLG_DOMAIN]
	if !domPresent {
		customerID, err := cfg.ReadConfigString(cfg.CONFIGCUSTID)
		if err != nil {
			lg.Error(err)
			return err
		}
		ulc = usrs.AddCustomer(ulc, customerID)
	}

	// Check that deleted and query flags are not both present
	err := lstUsrDeletedQuery(flgValMap)
	if err != nil {
		return err
	}

	// Cycle through flags that build the ulc excluding pages and count
	for key, val := range flgValMap {
		// Order by has dependent sort order so deal with that
		if key == flgnm.FLG_ORDERBY {
			retStruct, err := lstUsrSortOrderBy(flgValMap)
			if err != nil {
				return err
			}
			val = retStruct
		}
		// Projection has dependent custom field mask so deal with that
		if key == flgnm.FLG_PROJECTION {
			retStruct, err := lstUsrCustomProjection(flgValMap)
			if err != nil {
				return err
			}
			val = retStruct
		}

		luf, ok := lstUsrFuncMap[key]
		if !ok {
			continue
		}
		err := luf(ulc, val)
		if err != nil {
			return err
		}
	}
	return nil
}

func lstUsrProjection(ulc *admin.UsersListCall, inData interface{}) error {
	lg.Debug("starting lstUsrProjection()")
	defer lg.Debug("finished lstUsrProjection()")

	inStruct := inData.(gd.TwoStrStruct)
	projVal := inStruct.Element1

	if projVal != "" {
		ok := cmn.SliceContainsStr(usrs.ValidProjections, projVal)
		if !ok {
			err := fmt.Errorf(gmess.ERR_INVALIDPROJECTIONTYPE, projVal)
			lg.Error(err)
			return err
		}

		listCall := usrs.AddProjection(ulc, projVal)
		ulc = listCall.(*admin.UsersListCall)

		if projVal == "custom" {
			custVal := inStruct.Element2
			if custVal != "" {
				cFields := strings.Split(custVal, "~")
				mask := strings.Join(cFields, ",")
				listCall := usrs.AddCustomFieldMask(ulc, mask)
				ulc = listCall.(*admin.UsersListCall)
			} else {
				err := errors.New(gmess.ERR_NOCUSTOMFIELDMASK)
				lg.Error(err)
				return err
			}
		}
	}
	return nil
}

func lstUsrQuery(ulc *admin.UsersListCall, flagVal interface{}) error {
	lg.Debug("starting lstUsrQuery()")
	defer lg.Debug("finished lstUsrQuery()")

	qryVal := fmt.Sprintf("%v", flagVal)
	if qryVal != "" {
		formattedQuery, err := gpars.ParseQuery(qryVal, usrs.QueryAttrMap)
		if err != nil {
			lg.Error(err)
			return err
		}

		ulc = usrs.AddQuery(ulc, formattedQuery)
	}
	return nil
}

func lstUsrSortOrderBy(flgValMap map[string]interface{}) (gd.TwoStrStruct, error) {
	lg.Debug("starting lstUsrSortOrderBy()")
	defer lg.Debug("finished lstUsrSortOrderBy()")

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

func lstUsrViewType(ulc *admin.UsersListCall, flagVal interface{}) error {
	lg.Debug("starting lstUsrViewType()")
	defer lg.Debug("finished lstUsrViewType()")

	vtVal := fmt.Sprintf("%v", flagVal)
	if vtVal != "" {
		lowerVt := strings.ToLower(vtVal)
		ok := cmn.SliceContainsStr(usrs.ValidViewTypes, lowerVt)
		if !ok {
			err := fmt.Errorf(gmess.ERR_INVALIDVIEWTYPE, vtVal)
			lg.Error(err)
			return err
		}

		listCall := usrs.AddViewType(ulc, lowerVt)
		ulc = listCall.(*admin.UsersListCall)
	}
	return nil
}
