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

var batchCrtUserCmd = &cobra.Command{
	Use:     "users -i <input file path or google sheet id>",
	Aliases: []string{"user"},
	Short:   "Creates a batch of users",
	Long: `Creates a batch of users where user details are provided in a Google Sheet or CSV/JSON input file.
	
	Examples:	gmin batch-create users -i inputfile.json
			gmin bcrt user -i inputfile.csv -f csv
			gmin bcrt user -i 1odyAIp3jGspd3M4xeepxWD6aeQIUuHBgrZB2OHSu8MI -s 'Sheet1!A1:K25' -f gsheet
			
	The contents of a JSON file should look something like this:
	
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
	ds, err := cmn.CreateDirectoryService(admin.AdminDirectoryUserScope)
	if err != nil {
		return err
	}

	if inputFile == "" {
		err := errors.New("gmin: error - must provide inputfile")
		return err
	}

	lwrFmt := strings.ToLower(format)

	ok := cmn.SliceContainsStr(cmn.ValidFileFormats, lwrFmt)
	if !ok {
		return fmt.Errorf("gmin: error - %v is not a valid file format", format)
	}

	switch {
	case lwrFmt == "csv":
		err := btchCrtUsrProcessCSV(ds, inputFile)
		if err != nil {
			return err
		}
	case lwrFmt == "json":
		err := btchCrtUsrProcessJSON(ds, inputFile)
		if err != nil {
			return err
		}
	case lwrFmt == "gsheet":
		err := btchCrtUsrProcessSheet(ds, inputFile)
		if err != nil {
			return err
		}
	}

	return nil
}

func btchCrtJSONUser(ds *admin.Service, jsonData string) (*admin.User, error) {
	var user *admin.User

	user = new(admin.User)
	jsonBytes := []byte(jsonData)

	if !json.Valid(jsonBytes) {
		return nil, errors.New("gmin: error - attribute string is not valid JSON")
	}

	outStr, err := cmn.ParseInputAttrs(jsonBytes)
	if err != nil {
		return nil, err
	}

	err = cmn.ValidateInputAttrs(outStr, usrs.UserAttrMap)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(jsonBytes, &user)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func btchInsertNewUsers(ds *admin.Service, users []*admin.User) error {
	wg := new(sync.WaitGroup)

	for _, u := range users {
		if u.PrimaryEmail == "" || u.Name.GivenName == "" || u.Name.FamilyName == "" || u.Password == "" {
			return errors.New("gmin: error - primaryEmail, givenName, familyName and password must all be provided")
		}

		u.HashFunction = cmn.HashFunction
		pwd, err := cmn.HashPassword(u.Password)
		if err != nil {
			return err
		}
		u.Password = pwd

		uic := ds.Users.Insert(u)

		wg.Add(1)

		go btchUsrInsertProcess(u, wg, uic)
	}

	wg.Wait()

	return nil
}

func btchUsrInsertProcess(user *admin.User, wg *sync.WaitGroup, uic *admin.UsersInsertCall) {
	defer wg.Done()

	b := backoff.NewExponentialBackOff()
	b.MaxElapsedTime = 32 * time.Second

	err := backoff.Retry(func() error {
		var err error
		newUser, err := uic.Do()
		if err == nil {
			fmt.Println(cmn.GminMessage("**** gmin: user " + newUser.PrimaryEmail + " created ****"))
			return err
		}
		if !cmn.IsErrRetryable(err) {
			return backoff.Permanent(errors.New(cmn.GminMessage("gmin: error - " + err.Error() + " " + user.PrimaryEmail)))
		}
		return errors.New(cmn.GminMessage("gmin: error - " + err.Error() + " " + user.PrimaryEmail))
	}, b)
	if err != nil {
		fmt.Println(err)
	}
}

func btchCrtUsrProcessCSV(ds *admin.Service, filePath string) error {
	var (
		iSlice []interface{}
		hdrMap = map[int]string{}
		users  []*admin.User
	)

	csvfile, err := os.Open(filePath)
	if err != nil {
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
				return err
			}
			count = count + 1
			continue
		}

		for idx, value := range record {
			iSlice[idx] = value
		}

		userVar, err := btchCrtProcessUser(hdrMap, iSlice)
		if err != nil {
			fmt.Println(err.Error())
		}

		users = append(users, userVar)

		count = count + 1
	}

	err = btchInsertNewUsers(ds, users)
	if err != nil {
		return err
	}
	return nil
}

func btchCrtUsrProcessJSON(ds *admin.Service, filePath string) error {
	var users []*admin.User

	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		jsonData := scanner.Text()

		userVar, err := btchCrtJSONUser(ds, jsonData)
		if err != nil {
			return err
		}

		users = append(users, userVar)
	}
	err = scanner.Err()
	if err != nil {
		return err
	}

	err = btchInsertNewUsers(ds, users)
	if err != nil {
		return err
	}

	return nil
}

func btchCrtUsrProcessSheet(ds *admin.Service, sheetID string) error {
	var users []*admin.User

	if sheetRange == "" {
		return errors.New("gmin: error - sheetrange must be provided")
	}

	ss, err := cmn.CreateSheetService(sheet.DriveReadonlyScope)
	if err != nil {
		return err
	}

	ssvgc := ss.Spreadsheets.Values.Get(sheetID, sheetRange)
	sValRange, err := ssvgc.Do()
	if err != nil {
		return err
	}

	if len(sValRange.Values) == 0 {
		return errors.New("gmin: error - no data found in sheet " + sheetID + " range: " + sheetRange)
	}

	hdrMap := cmn.ProcessHeader(sValRange.Values[0])
	err = cmn.ValidateHeader(hdrMap, usrs.UserAttrMap)
	if err != nil {
		return err
	}

	for idx, row := range sValRange.Values {
		if idx == 0 {
			continue
		}

		userVar, err := btchCrtProcessUser(hdrMap, row)
		if err != nil {
			return err
		}

		users = append(users, userVar)
	}

	err = btchInsertNewUsers(ds, users)
	if err != nil {
		return err
	}

	return nil
}

func btchCrtProcessUser(hdrMap map[int]string, userData []interface{}) (*admin.User, error) {
	var (
		name *admin.UserName
		user *admin.User
	)

	name = new(admin.UserName)
	user = new(admin.User)

	for idx, attr := range userData {
		attrName := hdrMap[idx]

		switch {
		case attrName == "changePasswordAtNextLogin":
			lwrAttr := strings.ToLower(fmt.Sprintf("%v", attr))
			if lwrAttr == "true" {
				user.ChangePasswordAtNextLogin = true
			}
		case attrName == "familyName":
			name.FamilyName = fmt.Sprintf("%v", attr)
		case attrName == "givenName":
			name.GivenName = fmt.Sprintf("%v", attr)
		case attrName == "includeInGlobalAddressList":
			lwrAttr := strings.ToLower(fmt.Sprintf("%v", attr))
			if lwrAttr == "false" {
				user.IncludeInGlobalAddressList = false
				user.ForceSendFields = append(user.ForceSendFields, "IncludeInGlobalAddressList")
			}
		case attrName == "ipWhitelisted":
			lwrAttr := strings.ToLower(fmt.Sprintf("%v", attr))
			if lwrAttr == "true" {
				user.IpWhitelisted = true
			}
		case attrName == "orgUnitPath":
			user.OrgUnitPath = fmt.Sprintf("%v", attr)
		case attrName == "password":
			user.Password = fmt.Sprintf("%v", attr)
		case attrName == "primaryEmail":
			user.PrimaryEmail = fmt.Sprintf("%v", attr)
		case attrName == "recoveryEmail":
			user.RecoveryEmail = fmt.Sprintf("%v", attr)
		case attrName == "recoveryPhone":
			sAttr := fmt.Sprintf("%v", attr)
			err := cmn.ValidateRecoveryPhone(sAttr)
			if err != nil {
				return nil, err
			}
			user.RecoveryPhone = sAttr
		case attrName == "suspended":
			lwrAttr := strings.ToLower(fmt.Sprintf("%v", attr))
			if lwrAttr == "true" {
				user.Suspended = true
			}
		}
	}

	user.Name = name

	return user, nil
}

func init() {
	batchCreateCmd.AddCommand(batchCrtUserCmd)

	batchCrtUserCmd.Flags().StringVarP(&inputFile, "inputfile", "i", "", "filepath to user data file or sheet id")
	batchCrtUserCmd.Flags().StringVarP(&format, "format", "f", "json", "user data file format")
	batchCrtUserCmd.Flags().StringVarP(&sheetRange, "sheetrange", "s", "", "user data gsheet range")
}
