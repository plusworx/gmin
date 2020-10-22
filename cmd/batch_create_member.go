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
	mems "github.com/plusworx/gmin/utils/members"
	"github.com/spf13/cobra"
	admin "google.golang.org/api/admin/directory/v1"
)

var batchCrtMemberCmd = &cobra.Command{
	Use:     "group-members <group email address or id> -i <input file path or google sheet id>",
	Aliases: []string{"group-member", "grp-members", "grp-member", "gmembers", "gmember", "gmems", "gmem"},
	Args:    cobra.ExactArgs(1),
	Example: `gmin batch-create group-members engineering@mycompany.com -i inputfile.json
gmin bcrt gmems sales@mycompany.com -i inputfile.csv -f csv
gmin bcrt gmem finance@mycompany.com -i 1odyAIp3jGspd3M4xeepxWD6aeQIUuHBgrZB2OHSu8MI -s 'Sheet1!A1:K25' -f gsheet`,
	Short: "Creates a batch of group members",
	Long: `Creates a batch of group members where group member details are provided in a Google Sheet, CSV/JSON input file or piped JSON.
			
The contents of the JSON file or piped input should look something like this:
	
{"delivery_settings":"DIGEST","email":"kayden.yundt@mycompany.com","role":"MEMBER"}
{"delivery_settings":"ALL_MAIL","email":"kenyatta.tillman@mycompany.com","role":"MANAGER"}
{"delivery_settings":"DAILY","email":"keon.stroman@mycompany.com","role":"MEMBER"}
	
CSV and Google sheets must have a header row with the following column names being the only ones that are valid:
	
delivery_settings
email [required]
role

The column names are case insensitive and can be in any order.`,
	RunE: doBatchCrtMember,
}

func doBatchCrtMember(cmd *cobra.Command, args []string) error {
	lg.Debugw("starting doBatchCrtMember()",
		"args", args)
	defer lg.Debug("finished doBatchCrtMember()")

	var (
		members []*admin.Member
		objs    []interface{}
	)

	srv, err := cmn.CreateService(cmn.SRVTYPEADMIN, admin.AdminDirectoryGroupMemberScope)
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

	groupKey := args[0]

	switch {
	case lwrFmt == "csv":
		objs, err = btch.CreateProcessCSVFile(cmn.OBJTYPEMEMBER, inputFlgVal, mems.MemberAttrMap)
		if err != nil {
			return err
		}
	case lwrFmt == "json":
		objs, err = btch.CreateProcessJSON(cmn.OBJTYPEMEMBER, inputFlgVal, scanner, mems.MemberAttrMap)
		if err != nil {
			return err
		}
	case lwrFmt == "gsheet":
		rangeFlgVal, err := cmd.Flags().GetString(flgnm.FLG_SHEETRANGE)
		if err != nil {
			return err
		}

		objs, err = btch.CreateProcessGSheet(cmn.OBJTYPEMEMBER, inputFlgVal, rangeFlgVal, mems.MemberAttrMap)
		if err != nil {
			return err
		}
	default:
		err = fmt.Errorf(gmess.ERR_INVALIDFILEFORMAT, formatFlgVal)
		lg.Error(err)
		return err
	}

	for _, memObj := range objs {
		members = append(members, memObj.(*admin.Member))
	}

	err = bcmProcessObjects(ds, groupKey, members)
	if err != nil {
		return err
	}

	return nil
}

func bcmCreate(member *admin.Member, groupKey string, wg *sync.WaitGroup, mic *admin.MembersInsertCall) {
	lg.Debugw("starting bcmCreate()",
		"groupKey", groupKey)
	defer lg.Debug("finished bcmCreate()")

	defer wg.Done()

	b := backoff.NewExponentialBackOff()
	b.MaxElapsedTime = 32 * time.Second

	err := backoff.Retry(func() error {
		var err error
		newMember, err := mic.Do()
		if err == nil {
			fmt.Println(cmn.GminMessage(fmt.Sprintf(gmess.INFO_MEMBERCREATED, newMember.Email, groupKey)))
			lg.Infof(gmess.INFO_MEMBERCREATED, newMember.Email, groupKey)
			return err
		}
		if !cmn.IsErrRetryable(err) {
			return backoff.Permanent(fmt.Errorf(gmess.ERR_BATCHMEMBER, err.Error(), member.Email, groupKey))
		}
		// Log the retries
		lg.Warnw(err.Error(),
			"retrying", b.GetElapsedTime().String(),
			"group", groupKey,
			"member", member.Email)
		return fmt.Errorf(gmess.ERR_BATCHMEMBER, err.Error(), member.Email, groupKey)
	}, b)
	if err != nil {
		// Log final error
		lg.Error(err)
		fmt.Println(cmn.GminMessage(err.Error()))
	}
}

func bcmProcessObjects(ds *admin.Service, groupKey string, members []*admin.Member) error {
	lg.Debugw("starting bcmProcessObjects()",
		"groupKey", groupKey)
	defer lg.Debug("finished bcmProcessObjects()")

	wg := new(sync.WaitGroup)

	for _, m := range members {
		if m.Email == "" {
			err := errors.New(gmess.ERR_NOMEMBEREMAILADDRESS)
			lg.Error(err)
			return err
		}

		mic := ds.Members.Insert(groupKey, m)

		wg.Add(1)

		go bcmCreate(m, groupKey, wg, mic)
	}

	wg.Wait()

	return nil
}

func init() {
	batchCreateCmd.AddCommand(batchCrtMemberCmd)

	batchCrtMemberCmd.Flags().StringVarP(&inputFile, flgnm.FLG_INPUTFILE, "i", "", "filepath to group member data file or sheet id")
	batchCrtMemberCmd.Flags().StringVarP(&format, flgnm.FLG_FORMAT, "f", "json", "user data file format")
	batchCrtMemberCmd.Flags().StringVarP(&sheetRange, flgnm.FLG_SHEETRANGE, "s", "", "user data gsheet range")
}
