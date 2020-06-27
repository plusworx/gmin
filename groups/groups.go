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

	cfg "github.com/plusworx/gmin/config"
	admin "google.golang.org/api/admin/directory/v1"
	"google.golang.org/api/googleapi"
)

const (
	endField         string = ")"
	startGroupsField string = "(groups:"
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

// AllDomain fetches groups for all domains
func AllDomain(glc *admin.GroupsListCall) (*admin.Groups, error) {
	groups, err := glc.Customer(cfg.CustomerID).Do()
	if err != nil {
		return nil, err
	}

	return groups, nil
}

// AllDomainAttrs fetches specified attributes for all domain groups
func AllDomainAttrs(glc *admin.GroupsListCall, attrs string) (*admin.Groups, error) {
	var fields googleapi.Field = googleapi.Field(attrs)

	groups, err := glc.Customer(cfg.CustomerID).Fields(fields).Do()
	if err != nil {
		return nil, err
	}

	return groups, nil
}

// AllDomainQuery fetches groups for all domains that satisfy query arguments
func AllDomainQuery(glc *admin.GroupsListCall, query string) (*admin.Groups, error) {
	groups, err := glc.Customer(cfg.CustomerID).Query(query).Do()
	if err != nil {
		return nil, err
	}

	return groups, nil
}

// AllDomainQueryAttrs fetches specified attributes for all domain groups that satisfy query arguments
func AllDomainQueryAttrs(glc *admin.GroupsListCall, query string, attrs string) (*admin.Groups, error) {
	var fields googleapi.Field = googleapi.Field(attrs)

	groups, err := glc.Customer(cfg.CustomerID).Query(query).Fields(fields).Do()
	if err != nil {
		return nil, err
	}

	return groups, nil
}

// Domain fetches groups for a particular domain
func Domain(domain string, glc *admin.GroupsListCall) (*admin.Groups, error) {
	groups, err := glc.Domain(domain).Do()
	if err != nil {
		return nil, err
	}

	return groups, nil
}

// DomainAttrs fetches specified attributes for domain groups
func DomainAttrs(domain string, glc *admin.GroupsListCall, attrs string) (*admin.Groups, error) {
	var fields googleapi.Field = googleapi.Field(attrs)

	groups, err := glc.Domain(domain).Fields(fields).Do()
	if err != nil {
		return nil, err
	}

	return groups, nil
}

// DomainQuery fetches groups for specified domain that satisfy query arguments
func DomainQuery(domain string, glc *admin.GroupsListCall, query string) (*admin.Groups, error) {
	groups, err := glc.Domain(domain).Query(query).Do()
	if err != nil {
		return nil, err
	}

	return groups, nil
}

// DomainQueryAttrs fetches specified attributes for domain groups that satisfy query arguments
func DomainQueryAttrs(domain string, glc *admin.GroupsListCall, query string, attrs string) (*admin.Groups, error) {
	var fields googleapi.Field = googleapi.Field(attrs)

	groups, err := glc.Domain(domain).Query(query).Fields(fields).Do()
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

// Single fetches a group
func Single(ggc *admin.GroupsGetCall) (*admin.Group, error) {
	group, err := ggc.Do()
	if err != nil {
		return nil, err
	}

	return group, nil
}

// SingleAttrs fetches specified attributes for group
func SingleAttrs(ggc *admin.GroupsGetCall, attrs string) (*admin.Group, error) {
	var fields googleapi.Field = googleapi.Field(attrs)

	group, err := ggc.Fields(fields).Do()
	if err != nil {
		return nil, err
	}

	return group, nil
}
