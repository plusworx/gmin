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

var batchUndelUserCmd = &cobra.Command{
	Use:     "users -i <input file path>",
	Aliases: []string{"user"},
	Example: `gmin batch-undelete users -i inputfile.json
gmin bund user -i inputfile.csv -f csv
gmin bund user -i 1odyAIp3jGspd3M4xeepxWD6aeQIUuHBgrZB2OHSu8MI -s 'Sheet1!A1:B25' -f gsheet`,
	Short: "Undeletes a batch of users",
	Long: `Undeletes a batch of users where user details are provided in a Google Sheet, CSV/JSON input file or piped JSON.
			
The contents of a JSON file or piped input should look something like this:

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
	logger.Debugw("starting doBatchUndelUser()",
		"args", args)

	var undelUsers []usrs.UndeleteUser

	ds, err := cmn.CreateDirectoryService(admin.AdminDirectoryUserScope)
	if err != nil {
		logger.Error(err)
		return err
	}

	inputFlgVal, err := cmd.Flags().GetString("input-file")
	if err != nil {
		logger.Error(err)
		return err
	}

	scanner, err := cmn.InputFromStdIn(inputFlgVal)
	if err != nil {
		logger.Error(err)
		return err
	}

	if inputFlgVal == "" && scanner == nil {
		err := errors.New(gmess.ERRNOINPUTFILE)
		logger.Error(err)
		return err
	}

	formatFlgVal, err := cmd.Flags().GetString("format")
	if err != nil {
		logger.Error(err)
		return err
	}
	lwrFmt := strings.ToLower(formatFlgVal)

	ok := cmn.SliceContainsStr(cmn.ValidFileFormats, lwrFmt)
	if !ok {
		err = fmt.Errorf(gmess.ERRINVALIDFILEFORMAT, formatFlgVal)
		logger.Error(err)
		return err
	}

	switch {
	case lwrFmt == "csv":
		undelUsers, err = bunduProcessCSVFile(ds, inputFlgVal)
		if err != nil {
			logger.Error(err)
			return err
		}
	case lwrFmt == "json":
		undelUsers, err = bunduProcessJSON(ds, inputFlgVal, scanner)
		if err != nil {
			logger.Error(err)
			return err
		}
	case lwrFmt == "gsheet":
		rangeFlgVal, err := cmd.Flags().GetString("sheet-range")
		if err != nil {
			logger.Error(err)
			return err
		}

		undelUsers, err = bunduProcessGSheet(ds, inputFlgVal, rangeFlgVal)
		if err != nil {
			logger.Error(err)
			return err
		}
	default:
		return fmt.Errorf(gmess.ERRINVALIDFILEFORMAT, formatFlgVal)
	}

	err = bunduProcessObjects(ds, undelUsers)
	if err != nil {
		logger.Error(err)
		return err
	}

	logger.Debug("finished doBatchUndelUser()")
	return nil
}

func bunduFromFileFactory(hdrMap map[int]string, userData []interface{}) (usrs.UndeleteUser, error) {
	logger.Debugw("starting bunduFromFileFactory()",
		"hdrMap", hdrMap)

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
	logger.Debug("finished bunduFromFileFactory()")
	return undelUser, nil
}

func bunduFromJSONFactory(ds *admin.Service, jsonData string) (usrs.UndeleteUser, error) {
	logger.Debugw("starting bunduFromJSONFactory()",
		"jsonData", jsonData)

	undelUser := usrs.UndeleteUser{}
	jsonBytes := []byte(jsonData)

	if !json.Valid(jsonBytes) {
		logger.Error(gmess.ERRINVALIDJSONATTR)
		return undelUser, errors.New(gmess.ERRINVALIDJSONATTR)
	}

	outStr, err := cmn.ParseInputAttrs(jsonBytes)
	if err != nil {
		logger.Error(err)
		return undelUser, err
	}

	err = cmn.ValidateInputAttrs(outStr, usrs.UserAttrMap)
	if err != nil {
		logger.Error(err)
		return undelUser, err
	}

	err = json.Unmarshal(jsonBytes, &undelUser)
	if err != nil {
		logger.Error(err)
		return undelUser, err
	}
	logger.Debug("finished bunduFromJSONFactory()")
	return undelUser, nil
}

func bunduProcessCSVFile(ds *admin.Service, filePath string) ([]usrs.UndeleteUser, error) {
	logger.Debugw("starting bunduProcessCSVFile()",
		"filePath", filePath)

	var (
		iSlice     []interface{}
		hdrMap     = map[int]string{}
		undelUsers []usrs.UndeleteUser
	)

	csvfile, err := os.Open(filePath)
	if err != nil {
		logger.Error(err)
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
			logger.Error(err)
			return nil, err
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
				return nil, err
			}
			count = count + 1
			continue
		}

		for idx, value := range record {
			iSlice[idx] = value
		}

		undelUserVar, err := bunduFromFileFactory(hdrMap, iSlice)
		if err != nil {
			logger.Error(err)
			return nil, err
		}

		undelUsers = append(undelUsers, undelUserVar)

		count = count + 1
	}

	logger.Debug("finished bunduProcessCSVFile()")
	return undelUsers, nil
}

func bunduProcessGSheet(ds *admin.Service, sheetID string, sheetrange string) ([]usrs.UndeleteUser, error) {
	logger.Debugw("starting bunduProcessGSheet()",
		"sheetID", sheetID,
		"sheetrange", sheetrange)

	var undelUsers []usrs.UndeleteUser

	if sheetrange == "" {
		err := errors.New(gmess.ERRNOSHEETRANGE)
		logger.Error(err)
		return nil, err
	}

	ss, err := cmn.CreateSheetService(sheet.DriveReadonlyScope)
	if err != nil {
		logger.Error(err)
		return nil, err
	}

	ssvgc := ss.Spreadsheets.Values.Get(sheetID, sheetrange)
	sValRange, err := ssvgc.Do()
	if err != nil {
		logger.Error(err)
		return nil, err
	}

	if len(sValRange.Values) == 0 {
		err = fmt.Errorf(gmess.ERRNOSHEETDATAFOUND, sheetID, sheetrange)
		logger.Error(err)
		return nil, err
	}

	hdrMap := cmn.ProcessHeader(sValRange.Values[0])
	err = cmn.ValidateHeader(hdrMap, usrs.UserAttrMap)
	if err != nil {
		logger.Error(err)
		return nil, err
	}

	for idx, row := range sValRange.Values {
		if idx == 0 {
			continue
		}

		undelUserVar, err := bunduFromFileFactory(hdrMap, row)
		if err != nil {
			logger.Error(err)
			return nil, err
		}

		undelUsers = append(undelUsers, undelUserVar)
	}

	logger.Debug("finished bunduProcessGSheet()")
	return undelUsers, nil
}

func bunduProcessJSON(ds *admin.Service, filePath string, scanner *bufio.Scanner) ([]usrs.UndeleteUser, error) {
	logger.Debugw("starting bunduProcessJSON()",
		"filePath", filePath)

	var undelUsers []usrs.UndeleteUser

	if filePath != "" {
		file, err := os.Open(filePath)
		if err != nil {
			logger.Error(err)
			return nil, err
		}
		defer file.Close()

		scanner = bufio.NewScanner(file)
	}

	for scanner.Scan() {
		jsonData := scanner.Text()

		undelUserVar, err := bunduFromJSONFactory(ds, jsonData)
		if err != nil {
			logger.Error(err)
			return nil, err
		}

		undelUsers = append(undelUsers, undelUserVar)
	}
	err := scanner.Err()
	if err != nil {
		logger.Error(err)
		return nil, err
	}

	logger.Debug("finished bunduProcessJSON()")
	return undelUsers, nil
}

func bunduProcessObjects(ds *admin.Service, undelUsers []usrs.UndeleteUser) error {
	logger.Debug("starting bunduProcessObjects()")
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

		go bunduUndelete(u.UserKey, wg, uuc)
	}

	wg.Wait()

	logger.Debug("finished bunduProcessObjects()")
	return nil
}

func bunduUndelete(userKey string, wg *sync.WaitGroup, uuc *admin.UsersUndeleteCall) {
	logger.Debugw("starting bunduUndelete()",
		"userKey", userKey)

	defer wg.Done()

	b := backoff.NewExponentialBackOff()
	b.MaxElapsedTime = 32 * time.Second

	err := backoff.Retry(func() error {
		var err error
		err = uuc.Do()
		if err == nil {
			logger.Infof(gmess.INFOUSERUNDELETED, userKey)
			fmt.Println(cmn.GminMessage(fmt.Sprintf(gmess.INFOUSERUNDELETED, userKey)))
			return err
		}
		if !cmn.IsErrRetryable(err) {
			return backoff.Permanent(fmt.Errorf(gmess.ERRBATCHUSER, err.Error(), userKey))
		}
		// Log the retries
		logger.Warnw(err.Error(),
			"retrying", b.GetElapsedTime().String(),
			"user", userKey)
		return fmt.Errorf(gmess.ERRBATCHUSER, err.Error(), userKey)
	}, b)
	if err != nil {
		// Log final error
		logger.Error(err)
		fmt.Println(cmn.GminMessage(err.Error()))
	}
	logger.Debug("finished bunduUndelete()")
}

func init() {
	batchUndeleteCmd.AddCommand(batchUndelUserCmd)

	batchUndelUserCmd.Flags().StringVarP(&inputFile, "input-file", "i", "", "filepath to user data file or sheet id")
	batchUndelUserCmd.Flags().StringVarP(&format, "format", "f", "json", "user data file format")
	batchUndelUserCmd.Flags().StringVarP(&sheetRange, "sheet-range", "s", "", "user data gsheet range")
}
