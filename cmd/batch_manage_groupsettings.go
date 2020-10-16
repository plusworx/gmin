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

	valid "github.com/asaskevich/govalidator"
	"github.com/cenkalti/backoff/v4"
	cmn "github.com/plusworx/gmin/utils/common"
	flgnm "github.com/plusworx/gmin/utils/flagnames"
	gmess "github.com/plusworx/gmin/utils/gminmessages"
	grpset "github.com/plusworx/gmin/utils/groupsettings"
	lg "github.com/plusworx/gmin/utils/logging"
	"github.com/spf13/cobra"
	gset "google.golang.org/api/groupssettings/v1"
	sheet "google.golang.org/api/sheets/v4"
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

	var (
		groupKeys   []string
		grpSettings []*gset.Groups
	)

	ds, err := cmn.CreateGroupSettingService(gset.AppsGroupsSettingsScope)
	if err != nil {
		lg.Error(err)
		return err
	}

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

	switch {
	case lwrFmt == "csv":
		groupKeys, grpSettings, err = bmnggProcessCSVFile(ds, inputFlgVal)
		if err != nil {
			lg.Error(err)
			return err
		}
	case lwrFmt == "json":
		groupKeys, grpSettings, err = bmnggProcessJSON(ds, inputFlgVal, scanner)
		if err != nil {
			lg.Error(err)
			return err
		}
	case lwrFmt == "gsheet":
		rangeFlgVal, err := cmd.Flags().GetString(flgnm.FLG_SHEETRANGE)
		if err != nil {
			lg.Error(err)
			return err
		}

		groupKeys, grpSettings, err = bmnggProcessGSheet(ds, inputFlgVal, rangeFlgVal)
		if err != nil {
			lg.Error(err)
			return err
		}
	default:
		return fmt.Errorf(gmess.ERR_INVALIDFILEFORMAT, formatFlgVal)
	}

	err = bmnggProcessObjects(ds, groupKeys, grpSettings)
	if err != nil {
		lg.Error(err)
		return err
	}

	lg.Debug("finished doBatchMngGrpSettings()")
	return nil
}

func bmnggApproveMemberVal(grpSetting *gset.Groups, attrName string, attrValue string) error {
	lg.Debug("starting bmnggApproveMemberVal()")
	validTxt, err := grpset.ValidateGroupSettingValue(grpset.ApproveMemberMap, attrName, attrValue)
	if err != nil {
		return err
	}
	grpSetting.WhoCanApproveMembers = validTxt
	lg.Debug("finished bmnggApproveMemberVal()")
	return nil
}

func bmnggAssistContentVal(grpSetting *gset.Groups, attrName string, attrValue string) error {
	lg.Debug("starting bmnggAssistContentVal()")
	validTxt, err := grpset.ValidateGroupSettingValue(grpset.AssistContentMap, attrName, attrValue)
	if err != nil {
		return err
	}
	grpSetting.WhoCanAssistContent = validTxt
	lg.Debug("finished bmnggAssistContentVal()")
	return nil
}

func bmnggBanUserVal(grpSetting *gset.Groups, attrName string, attrValue string) error {
	lg.Debug("starting bmnggBanUserVal()")
	validTxt, err := grpset.ValidateGroupSettingValue(grpset.BanUserMap, attrName, attrValue)
	if err != nil {
		return err
	}
	grpSetting.WhoCanBanUsers = validTxt
	lg.Debug("finished bmnggBanUserVal()")
	return nil
}

func bmnggContactOwnerVal(grpSetting *gset.Groups, attrName string, attrValue string) error {
	lg.Debug("starting bmnggContactOwnerVal()")
	validTxt, err := grpset.ValidateGroupSettingValue(grpset.ContactOwnerMap, attrName, attrValue)
	if err != nil {
		return err
	}
	grpSetting.WhoCanContactOwner = validTxt
	lg.Debug("finished bmnggContactOwnerVal()")
	return nil
}

func bmnggDiscoverGroupVal(grpSettings *gset.Groups, attrName string, attrValue string) error {
	lg.Debug("starting bmnggDiscoverGroupVal()")
	validTxt, err := grpset.ValidateGroupSettingValue(grpset.DiscoverGroupMap, attrName, discoverGroup)
	if err != nil {
		return err
	}
	grpSettings.WhoCanDiscoverGroup = validTxt
	lg.Debug("finished bmnggDiscoverGroupVal()")
	return nil
}

func bmnggFromFileFactory(hdrMap map[int]string, gsData []interface{}) (*gset.Groups, string, error) {
	lg.Debugw("starting bmnggFromFileFactory()",
		"hdrMap", hdrMap)

	var (
		groupKey   string
		grpSetting *gset.Groups
	)

	grpSetting = new(gset.Groups)

	for idx, attr := range gsData {
		attrName := hdrMap[idx]
		attrVal := fmt.Sprintf("%v", attr)
		lowerAttrName := strings.ToLower(hdrMap[idx])
		lowerAttrVal := strings.ToLower(fmt.Sprintf("%v", attr))

		if lowerAttrName == "allowexternalmembers" {
			if lowerAttrVal == "true" {
				grpSetting.AllowExternalMembers = "true"
			} else {
				grpSetting.AllowExternalMembers = "false"
				grpSetting.ForceSendFields = append(grpSetting.ForceSendFields, "AllowExternalMembers")
			}
		}
		if lowerAttrName == "allowwebposting" {
			if lowerAttrVal == "true" {
				grpSetting.AllowWebPosting = "true"
			} else {
				grpSetting.AllowWebPosting = "false"
				grpSetting.ForceSendFields = append(grpSetting.ForceSendFields, "AllowWebPosting")
			}
		}
		if lowerAttrName == "archiveonly" {
			if lowerAttrVal == "true" {
				grpSetting.ArchiveOnly = "true"
			} else {
				grpSetting.ArchiveOnly = "false"
				grpSetting.ForceSendFields = append(grpSetting.ForceSendFields, "ArchiveOnly")
			}
		}
		if lowerAttrName == "customfootertext" {
			if attrVal == "" {
				grpSetting.ForceSendFields = append(grpSetting.ForceSendFields, "CustomFooterText")
			}
			grpSetting.CustomFooterText = footerText
		}
		if lowerAttrName == "customreplyto" {
			err := bmnggReplyEmailVal(grpSetting, attrVal)
			if err != nil {
				return nil, "", err
			}
		}
		if lowerAttrName == "defaultmessagedenynotificationtext" {
			if attrVal == "" {
				grpSetting.ForceSendFields = append(grpSetting.ForceSendFields, "DefaultMessageDenyNotificationText")
			}
			grpSetting.DefaultMessageDenyNotificationText = attrVal
		}
		if lowerAttrName == "enablecollaborativeinbox" {
			if lowerAttrVal == "true" {
				grpSetting.EnableCollaborativeInbox = "true"
			} else {
				grpSetting.EnableCollaborativeInbox = "false"
				grpSetting.ForceSendFields = append(grpSetting.ForceSendFields, "EnableCollaborativeInbox")
			}
		}
		if lowerAttrName == "favoriterepliesontop" {
			if lowerAttrVal == "true" {
				grpSetting.FavoriteRepliesOnTop = "true"
			} else {
				grpSetting.FavoriteRepliesOnTop = "false"
				grpSetting.ForceSendFields = append(grpSetting.ForceSendFields, "FavoriteRepliesOnTop")
			}
		}
		if lowerAttrName == "groupkey" {
			if attrVal == "" {
				err := fmt.Errorf(gmess.ERR_EMPTYSTRING, attrName)
				return nil, "", err
			}
			groupKey = attrVal
		}
		if lowerAttrName == "includecustomfooter" {
			if lowerAttrVal == "true" {
				grpSetting.IncludeCustomFooter = "true"
			} else {
				grpSetting.IncludeCustomFooter = "false"
				grpSetting.ForceSendFields = append(grpSetting.ForceSendFields, "IncludeCustomFooter")
			}
		}
		if lowerAttrName == "isarchived" {
			if lowerAttrVal == "true" {
				grpSetting.IsArchived = "true"
			} else {
				grpSetting.IsArchived = "false"
				grpSetting.ForceSendFields = append(grpSetting.ForceSendFields, "IsArchived")
			}
		}
		if lowerAttrName == "memberscanpostasthegroup" {
			if attrVal == "" {
				grpSetting.MembersCanPostAsTheGroup = "true"
			} else {
				grpSetting.MembersCanPostAsTheGroup = "false"
				grpSetting.ForceSendFields = append(grpSetting.ForceSendFields, "MembersCanPostAsTheGroup")
			}
		}
		if lowerAttrName == "messagemoderationlevel" {
			err := bmnggMessageModVal(grpSetting, attrName, attrVal)
			if err != nil {
				return nil, "", err
			}
		}
		if lowerAttrName == "primarylanguage" {
			err := bmnggLanguageVal(grpSetting, attrName, attrVal)
			if err != nil {
				return nil, "", err
			}
		}
		if lowerAttrName == "replyto" {
			err := bmnggReplyToVal(grpSetting, attrName, attrVal)
			if err != nil {
				return nil, "", err
			}
		}
		if lowerAttrName == "sendmessagedenynotification" {
			if attrVal == "" {
				grpSetting.SendMessageDenyNotification = "true"
			} else {
				grpSetting.SendMessageDenyNotification = "false"
				grpSetting.ForceSendFields = append(grpSetting.ForceSendFields, "SendMessageDenyNotification")
			}
		}
		if lowerAttrName == "spammoderationlevel" {
			err := bmnggSpamModVal(grpSetting, attrName, attrVal)
			if err != nil {
				return nil, "", err
			}
		}
		if lowerAttrName == "whocanapprovemembers" {
			err := bmnggApproveMemberVal(grpSetting, attrName, attrVal)
			if err != nil {
				return nil, "", err
			}
		}
		if lowerAttrName == "whocanassistcontent" {
			err := bmnggAssistContentVal(grpSetting, attrName, attrVal)
			if err != nil {
				return nil, "", err
			}
		}
		if lowerAttrName == "whocanbanusers" {
			err := bmnggBanUserVal(grpSetting, attrName, attrVal)
			if err != nil {
				return nil, "", err
			}
		}
		if lowerAttrName == "whocancontactowner" {
			err := bmnggContactOwnerVal(grpSetting, attrName, attrVal)
			if err != nil {
				return nil, "", err
			}
		}
		if lowerAttrName == "whocandiscovergroup" {
			err := bmnggDiscoverGroupVal(grpSetting, attrName, attrVal)
			if err != nil {
				return nil, "", err
			}
		}
		if lowerAttrName == "whocanjoin" {
			err := bmnggJoinVal(grpSetting, attrName, attrVal)
			if err != nil {
				return nil, "", err
			}
		}
		if lowerAttrName == "whocanleavegroup" {
			err := bmnggLeaveVal(grpSetting, attrName, attrVal)
			if err != nil {
				return nil, "", err
			}
		}
		if lowerAttrName == "whocanmoderatecontent" {
			err := bmnggModContentVal(grpSetting, attrName, attrVal)
			if err != nil {
				return nil, "", err
			}
		}
		if lowerAttrName == "whocanmoderatemembers" {
			err := bmnggModMemberVal(grpSetting, attrName, attrVal)
			if err != nil {
				return nil, "", err
			}
		}
		if lowerAttrName == "whocanpostmessage" {
			err := bmnggPostMessageVal(grpSetting, attrName, attrVal)
			if err != nil {
				return nil, "", err
			}
		}
		if lowerAttrName == "whocanviewgroup" {
			err := bmnggViewGroupVal(grpSetting, attrName, attrVal)
			if err != nil {
				return nil, "", err
			}
		}
		if lowerAttrName == "whocanviewmembership" {
			err := bmnggViewMembershipVal(grpSetting, attrName, attrVal)
			if err != nil {
				return nil, "", err
			}
		}
	}
	lg.Debug("finished bmnggFromFileFactory()")
	return grpSetting, groupKey, nil
}

func bmnggFromJSONFactory(ds *gset.Service, jsonData string) (*gset.Groups, string, error) {
	lg.Debugw("starting bmnggFromJSONFactory()",
		"jsonData", jsonData)

	var (
		grpSettings *gset.Groups
		grpKey      = grpset.Key{}
	)

	grpSettings = new(gset.Groups)
	jsonBytes := []byte(jsonData)

	if !json.Valid(jsonBytes) {
		lg.Error(gmess.ERR_INVALIDJSONATTR)
		return nil, "", errors.New(gmess.ERR_INVALIDJSONATTR)
	}

	outStr, err := cmn.ParseInputAttrs(jsonBytes)
	if err != nil {
		lg.Error(err)
		return nil, "", err
	}

	err = cmn.ValidateInputAttrs(outStr, grpset.GroupSettingsAttrMap)
	if err != nil {
		lg.Error(err)
		return nil, "", err
	}

	err = json.Unmarshal(jsonBytes, &grpKey)
	if err != nil {
		lg.Error(err)
		return nil, "", err
	}

	if grpKey.GroupKey == "" {
		err = errors.New(gmess.ERR_NOJSONGROUPKEY)
		lg.Error(err)
		return nil, "", err
	}

	err = json.Unmarshal(jsonBytes, &grpSettings)
	if err != nil {
		lg.Error(err)
		return nil, "", err
	}
	lg.Debug("finished bmnggFromJSONFactory()")
	return grpSettings, grpKey.GroupKey, nil
}

func bmnggJoinVal(grpSetting *gset.Groups, attrName string, attrValue string) error {
	lg.Debug("starting bmnggJoinVal()")
	validTxt, err := grpset.ValidateGroupSettingValue(grpset.JoinMap, attrName, attrValue)
	if err != nil {
		return err
	}
	grpSetting.WhoCanJoin = validTxt
	lg.Debug("finished bmnggJoinVal()")
	return nil
}

func bmnggLanguageVal(grpSetting *gset.Groups, attrName string, attrValue string) error {
	lg.Debug("starting bmnggLanguageVal()")
	validTxt, err := grpset.ValidateGroupSettingValue(grpset.LanguageMap, attrName, attrValue)
	if err != nil {
		return err
	}
	grpSetting.PrimaryLanguage = validTxt
	lg.Debug("finished bmnggLanguageVal()")
	return nil
}

func bmnggLeaveVal(grpSetting *gset.Groups, attrName string, attrValue string) error {
	lg.Debug("starting bmnggLeaveVal()")
	validTxt, err := grpset.ValidateGroupSettingValue(grpset.LeaveMap, attrName, attrValue)
	if err != nil {
		return err
	}
	grpSetting.WhoCanLeaveGroup = validTxt
	lg.Debug("finished bmnggLeaveVal()")
	return nil
}

func bmnggMessageModVal(grpSetting *gset.Groups, attrName string, attrValue string) error {
	lg.Debug("starting bmnggMessageModVal()")
	validTxt, err := grpset.ValidateGroupSettingValue(grpset.MessageModMap, attrName, attrValue)
	if err != nil {
		return err
	}
	grpSetting.MessageModerationLevel = validTxt
	lg.Debug("finished bmnggMessageModVal()")
	return nil
}

func bmnggModContentVal(grpSetting *gset.Groups, attrName string, attrValue string) error {
	lg.Debug("starting bmnggModContentVal()")
	validTxt, err := grpset.ValidateGroupSettingValue(grpset.ModContentMap, attrName, attrValue)
	if err != nil {
		return err
	}
	grpSetting.WhoCanModerateContent = validTxt
	lg.Debug("finished bmnggModContentVal()")
	return nil
}

func bmnggModMemberVal(grpSetting *gset.Groups, attrName string, attrValue string) error {
	lg.Debug("starting bmnggModMemberVal()")
	validTxt, err := grpset.ValidateGroupSettingValue(grpset.ModMemberMap, attrName, attrValue)
	if err != nil {
		return err
	}
	grpSetting.WhoCanModerateMembers = validTxt
	lg.Debug("finished bmnggModMemberVal()")
	return nil
}

func bmnggPerformUpdate(grpSetting *gset.Groups, groupKey string, wg *sync.WaitGroup, gsuc *gset.GroupsUpdateCall) {
	lg.Debugw("starting bmnggPerformUpdate()",
		"groupKey", groupKey)

	defer wg.Done()

	b := backoff.NewExponentialBackOff()
	b.MaxElapsedTime = 32 * time.Second

	err := backoff.Retry(func() error {
		var err error
		_, err = gsuc.Do()
		if err == nil {
			lg.Infof(gmess.INFO_GROUPSETTINGSCHANGED, groupKey)
			fmt.Println(cmn.GminMessage(fmt.Sprintf(gmess.INFO_GROUPSETTINGSCHANGED, groupKey)))
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
	lg.Debug("finished bmnggPerformUpdate()")
}

func bmnggPostMessageVal(grpSettings *gset.Groups, attrName string, attrValue string) error {
	lg.Debug("starting bmnggPostMessageVal()")
	validTxt, err := grpset.ValidateGroupSettingValue(grpset.PostMessageMap, attrName, attrValue)
	if err != nil {
		return err
	}
	grpSettings.WhoCanPostMessage = validTxt
	lg.Debug("finished bmnggPostMessageVal()")
	return nil
}

func bmnggProcessCSVFile(ds *gset.Service, filePath string) ([]string, []*gset.Groups, error) {
	lg.Debugw("starting bmnggProcessCSVFile()",
		"filePath", filePath)

	var (
		iSlice      []interface{}
		hdrMap      = map[int]string{}
		groupKeys   []string
		grpSettings []*gset.Groups
	)

	csvfile, err := os.Open(filePath)
	if err != nil {
		lg.Error(err)
		return nil, nil, err
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
			lg.Error(err)
			return nil, nil, err
		}

		if count == 0 {
			iSlice = make([]interface{}, len(record))
			for idx, value := range record {
				iSlice[idx] = value
			}
			hdrMap = cmn.ProcessHeader(iSlice)
			err = cmn.ValidateHeader(hdrMap, grpset.GroupSettingsAttrMap)
			if err != nil {
				lg.Error(err)
				return nil, nil, err
			}
			count = count + 1
			continue
		}

		for idx, value := range record {
			iSlice[idx] = value
		}

		gsVar, groupKey, err := bmnggFromFileFactory(hdrMap, iSlice)
		if err != nil {
			lg.Error(err)
			return nil, nil, err
		}

		groupKeys = append(groupKeys, groupKey)
		grpSettings = append(grpSettings, gsVar)
		count = count + 1
	}

	lg.Debug("finished bmnggProcessCSVFile()")
	return groupKeys, grpSettings, nil
}

func bmnggProcessGSheet(ds *gset.Service, sheetID string, sheetrange string) ([]string, []*gset.Groups, error) {
	lg.Debugw("starting bmnggProcessGSheet()",
		"sheetID", sheetID,
		"sheetrange", sheetrange)

	var (
		err         error
		groupKeys   []string
		grpSettings []*gset.Groups
	)

	if sheetrange == "" {
		err := errors.New(gmess.ERR_NOSHEETRANGE)
		lg.Error(err)
		return nil, nil, err
	}

	ss, err := cmn.CreateSheetService(sheet.DriveReadonlyScope)
	if err != nil {
		lg.Error(err)
		return nil, nil, err
	}

	ssvgc := ss.Spreadsheets.Values.Get(sheetID, sheetrange)
	sValRange, err := ssvgc.Do()
	if err != nil {
		lg.Error(err)
		return nil, nil, err
	}

	if len(sValRange.Values) == 0 {
		err = fmt.Errorf(gmess.ERR_NOSHEETDATAFOUND, sheetID, sheetrange)
		lg.Error(err)
		return nil, nil, err
	}

	hdrMap := cmn.ProcessHeader(sValRange.Values[0])
	err = cmn.ValidateHeader(hdrMap, grpset.GroupSettingsAttrMap)
	if err != nil {
		lg.Error(err)
		return nil, nil, err
	}

	for idx, row := range sValRange.Values {
		if idx == 0 {
			continue
		}

		gsVar, groupKey, err := bmnggFromFileFactory(hdrMap, row)
		if err != nil {
			lg.Error(err)
			return nil, nil, err
		}

		groupKeys = append(groupKeys, groupKey)
		grpSettings = append(grpSettings, gsVar)
	}

	lg.Debug("finished bmnggProcessGSheet()")
	return groupKeys, grpSettings, nil
}

func bmnggProcessJSON(ds *gset.Service, filePath string, scanner *bufio.Scanner) ([]string, []*gset.Groups, error) {
	lg.Debugw("starting bmnggProcessJSON()",
		"filePath", filePath)

	var (
		err         error
		groupKeys   []string
		grpSettings []*gset.Groups
	)

	if filePath != "" {
		file, err := os.Open(filePath)
		if err != nil {
			lg.Error(err)
			return nil, nil, err
		}
		defer file.Close()

		scanner = bufio.NewScanner(file)
	}

	for scanner.Scan() {
		jsonData := scanner.Text()

		gsVar, groupKey, err := bmnggFromJSONFactory(ds, jsonData)
		if err != nil {
			lg.Error(err)
			return nil, nil, err
		}

		groupKeys = append(groupKeys, groupKey)
		grpSettings = append(grpSettings, gsVar)
	}
	err = scanner.Err()
	if err != nil {
		lg.Error(err)
		return nil, nil, err
	}

	lg.Debug("finished bmnggProcessJSON()")
	return groupKeys, grpSettings, nil
}

func bmnggProcessObjects(ds *gset.Service, groupKeys []string, grpSettings []*gset.Groups) error {
	lg.Debugw("starting bmnggProcessObjects()",
		"groupKeys", groupKeys)

	wg := new(sync.WaitGroup)

	for idx, gs := range grpSettings {
		gsuc := ds.Groups.Update(groupKeys[idx], gs)

		wg.Add(1)

		go bmnggPerformUpdate(gs, groupKeys[idx], wg, gsuc)
	}

	wg.Wait()

	lg.Debug("finished bmnggProcessObjects()")
	return nil
}

func bmnggReplyEmailVal(grpSettings *gset.Groups, attrValue string) error {
	lg.Debug("starting bmnggReplyEmailVal()")
	if attrValue == "" {
		grpSettings.CustomReplyTo = attrValue
		grpSettings.ForceSendFields = append(grpSettings.ForceSendFields, "CustomReplyTo")
		return nil
	}
	ok := valid.IsEmail(attrValue)
	if !ok {
		err := fmt.Errorf(gmess.ERR_INVALIDEMAILADDRESS, attrValue)
		return err
	}
	grpSettings.CustomReplyTo = attrValue
	lg.Debug("finished bmnggReplyEmailVal()")
	return nil
}

func bmnggReplyToVal(grpSetting *gset.Groups, attrName string, attrValue string) error {
	lg.Debug("starting bmnggReplyToVal()")
	validTxt, err := grpset.ValidateGroupSettingValue(grpset.ReplyToMap, attrName, attrValue)
	if err != nil {
		return err
	}
	grpSetting.ReplyTo = validTxt
	lg.Debug("finished bmnggReplyToVal()")
	return nil
}

func bmnggSpamModVal(grpSetting *gset.Groups, attrName string, attrValue string) error {
	lg.Debug("starting bmnggSpamModVal()")
	validTxt, err := grpset.ValidateGroupSettingValue(grpset.SpamModMap, attrName, attrValue)
	if err != nil {
		return err
	}
	grpSetting.SpamModerationLevel = validTxt
	lg.Debug("finished bmnggSpamModVal()")
	return nil
}

func bmnggViewGroupVal(grpSetting *gset.Groups, attrName string, attrValue string) error {
	lg.Debug("starting bmnggViewGroupVal()")
	validTxt, err := grpset.ValidateGroupSettingValue(grpset.ViewGroupMap, attrName, attrValue)
	if err != nil {
		return err
	}
	grpSetting.WhoCanViewGroup = validTxt
	lg.Debug("finished bmnggViewGroupVal()")
	return nil
}

func bmnggViewMembershipVal(grpSetting *gset.Groups, attrName string, attrValue string) error {
	lg.Debug("starting bmnggViewMembershipVal()")
	validTxt, err := grpset.ValidateGroupSettingValue(grpset.ViewMembershipMap, attrName, attrValue)
	if err != nil {
		return err
	}
	grpSetting.WhoCanViewMembership = validTxt
	lg.Debug("finished bmnggViewMembershipVal()")
	return nil
}

func init() {
	batchManageCmd.AddCommand(batchMngGrpSettingsCmd)

	batchMngGrpSettingsCmd.Flags().StringVarP(&inputFile, flgnm.FLG_INPUTFILE, "i", "", "filepath to device data file")
	batchMngGrpSettingsCmd.Flags().StringVarP(&format, flgnm.FLG_FORMAT, "f", "json", "user data file format")
	batchMngGrpSettingsCmd.Flags().StringVarP(&sheetRange, flgnm.FLG_SHEETRANGE, "s", "", "user data gsheet range")
}
