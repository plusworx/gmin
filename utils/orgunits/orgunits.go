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

package orgunits

import (
	"strings"

	admin "google.golang.org/api/admin/directory/v1"
	"google.golang.org/api/googleapi"
)

const (
	endField           string = ")"
	startOrgUnitsField string = "organizationUnits("
)

// OrgUnitAttrMap provides lowercase mappings to valid admin.OrgUnit attributes
var OrgUnitAttrMap = map[string]string{
	"blockinheritance":  "blockInheritance",
	"description":       "description",
	"etag":              "etag",
	"kind":              "kind",
	"name":              "name",
	"orgunitid":         "orgUnitId",
	"orgunitpath":       "orgUnitPath",
	"parentorgunitid":   "parentOrgUnitId",
	"parentorgunitpath": "parentOrgUnitPath",
}

// ValidSearchTypes provides list of valid types for admin.OrgunitsListCall
var ValidSearchTypes = []string{
	"all",
	"children",
}

// AddListFields adds fields to be returned to admin.OrgunitsListCall
func AddListFields(oulc *admin.OrgunitsListCall, attrs string) *admin.OrgunitsListCall {
	var fields googleapi.Field = googleapi.Field(attrs)
	var newOULC *admin.OrgunitsListCall

	newOULC = oulc.Fields(fields)

	return newOULC
}

// AddListOUPath adds OrgUnitPath or ID to admin.OrgunitsListCall
func AddListOUPath(oulc *admin.OrgunitsListCall, path string) *admin.OrgunitsListCall {
	var newOULC *admin.OrgunitsListCall

	newOULC = oulc.OrgUnitPath(path)

	return newOULC
}

// AddListType adds Type to admin.OrgunitsListCall
func AddListType(oulc *admin.OrgunitsListCall, searchType string) *admin.OrgunitsListCall {
	var newOULC *admin.OrgunitsListCall

	newOULC = oulc.Type(searchType)

	return newOULC
}

// DoList calls the .Do() function on the admin.OrgunitsListCall
func DoList(oulc *admin.OrgunitsListCall) (*admin.OrgUnits, error) {
	orgunits, err := oulc.Do()
	if err != nil {
		return nil, err
	}

	return orgunits, nil
}

// FormatAttrs formats attributes for admin.OrgunitsListCall.Fields call
func FormatAttrs(attrs []string) string {
	var (
		outputStr string
		ouFields  []string
	)

	for _, a := range attrs {
		ouFields = append(ouFields, a)
	}

	outputStr = startOrgUnitsField + strings.Join(ouFields, ",") + endField

	return outputStr
}

// Get fetches an orgunit
func Get(ougc *admin.OrgunitsGetCall) (*admin.OrgUnit, error) {
	orgUnit, err := ougc.Do()
	if err != nil {
		return nil, err
	}

	return orgUnit, nil
}

// GetAttrs fetches specified attributes for orgunit
func GetAttrs(ougc *admin.OrgunitsGetCall, attrs string) (*admin.OrgUnit, error) {
	var fields googleapi.Field = googleapi.Field(attrs)

	orgUnit, err := ougc.Fields(fields).Do()
	if err != nil {
		return nil, err
	}

	return orgUnit, nil
}
