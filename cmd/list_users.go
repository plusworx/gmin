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
	usrs "github.com/plusworx/gmin/utils/users"
	"github.com/spf13/cobra"
	admin "google.golang.org/api/admin/directory/v1"
)

var listUsersCmd = &cobra.Command{
	Use:     "users",
	Aliases: []string{"user", "usr"},
	Args:    cobra.NoArgs,
	Example: `gmin list users -a primaryemail~addresses
gmin ls user -q name:Fred`,
	Short: "Outputs a list of users",
	Long:  `Outputs a list of users.`,
	RunE:  doListUsers,
}

func doListUsers(cmd *cobra.Command, args []string) error {
	logger.Debugw("starting doListUsers()",
		"args", args)

	var (
		jsonData     []byte
		users        *admin.Users
		validOrderBy string
	)

	if strings.ToLower(projection) == "custom" && customField == "" {
		err := errors.New(cmn.ErrNoCustomFieldMask)
		logger.Error(err)
		return err
	}

	if customField != "" && strings.ToLower(projection) != "custom" {
		err := errors.New(cmn.ErrProjectionFlagNotCustom)
		logger.Error(err)
		return err
	}

	if query != "" && deleted {
		err := errors.New(cmn.ErrQueryAndDeletedFlags)
		logger.Error(err)
		return err
	}

	ds, err := cmn.CreateDirectoryService(admin.AdminDirectoryUserReadonlyScope)
	if err != nil {
		logger.Error(err)
		return err
	}

	ulc := ds.Users.List()

	if attrs != "" {
		listAttrs, err := cmn.ParseOutputAttrs(attrs, usrs.UserAttrMap)
		if err != nil {
			logger.Error(err)
			return err
		}
		formattedAttrs := usrs.StartUsersField + listAttrs + usrs.EndField

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
			logger.Error(err)
			return err
		}
		ulc = usrs.AddCustomer(ulc, customerID)
	}

	if projection != "" {
		proj := strings.ToLower(projection)
		ok := cmn.SliceContainsStr(usrs.ValidProjections, proj)
		if !ok {
			err = fmt.Errorf(cmn.ErrInvalidProjectionType, projection)
			logger.Error(err)
			return err
		}

		listCall := usrs.AddProjection(ulc, proj)
		ulc = listCall.(*admin.UsersListCall)

		if proj == "custom" {
			if customField != "" {
				cFields := cmn.ParseTildeField(customField)
				mask := strings.Join(cFields, ",")
				listCall := usrs.AddCustomFieldMask(ulc, mask)
				ulc = listCall.(*admin.UsersListCall)
			} else {
				err = errors.New(cmn.ErrNoCustomFieldMask)
				logger.Error(err)
				return err
			}
		}
	}

	if query != "" {
		formattedQuery, err := cmn.ParseQuery(query, usrs.QueryAttrMap)
		if err != nil {
			logger.Error(err)
			return err
		}

		ulc = usrs.AddQuery(ulc, formattedQuery)
	}

	if orderBy != "" {
		ob := strings.ToLower(orderBy)
		ok := cmn.SliceContainsStr(usrs.ValidOrderByStrs, ob)
		if !ok {
			err = fmt.Errorf(cmn.ErrInvalidOrderBy, orderBy)
			logger.Error(err)
			return err
		}

		validOrderBy = ob

		if ob != "email" {
			validOrderBy, err = cmn.IsValidAttr(ob, usrs.UserAttrMap)
			if err != nil {
				logger.Error(err)
				return err
			}
		}

		ulc = usrs.AddOrderBy(ulc, validOrderBy)

		if sortOrder != "" {
			so := strings.ToLower(sortOrder)
			validSortOrder, err := cmn.IsValidAttr(so, cmn.ValidSortOrders)
			if err != nil {
				logger.Error(err)
				return err
			}

			ulc = usrs.AddSortOrder(ulc, validSortOrder)
		}
	}

	if viewType != "" {
		vt := strings.ToLower(viewType)
		ok := cmn.SliceContainsStr(usrs.ValidViewTypes, vt)
		if !ok {
			err = fmt.Errorf(cmn.ErrInvalidViewType, viewType)
			logger.Error(err)
			return err
		}

		listCall := usrs.AddViewType(ulc, vt)
		ulc = listCall.(*admin.UsersListCall)
	}

	ulc = usrs.AddMaxResults(ulc, maxResults)

	users, err = usrs.DoList(ulc)
	if err != nil {
		logger.Error(err)
		return err
	}

	if pages != "" {
		err = doUserPages(ulc, users, pages)
		if err != nil {
			logger.Error(err)
			return err
		}
	}

	jsonData, err = json.MarshalIndent(users, "", "    ")
	if err != nil {
		logger.Error(err)
		return err
	}

	if count {
		fmt.Println(len(users.Users))
	} else {
		fmt.Println(string(jsonData))
	}

	logger.Debug("finished doListUsers()")
	return nil
}

func doUserAllPages(ulc *admin.UsersListCall, users *admin.Users) error {
	logger.Debug("starting doUserAllPages()")

	if users.NextPageToken != "" {
		ulc = usrs.AddPageToken(ulc, users.NextPageToken)
		nxtUsers, err := usrs.DoList(ulc)
		if err != nil {
			logger.Error(err)
			return err
		}
		users.Users = append(users.Users, nxtUsers.Users...)
		users.Etag = nxtUsers.Etag
		users.NextPageToken = nxtUsers.NextPageToken

		if nxtUsers.NextPageToken != "" {
			doUserAllPages(ulc, users)
		}
	}

	logger.Debug("finished doUserAllPages()")
	return nil
}

func doUserNumPages(ulc *admin.UsersListCall, users *admin.Users, numPages int) error {
	logger.Debugw("starting doUserNumPages()",
		"numPages", numPages)

	if users.NextPageToken != "" && numPages > 0 {
		ulc = usrs.AddPageToken(ulc, users.NextPageToken)
		nxtUsers, err := usrs.DoList(ulc)
		if err != nil {
			logger.Error(err)
			return err
		}
		users.Users = append(users.Users, nxtUsers.Users...)
		users.Etag = nxtUsers.Etag
		users.NextPageToken = nxtUsers.NextPageToken

		if nxtUsers.NextPageToken != "" {
			doUserNumPages(ulc, users, numPages-1)
		}
	}

	logger.Debug("finished doUserNumPages()")
	return nil
}

func doUserPages(ulc *admin.UsersListCall, users *admin.Users, pages string) error {
	logger.Debugw("starting doUserPages()",
		"pages", pages)

	if pages == "all" {
		err := doUserAllPages(ulc, users)
		if err != nil {
			logger.Error(err)
			return err
		}
	} else {
		numPages, err := strconv.Atoi(pages)
		if err != nil {
			err = errors.New(cmn.ErrInvalidPagesArgument)
			logger.Error(err)
			return err
		}

		if numPages > 1 {
			err = doUserNumPages(ulc, users, numPages-1)
			if err != nil {
				logger.Error(err)
				return err
			}
		}
	}

	logger.Debug("finished doUserPages()")
	return nil
}

func init() {
	listCmd.AddCommand(listUsersCmd)

	listUsersCmd.Flags().StringVarP(&attrs, "attributes", "a", "", "required user attributes (separated by ~)")
	listUsersCmd.Flags().BoolVarP(&count, "count", "", false, "count number of entities returned")
	listUsersCmd.Flags().StringVarP(&customField, "custom-field-mask", "c", "", "custom field mask schemas (separated by ~)")
	listUsersCmd.Flags().StringVarP(&domain, "domain", "d", "", "domain from which to get users")
	listUsersCmd.Flags().Int64VarP(&maxResults, "max-results", "m", 500, "maximum number of results to return per page")
	listUsersCmd.Flags().StringVarP(&orderBy, "order-by", "o", "", "field by which results will be ordered")
	listUsersCmd.Flags().StringVarP(&pages, "pages", "p", "", "number of pages of results to be returned ('all' or a number)")
	listUsersCmd.Flags().StringVarP(&projection, "projection", "j", "", "type of projection")
	listUsersCmd.Flags().StringVarP(&query, "query", "q", "", "selection criteria to get users (separated by ~)")
	listUsersCmd.Flags().StringVarP(&sortOrder, "sort-order", "s", "", "sort order of returned results")
	listUsersCmd.Flags().StringVarP(&viewType, "view-type", "v", "", "data view type")
	listUsersCmd.Flags().BoolVarP(&deleted, "deleted", "x", false, "show deleted users")

}
