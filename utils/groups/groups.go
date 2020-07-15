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
	"errors"
	"strings"

	admin "google.golang.org/api/admin/directory/v1"
	"google.golang.org/api/googleapi"
)

const (
	endField         string = ")"
	startGroupsField string = "groups("
)

// GroupAttrMap provides lowercase mappings to valid admin.Group attributes
var GroupAttrMap = map[string]string{
	"admincreated":       "adminCreated",
	"description":        "description",
	"directmemberscount": "directMembersCount",
	"email":              "email",
	"etag":               "etag",
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
	var newGLC *admin.GroupsListCall

	newGLC = glc.Customer(customerID)

	return newGLC
}

// AddDomain adds domain to admin calls
func AddDomain(glc *admin.GroupsListCall, domain string) *admin.GroupsListCall {
	var newGLC *admin.GroupsListCall

	newGLC = glc.Domain(domain)

	return newGLC
}

// AddFields adds fields to be returned from admin calls
func AddFields(callObj interface{}, attrs string) interface{} {
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
	var newGLC *admin.GroupsListCall

	newGLC = glc.MaxResults(maxResults)

	return newGLC
}

// AddOrderBy adds OrderBy to admin calls
func AddOrderBy(glc *admin.GroupsListCall, orderBy string) *admin.GroupsListCall {
	var newGLC *admin.GroupsListCall

	newGLC = glc.OrderBy(orderBy)

	return newGLC
}

// AddPageToken adds PageToken to admin calls
func AddPageToken(glc *admin.GroupsListCall, token string) *admin.GroupsListCall {
	var newGLC *admin.GroupsListCall

	newGLC = glc.PageToken(token)

	return newGLC
}

// AddQuery adds query to admin calls
func AddQuery(glc *admin.GroupsListCall, query string) *admin.GroupsListCall {
	var newGLC *admin.GroupsListCall

	newGLC = glc.Query(query)

	return newGLC
}

// AddSortOrder adds SortOrder to admin calls
func AddSortOrder(glc *admin.GroupsListCall, sortorder string) *admin.GroupsListCall {
	var newGLC *admin.GroupsListCall

	newGLC = glc.SortOrder(sortorder)

	return newGLC
}

// AddUserKey adds UserKey to admin calls
func AddUserKey(glc *admin.GroupsListCall, key string) *admin.GroupsListCall {
	var newGLC *admin.GroupsListCall

	newGLC = glc.UserKey(key)

	return newGLC
}

// DoGet calls the .Do() function on the admin.GroupsGetCall
func DoGet(ggc *admin.GroupsGetCall) (*admin.Group, error) {
	group, err := ggc.Do()
	if err != nil {
		return nil, err
	}

	return group, nil
}

// DoList calls the .Do() function on the admin.GroupsListCall
func DoList(glc *admin.GroupsListCall) (*admin.Groups, error) {
	groups, err := glc.Do()
	if err != nil {
		return nil, err
	}

	return groups, nil
}

// FormatAttrs formats attributes for admin.GroupsListCall.Fields and admin.GroupsGetCall.Fields call
func FormatAttrs(attrs []string, get bool) string {
	var (
		outputStr   string
		groupFields []string
	)

	for _, a := range attrs {
		groupFields = append(groupFields, a)
	}

	if get {
		outputStr = strings.Join(groupFields, ",")
	} else {
		outputStr = startGroupsField + strings.Join(groupFields, ",") + endField
	}

	return outputStr
}

// FormatQuery produces a query string from multiple query elements
func FormatQuery(queryParts []string) (string, error) {
	formattedQuery := strings.Join(queryParts, " ")

	if strings.Contains(formattedQuery, "memberKey") &&
		(strings.Contains(formattedQuery, "name") ||
			strings.Contains(formattedQuery, "email")) {
		err := errors.New("gmin: error - memberKey must be used on its own in a query")
		return "", err
	}

	return formattedQuery, nil
}
