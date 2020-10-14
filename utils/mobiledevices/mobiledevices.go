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

package mobiledevices

import (
	"fmt"
	"sort"
	"strings"

	cmn "github.com/plusworx/gmin/utils/common"
	gmess "github.com/plusworx/gmin/utils/gminmessages"
	admin "google.golang.org/api/admin/directory/v1"
	"google.golang.org/api/googleapi"
)

const (
	// ENDFIELD is List call attribute string terminator
	ENDFIELD string = ")"
	// STARTMOBDEVICESFIELD is List call attribute string prefix
	STARTMOBDEVICESFIELD string = "mobiledevices("
)

// ManagedDevice is struct to extract device data
type ManagedDevice struct {
	Action     string
	ResourceId string
}

var attrValues = []string{
	"action",
}

var flagValues = []string{
	"order-by",
	"projection",
	"sort-order",
}

var mobDevApplicationsAttrs = []string{
	"displayname",
	"packagename",
	"permission",
	"versioncode",
	"versionname",
}

// MobDevAttrMap provides lowercase mappings to valid admin.MobileDevice attributes
var MobDevAttrMap = map[string]string{
	"adbstatus":                      "adbStatus",
	"action":                         "action", // used in batch commands
	"applications":                   "applications",
	"basebandversion":                "basebandVersion",
	"bootloaderversion":              "bootloaderVersion",
	"brand":                          "brand",
	"buildnumber":                    "buildNumber",
	"defaultlanguage":                "defaultLanguage",
	"developeroptionsstatus":         "developerOptionsStatus",
	"devicecompromisedstatus":        "deviceCompromisedStatus",
	"deviceid":                       "deviceId",
	"devicepasswordstatus":           "devicePasswordStatus",
	"displayname":                    "displayName",
	"email":                          "email",
	"encryptionstatus":               "encryptionStatus",
	"etag":                           "etag",
	"firstsync":                      "firstSync",
	"forcesendfields":                "forceSendFields",
	"hardware":                       "hardware",
	"hardwareid":                     "hardwareId",
	"imei":                           "imei",
	"kernelversion":                  "kernelVersion",
	"kind":                           "kind",
	"lastsync":                       "lastSync",
	"managedaccountisonownerprofile": "managedAccountIsOnOwnerProfile",
	"manufacturer":                   "manufacturer",
	"meid":                           "meid",
	"model":                          "model",
	"name":                           "name",
	"networkoperator":                "networkOperator",
	"os":                             "os",
	"otheraccountsinfo":              "otherAccountsInfo",
	"packagename":                    "packageName",
	"permission":                     "permission",
	"privilege":                      "privilege",
	"releaseversion":                 "releaseVersion",
	"resourceid":                     "resourceId",
	"securitypatchlevel":             "securityPatchLevel",
	"serialnumber":                   "serialNumber",
	"status":                         "status",
	"supportsworkprofile":            "supportsWorkProfile",
	"type":                           "type",
	"unknownsourcesstatus":           "unknownSourcesStatus",
	"useragent":                      "userAgent",
	"versioncode":                    "versionCode",
	"versionname":                    "versionName",
	"wifimacaddress":                 "wifiMacAddress",
}

var mobDevAttrs = []string{
	"adbStatus",
	"applications",
	"basebandVersion",
	"bootloaderVersion",
	"brand",
	"buildNumber",
	"defaultLanguage",
	"developerOptionsStatus",
	"deviceCompromisedStatus",
	"deviceId",
	"devicePasswordStatus",
	"email",
	"encryptionStatus",
	"etag",
	"firstSync",
	"forceSendFields",
	"hardware",
	"hardwareId",
	"imei",
	"kernelVersion",
	"kind",
	"lastSync",
	"managedAccountIsOnOwnerProfile",
	"manufacturer",
	"meid",
	"model",
	"name",
	"networkOperator",
	"os",
	"otherAccountsInfo",
	"privilege",
	"releaseVersion",
	"resourceId",
	"securityPatchLevel",
	"serialNumber",
	"status",
	"supportsWorkProfile",
	"type",
	"unknownSourcesStatus",
	"userAgent",
	"wifiMacAddress",
}

var mobDevCompAttrs = map[string]string{
	"applications": "applications",
}

// QueryAttrMap provides lowercase mappings to valid admin.MobileDevice query attributes
var QueryAttrMap = map[string]string{
	"brand":        "brand",
	"email":        "email",
	"hardware":     "hardware",
	"id":           "id",
	"imei":         "imei",
	"manufacturer": "manufacturer",
	"meid":         "meid",
	"model":        "model",
	"name":         "name",
	"os":           "os",
	"serial":       "serial",
	"status":       "status",
	"sync":         "sync",
	"type":         "type",
}

// ValidActions provide valid strings to be used for admin.MobiledevicesActionCall
var ValidActions = []string{
	"admin_account_wipe",
	"admin_remote_wipe",
	"approve",
	"block",
	"cancel_remote_wipe_then_activate",
	"cancel_remote_wipe_then_block",
}

// ValidOrderByStrs provide valid strings to be used to set admin.MobiledevicesListCall OrderBy
var ValidOrderByStrs = []string{
	"deviceid",
	"email",
	"lastsync",
	"model",
	"name",
	"os",
	"status",
	"type",
}

// ValidProjections provide valid strings to be used to set admin.MobiledevicesListCall Projection
var ValidProjections = []string{
	"basic",
	"full",
}

// AddFields adds fields to be returned from admin calls
func AddFields(callObj interface{}, attrs string) interface{} {
	var fields googleapi.Field = googleapi.Field(attrs)

	switch callObj.(type) {
	case *admin.MobiledevicesActionCall:
		var newMDAC *admin.MobiledevicesActionCall
		mdac := callObj.(*admin.MobiledevicesActionCall)
		newMDAC = mdac.Fields(fields)

		return newMDAC
	case *admin.MobiledevicesGetCall:
		var newMDGC *admin.MobiledevicesGetCall
		mdgc := callObj.(*admin.MobiledevicesGetCall)
		newMDGC = mdgc.Fields(fields)

		return newMDGC
	case *admin.MobiledevicesListCall:
		var newMDLC *admin.MobiledevicesListCall
		mdlc := callObj.(*admin.MobiledevicesListCall)
		newMDLC = mdlc.Fields(fields)

		return newMDLC
	}
	return nil
}

// AddMaxResults adds MaxResults to admin calls
func AddMaxResults(mdlc *admin.MobiledevicesListCall, maxResults int64) *admin.MobiledevicesListCall {
	var newMDLC *admin.MobiledevicesListCall

	newMDLC = mdlc.MaxResults(maxResults)

	return newMDLC
}

// AddOrderBy adds OrderBy to admin calls
func AddOrderBy(mdlc *admin.MobiledevicesListCall, orderBy string) *admin.MobiledevicesListCall {
	var newMDLC *admin.MobiledevicesListCall

	newMDLC = mdlc.OrderBy(orderBy)

	return newMDLC
}

// AddPageToken adds PageToken to admin calls
func AddPageToken(mdlc *admin.MobiledevicesListCall, token string) *admin.MobiledevicesListCall {
	var newMDLC *admin.MobiledevicesListCall

	newMDLC = mdlc.PageToken(token)

	return newMDLC
}

// AddProjection adds Projection to admin calls
func AddProjection(callObj interface{}, projection string) interface{} {
	switch callObj.(type) {
	case *admin.MobiledevicesGetCall:
		var newMDGC *admin.MobiledevicesGetCall
		mdgc := callObj.(*admin.MobiledevicesGetCall)
		newMDGC = mdgc.Projection(projection)

		return newMDGC
	case *admin.MobiledevicesListCall:
		var newMDLC *admin.MobiledevicesListCall
		mdlc := callObj.(*admin.MobiledevicesListCall)
		newMDLC = mdlc.Projection(projection)

		return newMDLC
	}

	return nil
}

// AddQuery adds query to admin calls
func AddQuery(mdlc *admin.MobiledevicesListCall, query string) *admin.MobiledevicesListCall {
	var newMDLC *admin.MobiledevicesListCall

	newMDLC = mdlc.Query(query)

	return newMDLC
}

// AddSortOrder adds SortOrder to admin calls
func AddSortOrder(mdlc *admin.MobiledevicesListCall, sortorder string) *admin.MobiledevicesListCall {
	var newMDLC *admin.MobiledevicesListCall

	newMDLC = mdlc.SortOrder(sortorder)

	return newMDLC
}

// DoGet calls the .Do() function on the admin.MobiledevicesGetCall
func DoGet(mdgc *admin.MobiledevicesGetCall) (*admin.MobileDevice, error) {
	mobdev, err := mdgc.Do()
	if err != nil {
		return nil, err
	}

	return mobdev, nil
}

// DoList calls the .Do() function on the admin.MobiledevicesListCall
func DoList(mdlc *admin.MobiledevicesListCall) (*admin.MobileDevices, error) {
	mobdevs, err := mdlc.Do()
	if err != nil {
		return nil, err
	}

	return mobdevs, nil
}

// ShowAttrs displays requested chromeOS device attributes
func ShowAttrs(filter string) {
	for _, a := range mobDevAttrs {
		lwrA := strings.ToLower(a)
		comp, _ := cmn.IsValidAttr(lwrA, mobDevCompAttrs)
		if filter == "" {
			if comp != "" {
				fmt.Println("* ", a)
			} else {
				fmt.Println(a)
			}
			continue
		}

		if strings.Contains(lwrA, strings.ToLower(filter)) {
			if comp != "" {
				fmt.Println("* ", a)
			} else {
				fmt.Println(a)
			}
		}

	}
}

// ShowAttrValues displays enumerated attribute values
func ShowAttrValues(lenArgs int, args []string, filter string) error {
	if lenArgs > 2 {
		return fmt.Errorf(gmess.ERRTOOMANYARGSMAX2, "mobiledevice")
	}

	if lenArgs == 1 {
		cmn.ShowAttrVals(attrValues, filter)
	}

	if lenArgs == 2 {
		attr := strings.ToLower(args[1])

		if attr == "action" {
			cmn.ShowAttrVals(ValidActions, filter)
		} else {
			return fmt.Errorf(gmess.ERRATTRNOTRECOGNIZED, args[1])
		}
	}

	return nil
}

// ShowCompAttrs displays chromeOS device composite attributes
func ShowCompAttrs(filter string) {
	keys := make([]string, 0, len(mobDevCompAttrs))
	for k := range mobDevCompAttrs {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		if filter == "" {
			fmt.Println(mobDevCompAttrs[k])
			continue
		}

		if strings.Contains(k, strings.ToLower(filter)) {
			fmt.Println(mobDevCompAttrs[k])
		}

	}
}

// ShowFlagValues displays enumerated flag values
func ShowFlagValues(lenArgs int, args []string, filter string) error {
	if lenArgs == 1 {
		cmn.ShowFlagValues(flagValues, filter)
	}

	if lenArgs == 2 {
		flag := strings.ToLower(args[1])
		valSlice := []string{}

		switch {
		case flag == "order-by":
			for _, val := range ValidOrderByStrs {
				s, _ := cmn.IsValidAttr(val, MobDevAttrMap)
				if s == "" {
					s = val
				}
				valSlice = append(valSlice, s)
			}
			uniqueSlice := cmn.UniqueStrSlice(valSlice)
			cmn.ShowFlagValues(uniqueSlice, filter)
		case flag == "projection":
			cmn.ShowFlagValues(ValidProjections, filter)
		case flag == "sort-order":
			for _, v := range cmn.ValidSortOrders {
				valSlice = append(valSlice, v)
			}
			uniqueSlice := cmn.UniqueStrSlice(valSlice)
			cmn.ShowFlagValues(uniqueSlice, filter)
		default:
			return fmt.Errorf(gmess.ERRFLAGNOTRECOGNIZED, args[1])
		}
	}

	return nil
}

// ShowSubAttrs displays attributes of composite attributes
func ShowSubAttrs(compAttr string, filter string) error {
	lwrCompAttr := strings.ToLower(compAttr)
	switch lwrCompAttr {
	case "applications":
		cmn.ShowAttrs(mobDevApplicationsAttrs, MobDevAttrMap, filter)
	default:
		return fmt.Errorf(gmess.ERRNOTCOMPOSITEATTR, compAttr)
	}

	return nil
}
