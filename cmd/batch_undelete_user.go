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
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/cenkalti/backoff/v4"
	btch "github.com/plusworx/gmin/utils/batch"
	cmn "github.com/plusworx/gmin/utils/common"
	flgnm "github.com/plusworx/gmin/utils/flagnames"
	gmess "github.com/plusworx/gmin/utils/gminmessages"
	lg "github.com/plusworx/gmin/utils/logging"
	usrs "github.com/plusworx/gmin/utils/users"
	"github.com/spf13/cobra"
	admin "google.golang.org/api/admin/directory/v1"
)

var batchUndelUserCmd = &cobra.Command{
	Use:     "users -i <input file path>",
	Aliases: []string{"user", "usrs", "usr"},
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
	lg.Debugw("starting doBatchUndelUser()",
		"args", args)
	defer lg.Debug("finished doBatchUndelUser()")

	var (
		objs       []interface{}
		undelUsers []usrs.UndeleteUser
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

	callParams := btch.CallParams{CallType: cmn.CALLTYPEUNDELETE, ObjectType: cmn.OBJTYPEUSER}

	switch {
	case lwrFmt == "csv":
		objs, err = btch.ProcessCSVFile(callParams, inputFlgVal, usrs.UserAttrMap)
		if err != nil {
			return err
		}
	case lwrFmt == "json":
		objs, err = btch.ProcessJSON(callParams, inputFlgVal, scanner, usrs.UserAttrMap)
		if err != nil {
			return err
		}
	case lwrFmt == "gsheet":
		rangeFlgVal, err := cmd.Flags().GetString(flgnm.FLG_SHEETRANGE)
		if err != nil {
			lg.Error(err)
			return err
		}

		objs, err = btch.ProcessGSheet(callParams, inputFlgVal, rangeFlgVal, usrs.UserAttrMap)
		if err != nil {
			return err
		}
	default:
		err = fmt.Errorf(gmess.ERR_INVALIDFILEFORMAT, formatFlgVal)
		lg.Error(err)
		return err
	}

	for _, uuObj := range objs {
		undelUsers = append(undelUsers, uuObj.(usrs.UndeleteUser))
	}

	err = bunduProcessObjects(ds, undelUsers)
	if err != nil {
		return err
	}

	return nil
}

func bunduProcessObjects(ds *admin.Service, undelUsers []usrs.UndeleteUser) error {
	lg.Debug("starting bunduProcessObjects()")
	defer lg.Debug("finished bunduProcessObjects()")

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

	return nil
}

func bunduUndelete(userKey string, wg *sync.WaitGroup, uuc *admin.UsersUndeleteCall) {
	lg.Debugw("starting bunduUndelete()",
		"userKey", userKey)
	defer lg.Debug("finished bunduUndelete()")

	defer wg.Done()

	b := backoff.NewExponentialBackOff()
	b.MaxElapsedTime = 32 * time.Second

	err := backoff.Retry(func() error {
		var err error
		err = uuc.Do()
		if err == nil {
			fmt.Println(cmn.GminMessage(fmt.Sprintf(gmess.INFO_USERUNDELETED, userKey)))
			lg.Infof(gmess.INFO_USERUNDELETED, userKey)
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
}

func init() {
	batchUndeleteCmd.AddCommand(batchUndelUserCmd)

	batchUndelUserCmd.Flags().StringVarP(&inputFile, flgnm.FLG_INPUTFILE, "i", "", "filepath to user data file or sheet id")
	batchUndelUserCmd.Flags().StringVarP(&format, flgnm.FLG_FORMAT, "f", "json", "user data file format")
	batchUndelUserCmd.Flags().StringVarP(&sheetRange, flgnm.FLG_SHEETRANGE, "s", "", "user data gsheet range")
}
