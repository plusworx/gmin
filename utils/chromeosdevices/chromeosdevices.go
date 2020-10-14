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

package chromeosdevices

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
	// STARTCHROMEDEVICESFIELD is List call attribute string prefix
	STARTCHROMEDEVICESFIELD string = "chromeosdevices("
)

// ManagedDevice is struct to extract device data
type ManagedDevice struct {
	Action            string
	DeviceId          string
	DeprovisionReason string
}

// MovedDevice is struct to extract device data
type MovedDevice struct {
	DeviceId    string
	OrgUnitPath string
}

var attrValues = []string{
	"action",
}

// crOsDevActiveTimeRangesAttrs contains names of all the addressable admin.ChromeOsDeviceActiveTimeRanges attributes
var crOsDevActiveTimeRangesAttrs = []string{
	"activetime",
	"date",
}

// crOsDevCPUStatusReportsAttrs contains names of all the addressable admin.ChromeOsDeviceCpuStatusReports attributes
var crOsDevCPUStatusReportsAttrs = []string{
	"cputemperatureinfo",
	"cpuutilizationpercentageinfo",
	"reporttime",
}

// crOsDevCPUStatusReportsCPUTempInfoAttrs contains names of all the addressable admin.ChromeOsDeviceCpuStatusReportsCpuTemperatureInfo attributes
var crOsDevCPUStatusReportsCPUTempInfoAttrs = []string{
	"label",
	"temperature",
}

// crOsDevDeviceFilesAttrs contains names of all the addressable admin.ChromeOsDeviceDeviceFiles attributes
var crOsDevDeviceFilesAttrs = []string{
	"createtime",
	"downloadurl",
	"name",
	"type",
}

// crOsDevDiskVolReportsAttrs contains names of all the addressable admin.ChromeOsDeviceDiskVolumeReports attributes
var crOsDevDiskVolReportsAttrs = []string{
	"volumeinfo",
}

// crOsDevDiskVolReportsVolInfoAttrs contains names of all the addressable admin.ChromeOsDeviceDiskVolumeReportsVolumeInfo attributes
var crOsDevDiskVolReportsVolInfoAttrs = []string{
	"storagefree",
	"storagetotal",
	"volumeid",
}

// crOsDevLastKnownNetwork contains names of all the addressable admin.ChromeOsDeviceLastKnownNetwork attributes
var crOsDevLastKnownNetworkAttrs = []string{
	"ipaddress",
	"wanipaddress",
}

// crOsDevRecentUsersAttrs contains names of all the addressable admin.ChromeOsDeviceRecentUsers attributes
var crOsDevRecentUsersAttrs = []string{
	"email",
	"type",
}

// crOsDevSystemRAMFreeReportsAttrs contains names of all the addressable admin.ChromeOsDeviceSystemRamFreeReports attributes
var crOsDevSystemRAMFreeReportsAttrs = []string{
	"reporttime",
	"systemramfreeinfo",
}

// crOsDevTpmVersionInfoAttrs contains names of all the addressable admin.ChromeOsDeviceTpmVersionInfo attributes
var crOsDevTpmVersionInfoAttrs = []string{
	"family",
	"firmwareversion",
	"manufacturer",
	"speclevel",
	"tpmmodel",
	"vendorspecific",
}

// QueryAttrMap provides lowercase mappings to valid admin.ChromeOsDevice query attributes
var QueryAttrMap = map[string]string{
	"assetid":      "asset_id",
	"asset_id":     "asset_id",
	"ethernetmac":  "ethernet_mac",
	"ethernet_mac": "ethernet_mac",
	"id":           "id",
	"location":     "location",
	"note":         "note",
	"recentuser":   "recent_user",
	"recent_user":  "recent_user",
	"register":     "register",
	"status":       "status",
	"sync":         "sync",
	"user":         "user",
	"wifimac":      "wifi_mac",
	"wifi_mac":     "wifi_mac",
}

// CrOSDevAttrMap provides lowercase mappings to valid admin.ChromeOsDevice attributes
var CrOSDevAttrMap = map[string]string{
	"action":                       "action", // used in batch manage
	"activetime":                   "activeTime",
	"activetimeranges":             "activeTimeRanges",
	"annotatedassetid":             "annotatedAssetId",
	"annotatedlocation":            "annotatedLocation",
	"annotateduser":                "annotatedUser",
	"autoupdateexpiration":         "autoUpdateExpiration",
	"bootmode":                     "bootMode",
	"cputemperatureinfo":           "cpuTemperatureInfo",
	"cpuutilizationpercentageinfo": "cpuUtilizationPercentageInfo",
	"cpustatusreports":             "cpuStatusReports",
	"createtime":                   "createTime",
	"date":                         "date",
	"deprovisionreason":            "deprovisionReason", // used in batch manage
	"devicefiles":                  "deviceFiles",
	"deviceid":                     "deviceId",
	"diskvolumereports":            "diskVolumeReports",
	"dockmacaddress":               "dockMacAddress",
	"downloadurl":                  "downloadUrl",
	"email":                        "email",
	"etag":                         "etag",
	"ethernetmacaddress":           "ethernetMacAddress",
	"ethernetmacaddress0":          "ethernetMacAddress0",
	"family":                       "family",
	"firmwareversion":              "firmwareVersion",
	"forcesendfields":              "forceSendFields",
	"ipaddress":                    "ipAddress",
	"kind":                         "kind",
	"label":                        "label",
	"lastenrollmenttime":           "lastEnrollmentTime",
	"lastknownnetwork":             "lastKnownNetwork",
	"lastsync":                     "lastSync",
	"macaddress":                   "macAddress",
	"manufacturedate":              "manufactureDate",
	"manufacturer":                 "manufacturer",
	"meid":                         "meid",
	"model":                        "model",
	"name":                         "name",
	"notes":                        "notes",
	"ordernumber":                  "orderNumber",
	"orgunitpath":                  "orgUnitPath",
	"osversion":                    "osVersion",
	"platformversion":              "platformVersion",
	"recentusers":                  "recentUsers",
	"reporttime":                   "reportTime",
	"serialnumber":                 "serialNumber",
	"speclevel":                    "specLevel",
	"status":                       "status",
	"storagefree":                  "storageFree",
	"storagetotal":                 "storageTotal",
	"supportenddate":               "supportEndDate",
	"systemramfreeinfo":            "systemRamFreeInfo",
	"systemramfreereports":         "systemRamFreeReports",
	"systemramtotal":               "systemRamTotal",
	"suspensionreason":             "suspensionReason",
	"temperature":                  "temperature",
	"tpmmodel":                     "tpmModel",
	"tpmversioninfo":               "tpmVersionInfo",
	"type":                         "type",
	"vendorspecific":               "vendorSpecific",
	"volumeid":                     "volumeId",
	"volumeinfo":                   "volumeInfo",
	"wanipaddress":                 "wanIpAddress",
	"willautorenew":                "willAutoRenew",
}

var crOSDevAttrs = []string{
	"activeTimeRanges",
	"annotatedAssetId",
	"annotatedLocation",
	"annotatedUser",
	"autoUpdateExpiration",
	"bootMode",
	"cpuStatusReports",
	"deviceFiles",
	"deviceId",
	"diskVolumeReports",
	"dockMacAddress",
	"etag",
	"ethernetMacAddress",
	"ethernetMacAddress0",
	"firmwareVersion",
	"forceSendFields",
	"kind",
	"lastEnrollmentTime",
	"lastKnownNetwork",
	"lastSync",
	"macAddress",
	"manufactureDate",
	"meid",
	"model",
	"notes",
	"orderNumber",
	"orgUnitPath",
	"osVersion",
	"platformVersion",
	"recentUsers",
	"serialNumber",
	"status",
	"supportEndDate",
	"systemRamFreeReports",
	"systemRamTotal",
	"tpmVersionInfo",
	"willAutoRenew",
}

var crOSDevCompAttrs = map[string]string{
	"activetimeranges":     "activeTimeRanges",
	"cpustatusreports":     "cpuStatusReports",
	"devicefiles":          "deviceFiles",
	"diskvolumereports":    "diskVolumeReports",
	"lastknownnetwork":     "lastKnownNetwork",
	"recentusers":          "recentUsers",
	"systemramfreereports": "systemRamFreeReports",
	"tpmversioninfo":       "tpmVersionInfo",
}

var flagValues = []string{
	"order-by",
	"projection",
	"reason",
	"sort-order",
}

// ValidActions provide valid strings to be used for admin.ChromeosdevicesActionCall
var ValidActions = []string{
	"deprovision",
	"disable",
	"reenable",
}

// ValidDeprovisionReasons provide valid strings to be used for admin.ChromeosdevicesActionCall
var ValidDeprovisionReasons = []string{
	"different_model_replacement",
	"retiring_device",
	"same_model_replacement",
	"upgrade_transfer",
}

// ValidOrderByStrs provide valid strings to be used to set admin.ChromeosdevicesListCall OrderBy
var ValidOrderByStrs = []string{
	"annotatedlocation",
	"annotateduser",
	"lastsync",
	"notes",
	"serialnumber",
	"status",
	"supportenddate",
}

// ValidProjections provide valid strings to be used to set admin.ChromeosdevicesListCall Projection
var ValidProjections = []string{
	"basic",
	"full",
}

// AddFields adds fields to be returned from admin calls
func AddFields(callObj interface{}, attrs string) interface{} {
	var fields googleapi.Field = googleapi.Field(attrs)

	switch callObj.(type) {
	case *admin.ChromeosdevicesActionCall:
		var newCDAC *admin.ChromeosdevicesActionCall
		cdac := callObj.(*admin.ChromeosdevicesActionCall)
		newCDAC = cdac.Fields(fields)

		return newCDAC
	case *admin.ChromeosdevicesGetCall:
		var newCDGC *admin.ChromeosdevicesGetCall
		cdgc := callObj.(*admin.ChromeosdevicesGetCall)
		newCDGC = cdgc.Fields(fields)

		return newCDGC
	case *admin.ChromeosdevicesListCall:
		var newCDLC *admin.ChromeosdevicesListCall
		cdlc := callObj.(*admin.ChromeosdevicesListCall)
		newCDLC = cdlc.Fields(fields)

		return newCDLC
	case *admin.ChromeosdevicesMoveDevicesToOuCall:
		var newCDMC *admin.ChromeosdevicesMoveDevicesToOuCall
		cdmc := callObj.(*admin.ChromeosdevicesMoveDevicesToOuCall)
		newCDMC = cdmc.Fields(fields)

		return newCDMC
	case *admin.ChromeosdevicesUpdateCall:
		var newCDUC *admin.ChromeosdevicesUpdateCall
		cduc := callObj.(*admin.ChromeosdevicesUpdateCall)
		newCDUC = cduc.Fields(fields)

		return newCDUC
	}
	return nil
}

// AddMaxResults adds MaxResults to admin calls
func AddMaxResults(cdlc *admin.ChromeosdevicesListCall, maxResults int64) *admin.ChromeosdevicesListCall {
	var newCDLC *admin.ChromeosdevicesListCall

	newCDLC = cdlc.MaxResults(maxResults)

	return newCDLC
}

// AddOrderBy adds OrderBy to admin calls
func AddOrderBy(cdlc *admin.ChromeosdevicesListCall, orderBy string) *admin.ChromeosdevicesListCall {
	var newCDLC *admin.ChromeosdevicesListCall

	newCDLC = cdlc.OrderBy(orderBy)

	return newCDLC
}

// AddOrgUnitPath adds OrgUnitPath to admin calls
func AddOrgUnitPath(cdlc *admin.ChromeosdevicesListCall, orgUnitPath string) *admin.ChromeosdevicesListCall {
	var newCDLC *admin.ChromeosdevicesListCall

	newCDLC = cdlc.OrgUnitPath(orgUnitPath)

	return newCDLC
}

// AddPageToken adds PageToken to admin calls
func AddPageToken(cdlc *admin.ChromeosdevicesListCall, token string) *admin.ChromeosdevicesListCall {
	var newCDLC *admin.ChromeosdevicesListCall

	newCDLC = cdlc.PageToken(token)

	return newCDLC
}

// AddProjection adds Projection to admin calls
func AddProjection(callObj interface{}, projection string) interface{} {
	switch callObj.(type) {
	case *admin.ChromeosdevicesGetCall:
		var newCDGC *admin.ChromeosdevicesGetCall
		cdgc := callObj.(*admin.ChromeosdevicesGetCall)
		newCDGC = cdgc.Projection(projection)

		return newCDGC
	case *admin.ChromeosdevicesListCall:
		var newCDLC *admin.ChromeosdevicesListCall
		cdlc := callObj.(*admin.ChromeosdevicesListCall)
		newCDLC = cdlc.Projection(projection)

		return newCDLC
	case *admin.ChromeosdevicesUpdateCall:
		var newCDUC *admin.ChromeosdevicesUpdateCall
		cduc := callObj.(*admin.ChromeosdevicesUpdateCall)
		newCDUC = cduc.Projection(projection)

		return newCDUC
	}

	return nil
}

// AddQuery adds query to admin calls
func AddQuery(cdlc *admin.ChromeosdevicesListCall, query string) *admin.ChromeosdevicesListCall {
	var newCDLC *admin.ChromeosdevicesListCall

	newCDLC = cdlc.Query(query)

	return newCDLC
}

// AddSortOrder adds SortOrder to admin calls
func AddSortOrder(cdlc *admin.ChromeosdevicesListCall, sortorder string) *admin.ChromeosdevicesListCall {
	var newCDLC *admin.ChromeosdevicesListCall

	newCDLC = cdlc.SortOrder(sortorder)

	return newCDLC
}

// DoGet calls the .Do() function on the admin.ChromeosdevicesGetCall
func DoGet(cdgc *admin.ChromeosdevicesGetCall) (*admin.ChromeOsDevice, error) {
	crosdev, err := cdgc.Do()
	if err != nil {
		return nil, err
	}

	return crosdev, nil
}

// DoList calls the .Do() function on the admin.ChromeosdevicesListCall
func DoList(cdlc *admin.ChromeosdevicesListCall) (*admin.ChromeOsDevices, error) {
	crosdevs, err := cdlc.Do()
	if err != nil {
		return nil, err
	}

	return crosdevs, nil
}

// ShowAttrs displays requested chromeOS device attributes
func ShowAttrs(filter string) {
	for _, a := range crOSDevAttrs {
		lwrA := strings.ToLower(a)
		comp, _ := cmn.IsValidAttr(lwrA, crOSDevCompAttrs)
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
		return fmt.Errorf(gmess.ERR_TOOMANYARGSMAX1, args[0])
	}

	if lenArgs == 1 {
		cmn.ShowAttrVals(attrValues, filter)
	}

	if lenArgs == 2 {
		attr := strings.ToLower(args[1])

		if attr == "action" {
			cmn.ShowAttrVals(ValidActions, filter)
		} else {
			return fmt.Errorf(gmess.ERR_ATTRNOTRECOGNIZED, args[1])
		}
	}

	return nil
}

// ShowCompAttrs displays chromeOS device composite attributes
func ShowCompAttrs(filter string) {
	keys := make([]string, 0, len(crOSDevCompAttrs))
	for k := range crOSDevCompAttrs {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		if filter == "" {
			fmt.Println(crOSDevCompAttrs[k])
			continue
		}

		if strings.Contains(k, strings.ToLower(filter)) {
			fmt.Println(crOSDevCompAttrs[k])
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
				s, _ := cmn.IsValidAttr(val, CrOSDevAttrMap)
				if s == "" {
					s = val
				}
				valSlice = append(valSlice, s)
			}
			uniqueSlice := cmn.UniqueStrSlice(valSlice)
			cmn.ShowFlagValues(uniqueSlice, filter)
		case flag == "projection":
			cmn.ShowFlagValues(ValidProjections, filter)
		case flag == "reason":
			cmn.ShowFlagValues(ValidDeprovisionReasons, filter)
		case flag == "sort-order":
			for _, v := range cmn.ValidSortOrders {
				valSlice = append(valSlice, v)
			}
			uniqueSlice := cmn.UniqueStrSlice(valSlice)
			cmn.ShowFlagValues(uniqueSlice, filter)
		default:
			return fmt.Errorf(gmess.ERR_FLAGNOTRECOGNIZED, args[1])
		}
	}

	return nil
}

// ShowSubAttrs displays attributes of composite attributes
func ShowSubAttrs(compAttr string, filter string) error {
	lwrCompAttr := strings.ToLower(compAttr)
	switch lwrCompAttr {
	case "activetimeranges":
		cmn.ShowAttrs(crOsDevActiveTimeRangesAttrs, CrOSDevAttrMap, filter)
	case "cpustatusreports":
		cmn.ShowAttrs(crOsDevCPUStatusReportsAttrs, CrOSDevAttrMap, filter)
	case "devicefiles":
		cmn.ShowAttrs(crOsDevDeviceFilesAttrs, CrOSDevAttrMap, filter)
	case "diskvolumereports":
		cmn.ShowAttrs(crOsDevDiskVolReportsAttrs, CrOSDevAttrMap, filter)
	case "lastknownnetwork":
		cmn.ShowAttrs(crOsDevLastKnownNetworkAttrs, CrOSDevAttrMap, filter)
	case "recentusers":
		cmn.ShowAttrs(crOsDevRecentUsersAttrs, CrOSDevAttrMap, filter)
	case "systemramfreereports":
		cmn.ShowAttrs(crOsDevSystemRAMFreeReportsAttrs, CrOSDevAttrMap, filter)
	case "tpmversioninfo":
		cmn.ShowAttrs(crOsDevTpmVersionInfoAttrs, CrOSDevAttrMap, filter)
	default:
		return fmt.Errorf(gmess.ERR_NOTCOMPOSITEATTR, compAttr)
	}

	return nil
}
