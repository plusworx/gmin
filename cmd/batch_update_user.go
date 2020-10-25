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

package cmd

import (
	"bufio"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/cenkalti/backoff/v4"
	cmn "github.com/plusworx/gmin/utils/common"
	flgnm "github.com/plusworx/gmin/utils/flagnames"
	gmess "github.com/plusworx/gmin/utils/gminmessages"
	lg "github.com/plusworx/gmin/utils/logging"
	usrs "github.com/plusworx/gmin/utils/users"
	"github.com/spf13/cobra"
	admin "google.golang.org/api/admin/directory/v1"
	sheet "google.golang.org/api/sheets/v4"
)

var batchUpdUserCmd = &cobra.Command{
	Use:     "users -i <input file path>",
	Aliases: []string{"user"},
	Example: `gmin batch-update users -i inputfile.json
gmin bupd users -i inputfile.csv -f csv
gmin bupd user -i 1odyAIp3jGspd3M4xeepxWD6aeQIUuHBgrZB2OHSu8MI -s 'Sheet1!A1:K25' -f gsheet`,
	Short: "Updates a batch of users",
	Long: `Updates a batch of users where user details are provided in a Google Sheet,CSV/JSON input file or piped JSON.
			  
The contents of the JSON file or piped input should look something like this:

{"userKey":"stan.laurel@myorg.org","name":{"givenName":"Stanislav","familyName":"Laurelius"},"primaryEmail":"stanislav.laurelius@myorg.org","password":"SuperSuperSecretPassword","changePasswordAtNextLogin":true}
{"userKey":"oliver.hardy@myorg.org","name":{"givenName":"Oliviatus","familyName":"Hardium"},"primaryEmail":"oliviatus.hardium@myorg.org","password":"StealthySuperSecretPassword","changePasswordAtNextLogin":true}
{"userKey":"harold.lloyd@myorg.org","name":{"givenName":"Haroldus","familyName":"Lloydius"},"primaryEmail":"haroldus.lloydius@myorg.org","password":"MightySuperSecretPassword","changePasswordAtNextLogin":true}

N.B. userKey (user email address, alias or id) must be provided.

CSV and Google sheets must have a header row with the following column names being the only ones that are valid:

changePasswordAtNextLogin [value true or false]
firstName
includeInGlobalAddressList [value true or false]
ipWhitelisted [value true or false]
lastName
orgUnitPath
password
primaryEmail
recoveryEmail
recoveryPhone [must start with '+' in E.164 format]
suspended [value true or false]
userKey [required]

The column names are case insensitive and can be in any order. firstName can be replaced by givenName and lastName can be replaced by familyName.`,
	RunE: doBatchUpdUser,
}

func doBatchUpdUser(cmd *cobra.Command, args []string) error {
	lg.Debugw("starting doBatchUpdUser()",
		"args", args)

	var (
		userKeys []string
		users    []*admin.User
	)

	srv, err := cmn.CreateService(cmn.SRVTYPEADMIN, admin.AdminDirectoryUserScope)
	if err != nil {
		return err
	}
	ds := srv.(*admin.Service)

	inputFlgVal, err := cmd.Flags().GetString(flgnm.FLG_INPUTFILE)
	if err != nil {
		lg.Error(err)
		return err
	}

	scanner, err := cmn.InputFromStdIn(inputFlgVal)
	if err != nil {
		lg.Error(err)
		return err
	}

	if inputFlgVal == "" && scanner == nil {
		err := errors.New(gmess.ERR_NOINPUTFILE)
		lg.Error(err)
		return err
	}

	formatFlgVal, err := cmd.Flags().GetString(flgnm.FLG_FORMAT)
	if err != nil {
		lg.Error(err)
		return err
	}
	lwrFmt := strings.ToLower(formatFlgVal)

	ok := cmn.SliceContainsStr(cmn.ValidFileFormats, lwrFmt)
	if !ok {
		err = fmt.Errorf(gmess.ERR_INVALIDFILEFORMAT, formatFlgVal)
		lg.Error(err)
		return err
	}

	switch {
	case lwrFmt == "csv":
		userKeys, users, err = bupduProcessCSVFile(ds, inputFlgVal)
		if err != nil {
			lg.Error(err)
			return err
		}
	case lwrFmt == "json":
		userKeys, users, err = bupduProcessJSON(ds, inputFlgVal, scanner)
		if err != nil {
			lg.Error(err)
			return err
		}
	case lwrFmt == "gsheet":
		rangeFlgVal, err := cmd.Flags().GetString(flgnm.FLG_SHEETRANGE)
		if err != nil {
			lg.Error(err)
			return err
		}

		userKeys, users, err = bupduProcessGSheet(ds, inputFlgVal, rangeFlgVal)
		if err != nil {
			lg.Error(err)
			return err
		}
	default:
		return fmt.Errorf(gmess.ERR_INVALIDFILEFORMAT, formatFlgVal)
	}

	err = bupduProcessObjects(ds, users, userKeys)
	if err != nil {
		lg.Error(err)
		return err
	}

	lg.Debug("finished doBatchUpdUser()")
	return nil
}

func bupduFromFileFactory(hdrMap map[int]string, userData []interface{}) (*admin.User, string, error) {
	lg.Debugw("starting bupduFromFileFactory()",
		"hdrMap", hdrMap)

	var (
		name    *admin.UserName
		user    *admin.User
		userKey string
	)

	name = new(admin.UserName)
	user = new(admin.User)

	for idx, attr := range userData {
		attrName := hdrMap[idx]
		attrVal := fmt.Sprintf("%v", attr)
		lowerAttrVal := strings.ToLower(fmt.Sprintf("%v", attr))

		if attrName == "changePasswordAtNextLogin" {
			if lowerAttrVal == "true" {
				user.ChangePasswordAtNextLogin = true
			} else {
				user.ChangePasswordAtNextLogin = false
				user.ForceSendFields = append(user.ForceSendFields, "ChangePasswordAtNextLogin")
			}
		}
		if attrName == "familyName" {
			name.FamilyName = attrVal
		}
		if attrName == "givenName" {
			name.GivenName = attrVal
		}
		if attrName == "includeInGlobalAddressList" {
			if lowerAttrVal == "true" {
				user.IncludeInGlobalAddressList = true
			} else {
				user.IncludeInGlobalAddressList = false
				user.ForceSendFields = append(user.ForceSendFields, "IncludeInGlobalAddressList")
			}
		}
		if attrName == "ipWhitelisted" {
			if lowerAttrVal == "true" {
				user.IpWhitelisted = true
			} else {
				user.IpWhitelisted = false
				user.ForceSendFields = append(user.ForceSendFields, "IpWhitelisted")
			}
		}
		if attrName == "orgUnitPath" {
			if attrVal == "" {
				err := fmt.Errorf(gmess.ERR_EMPTYSTRING, attrName)
				return nil, "", err
			}
			user.OrgUnitPath = attrVal
		}
		if attrName == "password" {
			if attrVal != "" {
				pwd, err := usrs.HashPassword(attrVal)
				if err != nil {
					return nil, "", err
				}
				user.Password = pwd
				user.HashFunction = usrs.HASHFUNCTION
			}
		}
		if attrName == "primaryEmail" {
			if attrVal == "" {
				err := fmt.Errorf(gmess.ERR_EMPTYSTRING, attrName)
				return nil, "", err
			}
			user.PrimaryEmail = attrVal
		}
		if attrName == "recoveryEmail" {
			user.RecoveryEmail = attrVal
			if attrVal == "" {
				user.ForceSendFields = append(user.ForceSendFields, "RecoveryEmail")
			}
		}
		if attrName == "recoveryPhone" {
			if attrVal != "" {
				err := cmn.ValidateRecoveryPhone(attrVal)
				if err != nil {
					lg.Error(err)
					return nil, "", err
				}
			}
			if attrVal == "" {
				user.ForceSendFields = append(user.ForceSendFields, "RecoveryPhone")
			}
			user.RecoveryPhone = attrVal
		}
		if attrName == "suspended" {
			if lowerAttrVal == "true" {
				user.Suspended = true
			} else {
				user.Suspended = false
				user.ForceSendFields = append(user.ForceSendFields, "Suspended")
			}
		}
		if attrName == "userKey" {
			if attrVal == "" {
				err := fmt.Errorf(gmess.ERR_EMPTYSTRING, attrName)
				return nil, "", err
			}
			userKey = attrVal
		}
	}

	if name.FamilyName != "" || name.GivenName != "" || name.FullName != "" {
		user.Name = name
	}
	lg.Debug("finished bupduFromFileFactory()")
	return user, userKey, nil
}

func bupduFromJSONFactory(ds *admin.Service, jsonData string) (*admin.User, string, error) {
	lg.Debugw("starting bupduFromJSONFactory()",
		"jsonData", jsonData)

	var (
		emptyVals = cmn.EmptyValues{}
		user      *admin.User
		usrKey    = usrs.Key{}
	)

	user = new(admin.User)
	jsonBytes := []byte(jsonData)

	if !json.Valid(jsonBytes) {
		lg.Error(gmess.ERR_INVALIDJSONATTR)
		return nil, "", errors.New(gmess.ERR_INVALIDJSONATTR)
	}

	outStr, err := cmn.ParseInputAttrs(jsonBytes)
	if err != nil {
		lg.Error(err)
		return nil, "", err
	}

	err = cmn.ValidateInputAttrs(outStr, usrs.UserAttrMap)
	if err != nil {
		lg.Error(err)
		return nil, "", err
	}

	err = json.Unmarshal(jsonBytes, &usrKey)
	if err != nil {
		lg.Error(err)
		return nil, "", err
	}

	if usrKey.UserKey == "" {
		err = errors.New(gmess.ERR_NOJSONUSERKEY)
		lg.Error(err)
		return nil, "", err
	}

	err = json.Unmarshal(jsonBytes, &user)
	if err != nil {
		lg.Error(err)
		return nil, "", err
	}

	err = json.Unmarshal(jsonBytes, &emptyVals)
	if err != nil {
		lg.Error(err)
		return nil, "", err
	}
	if len(emptyVals.ForceSendFields) > 0 {
		user.ForceSendFields = emptyVals.ForceSendFields
	}
	lg.Debug("finished bupduFromJSONFactory()")
	return user, usrKey.UserKey, nil
}

func bupduProcessCSVFile(ds *admin.Service, filePath string) ([]string, []*admin.User, error) {
	lg.Debugw("starting bupduProcessCSVFile()",
		"filePath", filePath)

	var (
		iSlice   []interface{}
		hdrMap   = map[int]string{}
		userKeys []string
		users    []*admin.User
	)

	csvfile, err := os.Open(filePath)
	if err != nil {
		lg.Error(err)
		return nil, nil, err
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
			return nil, nil, err
		}

		if count == 0 {
			iSlice = make([]interface{}, len(record))
			for idx, value := range record {
				iSlice[idx] = value
			}
			hdrMap = cmn.ProcessHeader(iSlice)
			err = cmn.ValidateHeader(hdrMap, usrs.UserAttrMap)
			if err != nil {
				lg.Error(err)
				return nil, nil, err
			}
			count = count + 1
			continue
		}

		for idx, value := range record {
			iSlice[idx] = value
		}

		userVar, userKey, err := bupduFromFileFactory(hdrMap, iSlice)
		if err != nil {
			lg.Error(err)
			return nil, nil, err
		}

		users = append(users, userVar)
		userKeys = append(userKeys, userKey)

		count = count + 1
	}

	lg.Debug("finished bupduProcessCSVFile()")
	return userKeys, users, nil
}

func bupduProcessGSheet(ds *admin.Service, sheetID string, sheetrange string) ([]string, []*admin.User, error) {
	lg.Debugw("starting bupduProcessGSheet()",
		"sheetID", sheetID,
		"sheetrange", sheetrange)

	var (
		userKeys []string
		users    []*admin.User
	)

	if sheetrange == "" {
		err := errors.New(gmess.ERR_NOSHEETRANGE)
		lg.Error(err)
		return nil, nil, err
	}

	srv, err := cmn.CreateService(cmn.SRVTYPESHEET, sheet.DriveReadonlyScope)
	if err != nil {
		return nil, nil, err
	}
	ss := srv.(*sheet.Service)

	ssvgc := ss.Spreadsheets.Values.Get(sheetID, sheetrange)
	sValRange, err := ssvgc.Do()
	if err != nil {
		lg.Error(err)
		return nil, nil, err
	}

	if len(sValRange.Values) == 0 {
		err = fmt.Errorf(gmess.ERR_NOSHEETDATAFOUND, sheetID, sheetrange)
		lg.Error(err)
		return nil, nil, err
	}

	hdrMap := cmn.ProcessHeader(sValRange.Values[0])
	err = cmn.ValidateHeader(hdrMap, usrs.UserAttrMap)
	if err != nil {
		lg.Error(err)
		return nil, nil, err
	}

	for idx, row := range sValRange.Values {
		if idx == 0 {
			continue
		}

		userVar, userKey, err := bupduFromFileFactory(hdrMap, row)
		if err != nil {
			lg.Error(err)
			return nil, nil, err
		}

		userKeys = append(userKeys, userKey)
		users = append(users, userVar)
	}

	lg.Debug("finished bupduProcessGSheet()")
	return userKeys, users, nil
}

func bupduProcessJSON(ds *admin.Service, filePath string, scanner *bufio.Scanner) ([]string, []*admin.User, error) {
	lg.Debugw("starting bupduProcessJSON()",
		"filePath", filePath)

	var (
		userKeys []string
		users    []*admin.User
	)

	if filePath != "" {
		file, err := os.Open(filePath)
		if err != nil {
			lg.Error(err)
			return nil, nil, err
		}
		defer file.Close()

		scanner = bufio.NewScanner(file)
	}

	for scanner.Scan() {
		jsonData := scanner.Text()

		userVar, userKey, err := bupduFromJSONFactory(ds, jsonData)
		if err != nil {
			lg.Error(err)
			return nil, nil, err
		}

		userKeys = append(userKeys, userKey)
		users = append(users, userVar)
	}
	err := scanner.Err()
	if err != nil {
		lg.Error(err)
		return nil, nil, err
	}

	lg.Debug("finished bupduProcessJSON()")
	return userKeys, users, nil
}

func bupduProcessObjects(ds *admin.Service, users []*admin.User, userKeys []string) error {
	lg.Debugw("starting bupduProcessObjects()",
		"userKeys", userKeys)

	wg := new(sync.WaitGroup)

	for idx, u := range users {
		if u.Password != "" {
			u.HashFunction = usrs.HASHFUNCTION
			pwd, err := usrs.HashPassword(u.Password)
			if err != nil {
				lg.Error(err)
				return err
			}
			u.Password = pwd
		}

		uuc := ds.Users.Update(userKeys[idx], u)

		wg.Add(1)

		go bupduUpdate(u, wg, uuc, userKeys[idx])
	}

	wg.Wait()

	lg.Debug("finished bupduProcessObjects()")
	return nil
}

func bupduUpdate(user *admin.User, wg *sync.WaitGroup, uuc *admin.UsersUpdateCall, userKey string) {
	lg.Debugw("starting bupduUpdate()",
		"userKey", userKey)

	defer wg.Done()

	b := backoff.NewExponentialBackOff()
	b.MaxElapsedTime = 32 * time.Second

	err := backoff.Retry(func() error {
		var err error
		_, err = uuc.Do()
		if err == nil {
			fmt.Println(cmn.GminMessage(fmt.Sprintf(gmess.INFO_USERUPDATED, userKey)))
			lg.Infof(gmess.INFO_USERUPDATED, userKey)
			return err
		}
		if !cmn.IsErrRetryable(err) {
			return backoff.Permanent(fmt.Errorf(gmess.ERR_BATCHUSER, err.Error(), userKey))
		}
		// Log the retries
		lg.Warnw(err.Error(),
			"retrying", b.GetElapsedTime().String(),
			"user", userKey)
		return fmt.Errorf(gmess.ERR_BATCHUSER, err.Error(), userKey)
	}, b)
	if err != nil {
		// Log final error
		lg.Error(err)
		fmt.Println(cmn.GminMessage(err.Error()))
	}
	lg.Debug("finished bupduUpdate()")
}

func init() {
	batchUpdateCmd.AddCommand(batchUpdUserCmd)

	batchUpdUserCmd.Flags().StringVarP(&inputFile, flgnm.FLG_INPUTFILE, "i", "", "filepath to user data file or sheet id")
	batchUpdUserCmd.Flags().StringVarP(&format, flgnm.FLG_FORMAT, "f", "json", "user data file format")
	batchUpdUserCmd.Flags().StringVarP(&sheetRange, flgnm.FLG_SHEETRANGE, "s", "", "user data gsheet range")
}
