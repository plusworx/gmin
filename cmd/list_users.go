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
	lg "github.com/plusworx/gmin/utils/logging"
	usrs "github.com/plusworx/gmin/utils/users"
	"github.com/spf13/cobra"
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

	var (
		jsonData     []byte
		users        *admin.Users
		validOrderBy string
	)

	flgCustFldMaskVal, err := cmd.Flags().GetString(flgnm.FLG_CUSTFLDMASK)
	if err != nil {
		lg.Error(err)
		return err
	}

	flgProjectionVal, err := cmd.Flags().GetString(flgnm.FLG_PROJECTION)
	if err != nil {
		lg.Error(err)
		return err
	}
	if strings.ToLower(flgProjectionVal) == "custom" && flgCustFldMaskVal == "" {
		err := errors.New(gmess.ERR_NOCUSTOMFIELDMASK)
		lg.Error(err)
		return err
	}

	if flgCustFldMaskVal != "" && strings.ToLower(flgProjectionVal) != "custom" {
		err := errors.New(gmess.ERR_PROJECTIONFLAGNOTCUSTOM)
		lg.Error(err)
		return err
	}

	flgDeletedVal, err := cmd.Flags().GetBool(flgnm.FLG_DELETED)
	if err != nil {
		lg.Error(err)
		return err
	}

	flgQueryVal, err := cmd.Flags().GetString(flgnm.FLG_QUERY)
	if err != nil {
		lg.Error(err)
		return err
	}
	if flgQueryVal != "" && flgDeletedVal {
		err := errors.New(gmess.ERR_QUERYANDDELETEDFLAGS)
		lg.Error(err)
		return err
	}

	srv, err := cmn.CreateService(cmn.SRVTYPEADMIN, admin.AdminDirectoryUserReadonlyScope)
	if err != nil {
		return err
	}
	ds := srv.(*admin.Service)

	ulc := ds.Users.List()

	flgAttrsVal, err := cmd.Flags().GetString(flgnm.FLG_ATTRIBUTES)
	if err != nil {
		lg.Error(err)
		return err
	}
	if flgAttrsVal != "" {
		listAttrs, err := gpars.ParseOutputAttrs(flgAttrsVal, usrs.UserAttrMap)
		if err != nil {
			lg.Error(err)
			return err
		}
		formattedAttrs := usrs.STARTUSERSFIELD + listAttrs + usrs.ENDFIELD

		listCall := usrs.AddFields(ulc, formattedAttrs)
		ulc = listCall.(*admin.UsersListCall)
	}

	if flgDeletedVal {
		ulc = usrs.AddShowDeleted(ulc)
	}

	flgDomainVal, err := cmd.Flags().GetString(flgnm.FLG_DOMAIN)
	if err != nil {
		lg.Error(err)
		return err
	}
	if flgDomainVal != "" {
		ulc = usrs.AddDomain(ulc, flgDomainVal)
	} else {
		customerID, err := cfg.ReadConfigString(cfg.CONFIGCUSTID)
		if err != nil {
			lg.Error(err)
			return err
		}
		ulc = usrs.AddCustomer(ulc, customerID)
	}

	if flgProjectionVal != "" {
		proj := strings.ToLower(flgProjectionVal)
		ok := cmn.SliceContainsStr(usrs.ValidProjections, proj)
		if !ok {
			err = fmt.Errorf(gmess.ERR_INVALIDPROJECTIONTYPE, flgProjectionVal)
			lg.Error(err)
			return err
		}

		listCall := usrs.AddProjection(ulc, proj)
		ulc = listCall.(*admin.UsersListCall)

		if proj == "custom" {
			if flgCustFldMaskVal != "" {
				cFields := strings.Split(flgCustFldMaskVal, "~")
				mask := strings.Join(cFields, ",")
				listCall := usrs.AddCustomFieldMask(ulc, mask)
				ulc = listCall.(*admin.UsersListCall)
			} else {
				err = errors.New(gmess.ERR_NOCUSTOMFIELDMASK)
				lg.Error(err)
				return err
			}
		}
	}

	if flgQueryVal != "" {
		formattedQuery, err := gpars.ParseQuery(flgQueryVal, usrs.QueryAttrMap)
		if err != nil {
			lg.Error(err)
			return err
		}

		ulc = usrs.AddQuery(ulc, formattedQuery)
	}

	flgOrderByVal, err := cmd.Flags().GetString(flgnm.FLG_ORDERBY)
	if err != nil {
		lg.Error(err)
		return err
	}
	if flgOrderByVal != "" {
		ob := strings.ToLower(flgOrderByVal)
		ok := cmn.SliceContainsStr(usrs.ValidOrderByStrs, ob)
		if !ok {
			err = fmt.Errorf(gmess.ERR_INVALIDORDERBY, flgOrderByVal)
			lg.Error(err)
			return err
		}

		validOrderBy = ob

		if ob != "email" {
			validOrderBy, err = cmn.IsValidAttr(ob, usrs.UserAttrMap)
			if err != nil {
				lg.Error(err)
				return err
			}
		}

		ulc = usrs.AddOrderBy(ulc, validOrderBy)

		flgSrtOrdByVal, err := cmd.Flags().GetString(flgnm.FLG_SORTORDER)
		if err != nil {
			lg.Error(err)
			return err
		}
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

	flgViewTypeVal, err := cmd.Flags().GetString(flgnm.FLG_VIEWTYPE)
	if err != nil {
		lg.Error(err)
		return err
	}
	if flgViewTypeVal != "" {
		vt := strings.ToLower(flgViewTypeVal)
		ok := cmn.SliceContainsStr(usrs.ValidViewTypes, vt)
		if !ok {
			err = fmt.Errorf(gmess.ERR_INVALIDVIEWTYPE, flgViewTypeVal)
			lg.Error(err)
			return err
		}

		listCall := usrs.AddViewType(ulc, vt)
		ulc = listCall.(*admin.UsersListCall)
	}

	flgMaxResultsVal, err := cmd.Flags().GetInt64(flgnm.FLG_MAXRESULTS)
	if err != nil {
		lg.Error(err)
		return err
	}
	ulc = usrs.AddMaxResults(ulc, flgMaxResultsVal)

	users, err = usrs.DoList(ulc)
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
		err = doUserPages(ulc, users, flgPagesVal)
		if err != nil {
			lg.Error(err)
			return err
		}
	}

	jsonData, err = json.MarshalIndent(users, "", "    ")
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
		fmt.Println(len(users.Users))
	} else {
		fmt.Println(string(jsonData))
	}

	lg.Debug("finished doListUsers()")
	return nil
}

func doUserAllPages(ulc *admin.UsersListCall, users *admin.Users) error {
	lg.Debug("starting doUserAllPages()")

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

	lg.Debug("finished doUserAllPages()")
	return nil
}

func doUserNumPages(ulc *admin.UsersListCall, users *admin.Users, numPages int) error {
	lg.Debugw("starting doUserNumPages()",
		"numPages", numPages)

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

	lg.Debug("finished doUserNumPages()")
	return nil
}

func doUserPages(ulc *admin.UsersListCall, users *admin.Users, pages string) error {
	lg.Debugw("starting doUserPages()",
		"pages", pages)

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

	lg.Debug("finished doUserPages()")
	return nil
}

func init() {
	listCmd.AddCommand(listUsersCmd)

	listUsersCmd.Flags().StringVarP(&attrs, flgnm.FLG_ATTRIBUTES, "a", "", "required user attributes (separated by ~)")
	listUsersCmd.Flags().BoolVarP(&count, flgnm.FLG_COUNT, "", false, "count number of entities returned")
	listUsersCmd.Flags().StringVarP(&customField, flgnm.FLG_CUSTFLDMASK, "c", "", "custom field mask schemas (separated by ~)")
	listUsersCmd.Flags().StringVarP(&domain, flgnm.FLG_DOMAIN, "d", "", "domain from which to get users")
	listUsersCmd.Flags().Int64VarP(&maxResults, flgnm.FLG_MAXRESULTS, "m", 500, "maximum number of results to return per page")
	listUsersCmd.Flags().StringVarP(&orderBy, flgnm.FLG_ORDERBY, "o", "", "field by which results will be ordered")
	listUsersCmd.Flags().StringVarP(&pages, flgnm.FLG_PAGES, "p", "", "number of pages of results to be returned ('all' or a number)")
	listUsersCmd.Flags().StringVarP(&projection, flgnm.FLG_PROJECTION, "j", "", "type of projection")
	listUsersCmd.Flags().StringVarP(&query, flgnm.FLG_QUERY, "q", "", "selection criteria to get users (separated by ~)")
	listUsersCmd.Flags().StringVarP(&sortOrder, flgnm.FLG_SORTORDER, "s", "", "sort order of returned results")
	listUsersCmd.Flags().StringVarP(&viewType, flgnm.FLG_VIEWTYPE, "v", "", "data view type")
	listUsersCmd.Flags().BoolVarP(&deleted, flgnm.FLG_DELETED, "x", false, "show deleted users")

}
