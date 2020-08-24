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
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/cenkalti/backoff/v4"
	cmn "github.com/plusworx/gmin/utils/common"
	usrs "github.com/plusworx/gmin/utils/users"
	"github.com/spf13/cobra"
	admin "google.golang.org/api/admin/directory/v1"
	sheet "google.golang.org/api/sheets/v4"
)

var batchCrtUserCmd = &cobra.Command{
	Use:     "users -i <input file path>",
	Aliases: []string{"user"},
	Short:   "Creates a batch of users",
	Long: `Creates a batch of users where user details are provided in a JSON input file.
	
	Examples:	gmin batch-create users -i inputfile.json
			gmin bcrt user -i inputfile.json
			
	The contents of the JSON file should look something like this:
	
	{"name":{"firstName":"Stan","familyName":"Laurel"},"primaryEmail":"stan.laurel@company.com","password":"SecretPassword","changePasswordAtNextLogin":true}
	{"name":{"givenName":"Oliver","familyName":"Hardy"},"primaryEmail":"oliver.hardy@company.com","password":"SecretPassword","changePasswordAtNextLogin":true}
	{"name":{"givenName":"Harold","familyName":"Lloyd"},"primaryEmail":"harold.lloyd@company.com","password":"SecretPassword","changePasswordAtNextLogin":true}`,
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

	if format == "gsheet" {
		if sheetRange == "" {
			return errors.New("gmin: error - sheetrange must be provided")
		}

		err := processSheet(ds, inputFile)
		if err != nil {
			return err
		}
		return nil
	}

	file, err := os.Open(inputFile)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		jsonData := scanner.Text()

		err = createJSONUser(ds, jsonData)
		if err != nil {
			return err
		}
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	return nil
}

func createJSONUser(ds *admin.Service, jsonData string) error {
	var user *admin.User

	user = new(admin.User)
	jsonBytes := []byte(jsonData)

	if !json.Valid(jsonBytes) {
		return errors.New("gmin: error - attribute string is not valid JSON")
	}

	outStr, err := cmn.ParseInputAttrs(jsonBytes)
	if err != nil {
		return err
	}

	err = cmn.ValidateInputAttrs(outStr, usrs.UserAttrMap)
	if err != nil {
		return err
	}

	err = json.Unmarshal(jsonBytes, &user)
	if err != nil {
		return err
	}

	err = insertNewUser(ds, user)
	if err != nil {
		return err
	}

	return nil
}

func init() {
	batchCreateCmd.AddCommand(batchCrtUserCmd)

	batchCrtUserCmd.Flags().StringVarP(&inputFile, "inputfile", "i", "", "filepath to user data file")
	batchCrtUserCmd.Flags().StringVarP(&format, "format", "f", "json", "user data file format")
	batchCrtUserCmd.Flags().StringVarP(&sheetRange, "sheetrange", "s", "", "user data gsheet range")
}

func insertNewUser(ds *admin.Service, user *admin.User) error {
	var newUser *admin.User

	if user.PrimaryEmail == "" || user.Name.GivenName == "" || user.Name.FamilyName == "" || user.Password == "" {
		return errors.New("gmin: error - primaryEmail, givenName, familyName and password must all be provided")
	}

	user.HashFunction = cmn.HashFunction
	pwd, err := cmn.HashPassword(user.Password)
	if err != nil {
		return err
	}
	user.Password = pwd

	uic := ds.Users.Insert(user)

	b := backoff.NewExponentialBackOff()
	b.MaxElapsedTime = 30 * time.Second

	err = backoff.Retry(func() error {
		var err error
		newUser, err = uic.Do()
		if err == nil {
			return err
		}

		if strings.Contains(err.Error(), "Missing required field") ||
			strings.Contains(err.Error(), "invalid") ||
			strings.Contains(err.Error(), "Entity already exists") ||
			strings.Contains(err.Error(), "should be") {
			return backoff.Permanent(err)
		}

		return err
	}, b)
	if err != nil {
		return err
	}

	fmt.Println(cmn.GminMessage("**** gmin: user " + newUser.PrimaryEmail + " created ****"))

	return nil
}

func processHeader(hdr []interface{}) map[int]string {
	hdrMap := make(map[int]string)
	for idx, attr := range hdr {
		strAttr := fmt.Sprintf("%v", attr)
		hdrMap[idx] = strings.ToLower(strAttr)
	}

	return hdrMap
}

func processSheet(ds *admin.Service, sheetID string) error {
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

	hdrMap := processHeader(sValRange.Values[0])
	err = validateHeader(hdrMap)
	if err != nil {
		return err
	}

	for idx, row := range sValRange.Values {
		if idx == 0 {
			continue
		}

		err := processUser(ds, hdrMap, row)
		if err != nil {
			return err
		}
	}

	return nil
}

func processUser(ds *admin.Service, hdrMap map[int]string, userData []interface{}) error {
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
				return err
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

	err := insertNewUser(ds, user)
	if err != nil {
		return err
	}

	return nil
}

func validateHeader(hdr map[int]string) error {
	for idx, hdrAttr := range hdr {
		correctVal, err := cmn.IsValidAttr(hdrAttr, usrs.UserAttrMap)
		if err != nil {
			return err
		}
		hdr[idx] = correctVal
	}
	return nil
}
