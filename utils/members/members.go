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
	endField          string = ")"
	startMembersField string = "members("
)

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
	"deliverysettings": "deliverySettings",
	"email":            "email",
	"etag":             "etag",
	"id":               "id",
	"kind":             "kind",
	"role":             "role",
	"status":           "status",
	"type":             "type",
}

// roleMap provides lowercase mappings to valid admin.Member roles
var roleMap = map[string]string{
	"owner":   "OWNER",
	"manager": "MANAGER",
	"member":  "MEMBER",
}

// FormatAttrs formats attributes for admin.MembersListCall.Fields and admin.MembersGetCall.Fields call
func FormatAttrs(attrs []string, get bool) string {
	var (
		outputStr    string
		memberFields []string
	)

	for _, a := range attrs {
		memberFields = append(memberFields, a)
	}

	if get {
		outputStr = strings.Join(memberFields, ",")
	} else {
		outputStr = startMembersField + strings.Join(memberFields, ",") + endField
	}

	return outputStr
}

// Get fetches member of a particular group
func Get(mgc *admin.MembersGetCall) (*admin.Member, error) {
	member, err := mgc.Do()
	if err != nil {
		return nil, err
	}

	return member, nil
}

// GetAttrs fetches specified attributes for member
func GetAttrs(mgc *admin.MembersGetCall, attrs string) (*admin.Member, error) {
	var fields googleapi.Field = googleapi.Field(attrs)

	member, err := mgc.Fields(fields).Do()
	if err != nil {
		return nil, err
	}

	return member, nil
}

// ListMemberAttrs fetches specified attributes for members
func ListMemberAttrs(mlc *admin.MembersListCall, attrs string, maxResults int64) (*admin.Members, error) {
	var fields googleapi.Field = googleapi.Field(attrs)

	members, err := mlc.Fields(fields).MaxResults(maxResults).Do()
	if err != nil {
		return nil, err
	}

	return members, nil
}

// ListMembers fetches members of a particular group for admin.MembersListCall
func ListMembers(mlc *admin.MembersListCall, maxResults int64) (*admin.Members, error) {
	members, err := mlc.MaxResults(maxResults).Do()
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

	validRole := roleMap[lowerRole]
	if validRole == "" {
		return "", fmt.Errorf("gmin: error - %v is not a valid role", role)
	}

	return validRole, nil
}
