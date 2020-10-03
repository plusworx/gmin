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
	grpset "github.com/plusworx/gmin/utils/groupsettings"
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
	logger.Debugw("starting doBatchMngGrpSettings()",
		"args", args)
	ds, err := cmn.CreateGroupSettingService(gset.AppsGroupsSettingsScope)
	if err != nil {
		logger.Error(err)
		return err
	}

	scanner, err := cmn.InputFromStdIn(inputFile)
	if err != nil {
		logger.Error(err)
		return err
	}

	if inputFile == "" && scanner == nil {
		err := errors.New(cmn.ErrNoInputFile)
		logger.Error(err)
		return err
	}

	lwrFmt := strings.ToLower(format)

	ok := cmn.SliceContainsStr(cmn.ValidFileFormats, lwrFmt)
	if !ok {
		err = fmt.Errorf(cmn.ErrInvalidFileFormat, format)
		logger.Error(err)
		return err
	}

	switch {
	case lwrFmt == "csv":
		err := btchMngGrpSettingsProcessCSV(ds, inputFile)
		if err != nil {
			logger.Error(err)
			return err
		}
	case lwrFmt == "json":
		err := btchMngGrpSettingsProcessJSON(ds, inputFile, scanner)
		if err != nil {
			logger.Error(err)
			return err
		}
	case lwrFmt == "gsheet":
		err := btchMngGrpSettingsProcessSheet(ds, inputFile)
		if err != nil {
			logger.Error(err)
			return err
		}
	}
	logger.Debug("finished doBatchMngGrpSettings()")
	return nil
}

func btchMngJSONGrpSettings(ds *gset.Service, jsonData string) (*gset.Groups, string, error) {
	logger.Debugw("starting btchMngJSONGrpSettings()",
		"jsonData", jsonData)

	var (
		grpSettings *gset.Groups
		grpKey      = grpset.Key{}
	)

	grpSettings = new(gset.Groups)
	jsonBytes := []byte(jsonData)

	if !json.Valid(jsonBytes) {
		logger.Error(cmn.ErrInvalidJSONAttr)
		return nil, "", errors.New(cmn.ErrInvalidJSONAttr)
	}

	outStr, err := cmn.ParseInputAttrs(jsonBytes)
	if err != nil {
		logger.Error(err)
		return nil, "", err
	}

	err = cmn.ValidateInputAttrs(outStr, grpset.GroupSettingsAttrMap)
	if err != nil {
		logger.Error(err)
		return nil, "", err
	}

	err = json.Unmarshal(jsonBytes, &grpKey)
	if err != nil {
		logger.Error(err)
		return nil, "", err
	}

	if grpKey.GroupKey == "" {
		err = errors.New(cmn.ErrNoJSONGroupKey)
		logger.Error(err)
		return nil, "", err
	}

	err = json.Unmarshal(jsonBytes, &grpSettings)
	if err != nil {
		logger.Error(err)
		return nil, "", err
	}
	logger.Debug("finished btchMngJSONGrpSettings()")
	return grpSettings, grpKey.GroupKey, nil
}

func btchManageGrpSettings(ds *gset.Service, groupKeys []string, grpSettings []*gset.Groups) error {
	logger.Debugw("starting btchManageGrpSettings()",
		"groupKeys", groupKeys)

	wg := new(sync.WaitGroup)

	for idx, gs := range grpSettings {
		gsuc := ds.Groups.Update(groupKeys[idx], gs)

		wg.Add(1)

		go btchGrpSettingsMngProcess(gs, groupKeys[idx], wg, gsuc)
	}

	wg.Wait()

	logger.Debug("finished btchManageGrpSettings()")
	return nil
}

func btchGrpSettingsMngProcess(grpSetting *gset.Groups, groupKey string, wg *sync.WaitGroup, gsuc *gset.GroupsUpdateCall) {
	logger.Debugw("starting btchGrpSettingsMngProcess()",
		"groupKey", groupKey)

	defer wg.Done()

	b := backoff.NewExponentialBackOff()
	b.MaxElapsedTime = 32 * time.Second

	err := backoff.Retry(func() error {
		var err error
		_, err = gsuc.Do()
		if err == nil {
			logger.Infof(cmn.InfoGroupSettingsChanged, groupKey)
			fmt.Println(cmn.GminMessage(fmt.Sprintf(cmn.InfoGroupSettingsChanged, groupKey)))
			return err
		}
		if !cmn.IsErrRetryable(err) {
			return backoff.Permanent(fmt.Errorf(cmn.ErrBatchGroupSettings, err.Error(), groupKey))
		}
		// Log the retries
		logger.Warnw(err.Error(),
			"retrying", b.GetElapsedTime().String(),
			"group", groupKey)
		return fmt.Errorf(cmn.ErrBatchGroupSettings, err.Error(), groupKey)
	}, b)
	if err != nil {
		// Log final error
		logger.Error(err)
		fmt.Println(cmn.GminMessage(err.Error()))
	}
	logger.Debug("finished btchGrpSettingsMngProcess()")
}

func btchMngGrpSettingsProcessCSV(ds *gset.Service, filePath string) error {
	logger.Debugw("starting btchMngGrpSettingsProcessCSV()",
		"filePath", filePath)

	var (
		iSlice      []interface{}
		hdrMap      = map[int]string{}
		groupKeys   []string
		grpSettings []*gset.Groups
	)

	csvfile, err := os.Open(filePath)
	if err != nil {
		logger.Error(err)
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
			logger.Error(err)
			return err
		}

		if count == 0 {
			iSlice = make([]interface{}, len(record))
			for idx, value := range record {
				iSlice[idx] = value
			}
			hdrMap = cmn.ProcessHeader(iSlice)
			err = cmn.ValidateHeader(hdrMap, grpset.GroupSettingsAttrMap)
			if err != nil {
				logger.Error(err)
				return err
			}
			count = count + 1
			continue
		}

		for idx, value := range record {
			iSlice[idx] = value
		}

		gsVar, groupKey, err := btchMngProcessGrpSettings(hdrMap, iSlice)
		if err != nil {
			logger.Error(err)
			return err
		}

		groupKeys = append(groupKeys, groupKey)
		grpSettings = append(grpSettings, gsVar)
		count = count + 1
	}

	err = btchManageGrpSettings(ds, groupKeys, grpSettings)
	if err != nil {
		logger.Error(err)
		return err
	}
	logger.Debug("finished btchMngGrpSettingsProcessCSV()")
	return nil
}

func btchMngGrpSettingsProcessJSON(ds *gset.Service, filePath string, scanner *bufio.Scanner) error {
	logger.Debugw("starting btchUpdMemProcessJSON()",
		"filePath", filePath)

	var (
		err         error
		groupKeys   []string
		grpSettings []*gset.Groups
	)

	if filePath != "" {
		file, err := os.Open(filePath)
		if err != nil {
			logger.Error(err)
			return err
		}
		defer file.Close()

		scanner = bufio.NewScanner(file)
	}

	for scanner.Scan() {
		jsonData := scanner.Text()

		gsVar, groupKey, err := btchMngJSONGrpSettings(ds, jsonData)
		if err != nil {
			logger.Error(err)
			return err
		}

		groupKeys = append(groupKeys, groupKey)
		grpSettings = append(grpSettings, gsVar)
	}
	err = scanner.Err()
	if err != nil {
		logger.Error(err)
		return err
	}

	err = btchManageGrpSettings(ds, groupKeys, grpSettings)
	if err != nil {
		logger.Error(err)
		return err
	}
	logger.Debug("finished btchUpdMemProcessJSON()")
	return nil
}

func btchMngGrpSettingsProcessSheet(ds *gset.Service, sheetID string) error {
	logger.Debugw("starting btchMngGrpSettingsProcessSheet()",
		"sheetID", sheetID)

	var (
		err         error
		groupKeys   []string
		grpSettings []*gset.Groups
	)

	if sheetRange == "" {
		err := errors.New(cmn.ErrNoSheetRange)
		logger.Error(err)
		return err
	}

	ss, err := cmn.CreateSheetService(sheet.DriveReadonlyScope)
	if err != nil {
		logger.Error(err)
		return err
	}

	ssvgc := ss.Spreadsheets.Values.Get(sheetID, sheetRange)
	sValRange, err := ssvgc.Do()
	if err != nil {
		logger.Error(err)
		return err
	}

	if len(sValRange.Values) == 0 {
		err = fmt.Errorf(cmn.ErrNoSheetDataFound, sheetID, sheetRange)
		logger.Error(err)
		return err
	}

	hdrMap := cmn.ProcessHeader(sValRange.Values[0])
	err = cmn.ValidateHeader(hdrMap, grpset.GroupSettingsAttrMap)
	if err != nil {
		logger.Error(err)
		return err
	}

	for idx, row := range sValRange.Values {
		if idx == 0 {
			continue
		}

		gsVar, groupKey, err := btchMngProcessGrpSettings(hdrMap, row)
		if err != nil {
			logger.Error(err)
			return err
		}

		groupKeys = append(groupKeys, groupKey)
		grpSettings = append(grpSettings, gsVar)
	}

	err = btchManageGrpSettings(ds, groupKeys, grpSettings)
	if err != nil {
		logger.Error(err)
		return err
	}
	logger.Debug("finished btchMngGrpSettingsProcessSheet()")
	return nil
}

func btchMngProcessGrpSettings(hdrMap map[int]string, gsData []interface{}) (*gset.Groups, string, error) {
	logger.Debugw("starting btchMngProcessGrpSettings()",
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
			err := bmgsReplyEmailVal(grpSetting, attrVal)
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
				err := fmt.Errorf(cmn.ErrEmptyString, attrName)
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
			err := bmgsMessageModVal(grpSetting, attrName, attrVal)
			if err != nil {
				return nil, "", err
			}
		}
		if lowerAttrName == "primarylanguage" {
			err := bmgsLanguageVal(grpSetting, attrName, attrVal)
			if err != nil {
				return nil, "", err
			}
		}
		if lowerAttrName == "replyto" {
			err := bmgsReplyToVal(grpSetting, attrName, attrVal)
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
			err := bmgsSpamModVal(grpSetting, attrName, attrVal)
			if err != nil {
				return nil, "", err
			}
		}
		if lowerAttrName == "whocanapprovemembers" {
			err := bmgsApproveMemberVal(grpSetting, attrName, attrVal)
			if err != nil {
				return nil, "", err
			}
		}
		if lowerAttrName == "whocanassistcontent" {
			err := bmgsAssistContentVal(grpSetting, attrName, attrVal)
			if err != nil {
				return nil, "", err
			}
		}
		if lowerAttrName == "whocanbanusers" {
			err := bmgsBanUserVal(grpSetting, attrName, attrVal)
			if err != nil {
				return nil, "", err
			}
		}
		if lowerAttrName == "whocancontactowner" {
			err := bmgsContactOwnerVal(grpSetting, attrName, attrVal)
			if err != nil {
				return nil, "", err
			}
		}
		if lowerAttrName == "whocandiscovergroup" {
			err := bmgsDiscoverGroupVal(grpSetting, attrName, attrVal)
			if err != nil {
				return nil, "", err
			}
		}
		if lowerAttrName == "whocanjoin" {
			err := bmgsJoinVal(grpSetting, attrName, attrVal)
			if err != nil {
				return nil, "", err
			}
		}
		if lowerAttrName == "whocanleavegroup" {
			err := bmgsLeaveVal(grpSetting, attrName, attrVal)
			if err != nil {
				return nil, "", err
			}
		}
		if lowerAttrName == "whocanmoderatecontent" {
			err := bmgsModContentVal(grpSetting, attrName, attrVal)
			if err != nil {
				return nil, "", err
			}
		}
		if lowerAttrName == "whocanmoderatemembers" {
			err := bmgsModMemberVal(grpSetting, attrName, attrVal)
			if err != nil {
				return nil, "", err
			}
		}
		if lowerAttrName == "whocanpostmessage" {
			err := bmgsPostMessageVal(grpSetting, attrName, attrVal)
			if err != nil {
				return nil, "", err
			}
		}
		if lowerAttrName == "whocanviewgroup" {
			err := bmgsViewGroupVal(grpSetting, attrName, attrVal)
			if err != nil {
				return nil, "", err
			}
		}
		if lowerAttrName == "whocanviewmembership" {
			err := bmgsViewMembershipVal(grpSetting, attrName, attrVal)
			if err != nil {
				return nil, "", err
			}
		}
	}
	logger.Debug("finished btchMngProcessGrpSettings()")
	return grpSetting, groupKey, nil
}

func bmgsApproveMemberVal(grpSetting *gset.Groups, attrName string, attrValue string) error {
	logger.Debug("starting bmgsApproveMemberVal()")
	validTxt, err := grpset.ValidateGroupSettingValue(grpset.ApproveMemberMap, attrName, attrValue)
	if err != nil {
		return err
	}
	grpSetting.WhoCanApproveMembers = validTxt
	logger.Debug("finished bmgsApproveMemberVal()")
	return nil
}

func bmgsAssistContentVal(grpSetting *gset.Groups, attrName string, attrValue string) error {
	logger.Debug("starting bmgsAssistContentVal()")
	validTxt, err := grpset.ValidateGroupSettingValue(grpset.AssistContentMap, attrName, attrValue)
	if err != nil {
		return err
	}
	grpSetting.WhoCanAssistContent = validTxt
	logger.Debug("finished bmgsAssistContentVal()")
	return nil
}

func bmgsBanUserVal(grpSetting *gset.Groups, attrName string, attrValue string) error {
	logger.Debug("starting bmgsBanUserVal()")
	validTxt, err := grpset.ValidateGroupSettingValue(grpset.BanUserMap, attrName, attrValue)
	if err != nil {
		return err
	}
	grpSetting.WhoCanBanUsers = validTxt
	logger.Debug("finished bmgsBanUserVal()")
	return nil
}

func bmgsContactOwnerVal(grpSetting *gset.Groups, attrName string, attrValue string) error {
	logger.Debug("starting bmgsContactOwnerVal()")
	validTxt, err := grpset.ValidateGroupSettingValue(grpset.ContactOwnerMap, attrName, attrValue)
	if err != nil {
		return err
	}
	grpSetting.WhoCanContactOwner = validTxt
	logger.Debug("finished bmgsContactOwnerVal()")
	return nil
}

func bmgsDiscoverGroupVal(grpSettings *gset.Groups, attrName string, attrValue string) error {
	logger.Debug("starting bmgsDiscoverGroupVal()")
	validTxt, err := grpset.ValidateGroupSettingValue(grpset.DiscoverGroupMap, attrName, discoverGroup)
	if err != nil {
		return err
	}
	grpSettings.WhoCanDiscoverGroup = validTxt
	logger.Debug("finished bmgsDiscoverGroupVal()")
	return nil
}

func bmgsJoinVal(grpSetting *gset.Groups, attrName string, attrValue string) error {
	logger.Debug("starting bmgsJoinVal()")
	validTxt, err := grpset.ValidateGroupSettingValue(grpset.JoinMap, attrName, attrValue)
	if err != nil {
		return err
	}
	grpSetting.WhoCanJoin = validTxt
	logger.Debug("finished bmgsJoinVal()")
	return nil
}

func bmgsLanguageVal(grpSetting *gset.Groups, attrName string, attrValue string) error {
	logger.Debug("starting bmgsLanguageVal()")
	validTxt, err := grpset.ValidateGroupSettingValue(grpset.LanguageMap, attrName, attrValue)
	if err != nil {
		return err
	}
	grpSetting.PrimaryLanguage = validTxt
	logger.Debug("finished bmgsLanguageVal()")
	return nil
}

func bmgsLeaveVal(grpSetting *gset.Groups, attrName string, attrValue string) error {
	logger.Debug("starting bmgsLeaveVal()")
	validTxt, err := grpset.ValidateGroupSettingValue(grpset.LeaveMap, attrName, attrValue)
	if err != nil {
		return err
	}
	grpSetting.WhoCanLeaveGroup = validTxt
	logger.Debug("finished bmgsLeaveVal()")
	return nil
}

func bmgsMessageModVal(grpSetting *gset.Groups, attrName string, attrValue string) error {
	logger.Debug("starting bmgsMessageModVal()")
	validTxt, err := grpset.ValidateGroupSettingValue(grpset.MessageModMap, attrName, attrValue)
	if err != nil {
		return err
	}
	grpSetting.MessageModerationLevel = validTxt
	logger.Debug("finished bmgsMessageModVal()")
	return nil
}

func bmgsModContentVal(grpSetting *gset.Groups, attrName string, attrValue string) error {
	logger.Debug("starting bmgsModContentVal()")
	validTxt, err := grpset.ValidateGroupSettingValue(grpset.ModContentMap, attrName, attrValue)
	if err != nil {
		return err
	}
	grpSetting.WhoCanModerateContent = validTxt
	logger.Debug("finished bmgsModContentVal()")
	return nil
}

func bmgsModMemberVal(grpSetting *gset.Groups, attrName string, attrValue string) error {
	logger.Debug("starting bmgsModMemberVal()")
	validTxt, err := grpset.ValidateGroupSettingValue(grpset.ModMemberMap, attrName, attrValue)
	if err != nil {
		return err
	}
	grpSetting.WhoCanModerateMembers = validTxt
	logger.Debug("finished bmgsModMemberVal()")
	return nil
}

func bmgsPostMessageVal(grpSettings *gset.Groups, attrName string, attrValue string) error {
	logger.Debug("starting bmgsPostMessageVal()")
	validTxt, err := grpset.ValidateGroupSettingValue(grpset.PostMessageMap, attrName, attrValue)
	if err != nil {
		return err
	}
	grpSettings.WhoCanPostMessage = validTxt
	logger.Debug("finished bmgsPostMessageVal()")
	return nil
}

func bmgsReplyEmailVal(grpSettings *gset.Groups, attrValue string) error {
	logger.Debug("starting bmgsReplyEmailVal()")
	if attrValue == "" {
		grpSettings.CustomReplyTo = attrValue
		grpSettings.ForceSendFields = append(grpSettings.ForceSendFields, "CustomReplyTo")
		return nil
	}
	ok := valid.IsEmail(attrValue)
	if !ok {
		err := fmt.Errorf(cmn.ErrInvalidEmailAddress, attrValue)
		return err
	}
	grpSettings.CustomReplyTo = attrValue
	logger.Debug("finished bmgsReplyEmailVal()")
	return nil
}

func bmgsReplyToVal(grpSetting *gset.Groups, attrName string, attrValue string) error {
	logger.Debug("starting bmgsReplyToVal()")
	validTxt, err := grpset.ValidateGroupSettingValue(grpset.ReplyToMap, attrName, attrValue)
	if err != nil {
		return err
	}
	grpSetting.ReplyTo = validTxt
	logger.Debug("finished bmgsReplyToVal()")
	return nil
}

func bmgsSpamModVal(grpSetting *gset.Groups, attrName string, attrValue string) error {
	logger.Debug("starting bmgsSpamModVal()")
	validTxt, err := grpset.ValidateGroupSettingValue(grpset.SpamModMap, attrName, attrValue)
	if err != nil {
		return err
	}
	grpSetting.SpamModerationLevel = validTxt
	logger.Debug("finished bmgsSpamModVal()")
	return nil
}

func bmgsViewGroupVal(grpSetting *gset.Groups, attrName string, attrValue string) error {
	logger.Debug("starting bmgsViewGroupVal()")
	validTxt, err := grpset.ValidateGroupSettingValue(grpset.ViewGroupMap, attrName, attrValue)
	if err != nil {
		return err
	}
	grpSetting.WhoCanViewGroup = validTxt
	logger.Debug("finished bmgsViewGroupVal()")
	return nil
}

func bmgsViewMembershipVal(grpSetting *gset.Groups, attrName string, attrValue string) error {
	logger.Debug("starting bmgsViewMembershipVal()")
	validTxt, err := grpset.ValidateGroupSettingValue(grpset.ViewMembershipMap, attrName, attrValue)
	if err != nil {
		return err
	}
	grpSetting.WhoCanViewMembership = validTxt
	logger.Debug("finished bmgsViewMembershipVal()")
	return nil
}

func init() {
	batchManageCmd.AddCommand(batchMngGrpSettingsCmd)

	batchMngGrpSettingsCmd.Flags().StringVarP(&inputFile, "inputfile", "i", "", "filepath to device data file")
	batchMngGrpSettingsCmd.Flags().StringVarP(&format, "format", "f", "json", "user data file format")
	batchMngGrpSettingsCmd.Flags().StringVarP(&sheetRange, "sheetrange", "s", "", "user data gsheet range")
}
