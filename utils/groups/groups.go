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

package groups

import (
	"fmt"
	"sort"
	"strings"

	cmn "github.com/plusworx/gmin/utils/common"
	gmess "github.com/plusworx/gmin/utils/gminmessages"
	lg "github.com/plusworx/gmin/utils/logging"
	admin "google.golang.org/api/admin/directory/v1"
	"google.golang.org/api/googleapi"
)

const (
	// ENDFIELD is List call attribute string terminator
	ENDFIELD string = ")"
	// KEYNAME is name of key for processing
	KEYNAME string = "groupKey"
	// STARTGROUPSFIELD is List call attribute string prefix
	STARTGROUPSFIELD string = "groups("
)

// GroupParams holds group data for batch processing
type GroupParams struct {
	GroupKey string
	Group    *admin.Group
}

// Key is struct used to extract groupKey
type Key struct {
	GroupKey string
}

var flagValues = []string{
	"order-by",
	"sort-order",
}

// GroupAttrMap provides lowercase mappings to valid admin.Group attributes
var GroupAttrMap = map[string]string{
	"admincreated":       "adminCreated",
	"description":        "description",
	"directmemberscount": "directMembersCount",
	"email":              "email",
	"etag":               "etag",
	"forcesendfields":    "forceSendFields",
	"groupkey":           "groupKey", // Used in batch commands
	"id":                 "id",
	"kind":               "kind",
	"name":               "name",
	"noneditablealiases": "nonEditableAliases",
}

// QueryAttrMap provides lowercase mappings to valid admin.Group query attributes
var QueryAttrMap = map[string]string{
	"email":     "email",
	"name":      "name",
	"memberkey": "memberKey",
}

// ValidOrderByStrs provide valid strings to be used to set admin.GroupsListCall OrderBy
var ValidOrderByStrs = []string{
	"email",
}

// AddCustomer adds Customer to admin calls
func AddCustomer(glc *admin.GroupsListCall, customerID string) *admin.GroupsListCall {
	lg.Debugw("starting AddCustomer()",
		"customerID", customerID)
	defer lg.Debug("finished AddCustomer()")

	var newGLC *admin.GroupsListCall

	newGLC = glc.Customer(customerID)

	return newGLC
}

// AddDomain adds domain to admin calls
func AddDomain(glc *admin.GroupsListCall, domain string) *admin.GroupsListCall {
	lg.Debugw("starting AddDomain()",
		"domain", domain)
	defer lg.Debug("finished AddDomain()")

	var newGLC *admin.GroupsListCall

	newGLC = glc.Domain(domain)

	return newGLC
}

// AddFields adds fields to be returned from admin calls
func AddFields(callObj interface{}, attrs string) interface{} {
	lg.Debugw("starting AddFields()",
		"attrs", attrs)
	defer lg.Debug("finished AddFields()")

	var fields googleapi.Field = googleapi.Field(attrs)

	switch callObj.(type) {
	case *admin.GroupsListCall:
		var newGLC *admin.GroupsListCall
		glc := callObj.(*admin.GroupsListCall)
		newGLC = glc.Fields(fields)

		return newGLC
	case *admin.GroupsGetCall:
		var newGGC *admin.GroupsGetCall
		ggc := callObj.(*admin.GroupsGetCall)
		newGGC = ggc.Fields(fields)

		return newGGC
	}

	return nil
}

// AddMaxResults adds MaxResults to admin calls
func AddMaxResults(glc *admin.GroupsListCall, maxResults int64) *admin.GroupsListCall {
	lg.Debugw("starting AddMaxResults()",
		"maxResults", maxResults)
	defer lg.Debug("finished AddMaxResults()")

	var newGLC *admin.GroupsListCall

	newGLC = glc.MaxResults(maxResults)

	return newGLC
}

// AddOrderBy adds OrderBy to admin calls
func AddOrderBy(glc *admin.GroupsListCall, orderBy string) *admin.GroupsListCall {
	lg.Debugw("starting AddOrderBy()",
		"orderBy", orderBy)
	defer lg.Debug("finished AddOrderBy()")

	var newGLC *admin.GroupsListCall

	newGLC = glc.OrderBy(orderBy)

	return newGLC
}

// AddPageToken adds PageToken to admin calls
func AddPageToken(glc *admin.GroupsListCall, token string) *admin.GroupsListCall {
	lg.Debugw("starting AddPageToken()",
		"token", token)
	defer lg.Debug("finished AddPageToken()")

	var newGLC *admin.GroupsListCall

	newGLC = glc.PageToken(token)

	return newGLC
}

// AddQuery adds query to admin calls
func AddQuery(glc *admin.GroupsListCall, query string) *admin.GroupsListCall {
	lg.Debugw("starting AddQuery()",
		"query", query)
	defer lg.Debug("finished AddQuery()")

	var newGLC *admin.GroupsListCall

	newGLC = glc.Query(query)

	return newGLC
}

// AddSortOrder adds SortOrder to admin calls
func AddSortOrder(glc *admin.GroupsListCall, sortorder string) *admin.GroupsListCall {
	lg.Debugw("starting AddSortOrder()",
		"sortorder", sortorder)
	defer lg.Debug("finished AddSortOrder()")

	var newGLC *admin.GroupsListCall

	newGLC = glc.SortOrder(sortorder)

	return newGLC
}

// AddUserKey adds UserKey to admin calls
func AddUserKey(glc *admin.GroupsListCall, key string) *admin.GroupsListCall {
	lg.Debugw("starting AddUserKey()",
		"key", key)
	defer lg.Debug("finished AddUserKey()")

	var newGLC *admin.GroupsListCall

	newGLC = glc.UserKey(key)

	return newGLC
}

// DoGet calls the .Do() function on the admin.GroupsGetCall
func DoGet(ggc *admin.GroupsGetCall) (*admin.Group, error) {
	lg.Debug("starting DoGet()")
	defer lg.Debug("finished DoGet()")

	group, err := ggc.Do()
	if err != nil {
		lg.Error(err)
		return nil, err
	}

	return group, nil
}

// DoList calls the .Do() function on the admin.GroupsListCall
func DoList(glc *admin.GroupsListCall) (*admin.Groups, error) {
	lg.Debug("starting DoList()")
	defer lg.Debug("finished DoList()")

	groups, err := glc.Do()
	if err != nil {
		lg.Error(err)
		return nil, err
	}

	return groups, nil
}

// PopulateGroup is used in batch processing
func PopulateGroup(group *admin.Group, hdrMap map[int]string, objData []interface{}) error {
	lg.Debugw("starting populateGroup()",
		"hdrMap", hdrMap)
	defer lg.Debug("finished populateGroup()")

	for idx, attr := range objData {
		attrName := hdrMap[idx]
		attrVal := fmt.Sprintf("%v", attr)

		switch {
		case (attrName == "email" || attrName == "name") && attrVal == "":
			err := fmt.Errorf(gmess.ERR_EMPTYSTRING, attrName)
			lg.Error(err)
			return err
		case attrName == "description":
			group.Description = attrVal
		case attrName == "email":
			group.Email = attrVal
		case attrName == "name":
			group.Name = attrVal
		default:
			err := fmt.Errorf(gmess.ERR_ATTRNOTRECOGNIZED, attrName)
			return err
		}
	}
	return nil
}

// PopulateGroupForUpdate is used in batch processing
func PopulateGroupForUpdate(grpParams *GroupParams, hdrMap map[int]string, objData []interface{}) error {
	lg.Debugw("starting PopulateGroupForUpdate()",
		"hdrMap", hdrMap)
	defer lg.Debug("finished PopulateGroupForUpdate()")
	for idx, attr := range objData {
		attrName := hdrMap[idx]
		attrVal := fmt.Sprintf("%v", attr)

		switch {
		case attrName == "description":
			grpParams.Group.Description = attrVal
			if attrVal == "" {
				grpParams.Group.ForceSendFields = append(grpParams.Group.ForceSendFields, "Description")
			}
		case attrName == "email":
			if attrVal == "" {
				err := fmt.Errorf(gmess.ERR_EMPTYSTRING, attrName)
				lg.Error(err)
				return err
			}
			grpParams.Group.Email = attrVal
		case attrName == "name":
			if attrVal == "" {
				err := fmt.Errorf(gmess.ERR_EMPTYSTRING, attrName)
				lg.Error(err)
				return err
			}
			grpParams.Group.Name = attrVal
		case attrName == "groupKey":
			if attrVal == "" {
				err := fmt.Errorf(gmess.ERR_EMPTYSTRING, attrName)
				lg.Error(err)
				return err
			}
			grpParams.GroupKey = attrVal
		default:
			err := fmt.Errorf(gmess.ERR_ATTRNOTRECOGNIZED, attrName)
			return err
		}
	}
	return nil
}

// ShowAttrs displays requested group attributes
func ShowAttrs(filter string) {
	lg.Debugw("starting ShowAttrs()",
		"filter", filter)
	defer lg.Debug("finished ShowAttrs()")

	keys := make([]string, 0, len(GroupAttrMap))
	for k := range GroupAttrMap {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		if filter == "" {
			fmt.Println(GroupAttrMap[k])
			continue
		}

		if strings.Contains(k, strings.ToLower(filter)) {
			fmt.Println(GroupAttrMap[k])
		}

	}
}

// ShowFlagValues displays enumerated flag values
func ShowFlagValues(lenArgs int, args []string, filter string) error {
	lg.Debugw("starting ShowFlagValues()",
		"lenArgs", lenArgs,
		"args", args,
		"filter", filter)
	defer lg.Debug("finished ShowFlagValues()")

	if lenArgs == 1 {
		cmn.ShowFlagValues(flagValues, filter)
	}

	if lenArgs == 2 {
		flag := strings.ToLower(args[1])
		valSlice := []string{}

		switch {
		case flag == "order-by":
			cmn.ShowFlagValues(ValidOrderByStrs, filter)
		case flag == "sort-order":
			for _, v := range cmn.ValidSortOrders {
				valSlice = append(valSlice, v)
			}
			uniqueSlice := cmn.UniqueStrSlice(valSlice)
			cmn.ShowFlagValues(uniqueSlice, filter)
		default:
			err := fmt.Errorf(gmess.ERR_FLAGNOTRECOGNIZED, args[1])
			lg.Error(err)
			return err
		}
	}
	return nil
}
