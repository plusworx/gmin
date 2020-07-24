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
	admin "google.golang.org/api/admin/directory/v1"
	"google.golang.org/api/googleapi"
)

const (
	// EndField is List call attribute string terminator
	EndField string = ")"
	// StartOrgUnitsField is List call attribute string prefix
	StartOrgUnitsField string = "organizationUnits("
)

// GminOrgUnit is custom admin.OrgUnit struct with no omitempty tags
type GminOrgUnit struct {
	// BlockInheritance: Should block inheritance
	BlockInheritance bool `json:"blockInheritance"`

	// Description: Description of OrgUnit
	Description string `json:"description"`

	// Etag: ETag of the resource.
	Etag string `json:"etag"`

	// Kind: Kind of resource this is.
	Kind string `json:"kind"`

	// Name: Name of OrgUnit
	Name string `json:"name"`

	// OrgUnitId: Id of OrgUnit
	OrgUnitId string `json:"orgUnitId"`

	// OrgUnitPath: Path of OrgUnit
	OrgUnitPath string `json:"orgUnitPath"`

	// ParentOrgUnitId: Id of parent OrgUnit
	ParentOrgUnitId string `json:"parentOrgUnitId"`

	// ParentOrgUnitPath: Path of parent OrgUnit
	ParentOrgUnitPath string `json:"parentOrgUnitPath"`

	// ServerResponse contains the HTTP response code and headers from the
	// server.
	googleapi.ServerResponse `json:"-"`

	// ForceSendFields is a list of field names (e.g. "BlockInheritance") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`

	// NullFields is a list of field names (e.g. "BlockInheritance") to
	// include in API requests with the JSON null value. By default, fields
	// with empty values are omitted from API requests. However, any field
	// with an empty value appearing in NullFields will be sent to the
	// server as null. It is an error if a field in this list has a
	// non-empty value. This may be used to include null fields in Patch
	// requests.
	NullFields []string `json:"-"`
}

// GminOrgUnits is custom admin.OrgUnits struct containing GminOrgUnit
type GminOrgUnits struct {
	// Etag: ETag of the resource.
	Etag string `json:"etag,omitempty"`

	// Kind: Kind of resource this is.
	Kind string `json:"kind,omitempty"`

	// OrganizationUnits: List of user objects.
	OrganizationUnits []*GminOrgUnit `json:"organizationUnits,omitempty"`

	// ServerResponse contains the HTTP response code and headers from the
	// server.
	googleapi.ServerResponse `json:"-"`

	// ForceSendFields is a list of field names (e.g. "Etag") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`

	// NullFields is a list of field names (e.g. "Etag") to include in API
	// requests with the JSON null value. By default, fields with empty
	// values are omitted from API requests. However, any field with an
	// empty value appearing in NullFields will be sent to the server as
	// null. It is an error if a field in this list has a non-empty value.
	// This may be used to include null fields in Patch requests.
	NullFields []string `json:"-"`
}

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

// AddFields adds fields to be returned to admin calls
func AddFields(callObj interface{}, attrs string) interface{} {
	var fields googleapi.Field = googleapi.Field(attrs)

	switch callObj.(type) {
	case *admin.OrgunitsListCall:
		var newOULC *admin.OrgunitsListCall
		oulc := callObj.(*admin.OrgunitsListCall)
		newOULC = oulc.Fields(fields)

		return newOULC
	case *admin.OrgunitsGetCall:
		var newOUGC *admin.OrgunitsGetCall
		ougc := callObj.(*admin.OrgunitsGetCall)
		newOUGC = ougc.Fields(fields)

		return newOUGC
	}

	return nil
}

// AddOUPath adds OrgUnitPath or ID to admin calls
func AddOUPath(oulc *admin.OrgunitsListCall, path string) *admin.OrgunitsListCall {
	var newOULC *admin.OrgunitsListCall

	newOULC = oulc.OrgUnitPath(path)

	return newOULC
}

// AddType adds Type to admin calls
func AddType(oulc *admin.OrgunitsListCall, searchType string) *admin.OrgunitsListCall {
	var newOULC *admin.OrgunitsListCall

	newOULC = oulc.Type(searchType)

	return newOULC
}

// DoGet calls the .Do() function on the admin.OrgunitsGetCall
func DoGet(ougc *admin.OrgunitsGetCall) (*admin.OrgUnit, error) {
	orgUnit, err := ougc.Do()
	if err != nil {
		return nil, err
	}

	return orgUnit, nil
}

// DoList calls the .Do() function on the admin.OrgunitsListCall
func DoList(oulc *admin.OrgunitsListCall) (*admin.OrgUnits, error) {
	orgunits, err := oulc.Do()
	if err != nil {
		return nil, err
	}

	return orgunits, nil
}
