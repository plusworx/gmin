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
	admin "google.golang.org/api/admin/directory/v1"
	"google.golang.org/api/googleapi"
)

const (
	// EndField is List call attribute string terminator
	EndField string = ")"
	// StartChromeDevicesField is List call attribute string prefix
	StartChromeDevicesField string = "chromeosdevices("
)

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
