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

	cmn "github.com/plusworx/gmin/utils/common"
	gmess "github.com/plusworx/gmin/utils/gminmessages"
	grps "github.com/plusworx/gmin/utils/groups"
	lg "github.com/plusworx/gmin/utils/logging"
	mems "github.com/plusworx/gmin/utils/members"
	ous "github.com/plusworx/gmin/utils/orgunits"
	usrs "github.com/plusworx/gmin/utils/users"
	admin "google.golang.org/api/admin/directory/v1"
	sheet "google.golang.org/api/sheets/v4"
)

// CreateFromFileFactory produces objects from input file data
func CreateFromFileFactory(outType int, hdrMap map[int]string, objData []interface{}) (interface{}, error) {
	lg.Debugw("starting FromFileFactory()",
		"hdrMap", hdrMap)
	defer lg.Debug("finished FromFileFactory()")

	switch outType {
	case cmn.OBJTYPEGROUP:
		group := new(admin.Group)
		err := grps.PopulateGroup(group, hdrMap, objData)
		if err != nil {
			return nil, err
		}
		return group, nil
	case cmn.OBJTYPEMEMBER:
		member := new(admin.Member)
		err := mems.PopulateMember(member, hdrMap, objData)
		if err != nil {
			return nil, err
		}
		return member, nil
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
		err := fmt.Errorf(gmess.ERR_OBJECTNOTRECOGNIZED, outType)
		lg.Error(err)
		return nil, err
	}
}

// CreateFromJSONFactory creates object from JSON data
func CreateFromJSONFactory(outType int, jsonData string, attrMap map[string]string) (interface{}, error) {
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

	switch outType {
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
		err := fmt.Errorf(gmess.ERR_OBJECTNOTRECOGNIZED, outType)
		lg.Error(err)
		return nil, err
	}
}

// CreateProcessCSVFile does batch processing of CSV input files
func CreateProcessCSVFile(outType int, filePath string, attrMap map[string]string) ([]interface{}, error) {
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

		objVar, err := CreateFromFileFactory(outType, hdrMap, iSlice)
		if err != nil {
			return nil, err
		}

		outputObjs = append(outputObjs, objVar)

		count = count + 1
	}

	return outputObjs, nil
}

// CreateProcessGSheet does batch processing of Google Sheet input
func CreateProcessGSheet(outType int, sheetID string, sheetrange string, attrMap map[string]string) ([]interface{}, error) {
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

	ss, err := cmn.CreateSheetService(sheet.DriveReadonlyScope)
	if err != nil {
		return nil, err
	}

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

		objVar, err := CreateFromFileFactory(outType, hdrMap, row)
		if err != nil {
			return nil, err
		}

		outputObjs = append(outputObjs, objVar)
	}

	return outputObjs, nil
}

// CreateProcessJSON does batch processing of JSON file input
func CreateProcessJSON(outType int, filePath string, scanner *bufio.Scanner, attrMap map[string]string) ([]interface{}, error) {
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

		objVar, err := CreateFromJSONFactory(outType, jsonData, attrMap)
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
