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
	gmess "github.com/plusworx/gmin/utils/gminmessages"
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
		err := errors.New(gmess.ErrNoInputFile)
		logger.Error(err)
		return err
	}

	lwrFmt := strings.ToLower(format)

	ok := cmn.SliceContainsStr(cmn.ValidFileFormats, lwrFmt)
	if !ok {
		err = fmt.Errorf(gmess.ErrInvalidFileFormat, format)
		logger.Error(err)
		return err
	}

	switch {
	case lwrFmt == "csv":
		err := bmnggProcessCSVFile(ds, inputFile)
		if err != nil {
			logger.Error(err)
			return err
		}
	case lwrFmt == "json":
		err := bmnggProcessJSON(ds, inputFile, scanner)
		if err != nil {
			logger.Error(err)
			return err
		}
	case lwrFmt == "gsheet":
		err := bmnggProcessGSheet(ds, inputFile)
		if err != nil {
			logger.Error(err)
			return err
		}
	default:
		return fmt.Errorf(gmess.ErrInvalidFileFormat, format)
	}
	logger.Debug("finished doBatchMngGrpSettings()")
	return nil
}

func bmnggApproveMemberVal(grpSetting *gset.Groups, attrName string, attrValue string) error {
	logger.Debug("starting bmnggApproveMemberVal()")
	validTxt, err := grpset.ValidateGroupSettingValue(grpset.ApproveMemberMap, attrName, attrValue)
	if err != nil {
		return err
	}
	grpSetting.WhoCanApproveMembers = validTxt
	logger.Debug("finished bmnggApproveMemberVal()")
	return nil
}

func bmnggAssistContentVal(grpSetting *gset.Groups, attrName string, attrValue string) error {
	logger.Debug("starting bmnggAssistContentVal()")
	validTxt, err := grpset.ValidateGroupSettingValue(grpset.AssistContentMap, attrName, attrValue)
	if err != nil {
		return err
	}
	grpSetting.WhoCanAssistContent = validTxt
	logger.Debug("finished bmnggAssistContentVal()")
	return nil
}

func bmnggBanUserVal(grpSetting *gset.Groups, attrName string, attrValue string) error {
	logger.Debug("starting bmnggBanUserVal()")
	validTxt, err := grpset.ValidateGroupSettingValue(grpset.BanUserMap, attrName, attrValue)
	if err != nil {
		return err
	}
	grpSetting.WhoCanBanUsers = validTxt
	logger.Debug("finished bmnggBanUserVal()")
	return nil
}

func bmnggContactOwnerVal(grpSetting *gset.Groups, attrName string, attrValue string) error {
	logger.Debug("starting bmnggContactOwnerVal()")
	validTxt, err := grpset.ValidateGroupSettingValue(grpset.ContactOwnerMap, attrName, attrValue)
	if err != nil {
		return err
	}
	grpSetting.WhoCanContactOwner = validTxt
	logger.Debug("finished bmnggContactOwnerVal()")
	return nil
}

func bmnggDiscoverGroupVal(grpSettings *gset.Groups, attrName string, attrValue string) error {
	logger.Debug("starting bmnggDiscoverGroupVal()")
	validTxt, err := grpset.ValidateGroupSettingValue(grpset.DiscoverGroupMap, attrName, discoverGroup)
	if err != nil {
		return err
	}
	grpSettings.WhoCanDiscoverGroup = validTxt
	logger.Debug("finished bmnggDiscoverGroupVal()")
	return nil
}

func bmnggFromFileFactory(hdrMap map[int]string, gsData []interface{}) (*gset.Groups, string, error) {
	logger.Debugw("starting bmnggFromFileFactory()",
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
				err := fmt.Errorf(gmess.ErrEmptyString, attrName)
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
	logger.Debug("finished bmnggFromFileFactory()")
	return grpSetting, groupKey, nil
}

func bmnggFromJSONFactory(ds *gset.Service, jsonData string) (*gset.Groups, string, error) {
	logger.Debugw("starting bmnggFromJSONFactory()",
		"jsonData", jsonData)

	var (
		grpSettings *gset.Groups
		grpKey      = grpset.Key{}
	)

	grpSettings = new(gset.Groups)
	jsonBytes := []byte(jsonData)

	if !json.Valid(jsonBytes) {
		logger.Error(gmess.ErrInvalidJSONAttr)
		return nil, "", errors.New(gmess.ErrInvalidJSONAttr)
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
		err = errors.New(gmess.ErrNoJSONGroupKey)
		logger.Error(err)
		return nil, "", err
	}

	err = json.Unmarshal(jsonBytes, &grpSettings)
	if err != nil {
		logger.Error(err)
		return nil, "", err
	}
	logger.Debug("finished bmnggFromJSONFactory()")
	return grpSettings, grpKey.GroupKey, nil
}

func bmnggJoinVal(grpSetting *gset.Groups, attrName string, attrValue string) error {
	logger.Debug("starting bmnggJoinVal()")
	validTxt, err := grpset.ValidateGroupSettingValue(grpset.JoinMap, attrName, attrValue)
	if err != nil {
		return err
	}
	grpSetting.WhoCanJoin = validTxt
	logger.Debug("finished bmnggJoinVal()")
	return nil
}

func bmnggLanguageVal(grpSetting *gset.Groups, attrName string, attrValue string) error {
	logger.Debug("starting bmnggLanguageVal()")
	validTxt, err := grpset.ValidateGroupSettingValue(grpset.LanguageMap, attrName, attrValue)
	if err != nil {
		return err
	}
	grpSetting.PrimaryLanguage = validTxt
	logger.Debug("finished bmnggLanguageVal()")
	return nil
}

func bmnggLeaveVal(grpSetting *gset.Groups, attrName string, attrValue string) error {
	logger.Debug("starting bmnggLeaveVal()")
	validTxt, err := grpset.ValidateGroupSettingValue(grpset.LeaveMap, attrName, attrValue)
	if err != nil {
		return err
	}
	grpSetting.WhoCanLeaveGroup = validTxt
	logger.Debug("finished bmnggLeaveVal()")
	return nil
}

func bmnggMessageModVal(grpSetting *gset.Groups, attrName string, attrValue string) error {
	logger.Debug("starting bmnggMessageModVal()")
	validTxt, err := grpset.ValidateGroupSettingValue(grpset.MessageModMap, attrName, attrValue)
	if err != nil {
		return err
	}
	grpSetting.MessageModerationLevel = validTxt
	logger.Debug("finished bmnggMessageModVal()")
	return nil
}

func bmnggModContentVal(grpSetting *gset.Groups, attrName string, attrValue string) error {
	logger.Debug("starting bmnggModContentVal()")
	validTxt, err := grpset.ValidateGroupSettingValue(grpset.ModContentMap, attrName, attrValue)
	if err != nil {
		return err
	}
	grpSetting.WhoCanModerateContent = validTxt
	logger.Debug("finished bmnggModContentVal()")
	return nil
}

func bmnggModMemberVal(grpSetting *gset.Groups, attrName string, attrValue string) error {
	logger.Debug("starting bmnggModMemberVal()")
	validTxt, err := grpset.ValidateGroupSettingValue(grpset.ModMemberMap, attrName, attrValue)
	if err != nil {
		return err
	}
	grpSetting.WhoCanModerateMembers = validTxt
	logger.Debug("finished bmnggModMemberVal()")
	return nil
}

func bmnggPerformUpdate(grpSetting *gset.Groups, groupKey string, wg *sync.WaitGroup, gsuc *gset.GroupsUpdateCall) {
	logger.Debugw("starting bmnggPerformUpdate()",
		"groupKey", groupKey)

	defer wg.Done()

	b := backoff.NewExponentialBackOff()
	b.MaxElapsedTime = 32 * time.Second

	err := backoff.Retry(func() error {
		var err error
		_, err = gsuc.Do()
		if err == nil {
			logger.Infof(gmess.InfoGroupSettingsChanged, groupKey)
			fmt.Println(cmn.GminMessage(fmt.Sprintf(gmess.InfoGroupSettingsChanged, groupKey)))
			return err
		}
		if !cmn.IsErrRetryable(err) {
			return backoff.Permanent(fmt.Errorf(gmess.ErrBatchGroupSettings, err.Error(), groupKey))
		}
		// Log the retries
		logger.Warnw(err.Error(),
			"retrying", b.GetElapsedTime().String(),
			"group", groupKey)
		return fmt.Errorf(gmess.ErrBatchGroupSettings, err.Error(), groupKey)
	}, b)
	if err != nil {
		// Log final error
		logger.Error(err)
		fmt.Println(cmn.GminMessage(err.Error()))
	}
	logger.Debug("finished bmnggPerformUpdate()")
}

func bmnggPostMessageVal(grpSettings *gset.Groups, attrName string, attrValue string) error {
	logger.Debug("starting bmnggPostMessageVal()")
	validTxt, err := grpset.ValidateGroupSettingValue(grpset.PostMessageMap, attrName, attrValue)
	if err != nil {
		return err
	}
	grpSettings.WhoCanPostMessage = validTxt
	logger.Debug("finished bmnggPostMessageVal()")
	return nil
}

func bmnggProcessCSVFile(ds *gset.Service, filePath string) error {
	logger.Debugw("starting bmnggProcessCSVFile()",
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

		gsVar, groupKey, err := bmnggFromFileFactory(hdrMap, iSlice)
		if err != nil {
			logger.Error(err)
			return err
		}

		groupKeys = append(groupKeys, groupKey)
		grpSettings = append(grpSettings, gsVar)
		count = count + 1
	}

	err = bmnggProcessObjects(ds, groupKeys, grpSettings)
	if err != nil {
		logger.Error(err)
		return err
	}
	logger.Debug("finished bmnggProcessCSVFile()")
	return nil
}

func bmnggProcessGSheet(ds *gset.Service, sheetID string) error {
	logger.Debugw("starting bmnggProcessGSheet()",
		"sheetID", sheetID)

	var (
		err         error
		groupKeys   []string
		grpSettings []*gset.Groups
	)

	if sheetRange == "" {
		err := errors.New(gmess.ErrNoSheetRange)
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
		err = fmt.Errorf(gmess.ErrNoSheetDataFound, sheetID, sheetRange)
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

		gsVar, groupKey, err := bmnggFromFileFactory(hdrMap, row)
		if err != nil {
			logger.Error(err)
			return err
		}

		groupKeys = append(groupKeys, groupKey)
		grpSettings = append(grpSettings, gsVar)
	}

	err = bmnggProcessObjects(ds, groupKeys, grpSettings)
	if err != nil {
		logger.Error(err)
		return err
	}
	logger.Debug("finished bmnggProcessGSheet()")
	return nil
}

func bmnggProcessJSON(ds *gset.Service, filePath string, scanner *bufio.Scanner) error {
	logger.Debugw("starting bmnggProcessJSON()",
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

		gsVar, groupKey, err := bmnggFromJSONFactory(ds, jsonData)
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

	err = bmnggProcessObjects(ds, groupKeys, grpSettings)
	if err != nil {
		logger.Error(err)
		return err
	}
	logger.Debug("finished bmnggProcessJSON()")
	return nil
}

func bmnggProcessObjects(ds *gset.Service, groupKeys []string, grpSettings []*gset.Groups) error {
	logger.Debugw("starting bmnggProcessObjects()",
		"groupKeys", groupKeys)

	wg := new(sync.WaitGroup)

	for idx, gs := range grpSettings {
		gsuc := ds.Groups.Update(groupKeys[idx], gs)

		wg.Add(1)

		go bmnggPerformUpdate(gs, groupKeys[idx], wg, gsuc)
	}

	wg.Wait()

	logger.Debug("finished bmnggProcessObjects()")
	return nil
}

func bmnggReplyEmailVal(grpSettings *gset.Groups, attrValue string) error {
	logger.Debug("starting bmnggReplyEmailVal()")
	if attrValue == "" {
		grpSettings.CustomReplyTo = attrValue
		grpSettings.ForceSendFields = append(grpSettings.ForceSendFields, "CustomReplyTo")
		return nil
	}
	ok := valid.IsEmail(attrValue)
	if !ok {
		err := fmt.Errorf(gmess.ErrInvalidEmailAddress, attrValue)
		return err
	}
	grpSettings.CustomReplyTo = attrValue
	logger.Debug("finished bmnggReplyEmailVal()")
	return nil
}

func bmnggReplyToVal(grpSetting *gset.Groups, attrName string, attrValue string) error {
	logger.Debug("starting bmnggReplyToVal()")
	validTxt, err := grpset.ValidateGroupSettingValue(grpset.ReplyToMap, attrName, attrValue)
	if err != nil {
		return err
	}
	grpSetting.ReplyTo = validTxt
	logger.Debug("finished bmnggReplyToVal()")
	return nil
}

func bmnggSpamModVal(grpSetting *gset.Groups, attrName string, attrValue string) error {
	logger.Debug("starting bmnggSpamModVal()")
	validTxt, err := grpset.ValidateGroupSettingValue(grpset.SpamModMap, attrName, attrValue)
	if err != nil {
		return err
	}
	grpSetting.SpamModerationLevel = validTxt
	logger.Debug("finished bmnggSpamModVal()")
	return nil
}

func bmnggViewGroupVal(grpSetting *gset.Groups, attrName string, attrValue string) error {
	logger.Debug("starting bmnggViewGroupVal()")
	validTxt, err := grpset.ValidateGroupSettingValue(grpset.ViewGroupMap, attrName, attrValue)
	if err != nil {
		return err
	}
	grpSetting.WhoCanViewGroup = validTxt
	logger.Debug("finished bmnggViewGroupVal()")
	return nil
}

func bmnggViewMembershipVal(grpSetting *gset.Groups, attrName string, attrValue string) error {
	logger.Debug("starting bmnggViewMembershipVal()")
	validTxt, err := grpset.ValidateGroupSettingValue(grpset.ViewMembershipMap, attrName, attrValue)
	if err != nil {
		return err
	}
	grpSetting.WhoCanViewMembership = validTxt
	logger.Debug("finished bmnggViewMembershipVal()")
	return nil
}

func init() {
	batchManageCmd.AddCommand(batchMngGrpSettingsCmd)

	batchMngGrpSettingsCmd.Flags().StringVarP(&inputFile, "input-file", "i", "", "filepath to device data file")
	batchMngGrpSettingsCmd.Flags().StringVarP(&format, "format", "f", "json", "user data file format")
	batchMngGrpSettingsCmd.Flags().StringVarP(&sheetRange, "sheet-range", "s", "", "user data gsheet range")
}
