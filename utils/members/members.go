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
	KEYNAME string = "memberKey"
	// STARTMEMBERSFIELD is List call attribute string prefix
	STARTMEMBERSFIELD string = "members("
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
	"memberkey":         "memberKey", // used in batch commands
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
	lg.Debugw("starting AddFields()",
		"attrs", attrs)
	defer lg.Debug("finished AddFields()")

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
	lg.Debugw("starting AddMaxResults()",
		"maxResults", maxResults)
	defer lg.Debug("finished AddMaxResults()")

	var newMLC *admin.MembersListCall

	newMLC = mlc.MaxResults(maxResults)

	return newMLC
}

// AddPageToken adds PageToken to admin calls
func AddPageToken(mlc *admin.MembersListCall, token string) *admin.MembersListCall {
	lg.Debugw("starting AddPageToken()",
		"token", token)
	defer lg.Debug("finished AddPageToken()")

	var newMLC *admin.MembersListCall

	newMLC = mlc.PageToken(token)

	return newMLC
}

// AddRoles adds Roles to admin calls
func AddRoles(mlc *admin.MembersListCall, roles string) *admin.MembersListCall {
	lg.Debugw("starting AddRoles()",
		"roles", roles)
	defer lg.Debug("finished AddRoles()")

	var newMLC *admin.MembersListCall

	newMLC = mlc.Roles(roles)

	return newMLC
}

// DoGet calls the .Do() function on the admin.MembersGetCall
func DoGet(mgc *admin.MembersGetCall) (*admin.Member, error) {
	lg.Debug("starting DoGet()")
	defer lg.Debug("finished DoGet()")

	member, err := mgc.Do()
	if err != nil {
		lg.Error(err)
		return nil, err
	}

	return member, nil
}

// DoList calls the .Do() function on the admin.MembersListCall
func DoList(mlc *admin.MembersListCall) (*admin.Members, error) {
	lg.Debug("starting DoList()")
	defer lg.Debug("finished DoList()")

	members, err := mlc.Do()
	if err != nil {
		lg.Error(err)
		return nil, err
	}

	return members, nil
}

// PopulateMember is used in batch processing
func PopulateMember(member *admin.Member, hdrMap map[int]string, objData []interface{}) error {
	lg.Debugw("starting populateMember()",
		"hdrMap", hdrMap)
	defer lg.Debug("finished populateMember()")

	for idx, attr := range objData {
		attrName := hdrMap[idx]
		attrVal := fmt.Sprintf("%v", attr)

		switch {
		case attrName == "delivery_settings":
			validDS, err := ValidateDeliverySetting(attrVal)
			if err != nil {
				return err
			}
			member.DeliverySettings = validDS
		case attrName == "email":
			if attrVal == "" {
				err := fmt.Errorf(gmess.ERR_EMPTYSTRING, attrName)
				lg.Error(err)
				return err
			}
			member.Email = attrVal
		case attrName == "role":
			validRole, err := ValidateRole(attrVal)
			if err != nil {
				return err
			}
			member.Role = validRole
		default:
			err := fmt.Errorf(gmess.ERR_ATTRNOTRECOGNIZED, attrName)
			return err
		}
	}
	return nil
}

// ShowAttrs displays requested group member attributes
func ShowAttrs(filter string) {
	lg.Debugw("starting ShowAttrs()",
		"filter", filter)
	defer lg.Debug("finished ShowAttrs()")

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
func ShowAttrValues(lenArgs int, args []string, filter string) error {
	lg.Debugw("starting ShowAttrValues()",
		"lenArgs", lenArgs,
		"args", args,
		"filter", filter)
	defer lg.Debug("finished ShowAttrValues()")

	if lenArgs > 2 {
		err := fmt.Errorf(gmess.ERR_TOOMANYARGSMAX1, args[0])
		lg.Error(err)
		return err
	}

	if lenArgs == 1 {
		cmn.ShowAttrVals(attrValues, filter)
	}

	if lenArgs == 2 {
		attr := strings.ToLower(args[1])
		values := []string{}

		switch {
		case attr == "delivery_settings":
			for _, val := range deliverySettingMap {
				values = append(values, val)
			}
		case attr == "role":
			for _, val := range RoleMap {
				values = append(values, val)
			}
		default:
			err := fmt.Errorf(gmess.ERR_ATTRNOTRECOGNIZED, args[1])
			lg.Error(err)
			return err
		}

		sort.Strings(values)
		cmn.ShowAttrVals(values, filter)
	}

	return nil
}

// ShowFlagValues displays enumerated flag values
func ShowFlagValues(lenArgs int, args []string, filter string) error {
	lg.Debugw("starting ShowFlagValues()",
		"lenArgs", lenArgs,
		"args", args,
		"filter", filter)
	defer lg.Debug("finished ShowFlagValues()")

	values := []string{}

	if lenArgs == 1 {
		cmn.ShowFlagValues(flagValues, filter)
	}

	if lenArgs == 2 {
		flag := strings.ToLower(args[1])

		if flag == "roles" {
			for _, val := range RoleMap {
				values = append(values, val)
			}
			sort.Strings(values)
			cmn.ShowFlagValues(values, filter)
		} else {
			err := fmt.Errorf(gmess.ERR_FLAGNOTRECOGNIZED, args[1])
			lg.Error(err)
			return err
		}
	}

	return nil
}

// ValidateDeliverySetting checks that a valid delivery setting has been provided
func ValidateDeliverySetting(ds string) (string, error) {
	lg.Debugw("starting ValidateDeliverySetting()",
		"ds", ds)
	defer lg.Debug("finished ValidateDeliverySetting()")

	lowerDS := strings.ToLower(ds)

	validSetting := deliverySettingMap[lowerDS]
	if validSetting == "" {
		err := fmt.Errorf(gmess.ERR_INVALIDDELIVERYSETTING, ds)
		lg.Error(err)
		return "", err
	}

	return validSetting, nil
}

// ValidateRole checks that a valid role has been provided
func ValidateRole(role string) (string, error) {
	lg.Debugw("starting ValidateRole()",
		"role", role)
	defer lg.Debug("finished ValidateRole()")

	lowerRole := strings.ToLower(role)

	validRole := RoleMap[lowerRole]
	if validRole == "" {
		return "", fmt.Errorf(gmess.ERR_INVALIDROLE, role)
	}

	return validRole, nil
}
