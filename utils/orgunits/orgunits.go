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
	KEYNAME string = "ouKey"
	// STARTORGUNITSFIELD is List call attribute string prefix
	STARTORGUNITSFIELD string = "organizationUnits("
)

var flagValues = []string{
	"type",
}

// OrgUnitAttrMap provides lowercase mappings to valid admin.OrgUnit attributes
var OrgUnitAttrMap = map[string]string{
	"blockinheritance":  "blockInheritance",
	"description":       "description",
	"etag":              "etag",
	"forcesendfields":   "forceSendFields",
	"kind":              "kind",
	"name":              "name",
	"orgunitid":         "orgUnitId",
	"orgunitpath":       "orgUnitPath",
	"oukey":             "ouKey", // Used in batch commands
	"parentorgunitid":   "parentOrgUnitId",
	"parentorgunitpath": "parentOrgUnitPath",
}

// ValidSearchTypes provides list of valid types for admin.OrgunitsListCall
var ValidSearchTypes = []string{
	"all",
	"children",
}

// Key is struct used to extract ouKey
type Key struct {
	OUKey string
}

// AddFields adds fields to be returned to admin calls
func AddFields(callObj interface{}, attrs string) interface{} {
	lg.Debugw("starting AddFields()",
		"attrs", attrs)
	defer lg.Debug("finished AddFields()")

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
	lg.Debugw("starting AddOUPath()",
		"path", path)
	defer lg.Debug("finished AddOUPath()")

	var newOULC *admin.OrgunitsListCall

	newOULC = oulc.OrgUnitPath(path)

	return newOULC
}

// AddType adds Type to admin calls
func AddType(oulc *admin.OrgunitsListCall, searchType string) *admin.OrgunitsListCall {
	lg.Debugw("starting AddType()",
		"searchType", searchType)
	defer lg.Debug("finished AddType()")

	var newOULC *admin.OrgunitsListCall

	newOULC = oulc.Type(searchType)

	return newOULC
}

// DoGet calls the .Do() function on the admin.OrgunitsGetCall
func DoGet(ougc *admin.OrgunitsGetCall) (*admin.OrgUnit, error) {
	lg.Debug("starting DoGet()")
	defer lg.Debug("finished DoGet()")

	orgUnit, err := ougc.Do()
	if err != nil {
		lg.Error(err)
		return nil, err
	}

	return orgUnit, nil
}

// DoList calls the .Do() function on the admin.OrgunitsListCall
func DoList(oulc *admin.OrgunitsListCall) (*admin.OrgUnits, error) {
	lg.Debug("starting DoList()")
	defer lg.Debug("finished DoList()")

	orgunits, err := oulc.Do()
	if err != nil {
		lg.Error(err)
		return nil, err
	}

	return orgunits, nil
}

// PopulateOrgUnit is used in batch processing
func PopulateOrgUnit(orgunit *admin.OrgUnit, hdrMap map[int]string, objData []interface{}) error {
	lg.Debugw("starting populateMember()",
		"hdrMap", hdrMap)
	defer lg.Debug("finished populateMember()")

	for idx, attr := range objData {
		attrName := hdrMap[idx]
		attrVal := fmt.Sprintf("%v", attr)
		lowerAttrVal := strings.ToLower(fmt.Sprintf("%v", attr))

		switch {
		case attrName == "blockInheritance":
			if lowerAttrVal == "true" {
				orgunit.BlockInheritance = true
			}
		case attrName == "description":
			orgunit.Description = attrVal
		case attrName == "name":
			if attrVal == "" {
				err := fmt.Errorf(gmess.ERR_EMPTYSTRING, attrName)
				lg.Error(err)
				return err
			}
			orgunit.Name = attrVal
		case attrName == "parentOrgUnitPath":
			if attrVal == "" {
				err := fmt.Errorf(gmess.ERR_EMPTYSTRING, attrName)
				lg.Error(err)
				return err
			}
			orgunit.ParentOrgUnitPath = attrVal
		default:
			err := fmt.Errorf(gmess.ERR_ATTRNOTRECOGNIZED, attrName)
			return err
		}
	}
	return nil
}

// ShowAttrs displays requested orgunit attributes
func ShowAttrs(filter string) {
	lg.Debug("starting ShowAttrs()",
		"filter", filter)
	defer lg.Debug("finished ShowAttrs()")

	keys := make([]string, 0, len(OrgUnitAttrMap))
	for k := range OrgUnitAttrMap {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		if filter == "" {
			fmt.Println(OrgUnitAttrMap[k])
			continue
		}

		if strings.Contains(k, strings.ToLower(filter)) {
			fmt.Println(OrgUnitAttrMap[k])
		}

	}
}

// ShowFlagValues displays enumerated flag values
func ShowFlagValues(lenArgs int, args []string, filter string) error {
	lg.Debug("starting ShowFlagValues()",
		"lenArgs", lenArgs,
		"args", args,
		"filter", filter)
	defer lg.Debug("finished ShowFlagValues()")

	if lenArgs == 1 {
		cmn.ShowFlagValues(flagValues, filter)
	}

	if lenArgs == 2 {
		flag := strings.ToLower(args[1])

		switch {
		case flag == "type":
			cmn.ShowFlagValues(ValidSearchTypes, filter)
		default:
			err := fmt.Errorf(gmess.ERR_FLAGNOTRECOGNIZED, args[1])
			lg.Error(err)
			return err
		}
	}

	return nil
}
