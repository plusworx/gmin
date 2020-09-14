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
	"errors"
	"fmt"
	"sort"
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

var attrValues = []string{
	"delivery_settings",
	"role",
}

// deliverySettingMap provides lowercase mappings to valid admin.Member delivery settings
var deliverySettingMap = map[string]string{
	"all_mail": "ALL_MAIL",
	"daily":    "DAILY",
	"digest":   "DIGEST",
	"disabled": "DISABLED",
	"none":     "NONE",
}

var flagValues = []string{
	"roles",
}

// MemberAttrMap provides lowercase mappings to valid admin.Member attributes
var MemberAttrMap = map[string]string{
	"delivery_settings": "delivery_settings",
	"email":             "email",
	"etag":              "etag",
	"id":                "id",
	"kind":              "kind",
	"memberkey":         "memberKey", // used in batch update
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

// Key is struct used to extract memberKey
type Key struct {
	MemberKey string
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

// ShowAttrs displays requested group member attributes
func ShowAttrs(filter string) {
	keys := make([]string, 0, len(MemberAttrMap))
	for k := range MemberAttrMap {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		if filter == "" {
			fmt.Println(MemberAttrMap[k])
			continue
		}

		if strings.Contains(k, strings.ToLower(filter)) {
			fmt.Println(MemberAttrMap[k])
		}

	}
}

// ShowAttrValues displays enumerated attribute values
func ShowAttrValues(lenArgs int, args []string) error {
	if lenArgs > 2 {
		return errors.New("gmin: error - too many arguments, group-member has maximum of 2")
	}

	if lenArgs == 1 {
		for _, v := range attrValues {
			fmt.Println(v)
		}
	}

	if lenArgs == 2 {
		attr := strings.ToLower(args[1])
		values := []string{}

		switch {
		case attr == "delivery_settings":
			for _, val := range deliverySettingMap {
				values = append(values, val)
			}
			sort.Strings(values)
			for _, s := range values {
				fmt.Println(s)
			}
		case attr == "role":
			for _, val := range RoleMap {
				values = append(values, val)
			}
			sort.Strings(values)
			for _, s := range values {
				fmt.Println(s)
			}
		default:
			return fmt.Errorf("gmin: error - %v attribute not recognized", args[1])
		}
	}

	return nil
}

// ShowFlagValues displays enumerated flag values
func ShowFlagValues(lenArgs int, args []string) error {
	values := []string{}

	if lenArgs == 1 {
		for _, v := range flagValues {
			fmt.Println(v)
		}
	}

	if lenArgs == 2 {
		flag := strings.ToLower(args[1])

		if flag == "roles" {
			for _, val := range RoleMap {
				values = append(values, val)
			}
			sort.Strings(values)
			for _, s := range values {
				fmt.Println(s)
			}
		} else {
			return fmt.Errorf("gmin: error - %v flag not recognized", args[1])
		}
	}

	return nil
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
