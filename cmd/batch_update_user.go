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
	logger.Debugw("starting doBatchUpdUser()",
		"args", args)

	ds, err := cmn.CreateDirectoryService(admin.AdminDirectoryUserScope)
	if err != nil {
		logger.Error(err)
		return err
	}

	scanner, err := cmn.InputFromStdIn(inputFile)
	if err != nil {
		logger.Error(err)
		return err
	}

	if inputFile == "" && scanner == nil {
		err := errors.New(cmn.ErrNoInputFile)
		logger.Error(err)
		return err
	}

	lwrFmt := strings.ToLower(format)

	ok := cmn.SliceContainsStr(cmn.ValidFileFormats, lwrFmt)
	if !ok {
		err = fmt.Errorf(cmn.ErrInvalidFileFormat, format)
		logger.Error(err)
		return err
	}

	switch {
	case lwrFmt == "csv":
		err := btchUpdUsrProcessCSV(ds, inputFile)
		if err != nil {
			logger.Error(err)
			return err
		}
	case lwrFmt == "json":
		err := btchUpdUsrProcessJSON(ds, inputFile, scanner)
		if err != nil {
			logger.Error(err)
			return err
		}
	case lwrFmt == "gsheet":
		err := btchUpdUsrProcessSheet(ds, inputFile)
		if err != nil {
			logger.Error(err)
			return err
		}
	}
	logger.Debug("finished doBatchUpdUser()")
	return nil
}

func btchUpdJSONUser(ds *admin.Service, jsonData string) (*admin.User, string, error) {
	logger.Debugw("starting btchUpdJSONUser()",
		"jsonData", jsonData)

	var (
		emptyVals = cmn.EmptyValues{}
		user      *admin.User
		usrKey    = usrs.Key{}
	)

	user = new(admin.User)
	jsonBytes := []byte(jsonData)

	if !json.Valid(jsonBytes) {
		logger.Error(cmn.ErrInvalidJSONAttr)
		return nil, "", errors.New(cmn.ErrInvalidJSONAttr)
	}

	outStr, err := cmn.ParseInputAttrs(jsonBytes)
	if err != nil {
		logger.Error(err)
		return nil, "", err
	}

	err = cmn.ValidateInputAttrs(outStr, usrs.UserAttrMap)
	if err != nil {
		logger.Error(err)
		return nil, "", err
	}

	err = json.Unmarshal(jsonBytes, &usrKey)
	if err != nil {
		logger.Error(err)
		return nil, "", err
	}

	if usrKey.UserKey == "" {
		err = errors.New(cmn.ErrNoJSONUserKey)
		logger.Error(err)
		return nil, "", err
	}

	err = json.Unmarshal(jsonBytes, &user)
	if err != nil {
		logger.Error(err)
		return nil, "", err
	}

	err = json.Unmarshal(jsonBytes, &emptyVals)
	if err != nil {
		logger.Error(err)
		return nil, "", err
	}
	if len(emptyVals.ForceSendFields) > 0 {
		user.ForceSendFields = emptyVals.ForceSendFields
	}
	logger.Debug("finished btchUpdJSONUser()")
	return user, usrKey.UserKey, nil
}

func btchUpdateUsers(ds *admin.Service, users []*admin.User, userKeys []string) error {
	logger.Debugw("starting btchUpdateUsers()",
		"userKeys", userKeys)

	wg := new(sync.WaitGroup)

	for idx, u := range users {
		if u.Password != "" {
			u.HashFunction = cmn.HashFunction
			pwd, err := cmn.HashPassword(u.Password)
			if err != nil {
				logger.Error(err)
				return err
			}
			u.Password = pwd
		}

		uuc := ds.Users.Update(userKeys[idx], u)

		wg.Add(1)

		go btchUsrUpdateProcess(u, wg, uuc, userKeys[idx])
	}

	wg.Wait()

	logger.Debug("finished btchUpdateUsers()")
	return nil
}

func btchUsrUpdateProcess(user *admin.User, wg *sync.WaitGroup, uuc *admin.UsersUpdateCall, userKey string) {
	logger.Debugw("starting btchUsrUpdateProcess()",
		"userKey", userKey)

	defer wg.Done()

	b := backoff.NewExponentialBackOff()
	b.MaxElapsedTime = 32 * time.Second

	err := backoff.Retry(func() error {
		var err error
		_, err = uuc.Do()
		if err == nil {
			logger.Infof(cmn.InfoUserUpdated, userKey)
			fmt.Println(cmn.GminMessage(fmt.Sprintf(cmn.InfoUserUpdated, userKey)))
			return err
		}
		if !cmn.IsErrRetryable(err) {
			return backoff.Permanent(fmt.Errorf(cmn.ErrBatchUser, err.Error(), userKey))
		}
		// Log the retries
		logger.Warnw(err.Error(),
			"retrying", b.GetElapsedTime().String(),
			"user", userKey)
		return fmt.Errorf(cmn.ErrBatchUser, err.Error(), userKey)
	}, b)
	if err != nil {
		// Log final error
		logger.Error(err)
		fmt.Println(cmn.GminMessage(err.Error()))
	}
	logger.Debug("finished btchUsrUpdateProcess()")
}

func btchUpdUsrProcessCSV(ds *admin.Service, filePath string) error {
	logger.Debugw("starting btchUpdUsrProcessCSV()",
		"filePath", filePath)

	var (
		iSlice   []interface{}
		hdrMap   = map[int]string{}
		userKeys []string
		users    []*admin.User
	)

	csvfile, err := os.Open(filePath)
	if err != nil {
		logger.Error(err)
		return err
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
			logger.Error(err)
			return err
		}

		if count == 0 {
			iSlice = make([]interface{}, len(record))
			for idx, value := range record {
				iSlice[idx] = value
			}
			hdrMap = cmn.ProcessHeader(iSlice)
			err = cmn.ValidateHeader(hdrMap, usrs.UserAttrMap)
			if err != nil {
				logger.Error(err)
				return err
			}
			count = count + 1
			continue
		}

		for idx, value := range record {
			iSlice[idx] = value
		}

		userVar, userKey, err := btchUpdProcessUser(hdrMap, iSlice)
		if err != nil {
			logger.Error(err)
			return err
		}

		users = append(users, userVar)
		userKeys = append(userKeys, userKey)

		count = count + 1
	}

	err = btchUpdateUsers(ds, users, userKeys)
	if err != nil {
		logger.Error(err)
		return err
	}
	logger.Debug("finished btchUpdUsrProcessCSV()")
	return nil
}

func btchUpdUsrProcessJSON(ds *admin.Service, filePath string, scanner *bufio.Scanner) error {
	logger.Debugw("starting btchUpdUsrProcessJSON()",
		"filePath", filePath)

	var (
		userKeys []string
		users    []*admin.User
	)

	if filePath != "" {
		file, err := os.Open(filePath)
		if err != nil {
			logger.Error(err)
			return err
		}
		defer file.Close()

		scanner = bufio.NewScanner(file)
	}

	for scanner.Scan() {
		jsonData := scanner.Text()

		userVar, userKey, err := btchUpdJSONUser(ds, jsonData)
		if err != nil {
			logger.Error(err)
			return err
		}

		userKeys = append(userKeys, userKey)
		users = append(users, userVar)
	}
	err := scanner.Err()
	if err != nil {
		logger.Error(err)
		return err
	}

	err = btchUpdateUsers(ds, users, userKeys)
	if err != nil {
		logger.Error(err)
		return err
	}
	logger.Debug("finished btchUpdUsrProcessJSON()")
	return nil
}

func btchUpdUsrProcessSheet(ds *admin.Service, sheetID string) error {
	logger.Debugw("starting btchUpdUsrProcessSheet()",
		"sheetID", sheetID)

	var (
		userKeys []string
		users    []*admin.User
	)

	if sheetRange == "" {
		err := errors.New(cmn.ErrNoSheetRange)
		logger.Error(err)
		return err
	}

	ss, err := cmn.CreateSheetService(sheet.DriveReadonlyScope)
	if err != nil {
		logger.Error(err)
		return err
	}

	ssvgc := ss.Spreadsheets.Values.Get(sheetID, sheetRange)
	sValRange, err := ssvgc.Do()
	if err != nil {
		logger.Error(err)
		return err
	}

	if len(sValRange.Values) == 0 {
		err = fmt.Errorf(cmn.ErrNoSheetDataFound, sheetID, sheetRange)
		logger.Error(err)
		return err
	}

	hdrMap := cmn.ProcessHeader(sValRange.Values[0])
	err = cmn.ValidateHeader(hdrMap, usrs.UserAttrMap)
	if err != nil {
		logger.Error(err)
		return err
	}

	for idx, row := range sValRange.Values {
		if idx == 0 {
			continue
		}

		userVar, userKey, err := btchUpdProcessUser(hdrMap, row)
		if err != nil {
			logger.Error(err)
			return err
		}

		userKeys = append(userKeys, userKey)
		users = append(users, userVar)
	}

	err = btchUpdateUsers(ds, users, userKeys)
	if err != nil {
		logger.Error(err)
		return err
	}
	logger.Debug("finished btchUpdUsrProcessSheet()")
	return nil
}

func btchUpdProcessUser(hdrMap map[int]string, userData []interface{}) (*admin.User, string, error) {
	logger.Debugw("starting btchUpdProcessUser()",
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
				err := fmt.Errorf(cmn.ErrEmptyString, attrName)
				return nil, "", err
			}
			user.OrgUnitPath = attrVal
		}
		if attrName == "password" {
			if attrVal != "" {
				pwd, err := cmn.HashPassword(attrVal)
				if err != nil {
					return nil, "", err
				}
				user.Password = pwd
				user.HashFunction = cmn.HashFunction
			}
		}
		if attrName == "primaryEmail" {
			if attrVal == "" {
				err := fmt.Errorf(cmn.ErrEmptyString, attrName)
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
					logger.Error(err)
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
				err := fmt.Errorf(cmn.ErrEmptyString, attrName)
				return nil, "", err
			}
			userKey = attrVal
		}
	}

	if name.FamilyName != "" || name.GivenName != "" || name.FullName != "" {
		user.Name = name
	}
	logger.Debug("finished btchUpdProcessUser()")
	return user, userKey, nil
}

func init() {
	batchUpdateCmd.AddCommand(batchUpdUserCmd)

	batchUpdUserCmd.Flags().StringVarP(&inputFile, "inputfile", "i", "", "filepath to user data file or sheet id")
	batchUpdUserCmd.Flags().StringVarP(&format, "format", "f", "json", "user data file format")
	batchUpdUserCmd.Flags().StringVarP(&sheetRange, "sheetrange", "s", "", "user data gsheet range")
}
