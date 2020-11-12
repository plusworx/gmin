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
	grpset "github.com/plusworx/gmin/utils/groupsettings"
	lg "github.com/plusworx/gmin/utils/logging"
	"github.com/spf13/cobra"
	gset "google.golang.org/api/groupssettings/v1"
)

var batchMngGrpSettingsCmd = &cobra.Command{
	Use:     "group-settings -i <input file path or google sheet id>",
	Aliases: []string{"grp-settings", "grp-set", "gsettings", "gset"},
	Example: `gmin batch-manage group-settings -i inputfile.json
gmin bmng gsettings -i inputfile.csv -f csv
gmin bmng gset -i 1odyAIp3jGspd3M4xeepxWD6aeQIUuHBgrZB2OHSu8MI -s 'Sheet1!A1:K25' -f gsheet`,
	Short: "Manages a batch of group settings",
	Long: `Manages a batch of group settings where setting details are provided in a Google Sheet, CSV/JSON input file or piped JSON.
				  
The contents of the JSON file or piped input should look something like this:
	
{"groupKey":"finance@mycompany.com","allowExternalMembers":"true","allowWebPosting":"true","enableCollaborativeInbox":"true"}
{"groupKey":"marketing@mycompany.com","membersCanPostAsTheGroup":"true","whoCanJoin":"INVITED_CAN_JOIN"}
{"groupKey":"sales@mycompany.com","messageModerationLevel":"MODERATE_NEW_MEMBERS"}
	
CSV and Google sheets must have a header row with the following column names being the only ones that are valid:
	
allowExternalMembers
allowWebPosting
archiveOnly
customFooterText
customReplyTo
defaultMessageDenyNotificationText
enableCollaborativeInbox
favoriteRepliesOnTop
groupKey [required]
includeCustomFooter
includeInGlobalAddressList
isArchived
membersCanPostAsTheGroup
messageModerationLevel
primaryLanguage
replyTo
sendMessageDenyNotification
spamModerationLevel
whoCanApproveMembers
whoCanAssistContent
whoCanBanUsers
whoCanContactOwner
whoCanDiscoverGroup
whoCanJoin
whoCanLeaveGroup
whoCanModerateContent
whoCanModerateMembers
whoCanPostMessage
whoCanViewGroup
whoCanViewMembership

The column names are case insensitive and can be in any order.`,
	RunE: doBatchMngGrpSettings,
}

func doBatchMngGrpSettings(cmd *cobra.Command, args []string) error {
	lg.Debugw("starting doBatchMngGrpSettings()",
		"args", args)
	defer lg.Debug("finished doBatchMngGrpSettings()")

	var grpParams []grpset.GroupParams

	srv, err := cmn.CreateService(cmn.SRVTYPEGRPSETTING, gset.AppsGroupsSettingsScope)
	if err != nil {
		return err
	}
	gs := srv.(*gset.Service)

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

	callParams := btch.CallParams{CallType: cmn.CALLTYPEMANAGE, ObjectType: cmn.OBJTYPEGRPSET}
	inputParams := btch.ProcessInputParams{
		Format:      lwrFmt,
		InputFlgVal: inputFlgVal,
		Scanner:     scanner,
		SheetRange:  rangeFlgVal,
	}

	objs, err := bmnggProcessInput(callParams, inputParams)
	if err != nil {
		return err
	}

	for _, gpObj := range objs {
		grpParams = append(grpParams, gpObj.(grpset.GroupParams))
	}

	err = bmnggProcessObjects(gs, grpParams)
	if err != nil {
		return err
	}

	return nil
}

func bmnggPerformUpdate(grpSetting *gset.Groups, groupKey string, wg *sync.WaitGroup, gsuc *gset.GroupsUpdateCall) {
	lg.Debugw("starting bmnggPerformUpdate()",
		"groupKey", groupKey)
	defer lg.Debug("finished bmnggPerformUpdate()")

	defer wg.Done()

	b := backoff.NewExponentialBackOff()
	b.MaxElapsedTime = 32 * time.Second

	err := backoff.Retry(func() error {
		var err error
		_, err = gsuc.Do()
		if err == nil {
			fmt.Println(cmn.GminMessage(fmt.Sprintf(gmess.INFO_GROUPSETTINGSCHANGED, groupKey)))
			lg.Infof(gmess.INFO_GROUPSETTINGSCHANGED, groupKey)
			return err
		}
		if !cmn.IsErrRetryable(err) {
			return backoff.Permanent(fmt.Errorf(gmess.ERR_BATCHGROUPSETTINGS, err.Error(), groupKey))
		}
		// Log the retries
		lg.Warnw(err.Error(),
			"retrying", b.GetElapsedTime().String(),
			"group", groupKey)
		return fmt.Errorf(gmess.ERR_BATCHGROUPSETTINGS, err.Error(), groupKey)
	}, b)
	if err != nil {
		// Log final error
		lg.Error(err)
		fmt.Println(cmn.GminMessage(err.Error()))
	}
}

func bmnggProcessInput(callParams btch.CallParams, inputParams btch.ProcessInputParams) ([]interface{}, error) {
	lg.Debug("starting bmnggProcessInput()")
	defer lg.Debug("finished bmnggProcessInput()")

	var (
		err  error
		objs []interface{}
	)

	switch inputParams.Format {
	case "csv":
		objs, err = btch.ProcessCSVFile(callParams, inputParams.InputFlgVal, grpset.GroupSettingsAttrMap)
		if err != nil {
			return nil, err
		}
	case "json":
		objs, err = btch.ProcessJSON(callParams, inputParams.InputFlgVal, inputParams.Scanner, grpset.GroupSettingsAttrMap)
		if err != nil {
			return nil, err
		}
	case "gsheet":
		objs, err = btch.ProcessGSheet(callParams, inputParams.InputFlgVal, inputParams.SheetRange, grpset.GroupSettingsAttrMap)
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

func bmnggProcessObjects(gss *gset.Service, grpParams []grpset.GroupParams) error {
	lg.Debugw("starting bmnggProcessObjects()",
		"grpParams", grpParams)
	defer lg.Debug("finished bmnggProcessObjects()")

	wg := new(sync.WaitGroup)

	for _, gp := range grpParams {
		gsuc := gss.Groups.Update(gp.GroupKey, gp.Settings)

		wg.Add(1)

		go bmnggPerformUpdate(gp.Settings, gp.GroupKey, wg, gsuc)
	}

	wg.Wait()

	return nil
}

func init() {
	batchManageCmd.AddCommand(batchMngGrpSettingsCmd)

	batchMngGrpSettingsCmd.Flags().StringP(flgnm.FLG_INPUTFILE, "i", "", "filepath to device data file")
	batchMngGrpSettingsCmd.Flags().StringP(flgnm.FLG_FORMAT, "f", "json", "user data file format")
	batchMngGrpSettingsCmd.Flags().StringP(flgnm.FLG_SHEETRANGE, "s", "", "user data gsheet range")
}
