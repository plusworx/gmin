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
	gmess "github.com/plusworx/gmin/utils/gminmessages"
	usrs "github.com/plusworx/gmin/utils/users"
	"github.com/spf13/cobra"
	admin "google.golang.org/api/admin/directory/v1"
	sheet "google.golang.org/api/sheets/v4"
)

var batchCrtUserCmd = &cobra.Command{
	Use:     "users -i <input file path or google sheet id>",
	Aliases: []string{"user"},
	Example: `gmin batch-create users -i inputfile.json
gmin bcrt user -i inputfile.csv -f csv
gmin bcrt user -i 1odyAIp3jGspd3M4xeepxWD6aeQIUuHBgrZB2OHSu8MI -s 'Sheet1!A1:K25' -f gsheet`,
	Short: "Creates a batch of users",
	Long: `Creates a batch of users where user details are provided in a Google Sheet,CSV/JSON input file or piped JSON.
			
The contents of a JSON file or piped input should look something like this:

{"name":{"firstName":"Stan","familyName":"Laurel"},"primaryEmail":"stan.laurel@company.com","password":"SecretPassword","changePasswordAtNextLogin":true}
{"name":{"givenName":"Oliver","familyName":"Hardy"},"primaryEmail":"oliver.hardy@company.com","password":"SecretPassword","changePasswordAtNextLogin":true}
{"name":{"givenName":"Harold","familyName":"Lloyd"},"primaryEmail":"harold.lloyd@company.com","password":"SecretPassword","changePasswordAtNextLogin":true}

CSV and Google sheets must have a header row with the following column names being the only ones that are valid:

changePasswordAtNextLogin [value true or false]
firstName [required]
includeInGlobalAddressList [value true or false]
ipWhitelisted [value true or false]
lastName [required]
orgUnitPath
password [required]
primaryEmail [required]
recoveryEmail
recoveryPhone [must start with '+' in E.164 format]
suspended [value true or false]

The column names are case insensitive and can be in any order. firstName can be replaced by givenName and lastName can be replaced by familyName.`,
	RunE: doBatchCrtUser,
}

func doBatchCrtUser(cmd *cobra.Command, args []string) error {
	logger.Debugw("starting doBatchCrtUser()",
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
		err := errors.New(gmess.ErrNoInputFile)
		logger.Error(err)
		return err
	}

	lwrFmt := strings.ToLower(format)

	ok := cmn.SliceContainsStr(cmn.ValidFileFormats, lwrFmt)
	if !ok {
		err = fmt.Errorf(gmess.ErrInvalidFileFormat, format)
		logger.Error(err)
		return err
	}

	switch {
	case lwrFmt == "csv":
		err := bcuProcessCSVFile(ds, inputFile)
		if err != nil {
			logger.Error(err)
			return err
		}
	case lwrFmt == "json":
		err := bcuProcessJSON(ds, inputFile, scanner)
		if err != nil {
			logger.Error(err)
			return err
		}
	case lwrFmt == "gsheet":
		err := bcuProcessGSheet(ds, inputFile)
		if err != nil {
			logger.Error(err)
			return err
		}
	default:
		return fmt.Errorf(gmess.ErrInvalidFileFormat, format)
	}
	logger.Debug("finished doBatchCrtUser()")
	return nil
}

func bcuCreate(user *admin.User, wg *sync.WaitGroup, uic *admin.UsersInsertCall) {
	logger.Debug("starting bcuCreate()")

	defer wg.Done()

	b := backoff.NewExponentialBackOff()
	b.MaxElapsedTime = 32 * time.Second

	err := backoff.Retry(func() error {
		var err error
		newUser, err := uic.Do()
		if err == nil {
			logger.Infof(gmess.InfoUserCreated, newUser.PrimaryEmail)
			fmt.Println(cmn.GminMessage(fmt.Sprintf(gmess.InfoUserCreated, newUser.PrimaryEmail)))
			return err
		}
		if !cmn.IsErrRetryable(err) {
			return backoff.Permanent(fmt.Errorf(gmess.ErrBatchUser, err.Error(), user.PrimaryEmail))
		}
		// Log the retries
		logger.Warnw(err.Error(),
			"retrying", b.GetElapsedTime().String(),
			"user", user.PrimaryEmail)
		return fmt.Errorf(gmess.ErrBatchUser, err.Error(), user.PrimaryEmail)
	}, b)
	if err != nil {
		// Log final error
		logger.Error(err)
		fmt.Println(cmn.GminMessage(err.Error()))
	}
	logger.Debug("finished bcuCreate()")
}

func bcuFromFileFactory(hdrMap map[int]string, userData []interface{}) (*admin.User, error) {
	logger.Debugw("starting bcuFromFileFactory()",
		"hdrMap", hdrMap)

	var (
		name *admin.UserName
		user *admin.User
	)

	name = new(admin.UserName)
	user = new(admin.User)

	for idx, val := range userData {
		attrName := hdrMap[idx]
		attrVal := fmt.Sprintf("%v", val)
		lowerAttrVal := strings.ToLower(fmt.Sprintf("%v", val))

		if attrName == "changePasswordAtNextLogin" {
			if lowerAttrVal == "true" {
				user.ChangePasswordAtNextLogin = true
			}
		}
		if attrName == "familyName" {
			name.FamilyName = attrVal
		}
		if attrName == "givenName" {
			name.GivenName = attrVal
		}
		if attrName == "includeInGlobalAddressList" {
			if lowerAttrVal == "false" {
				user.IncludeInGlobalAddressList = false
				user.ForceSendFields = append(user.ForceSendFields, "IncludeInGlobalAddressList")
			}
		}
		if attrName == "ipWhitelisted" {
			if lowerAttrVal == "true" {
				user.IpWhitelisted = true
			}
		}
		if attrName == "orgUnitPath" {
			user.OrgUnitPath = attrVal
		}
		if attrName == "password" {
			if attrVal == "" {
				err := fmt.Errorf(gmess.ErrEmptyString, attrName)
				return nil, err
			}
			pwd, err := cmn.HashPassword(attrVal)
			if err != nil {
				return nil, err
			}
			user.Password = pwd
			user.HashFunction = cmn.HashFunction
		}
		if attrName == "primaryEmail" {
			if attrVal == "" {
				err := fmt.Errorf(gmess.ErrEmptyString, attrName)
				return nil, err
			}
			user.PrimaryEmail = attrVal
		}
		if attrName == "recoveryEmail" {
			user.RecoveryEmail = attrVal
		}
		if attrName == "recoveryPhone" {
			if attrVal != "" {
				err := cmn.ValidateRecoveryPhone(attrVal)
				if err != nil {
					logger.Error(err)
					return nil, err
				}
				user.RecoveryPhone = attrVal
			}
		}
		if attrName == "suspended" {
			if lowerAttrVal == "true" {
				user.Suspended = true
			}
		}
	}
	user.Name = name
	logger.Debug("finished bcuFromFileFactory()")
	return user, nil
}

func bcuFromJSONFactory(ds *admin.Service, jsonData string) (*admin.User, error) {
	logger.Debugw("starting bcuFromJSONFactory()",
		"jsonData", jsonData)

	var (
		emptyVals = cmn.EmptyValues{}
		user      *admin.User
	)

	user = new(admin.User)
	jsonBytes := []byte(jsonData)

	if !json.Valid(jsonBytes) {
		logger.Error(gmess.ErrInvalidJSONAttr)
		return nil, errors.New(gmess.ErrInvalidJSONAttr)
	}

	outStr, err := cmn.ParseInputAttrs(jsonBytes)
	if err != nil {
		logger.Error(err)
		return nil, err
	}

	err = cmn.ValidateInputAttrs(outStr, usrs.UserAttrMap)
	if err != nil {
		logger.Error(err)
		return nil, err
	}

	err = json.Unmarshal(jsonBytes, &user)
	if err != nil {
		logger.Error(err)
		return nil, err
	}

	err = json.Unmarshal(jsonBytes, &emptyVals)
	if err != nil {
		logger.Error(err)
		return nil, err
	}
	if len(emptyVals.ForceSendFields) > 0 {
		user.ForceSendFields = emptyVals.ForceSendFields
	}
	logger.Debug("finished bcuFromJSONFactory()")
	return user, nil
}

func bcuProcessCSVFile(ds *admin.Service, filePath string) error {
	logger.Debugw("starting bcuProcessCSVFile()",
		"filePath", filePath)

	var (
		iSlice []interface{}
		hdrMap = map[int]string{}
		users  []*admin.User
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

		userVar, err := bcuFromFileFactory(hdrMap, iSlice)
		if err != nil {
			logger.Error(err)
			return err
		}

		users = append(users, userVar)

		count = count + 1
	}

	err = bcuProcessObjects(ds, users)
	if err != nil {
		logger.Error(err)
		return err
	}
	logger.Debug("finished bcuProcessCSVFile()")
	return nil
}

func bcuProcessGSheet(ds *admin.Service, sheetID string) error {
	logger.Debugw("starting bcuProcessGSheet()",
		"sheetID", sheetID)

	var users []*admin.User

	if sheetRange == "" {
		err := errors.New(gmess.ErrNoSheetRange)
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
		err = fmt.Errorf(gmess.ErrNoSheetDataFound, sheetID, sheetRange)
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

		userVar, err := bcuFromFileFactory(hdrMap, row)
		if err != nil {
			logger.Error(err)
			return err
		}

		users = append(users, userVar)
	}

	err = bcuProcessObjects(ds, users)
	if err != nil {
		logger.Error(err)
		return err
	}
	logger.Debug("finished bcuProcessGSheet()")
	return nil
}

func bcuProcessJSON(ds *admin.Service, filePath string, scanner *bufio.Scanner) error {
	logger.Debugw("starting bcuProcessJSON()",
		"filePath", filePath)

	var users []*admin.User

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

		userVar, err := bcuFromJSONFactory(ds, jsonData)
		if err != nil {
			logger.Error(err)
			return err
		}

		users = append(users, userVar)
	}
	err := scanner.Err()
	if err != nil {
		logger.Error(err)
		return err
	}

	err = bcuProcessObjects(ds, users)
	if err != nil {
		logger.Error(err)
		return err
	}
	logger.Debug("finished bcuProcessJSON()")
	return nil
}

func bcuProcessObjects(ds *admin.Service, users []*admin.User) error {
	logger.Debug("starting bcuProcessObjects()")

	wg := new(sync.WaitGroup)

	for _, u := range users {
		if u.PrimaryEmail == "" || u.Name.GivenName == "" || u.Name.FamilyName == "" || u.Password == "" {
			err := errors.New(gmess.ErrBatchMissingUserData)
			logger.Error(err)
			return err
		}

		u.HashFunction = cmn.HashFunction
		pwd, err := cmn.HashPassword(u.Password)
		if err != nil {
			logger.Error(err)
			return err
		}
		u.Password = pwd

		uic := ds.Users.Insert(u)

		wg.Add(1)

		go bcuCreate(u, wg, uic)
	}

	wg.Wait()

	logger.Debug("finished bcuProcessObjects()")
	return nil
}

func init() {
	batchCreateCmd.AddCommand(batchCrtUserCmd)

	batchCrtUserCmd.Flags().StringVarP(&inputFile, "input-file", "i", "", "filepath to user data file or sheet id")
	batchCrtUserCmd.Flags().StringVarP(&format, "format", "f", "json", "user data file format")
	batchCrtUserCmd.Flags().StringVarP(&sheetRange, "sheet-range", "s", "", "user data gsheet range")
}
