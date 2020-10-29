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

var batchCrtUserCmd = &cobra.Command{
	Use:     "users -i <input file path or google sheet id>",
	Aliases: []string{"user", "usrs", "usr"},
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
	lg.Debugw("starting doBatchCrtUser()",
		"args", args)
	defer lg.Debug("finished doBatchCrtUser()")

	var (
		objs  []interface{}
		users []*admin.User
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

	callParams := btch.CallParams{CallType: cmn.CALLTYPECREATE, ObjectType: cmn.OBJTYPEUSER}

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

	for _, userObj := range objs {
		users = append(users, userObj.(*admin.User))
	}

	err = bcuProcessObjects(ds, users)
	if err != nil {
		return err
	}

	return nil
}

func bcuCreate(user *admin.User, wg *sync.WaitGroup, uic *admin.UsersInsertCall) {
	lg.Debug("starting bcuCreate()")
	defer lg.Debug("finished bcuCreate()")

	defer wg.Done()

	b := backoff.NewExponentialBackOff()
	b.MaxElapsedTime = 32 * time.Second

	err := backoff.Retry(func() error {
		var err error
		newUser, err := uic.Do()
		if err == nil {
			fmt.Println(cmn.GminMessage(fmt.Sprintf(gmess.INFO_USERCREATED, newUser.PrimaryEmail)))
			lg.Infof(gmess.INFO_USERCREATED, newUser.PrimaryEmail)
			return err
		}
		if !cmn.IsErrRetryable(err) {
			return backoff.Permanent(fmt.Errorf(gmess.ERR_BATCHUSER, err.Error(), user.PrimaryEmail))
		}
		// Log the retries
		lg.Warnw(err.Error(),
			"retrying", b.GetElapsedTime().String(),
			"user", user.PrimaryEmail)
		return fmt.Errorf(gmess.ERR_BATCHUSER, err.Error(), user.PrimaryEmail)
	}, b)
	if err != nil {
		// Log final error
		lg.Error(err)
		fmt.Println(cmn.GminMessage(err.Error()))
	}
}

func bcuProcessObjects(ds *admin.Service, users []*admin.User) error {
	lg.Debug("starting bcuProcessObjects()")
	defer lg.Debug("finished bcuProcessObjects()")

	wg := new(sync.WaitGroup)

	for _, u := range users {
		if u.PrimaryEmail == "" || u.Name.GivenName == "" || u.Name.FamilyName == "" || u.Password == "" {
			err := errors.New(gmess.ERR_BATCHMISSINGUSERDATA)
			lg.Error(err)
			return err
		}

		u.HashFunction = usrs.HASHFUNCTION
		pwd, err := usrs.HashPassword(u.Password)
		if err != nil {
			return err
		}
		u.Password = pwd

		uic := ds.Users.Insert(u)

		wg.Add(1)

		go bcuCreate(u, wg, uic)
	}

	wg.Wait()

	return nil
}

func init() {
	batchCreateCmd.AddCommand(batchCrtUserCmd)

	batchCrtUserCmd.Flags().StringVarP(&inputFile, flgnm.FLG_INPUTFILE, "i", "", "filepath to user data file or sheet id")
	batchCrtUserCmd.Flags().StringVarP(&format, flgnm.FLG_FORMAT, "f", "json", "user data file format")
	batchCrtUserCmd.Flags().StringVarP(&sheetRange, flgnm.FLG_SHEETRANGE, "s", "", "user data gsheet range")
}
