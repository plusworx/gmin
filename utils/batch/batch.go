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

package batch

import (
	"bufio"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"

	cdevs "github.com/plusworx/gmin/utils/chromeosdevices"
	cmn "github.com/plusworx/gmin/utils/common"
	gmess "github.com/plusworx/gmin/utils/gminmessages"
	grps "github.com/plusworx/gmin/utils/groups"
	grpset "github.com/plusworx/gmin/utils/groupsettings"
	lg "github.com/plusworx/gmin/utils/logging"
	mems "github.com/plusworx/gmin/utils/members"
	mdevs "github.com/plusworx/gmin/utils/mobiledevices"
	ous "github.com/plusworx/gmin/utils/orgunits"
	usrs "github.com/plusworx/gmin/utils/users"
	admin "google.golang.org/api/admin/directory/v1"
	sheet "google.golang.org/api/sheets/v4"
)

// CallParams holds batch call parameters
type CallParams struct {
	CallType   int
	ObjectType int
}

// DeleteFromFileFactory produces objects from input file data
func DeleteFromFileFactory(hdrMap map[int]string, objData []interface{}, keyName string) (string, error) {
	lg.Debugw("starting DeleteFromFileFactory()",
		"hdrMap", hdrMap)
	defer lg.Debug("finished DeleteFromFileFactory()")

	var outStr string

	for idx, val := range objData {
		attrName := hdrMap[idx]
		attrVal := fmt.Sprintf("%v", val)

		if attrName == keyName {
			outStr = attrVal
		}
	}
	return outStr, nil
}

// DeleteProcessGSheet does batch processing of Google Sheet input
func DeleteProcessGSheet(sheetID string, sheetrange string, attrMap map[string]string, keyName string) ([]string, error) {
	lg.Debugw("starting DeleteProcessGSheet()",
		"sheetID", sheetID,
		"sheetrange", sheetrange)
	defer lg.Debug("finished DeleteProcessGSheet()")

	var outputObjs []string

	if sheetrange == "" {
		err := errors.New(gmess.ERR_NOSHEETRANGE)
		lg.Error(err)
		return nil, err
	}

	srv, err := cmn.CreateService(cmn.SRVTYPESHEET, sheet.DriveReadonlyScope)
	if err != nil {
		lg.Error(err)
		return nil, err
	}
	ss := srv.(*sheet.Service)

	ssvgc := ss.Spreadsheets.Values.Get(sheetID, sheetrange)
	sValRange, err := ssvgc.Do()
	if err != nil {
		lg.Error(err)
		return nil, err
	}

	if len(sValRange.Values) == 0 {
		err = fmt.Errorf(gmess.ERR_NOSHEETDATAFOUND, sheetID, sheetrange)
		lg.Error(err)
		return nil, err
	}

	hdrMap := cmn.ProcessHeader(sValRange.Values[0])
	err = cmn.ValidateHeader(hdrMap, attrMap)
	if err != nil {
		return nil, err
	}

	for idx, row := range sValRange.Values {
		if idx == 0 {
			continue
		}

		objVar, err := DeleteFromFileFactory(hdrMap, row, keyName)
		if err != nil {
			return nil, err
		}

		outputObjs = append(outputObjs, objVar)
	}

	return outputObjs, nil
}

// DeleteProcessTextFile does batch processing of text input
func DeleteProcessTextFile(filePath string, scanner *bufio.Scanner) ([]string, error) {
	lg.Debugw("starting DeleteProcessTextFile()",
		"filePath", filePath)
	defer lg.Debug("finished DeleteProcessTextFile()")

	var outputObjs []string

	if filePath != "" {
		file, err := os.Open(filePath)
		if err != nil {
			lg.Error(err)
			return nil, err
		}
		defer file.Close()
		scanner = bufio.NewScanner(file)
	}

	for scanner.Scan() {
		obj := scanner.Text()
		outputObjs = append(outputObjs, obj)
	}

	return outputObjs, nil
}

// FromFileFactory produces objects from input file data
func FromFileFactory(callParams CallParams, hdrMap map[int]string, objData []interface{}) (interface{}, error) {
	lg.Debugw("starting FromFileFactory()",
		"hdrMap", hdrMap)
	defer lg.Debug("finished FromFileFactory()")

	switch callParams.ObjectType {
	case cmn.OBJTYPECROSDEV:
		if callParams.CallType == cmn.CALLTYPEMANAGE {
			mngdev := cdevs.ManagedDevice{}
			err := cdevs.PopulateManagedDev(&mngdev, hdrMap, objData)
			if err != nil {
				return nil, err
			}
			return mngdev, nil
		}
		if callParams.CallType == cmn.CALLTYPEMOVE {
			mvdev := cdevs.MovedDevice{}
			err := cdevs.PopulateMovedDev(&mvdev, hdrMap, objData)
			if err != nil {
				return nil, err
			}
			return mvdev, nil
		}
	case cmn.OBJTYPEGROUP:
		group := new(admin.Group)
		err := grps.PopulateGroup(group, hdrMap, objData)
		if err != nil {
			return nil, err
		}
		return group, nil
	case cmn.OBJTYPEGRPSET:
		grpParams := grpset.GroupParams{}
		err := grpset.PopulateGroupSettings(&grpParams, hdrMap, objData)
		if err != nil {
			return nil, err
		}
		return grpParams, nil
	case cmn.OBJTYPEMEMBER:
		member := new(admin.Member)
		err := mems.PopulateMember(member, hdrMap, objData)
		if err != nil {
			return nil, err
		}
		return member, nil
	case cmn.OBJTYPEMOBDEV:
		if callParams.CallType == cmn.CALLTYPEMANAGE {
			mngdev := mdevs.ManagedDevice{}
			err := mdevs.PopulateManagedDev(&mngdev, hdrMap, objData)
			if err != nil {
				return nil, err
			}
			return mngdev, nil
		}
	case cmn.OBJTYPEORGUNIT:
		orgunit := new(admin.OrgUnit)
		err := ous.PopulateOrgUnit(orgunit, hdrMap, objData)
		if err != nil {
			return nil, err
		}
		return orgunit, nil
	case cmn.OBJTYPEUSER:
		user := new(admin.User)
		err := usrs.PopulateUser(user, hdrMap, objData)
		if err != nil {
			return nil, err
		}
		return user, nil
	default:
		err := fmt.Errorf(gmess.ERR_OBJECTNOTRECOGNIZED, callParams.ObjectType)
		lg.Error(err)
		return nil, err
	}
	// Only gets here if ChromeOSDev or MobDev call type invalid
	err := fmt.Errorf(gmess.ERR_CALLTYPENOTRECOGNIZED, callParams.CallType)
	lg.Error(err)
	return nil, err
}

// FromJSONFactory creates object from JSON data
func FromJSONFactory(callParam CallParams, jsonData string, attrMap map[string]string) (interface{}, error) {
	lg.Debugw("starting FromJSONFactory()",
		"jsonData", jsonData)
	defer lg.Debug("finished FromJSONFactory()")

	emptyVals := cmn.EmptyValues{}
	jsonBytes := []byte(jsonData)

	if !json.Valid(jsonBytes) {
		err := errors.New(gmess.ERR_INVALIDJSONATTR)
		lg.Error(err)
		return nil, err
	}

	outStr, err := cmn.ParseInputAttrs(jsonBytes)
	if err != nil {
		return nil, err
	}

	err = cmn.ValidateInputAttrs(outStr, attrMap)
	if err != nil {
		return nil, err
	}

	switch callParam.ObjectType {
	case cmn.OBJTYPECROSDEV:
		if callParam.CallType == cmn.CALLTYPEMANAGE {
			mngDev := cdevs.ManagedDevice{}
			err = json.Unmarshal(jsonBytes, &mngDev)
			if err != nil {
				lg.Error(err)
				return nil, err
			}
			return mngDev, nil
		}
		if callParam.CallType == cmn.CALLTYPEMOVE {
			mvDev := cdevs.MovedDevice{}
			err = json.Unmarshal(jsonBytes, &mvDev)
			if err != nil {
				lg.Error(err)
				return nil, err
			}
			return mvDev, nil
		}
	case cmn.OBJTYPEGROUP:
		group := new(admin.Group)
		err = json.Unmarshal(jsonBytes, &group)
		if err != nil {
			lg.Error(err)
			return nil, err
		}
		err = json.Unmarshal(jsonBytes, &emptyVals)
		if err != nil {
			lg.Error(err)
			return nil, err
		}
		if len(emptyVals.ForceSendFields) > 0 {
			group.ForceSendFields = emptyVals.ForceSendFields
		}
		return group, nil
	case cmn.OBJTYPEGRPSET:
		var (
			grpKey    = grpset.Key{}
			grpParams = grpset.GroupParams{}
		)

		err = json.Unmarshal(jsonBytes, &grpKey)
		if err != nil {
			lg.Error(err)
			return nil, err
		}

		if grpKey.GroupKey == "" {
			err = errors.New(gmess.ERR_NOJSONGROUPKEY)
			lg.Error(err)
			return nil, err
		}
		grpParams.GroupKey = grpKey.GroupKey

		err = json.Unmarshal(jsonBytes, &grpParams.Settings)
		if err != nil {
			lg.Error(err)
			return nil, err
		}
		return grpParams, nil
	case cmn.OBJTYPEMEMBER:
		member := new(admin.Member)
		err = json.Unmarshal(jsonBytes, &member)
		if err != nil {
			lg.Error(err)
			return nil, err
		}
		err = json.Unmarshal(jsonBytes, &emptyVals)
		if err != nil {
			lg.Error(err)
			return nil, err
		}
		if len(emptyVals.ForceSendFields) > 0 {
			member.ForceSendFields = emptyVals.ForceSendFields
		}
		return member, nil
	case cmn.OBJTYPEMOBDEV:
		if callParam.CallType == cmn.CALLTYPEMANAGE {
			mngDev := mdevs.ManagedDevice{}
			err = json.Unmarshal(jsonBytes, &mngDev)
			if err != nil {
				lg.Error(err)
				return nil, err
			}
			return mngDev, nil
		}
	case cmn.OBJTYPEORGUNIT:
		orgunit := new(admin.OrgUnit)
		err = json.Unmarshal(jsonBytes, &orgunit)
		if err != nil {
			lg.Error(err)
			return nil, err
		}
		err = json.Unmarshal(jsonBytes, &emptyVals)
		if err != nil {
			lg.Error(err)
			return nil, err
		}
		if len(emptyVals.ForceSendFields) > 0 {
			orgunit.ForceSendFields = emptyVals.ForceSendFields
		}
		return orgunit, nil
	case cmn.OBJTYPEUSER:
		user := new(admin.User)
		err = json.Unmarshal(jsonBytes, &user)
		if err != nil {
			lg.Error(err)
			return nil, err
		}
		err = json.Unmarshal(jsonBytes, &emptyVals)
		if err != nil {
			lg.Error(err)
			return nil, err
		}
		if len(emptyVals.ForceSendFields) > 0 {
			user.ForceSendFields = emptyVals.ForceSendFields
		}
		return user, nil
	default:
		err := fmt.Errorf(gmess.ERR_OBJECTNOTRECOGNIZED, callParam.ObjectType)
		lg.Error(err)
		return nil, err
	}
	// Only gets here if ChromeOSDev or MobDev call type invalid
	err = fmt.Errorf(gmess.ERR_CALLTYPENOTRECOGNIZED, callParam.CallType)
	lg.Error(err)
	return nil, err
}

// ProcessCSVFile does batch processing of CSV input files
func ProcessCSVFile(callParams CallParams, filePath string, attrMap map[string]string) ([]interface{}, error) {
	lg.Debugw("starting ProcessCSVFile()",
		"filePath", filePath,
		"attrMap", attrMap)
	defer lg.Debug("finished ProcessCSVFile()")

	var (
		iSlice     []interface{}
		hdrMap     = map[int]string{}
		outputObjs []interface{}
	)

	csvfile, err := os.Open(filePath)
	if err != nil {
		lg.Error(err)
		return nil, err
	}
	defer csvfile.Close()

	r := csv.NewReader(csvfile)

	count := 0
	for {
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			lg.Error(err)
			return nil, err
		}

		if count == 0 {
			iSlice = make([]interface{}, len(record))
			for idx, value := range record {
				iSlice[idx] = value
			}
			hdrMap = cmn.ProcessHeader(iSlice)
			err = validateHeader(hdrMap, attrMap)
			if err != nil {
				return nil, err
			}
			count = count + 1
			continue
		}

		for idx, value := range record {
			iSlice[idx] = value
		}

		objVar, err := FromFileFactory(callParams, hdrMap, iSlice)
		if err != nil {
			return nil, err
		}

		outputObjs = append(outputObjs, objVar)

		count = count + 1
	}

	return outputObjs, nil
}

// ProcessGSheet does batch processing of Google Sheet input
func ProcessGSheet(callParams CallParams, sheetID string, sheetrange string, attrMap map[string]string) ([]interface{}, error) {
	lg.Debugw("starting ProcessGSheet()",
		"sheetID", sheetID,
		"sheetrange", sheetrange,
		"attrMap", attrMap)
	defer lg.Debug("finished ProcessGSheet()")

	var outputObjs []interface{}

	if sheetrange == "" {
		err := errors.New(gmess.ERR_NOSHEETRANGE)
		lg.Error(err)
		return nil, err
	}

	srv, err := cmn.CreateService(cmn.SRVTYPESHEET, sheet.DriveReadonlyScope)
	if err != nil {
		lg.Error(err)
		return nil, err
	}
	ss := srv.(*sheet.Service)

	ssvgc := ss.Spreadsheets.Values.Get(sheetID, sheetrange)
	sValRange, err := ssvgc.Do()
	if err != nil {
		lg.Error(err)
		return nil, err
	}

	if len(sValRange.Values) == 0 {
		err = fmt.Errorf(gmess.ERR_NOSHEETDATAFOUND, sheetID, sheetrange)
		lg.Error(err)
		return nil, err
	}

	hdrMap := cmn.ProcessHeader(sValRange.Values[0])
	err = cmn.ValidateHeader(hdrMap, attrMap)
	if err != nil {
		return nil, err
	}

	for idx, row := range sValRange.Values {
		if idx == 0 {
			continue
		}

		objVar, err := FromFileFactory(callParams, hdrMap, row)
		if err != nil {
			return nil, err
		}

		outputObjs = append(outputObjs, objVar)
	}

	return outputObjs, nil
}

// ProcessJSON does batch processing of JSON file input
func ProcessJSON(callParam CallParams, filePath string, scanner *bufio.Scanner, attrMap map[string]string) ([]interface{}, error) {
	lg.Debugw("starting ProcessJSON()",
		"filePath", filePath,
		"attrMap", attrMap)
	defer lg.Debug("finished ProcessJSON()")

	var outputObjs []interface{}

	if filePath != "" {
		file, err := os.Open(filePath)
		if err != nil {
			lg.Error(err)
			return nil, err
		}
		defer file.Close()

		scanner = bufio.NewScanner(file)
	}

	for scanner.Scan() {
		jsonData := scanner.Text()

		objVar, err := FromJSONFactory(callParam, jsonData, attrMap)
		if err != nil {
			return nil, err
		}

		outputObjs = append(outputObjs, objVar)
	}
	err := scanner.Err()
	if err != nil {
		lg.Error(err)
		return nil, err
	}

	return outputObjs, nil
}

// validateHeader validates header column names
func validateHeader(hdr map[int]string, attrMap map[string]string) error {
	lg.Debugw("starting ValidateHeader()",
		"hdr", hdr,
		"attrMap", attrMap)
	defer lg.Debug("finished ValidateHeader()")

	for idx, hdrAttr := range hdr {
		correctVal, err := cmn.IsValidAttr(hdrAttr, attrMap)
		if err != nil {
			return err
		}
		hdr[idx] = correctVal
	}
	return nil
}
