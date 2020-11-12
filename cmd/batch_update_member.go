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

var batchUpdMemberCmd = &cobra.Command{
	Use:     "group-members <group email address, alias or id> -i <input file path>",
	Aliases: []string{"group-member", "grp-members", "grp-member", "gmembers", "gmember", "gmems", "gmem"},
	Args:    cobra.ExactArgs(1),
	Example: `gmin batch-update group-members sales@mycompany.com -i inputfile.json
gmin bupd gmems sales@mycompany.com -i inputfile.csv -f csv
gmin bupd gmem finance@mycompany.com -i 1odyAIp3jGspd3M4xeepxWD6aeQIUuHBgrZB2OHSu8MI -s 'Sheet1!A1:K25' -f gsheet`,
	Short: "Updates a batch of group members",
	Long: `Updates a batch of group members where group member details are provided in a Google Sheet, CSV/JSON input file or piped JSON.
			  
The contents of the JSON file or piped input should look something like this:

{"memberKey":"rudolph.brown@mycompany.com","delivery_settings":"DIGEST","role":"MANAGER"}
{"memberKey":"emily.bronte@mycompany.com","delivery_settings":"DAILY","role":"MEMBER"}
{"memberKey":"charles.dickens@mycompany.com","delivery_settings":"NONE","role":"MEMBER"}

N.B. memberKey (member email address, alias or id) must be provided.

CSV and Google sheets must have a header row with the following column names being the only ones that are valid:

delivery_settings
memberKey [required]
role

The column names are case insensitive and can be in any order.`,
	RunE: doBatchUpdMember,
}

func doBatchUpdMember(cmd *cobra.Command, args []string) error {
	lg.Debugw("starting doBatchUpdMember()",
		"args", args)
	defer lg.Debug("finished doBatchUpdMember()")

	var memParams []mems.MemberParams

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

	rangeFlgVal, err := cmd.Flags().GetString(flgnm.FLG_SHEETRANGE)
	if err != nil {
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

	callParams := btch.CallParams{CallType: cmn.CALLTYPEUPDATE, ObjectType: cmn.OBJTYPEMEMBER}
	inputParams := btch.ProcessInputParams{
		Format:      lwrFmt,
		InputFlgVal: inputFlgVal,
		Scanner:     scanner,
		SheetRange:  rangeFlgVal,
	}

	objs, err := bumProcessInput(callParams, inputParams)
	if err != nil {
		return err
	}

	for _, memObj := range objs {
		memParams = append(memParams, memObj.(mems.MemberParams))
	}

	err = bumProcessObjects(ds, groupKey, memParams)
	if err != nil {
		return err
	}

	return nil
}

func bumProcessInput(callParams btch.CallParams, inputParams btch.ProcessInputParams) ([]interface{}, error) {
	lg.Debug("starting bumProcessInput()")
	defer lg.Debug("finished bumProcessInput()")

	var (
		err  error
		objs []interface{}
	)

	switch inputParams.Format {
	case "csv":
		objs, err = btch.ProcessCSVFile(callParams, inputParams.InputFlgVal, mems.MemberAttrMap)
		if err != nil {
			return nil, err
		}
	case "json":
		objs, err = btch.ProcessJSON(callParams, inputParams.InputFlgVal, inputParams.Scanner, mems.MemberAttrMap)
		if err != nil {
			return nil, err
		}
	case "gsheet":
		objs, err = btch.ProcessGSheet(callParams, inputParams.InputFlgVal, inputParams.SheetRange, mems.MemberAttrMap)
		if err != nil {
			return nil, err
		}
	default:
		err = fmt.Errorf(gmess.ERR_INVALIDFILEFORMAT, inputParams.Format)
		lg.Error(err)
		return nil, err
	}

	return objs, nil
}

func bumProcessObjects(ds *admin.Service, groupKey string, memParams []mems.MemberParams) error {
	lg.Debugw("starting bumProcessObjects()",
		"groupKey", groupKey,
		"memParams", memParams)
	defer lg.Debug("finished bumProcessObjects()")

	wg := new(sync.WaitGroup)

	for _, mp := range memParams {
		muc := ds.Members.Update(groupKey, mp.MemberKey, mp.Member)

		wg.Add(1)

		go bumUpdate(mp.Member, groupKey, wg, muc, mp.MemberKey)
	}

	wg.Wait()

	return nil
}

func bumUpdate(member *admin.Member, groupKey string, wg *sync.WaitGroup, muc *admin.MembersUpdateCall, memKey string) {
	lg.Debugw("starting bumUpdate()",
		"groupKey", groupKey,
		"memKey", memKey)
	defer lg.Debug("finished bumUpdate()")

	defer wg.Done()

	b := backoff.NewExponentialBackOff()
	b.MaxElapsedTime = 32 * time.Second

	err := backoff.Retry(func() error {
		var err error
		_, err = muc.Do()
		if err == nil {
			fmt.Println(cmn.GminMessage(fmt.Sprintf(gmess.INFO_MEMBERUPDATED, memKey, groupKey)))
			lg.Infof(gmess.INFO_MEMBERUPDATED, memKey, groupKey)
			return err
		}
		if !cmn.IsErrRetryable(err) {
			return backoff.Permanent(fmt.Errorf(gmess.ERR_BATCHMEMBER, err.Error(), memKey, groupKey))
		}
		// Log the retries
		lg.Warnw(err.Error(),
			"retrying", b.GetElapsedTime().String(),
			"group", groupKey,
			"member", memKey)
		return fmt.Errorf(gmess.ERR_BATCHMEMBER, err.Error(), memKey, groupKey)
	}, b)
	if err != nil {
		// Log final error
		lg.Error(err)
		fmt.Println(cmn.GminMessage(err.Error()))
	}
}

func init() {
	batchUpdateCmd.AddCommand(batchUpdMemberCmd)

	batchUpdMemberCmd.Flags().StringP(flgnm.FLG_INPUTFILE, "i", "", "filepath to group member data file or sheet id")
	batchUpdMemberCmd.Flags().StringP(flgnm.FLG_FORMAT, "f", "json", "user data file format")
	batchUpdMemberCmd.Flags().StringP(flgnm.FLG_SHEETRANGE, "s", "", "user data gsheet range")
}
