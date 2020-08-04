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

package members

import (
	"fmt"
	"strings"

	admin "google.golang.org/api/admin/directory/v1"
	"google.golang.org/api/googleapi"
)

const (
	// EndField is List call attribute string terminator
	EndField string = ")"
	// StartMembersField is List call attribute string prefix
	StartMembersField string = "members("
)

// GminMember is custom admin.Member struct with no omitempty tags
type GminMember struct {
	// DeliverySettings: Delivery settings of member
	DeliverySettings string `json:"delivery_settings"`

	// Email: Email of member (Read-only)
	Email string `json:"email"`

	// Etag: ETag of the resource.
	Etag string `json:"etag"`

	// Id: The unique ID of the group member. A member id can be used as a
	// member request URI's memberKey. Unique identifier of group
	// (Read-only) Unique identifier of member (Read-only)
	Id string `json:"id"`

	// Kind: Kind of resource this is.
	Kind string `json:"kind"`

	// Role: Role of member
	Role string `json:"role"`

	// Status: Status of member (Immutable)
	Status string `json:"status"`

	// Type: Type of member (Immutable)
	Type string `json:"type"`

	// ServerResponse contains the HTTP response code and headers from the
	// server.
	googleapi.ServerResponse `json:"-"`

	// ForceSendFields is a list of field names (e.g. "DeliverySettings") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`

	// NullFields is a list of field names (e.g. "DeliverySettings") to
	// include in API requests with the JSON null value. By default, fields
	// with empty values are omitted from API requests. However, any field
	// with an empty value appearing in NullFields will be sent to the
	// server as null. It is an error if a field in this list has a
	// non-empty value. This may be used to include null fields in Patch
	// requests.
	NullFields []string `json:"-"`
}

// GminMembers is custom admin.Members struct containing GminMember
type GminMembers struct {
	// Etag: ETag of the resource.
	Etag string `json:"etag,omitempty"`

	// Kind: Kind of resource this is.
	Kind string `json:"kind,omitempty"`

	// Members: List of member objects.
	Members []*GminMember `json:"members,omitempty"`

	// NextPageToken: Token used to access next page of this result.
	NextPageToken string `json:"nextPageToken,omitempty"`

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

// deliverySettingMap provides lowercase mappings to valid admin.Member delivery settings
var deliverySettingMap = map[string]string{
	"all_mail": "ALL_MAIL",
	"daily":    "DAILY",
	"digest":   "DIGEST",
	"disabled": "DISABLED",
	"none":     "NONE",
}

// MemberAttrMap provides lowercase mappings to valid admin.Member attributes
var MemberAttrMap = map[string]string{
	"delivery_settings": "delivery_settings",
	"email":             "email",
	"etag":              "etag",
	"id":                "id",
	"kind":              "kind",
	"role":              "role",
	"status":            "status",
	"type":              "type",
}

// RoleMap provides lowercase mappings to valid admin.Member roles
var RoleMap = map[string]string{
	"owner":   "OWNER",
	"manager": "MANAGER",
	"member":  "MEMBER",
}

// AddFields adds fields to be returned to admin calls
func AddFields(callObj interface{}, attrs string) interface{} {
	var fields googleapi.Field = googleapi.Field(attrs)

	switch callObj.(type) {
	case *admin.MembersListCall:
		var newMLC *admin.MembersListCall
		mlc := callObj.(*admin.MembersListCall)
		newMLC = mlc.Fields(fields)

		return newMLC
	case *admin.MembersGetCall:
		var newMGC *admin.MembersGetCall
		mgc := callObj.(*admin.MembersGetCall)
		newMGC = mgc.Fields(fields)

		return newMGC
	}

	return nil
}

// AddMaxResults adds MaxResults to admin calls
func AddMaxResults(mlc *admin.MembersListCall, maxResults int64) *admin.MembersListCall {
	var newMLC *admin.MembersListCall

	newMLC = mlc.MaxResults(maxResults)

	return newMLC
}

// AddPageToken adds PageToken to admin calls
func AddPageToken(mlc *admin.MembersListCall, token string) *admin.MembersListCall {
	var newMLC *admin.MembersListCall

	newMLC = mlc.PageToken(token)

	return newMLC
}

// AddRoles adds Roles to admin calls
func AddRoles(mlc *admin.MembersListCall, roles string) *admin.MembersListCall {
	var newMLC *admin.MembersListCall

	newMLC = mlc.Roles(roles)

	return newMLC
}

// DoGet calls the .Do() function on the admin.MembersGetCall
func DoGet(mgc *admin.MembersGetCall) (*admin.Member, error) {
	member, err := mgc.Do()
	if err != nil {
		return nil, err
	}

	return member, nil
}

// DoList calls the .Do() function on the admin.MembersListCall
func DoList(mlc *admin.MembersListCall) (*admin.Members, error) {
	members, err := mlc.Do()
	if err != nil {
		return nil, err
	}

	return members, nil
}

// ValidateDeliverySetting checks that a valid delivery setting has been provided
func ValidateDeliverySetting(ds string) (string, error) {
	lowerDS := strings.ToLower(ds)

	validSetting := deliverySettingMap[lowerDS]
	if validSetting == "" {
		return "", fmt.Errorf("gmin: error - %v is not a valid delivery setting", ds)
	}

	return validSetting, nil
}

// ValidateRole checks that a valid role has been provided
func ValidateRole(role string) (string, error) {
	lowerRole := strings.ToLower(role)

	validRole := RoleMap[lowerRole]
	if validRole == "" {
		return "", fmt.Errorf("gmin: error - %v is not a valid role", role)
	}

	return validRole, nil
}
