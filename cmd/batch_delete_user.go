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
	"errors"
	"fmt"
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

var batchDelUserCmd = &cobra.Command{
	Use:     "users [-i input file path]",
	Aliases: []string{"user", "usrs", "usr"},
	Example: `gmin batch-delete users -i inputfile.txt
gmin bdel user -i inputfile.txt
gmin ls user -a primaryemail -q orgunitpath=/TestOU | jq '.users[] | .primaryEmail' -r | gmin bdel user`,
	Short: "Deletes a batch of users",
	Long: `Deletes a batch of users where user details are provided in a text input file or from a pipe.
			
The input file or piped in data should provide the user email addresses, aliases or ids to be deleted on separate lines like this:

frank.castle@mycompany.com
bruce.wayne@mycompany.com
peter.parker@mycompany.com

An input Google sheet must have a header row with the following column names being the only ones that are valid:

userKey [required]

The column name is case insensitive.`,
	RunE: doBatchDelUser,
}

func doBatchDelUser(cmd *cobra.Command, args []string) error {
	logger.Debugw("starting doBatchDelUser()",
		"args", args)

	var users []string

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
		err := errors.New(gmess.ErrNoInputFile)
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
		err = fmt.Errorf(gmess.ErrInvalidFileFormat, formatFlgVal)
		logger.Error(err)
		return err
	}

	switch {
	case lwrFmt == "text":
		users, err = bduProcessTextFile(ds, inputFlgVal, scanner)
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

		users, err = bduProcessGSheet(ds, inputFlgVal, rangeFlgVal)
		if err != nil {
			logger.Error(err)
			return err
		}
	default:
		return fmt.Errorf(gmess.ErrInvalidFileFormat, formatFlgVal)
	}

	err = bduProcessDeletion(ds, users)
	if err != nil {
		logger.Error(err)
		return err
	}

	logger.Debug("finished doBatchDelUser()")
	return nil
}

func bduDelete(wg *sync.WaitGroup, udc *admin.UsersDeleteCall, user string) {
	logger.Debugw("starting bduDelete()",
		"user", user)

	defer wg.Done()

	b := backoff.NewExponentialBackOff()
	b.MaxElapsedTime = 32 * time.Second

	err := backoff.Retry(func() error {
		var err error

		err = udc.Do()
		if err == nil {
			logger.Infof(gmess.InfoUserDeleted, user)
			fmt.Println(cmn.GminMessage(fmt.Sprintf(gmess.InfoUserDeleted, user)))
			return err
		}
		if !cmn.IsErrRetryable(err) {
			return backoff.Permanent(fmt.Errorf(gmess.ErrBatchUser, err.Error(), user))
		}
		// Log the retries
		logger.Warnw(err.Error(),
			"retrying", b.GetElapsedTime().String(),
			"user", user)
		return fmt.Errorf(gmess.ErrBatchUser, err.Error(), user)
	}, b)
	if err != nil {
		// Log final error
		logger.Error(err)
		fmt.Println(cmn.GminMessage(err.Error()))
	}
	logger.Debug("finished bduDelete()")
}

func bduFromFileFactory(hdrMap map[int]string, userData []interface{}) (string, error) {
	logger.Debugw("starting bduFromFileFactory()",
		"hdrMap", hdrMap)

	var user string

	for idx, val := range userData {
		attrName := hdrMap[idx]
		attrVal := fmt.Sprintf("%v", val)

		if attrName == "userKey" {
			user = attrVal
		}
	}
	logger.Debug("finished bduFromFileFactory()")
	return user, nil
}

func bduProcessDeletion(ds *admin.Service, users []string) error {
	logger.Debug("starting bduProcessDeletion()")

	wg := new(sync.WaitGroup)

	for _, user := range users {
		udc := ds.Users.Delete(user)

		wg.Add(1)

		go bduDelete(wg, udc, user)
	}

	wg.Wait()

	logger.Debug("finished bduProcessDeletion()")
	return nil
}

func bduProcessGSheet(ds *admin.Service, sheetID string, sheetrange string) ([]string, error) {
	logger.Debugw("starting bduProcessGSheet()",
		"sheetID", sheetID,
		"sheetrange", sheetrange)

	var users []string

	if sheetrange == "" {
		err := errors.New(gmess.ErrNoSheetRange)
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
		err = fmt.Errorf(gmess.ErrNoSheetDataFound, sheetID, sheetrange)
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

		userVar, err := bduFromFileFactory(hdrMap, row)
		if err != nil {
			logger.Error(err)
			return nil, err
		}

		users = append(users, userVar)
	}

	logger.Debug("finished bduProcessGSheet()")
	return users, nil
}

func bduProcessTextFile(ds *admin.Service, filePath string, scanner *bufio.Scanner) ([]string, error) {
	logger.Debugw("starting bduProcessTextFile()",
		"filePath", filePath)

	var users []string

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
		user := scanner.Text()
		users = append(users, user)
	}

	logger.Debug("finished bduProcessTextFile()")
	return users, nil
}

func init() {
	batchDelCmd.AddCommand(batchDelUserCmd)

	batchDelUserCmd.Flags().StringVarP(&inputFile, "input-file", "i", "", "filepath to user data text file")
	batchDelUserCmd.Flags().StringVarP(&delFormat, "format", "f", "text", "user data file format (text or gsheet)")
	batchDelUserCmd.Flags().StringVarP(&sheetRange, "sheet-range", "s", "", "user data gsheet range")
}
