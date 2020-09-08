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

var batchUndelUserCmd = &cobra.Command{
	Use:     "users -i <input file path>",
	Aliases: []string{"user"},
	Short:   "Undeletes a batch of users",
	Long: `Undeletes a batch of users where user details are provided in a Google Sheet or CSV/JSON input file.
	
	Examples:	gmin batch-undelete users -i inputfile.json
			gmin bund user -i inputfile.csv -f csv
			gmin bund user -i 1odyAIp3jGspd3M4xeepxWD6aeQIUuHBgrZB2OHSu8MI -s 'Sheet1!A1:B25' -f gsheet
			
	The contents of a JSON file should look something like this:
	
	{"userKey":"417578192529765228417","orgUnitPath":"/Sales"}
	{"userKey":"308127142904731923463","orgUnitPath":"/"}
	{"userKey":"107967172367714327529","orgUnitPath":"/Engineering"}

	N.B. userKey must be the unique user id and NOT email address
	
	CSV and Google sheets must have a header row with the following column names being the only ones that are valid:
	
	orgUnitPath [required]
	userKey [required]
	
	The column names are case insensitive and can be in any order.`,
	RunE: doBatchUndelUser,
}

func doBatchUndelUser(cmd *cobra.Command, args []string) error {
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
		err := btchUndelUsrProcessCSV(ds, inputFile)
		if err != nil {
			return err
		}
	case lwrFmt == "json":
		err := btchUndelUsrProcessJSON(ds, inputFile)
		if err != nil {
			return err
		}
	case lwrFmt == "gsheet":
		err := btchUndelUsrProcessSheet(ds, inputFile)
		if err != nil {
			return err
		}
	}

	return nil
}

func btchUndelJSONUser(ds *admin.Service, jsonData string) (usrs.UndeleteUser, error) {
	undelUser := usrs.UndeleteUser{}
	jsonBytes := []byte(jsonData)

	if !json.Valid(jsonBytes) {
		return undelUser, errors.New("gmin: error - attribute string is not valid JSON")
	}

	outStr, err := cmn.ParseInputAttrs(jsonBytes)
	if err != nil {
		return undelUser, err
	}

	err = cmn.ValidateInputAttrs(outStr, usrs.UserAttrMap)
	if err != nil {
		return undelUser, err
	}

	err = json.Unmarshal(jsonBytes, &undelUser)
	if err != nil {
		return undelUser, err
	}

	return undelUser, nil
}

func btchUndelUsers(ds *admin.Service, undelUsers []usrs.UndeleteUser) error {
	wg := new(sync.WaitGroup)

	for _, u := range undelUsers {
		userUndelete := admin.UserUndelete{}

		if u.OrgUnitPath == "" {
			userUndelete.OrgUnitPath = "/"
		} else {
			userUndelete.OrgUnitPath = u.OrgUnitPath
		}

		uuc := ds.Users.Undelete(u.UserKey, &userUndelete)

		wg.Add(1)

		go btchUsrUndelProcess(u.UserKey, wg, uuc)
	}

	wg.Wait()

	return nil
}

func btchUsrUndelProcess(userKey string, wg *sync.WaitGroup, uuc *admin.UsersUndeleteCall) {
	defer wg.Done()

	b := backoff.NewExponentialBackOff()
	b.MaxElapsedTime = 32 * time.Second

	err := backoff.Retry(func() error {
		var err error
		err = uuc.Do()
		if err == nil {
			fmt.Println(cmn.GminMessage("**** gmin: user " + userKey + " undeleted ****"))
			return err
		}
		if !cmn.IsErrRetryable(err) {
			return backoff.Permanent(errors.New(cmn.GminMessage("gmin: error - " + err.Error() + " " + userKey)))
		}
		return errors.New(cmn.GminMessage("gmin: error - " + err.Error() + " " + userKey))
	}, b)
	if err != nil {
		fmt.Println(err)
	}
}

func btchUndelUsrProcessCSV(ds *admin.Service, filePath string) error {
	var (
		iSlice     []interface{}
		hdrMap     = map[int]string{}
		undelUsers []usrs.UndeleteUser
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

		undelUserVar, err := btchUndelProcessUser(hdrMap, iSlice)
		if err != nil {
			fmt.Println(err.Error())
		}

		undelUsers = append(undelUsers, undelUserVar)

		count = count + 1
	}

	err = btchUndelUsers(ds, undelUsers)
	if err != nil {
		return err
	}
	return nil
}

func btchUndelUsrProcessJSON(ds *admin.Service, filePath string) error {
	var undelUsers []usrs.UndeleteUser

	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		jsonData := scanner.Text()

		undelUserVar, err := btchUndelJSONUser(ds, jsonData)
		if err != nil {
			return err
		}

		undelUsers = append(undelUsers, undelUserVar)
	}
	err = scanner.Err()
	if err != nil {
		return err
	}

	err = btchUndelUsers(ds, undelUsers)
	if err != nil {
		return err
	}

	return nil
}

func btchUndelUsrProcessSheet(ds *admin.Service, sheetID string) error {
	var undelUsers []usrs.UndeleteUser

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

		undelUserVar, err := btchUndelProcessUser(hdrMap, row)
		if err != nil {
			return err
		}

		undelUsers = append(undelUsers, undelUserVar)
	}

	err = btchUndelUsers(ds, undelUsers)
	if err != nil {
		return err
	}

	return nil
}

func btchUndelProcessUser(hdrMap map[int]string, userData []interface{}) (usrs.UndeleteUser, error) {
	undelUser := usrs.UndeleteUser{}

	for idx, attr := range userData {
		attrName := hdrMap[idx]

		switch {
		case attrName == "userKey":
			undelUser.UserKey = fmt.Sprintf("%v", attr)
		case attrName == "orgUnitPath":
			undelUser.OrgUnitPath = fmt.Sprintf("%v", attr)
		}
	}

	return undelUser, nil
}

func init() {
	batchUndeleteCmd.AddCommand(batchUndelUserCmd)

	batchUndelUserCmd.Flags().StringVarP(&inputFile, "inputfile", "i", "", "filepath to user data file or sheet id")
	batchUndelUserCmd.Flags().StringVarP(&format, "format", "f", "json", "user data file format")
	batchUndelUserCmd.Flags().StringVarP(&sheetRange, "sheetrange", "s", "", "user data gsheet range")

	batchUndelUserCmd.MarkFlagRequired("inputfile")
}
