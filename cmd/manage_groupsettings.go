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
	"fmt"

	valid "github.com/asaskevich/govalidator"
	cmn "github.com/plusworx/gmin/utils/common"
	flgnm "github.com/plusworx/gmin/utils/flagnames"
	gmess "github.com/plusworx/gmin/utils/gminmessages"
	grpset "github.com/plusworx/gmin/utils/groupsettings"
	lg "github.com/plusworx/gmin/utils/logging"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	gset "google.golang.org/api/groupssettings/v1"
)

var manageGroupSettingsCmd = &cobra.Command{
	Use:     "group-settings <group email address>",
	Aliases: []string{"grp-settings", "grp-set", "gsettings", "gset"},
	Args:    cobra.ExactArgs(1),
	Example: `gmin manage group-settings sales@mycompany.org --collab-inbox --web-posting
gmin mng gset finance@mycompany.org -c -w`,
	Short: "Manages settings for a group",
	Long:  `Manages settings for a group.`,
	RunE:  doManageGroupSettings,
}

func doManageGroupSettings(cmd *cobra.Command, args []string) error {
	lg.Debugw("starting doManageGroupSettings()",
		"args", args)

	var flagsPassed []string

	grpSettings := new(gset.Groups)

	// Collect names of command flags passed in
	cmd.Flags().Visit(func(f *pflag.Flag) {
		flagsPassed = append(flagsPassed, f.Name)
	})

	// Process command flags
	err := processMngGrpSettingFlags(cmd, grpSettings, flagsPassed)
	if err != nil {
		lg.Error(err)
		return err
	}

	grpSettings.Email = args[0]

	srv, err := cmn.CreateService(cmn.SRVTYPEGRPSETTING, gset.AppsGroupsSettingsScope)
	if err != nil {
		return err
	}
	gss := srv.(*gset.Service)

	gsuc := gss.Groups.Update(args[0], grpSettings)
	newSettings, err := gsuc.Do()
	if err != nil {
		lg.Error(err)
		return err
	}

	fmt.Println(cmn.GminMessage(fmt.Sprintf(gmess.INFO_GROUPSETTINGSCHANGED, newSettings.Email)))
	lg.Infof(gmess.INFO_GROUPSETTINGSCHANGED, newSettings.Email)

	lg.Debug("finished doManageGroupSettings()")
	return nil
}

func init() {
	manageCmd.AddCommand(manageGroupSettingsCmd)

	manageGroupSettingsCmd.Flags().String(flgnm.FLG_APPROVEMEM, "", "who can approve members")
	manageGroupSettingsCmd.Flags().BoolP(flgnm.FLG_ARCHIVED, "r", false, "is archived")
	manageGroupSettingsCmd.Flags().BoolP(flgnm.FLG_ARCHIVEONLY, "a", false, "archive only")
	manageGroupSettingsCmd.Flags().String(flgnm.FLG_ASSISTCONTENT, "", "who can moderate metadata")
	manageGroupSettingsCmd.Flags().String(flgnm.FLG_BANUSER, "", "who can ban user")
	manageGroupSettingsCmd.Flags().BoolP(flgnm.FLG_COLLABINBOX, "c", false, "enable collaborative inbox")
	manageGroupSettingsCmd.Flags().String(flgnm.FLG_CONTACTOWNER, "", "who can contact owner")
	manageGroupSettingsCmd.Flags().String(flgnm.FLG_DENYTEXT, "", "default message deny notification text")
	manageGroupSettingsCmd.Flags().String(flgnm.FLG_DISCGROUP, "", "who can discover group")
	manageGroupSettingsCmd.Flags().BoolP(flgnm.FLG_EXTMEMBER, "e", false, "allow external members")
	manageGroupSettingsCmd.Flags().BoolP(flgnm.FLG_FOOTERON, "f", false, "include custom footer")
	manageGroupSettingsCmd.Flags().String(flgnm.FLG_FOOTERTEXT, "", "custom footer text")
	manageGroupSettingsCmd.Flags().BoolP(flgnm.FLG_GAL, "g", false, "include in Global Address List")
	manageGroupSettingsCmd.Flags().String(flgnm.FLG_JOIN, "", "who can join group")
	manageGroupSettingsCmd.Flags().String(flgnm.FLG_LANGUAGE, "", "primary language")
	manageGroupSettingsCmd.Flags().String(flgnm.FLG_LEAVE, "", "who can leave group")
	manageGroupSettingsCmd.Flags().String(flgnm.FLG_MESSAGEMOD, "", "message moderation level")
	manageGroupSettingsCmd.Flags().String(flgnm.FLG_MODCONTENT, "", "who can moderate content")
	manageGroupSettingsCmd.Flags().String(flgnm.FLG_MODMEMBER, "", "who can moderate members")
	manageGroupSettingsCmd.Flags().BoolP(flgnm.FLG_NOTIFYDENY, "n", false, "send message deny notification")
	manageGroupSettingsCmd.Flags().BoolP(flgnm.FLG_POSTASGROUP, "p", false, "members can post as group")
	manageGroupSettingsCmd.Flags().String(flgnm.FLG_POSTMESSAGE, "", "who can post messages")
	manageGroupSettingsCmd.Flags().BoolP(flgnm.FLG_REPLIESONTOP, "t", false, "favourite replies on top")
	manageGroupSettingsCmd.Flags().String(flgnm.FLG_REPLYEMAIL, "", "custom reply to email address")
	manageGroupSettingsCmd.Flags().String(flgnm.FLG_REPLYTO, "", "who receives the default reply")
	manageGroupSettingsCmd.Flags().String(flgnm.FLG_SPAMMOD, "", "spam moderation level")
	manageGroupSettingsCmd.Flags().String(flgnm.FLG_VIEWGROUP, "", "who can view group")
	manageGroupSettingsCmd.Flags().String(flgnm.FLG_VIEWMEMSHIP, "", "who can view membership")
	manageGroupSettingsCmd.Flags().BoolP(flgnm.FLG_WEBPOSTING, "w", false, "allow web posting")
}

func processMngGrpSettingFlags(cmd *cobra.Command, grpSettings *gset.Groups, flagNames []string) error {
	lg.Debugw("starting processMngGrpSettingFlags()",
		"flagNames", flagNames)
	defer lg.Debug("finished processMngGrpSettingFlags()")

	var (
		err     error
		flgBVal bool
		flgSVal string
	)

	boolFuncMap := map[string]func(*gset.Groups, bool){
		flgnm.FLG_ARCHIVED:     mgsArchivedFlag,
		flgnm.FLG_ARCHIVEONLY:  mgsArchiveOnlyFlag,
		flgnm.FLG_COLLABINBOX:  mgsCollabInboxFlag,
		flgnm.FLG_EXTMEMBER:    mgsExtMemberFlag,
		flgnm.FLG_FOOTERON:     mgsFooterOnFlag,
		flgnm.FLG_GAL:          mgsGalFlag,
		flgnm.FLG_NOTIFYDENY:   mgsNotifyDenyFlag,
		flgnm.FLG_POSTASGROUP:  mgsPostAsGroupFlag,
		flgnm.FLG_REPLIESONTOP: mgsRepliesOnTopFlag,
		flgnm.FLG_WEBPOSTING:   mgsWebPostingFlag,
	}

	oneStrFuncMap := map[string]func(*gset.Groups, string) error{
		flgnm.FLG_DENYTEXT:   mgsDenyTextFlag,
		flgnm.FLG_FOOTERTEXT: mgsFooterTextFlag,
		flgnm.FLG_REPLYEMAIL: mgsReplyEmailFlag,
	}

	twoStrFuncMap := map[string]func(*gset.Groups, string, string) error{
		flgnm.FLG_APPROVEMEM:    mgsApproveMemberFlag,
		flgnm.FLG_ASSISTCONTENT: mgsAssistContentFlag,
		flgnm.FLG_BANUSER:       mgsBanUserFlag,
		flgnm.FLG_CONTACTOWNER:  mgsContactOwnerFlag,
		flgnm.FLG_DISCGROUP:     mgsDiscoverGroupFlag,
		flgnm.FLG_JOIN:          mgsJoinFlag,
		flgnm.FLG_LANGUAGE:      mgsLanguageFlag,
		flgnm.FLG_LEAVE:         mgsLeaveFlag,
		flgnm.FLG_MESSAGEMOD:    mgsMessageModFlag,
		flgnm.FLG_MODCONTENT:    mgsModContentFlag,
		flgnm.FLG_MODMEMBER:     mgsModMemberFlag,
		flgnm.FLG_POSTMESSAGE:   mgsPostMessageFlag,
		flgnm.FLG_REPLYTO:       mgsReplyToFlag,
		flgnm.FLG_SPAMMOD:       mgsSpamModFlag,
		flgnm.FLG_VIEWGROUP:     mgsViewGroupFlag,
		flgnm.FLG_VIEWMEMSHIP:   mgsViewMembershipFlag,
	}

	for _, flName := range flagNames {
		// Try boolFuncMap
		bf, bExists := boolFuncMap[flName]
		if bExists {
			flgBVal, err = cmd.Flags().GetBool(flName)
			if err != nil {
				lg.Error(err)
				return err
			}
			bf(grpSettings, flgBVal)
			continue
		}

		// Get string flag value because it's not a bool
		flgSVal, err = cmd.Flags().GetString(flName)
		if err != nil {
			lg.Error(err)
			return err
		}

		// Try oneStrFuncMap
		osf, osExists := oneStrFuncMap[flName]
		if osExists {
			err := osf(grpSettings, flgSVal)
			if err != nil {
				return err
			}
			continue
		}

		// Try twoStrFuncMap
		tsf, tsExists := twoStrFuncMap[flName]
		if tsExists {
			err := tsf(grpSettings, "--"+flName, flgSVal)
			if err != nil {
				return err
			}
			continue
		}
		// Flag not recognized
		err = fmt.Errorf(gmess.ERR_FLAGNOTRECOGNIZED, flName)
		return err
	}
	return nil
}

func mgsApproveMemberFlag(grpSettings *gset.Groups, flagName string, flagVal string) error {
	lg.Debug("starting mgsApproveMemberFlag()")
	defer lg.Debug("finished mgsApproveMemberFlag()")

	validTxt, err := grpset.ValidateGroupSettingValue(grpset.ApproveMemberMap, flagName, flagVal)
	if err != nil {
		return err
	}
	grpSettings.WhoCanApproveMembers = validTxt
	return nil
}

func mgsArchivedFlag(grpSettings *gset.Groups, flagVal bool) {
	lg.Debug("starting mgsArchivedFlag()")
	defer lg.Debug("finished mgsArchivedFlag()")

	if flagVal {
		grpSettings.IsArchived = "true"
	} else {
		grpSettings.IsArchived = "false"
		grpSettings.ForceSendFields = append(grpSettings.ForceSendFields, "IsArchived")
	}
}

func mgsArchiveOnlyFlag(grpSettings *gset.Groups, flagVal bool) {
	lg.Debug("starting mgsArchiveOnlyFlag()")
	defer lg.Debug("finished mgsArchiveOnlyFlag()")

	if flagVal {
		grpSettings.ArchiveOnly = "true"
	} else {
		grpSettings.ArchiveOnly = "false"
		grpSettings.ForceSendFields = append(grpSettings.ForceSendFields, "ArchiveOnly")
	}
}

func mgsAssistContentFlag(grpSettings *gset.Groups, flagName string, flagVal string) error {
	lg.Debug("starting mgsAssistContentFlag()")
	defer lg.Debug("finished mgsAssistContentFlag()")

	validTxt, err := grpset.ValidateGroupSettingValue(grpset.AssistContentMap, flagName, flagVal)
	if err != nil {
		return err
	}
	grpSettings.WhoCanAssistContent = validTxt
	return nil
}

func mgsBanUserFlag(grpSettings *gset.Groups, flagName string, flagVal string) error {
	lg.Debug("starting mgsBanUserFlag()")
	defer lg.Debug("finished mgsBanUserFlag()")

	validTxt, err := grpset.ValidateGroupSettingValue(grpset.BanUserMap, flagName, flagVal)
	if err != nil {
		return err
	}
	grpSettings.WhoCanBanUsers = validTxt
	return nil
}

func mgsCollabInboxFlag(grpSettings *gset.Groups, flagVal bool) {
	lg.Debug("starting mgsCollabInboxFlag()")
	defer lg.Debug("finished mgsCollabInboxFlag()")

	if flagVal {
		grpSettings.EnableCollaborativeInbox = "true"
	} else {
		grpSettings.EnableCollaborativeInbox = "false"
		grpSettings.ForceSendFields = append(grpSettings.ForceSendFields, "EnableCollaborativeInbox")
	}
}

func mgsContactOwnerFlag(grpSettings *gset.Groups, flagName string, flagVal string) error {
	lg.Debug("starting mgsContactOwnerFlag()")
	defer lg.Debug("finished mgsContactOwnerFlag()")

	validTxt, err := grpset.ValidateGroupSettingValue(grpset.ContactOwnerMap, flagName, flagVal)
	if err != nil {
		return err
	}
	grpSettings.WhoCanContactOwner = validTxt
	return nil
}

func mgsDenyTextFlag(grpSettings *gset.Groups, flagVal string) error {
	lg.Debug("starting mgsDenyTextFlag()")
	defer lg.Debug("finished mgsDenyTextFlag()")

	if flagVal == "" {
		grpSettings.ForceSendFields = append(grpSettings.ForceSendFields, "DefaultMessageDenyNotificationText")
	}
	grpSettings.DefaultMessageDenyNotificationText = flagVal
	return nil
}

func mgsDiscoverGroupFlag(grpSettings *gset.Groups, flagName string, flagVal string) error {
	lg.Debug("starting mgsDiscoverGroupFlag()")
	defer lg.Debug("finished mgsDiscoverGroupFlag()")

	validTxt, err := grpset.ValidateGroupSettingValue(grpset.DiscoverGroupMap, flagName, flagVal)
	if err != nil {
		return err
	}
	grpSettings.WhoCanDiscoverGroup = validTxt
	return nil
}

func mgsExtMemberFlag(grpSettings *gset.Groups, flagVal bool) {
	lg.Debug("starting mgsExtMemberFlag()")
	defer lg.Debug("finished mgsExtMemberFlag()")

	if flagVal {
		grpSettings.AllowExternalMembers = "true"
	} else {
		grpSettings.AllowExternalMembers = "false"
		grpSettings.ForceSendFields = append(grpSettings.ForceSendFields, "AllowExternalMembers")
	}
}

func mgsFooterOnFlag(grpSettings *gset.Groups, flagVal bool) {
	lg.Debug("starting mgsFooterOnFlag()")
	defer lg.Debug("finished mgsFooterOnFlag()")

	if flagVal {
		grpSettings.IncludeCustomFooter = "true"
	} else {
		grpSettings.IncludeCustomFooter = "false"
		grpSettings.ForceSendFields = append(grpSettings.ForceSendFields, "IncludeCustomFooter")
	}
}

func mgsFooterTextFlag(grpSettings *gset.Groups, flagVal string) error {
	lg.Debug("starting mgsFooterTextFlag()")
	defer lg.Debug("finished mgsFooterTextFlag()")

	if flagVal == "" {
		grpSettings.ForceSendFields = append(grpSettings.ForceSendFields, "CustomFooterText")
	}
	grpSettings.CustomFooterText = flagVal
	return nil
}

func mgsGalFlag(grpSettings *gset.Groups, flagVal bool) {
	lg.Debug("starting mgsGroupDirFlag()")
	defer lg.Debug("finished mgsGroupDirFlag()")

	if flagVal {
		grpSettings.IncludeInGlobalAddressList = "true"
	} else {
		grpSettings.IncludeInGlobalAddressList = "false"
		grpSettings.ForceSendFields = append(grpSettings.ForceSendFields, "IncludeInGlobalAddressList")
	}
}

func mgsJoinFlag(grpSettings *gset.Groups, flagName string, flagVal string) error {
	lg.Debug("starting mgsJoinFlag()")
	defer lg.Debug("finished mgsJoinFlag()")

	validTxt, err := grpset.ValidateGroupSettingValue(grpset.JoinMap, flagName, flagVal)
	if err != nil {
		return err
	}
	grpSettings.WhoCanJoin = validTxt
	return nil
}

func mgsLanguageFlag(grpSettings *gset.Groups, flagName string, flagVal string) error {
	lg.Debug("starting mgsLanguageFlag()")
	defer lg.Debug("finished mgsLanguageFlag()")

	validTxt, err := grpset.ValidateGroupSettingValue(grpset.LanguageMap, flagName, flagVal)
	if err != nil {
		return err
	}
	grpSettings.PrimaryLanguage = validTxt
	return nil
}

func mgsLeaveFlag(grpSettings *gset.Groups, flagName string, flagVal string) error {
	lg.Debug("starting mgsLeaveFlag()")
	defer lg.Debug("finished mgsLeaveFlag()")

	validTxt, err := grpset.ValidateGroupSettingValue(grpset.LeaveMap, flagName, flagVal)
	if err != nil {
		return err
	}
	grpSettings.WhoCanLeaveGroup = validTxt
	return nil
}

func mgsMessageModFlag(grpSettings *gset.Groups, flagName string, flagVal string) error {
	lg.Debug("starting mgsMessageModFlag()")
	defer lg.Debug("finished mgsMessageModFlag()")

	validTxt, err := grpset.ValidateGroupSettingValue(grpset.MessageModMap, flagName, flagVal)
	if err != nil {
		return err
	}
	grpSettings.MessageModerationLevel = validTxt
	return nil
}

func mgsModContentFlag(grpSettings *gset.Groups, flagName string, flagVal string) error {
	lg.Debug("starting mgsModContentFlag()")
	defer lg.Debug("finished mgsModContentFlag()")

	validTxt, err := grpset.ValidateGroupSettingValue(grpset.ModContentMap, flagName, flagVal)
	if err != nil {
		return err
	}
	grpSettings.WhoCanModerateContent = validTxt
	return nil
}

func mgsModMemberFlag(grpSettings *gset.Groups, flagName string, flagVal string) error {
	lg.Debug("starting mgsModMemberFlag()")
	defer lg.Debug("finished mgsModMemberFlag()")

	validTxt, err := grpset.ValidateGroupSettingValue(grpset.ModMemberMap, flagName, flagVal)
	if err != nil {
		return err
	}
	grpSettings.WhoCanModerateMembers = validTxt
	return nil
}

func mgsNotifyDenyFlag(grpSettings *gset.Groups, flagVal bool) {
	lg.Debug("starting mgsNotifyDenyFlag()")
	defer lg.Debug("finished mgsNotifyDenyFlag()")

	if flagVal {
		grpSettings.SendMessageDenyNotification = "true"
	} else {
		grpSettings.SendMessageDenyNotification = "false"
		grpSettings.ForceSendFields = append(grpSettings.ForceSendFields, "SendMessageDenyNotification")
	}
}

func mgsPostAsGroupFlag(grpSettings *gset.Groups, flagVal bool) {
	lg.Debug("starting mgsPostAsGroupFlag()")
	defer lg.Debug("finished mgsPostAsGroupFlag()")

	if flagVal {
		grpSettings.MembersCanPostAsTheGroup = "true"
	} else {
		grpSettings.MembersCanPostAsTheGroup = "false"
		grpSettings.ForceSendFields = append(grpSettings.ForceSendFields, "MembersCanPostAsTheGroup")
	}
}

func mgsPostMessageFlag(grpSettings *gset.Groups, flagName string, flagVal string) error {
	lg.Debug("starting mgsPostMessageFlag()")
	defer lg.Debug("finished mgsPostMessageFlag()")

	validTxt, err := grpset.ValidateGroupSettingValue(grpset.PostMessageMap, flagName, flagVal)
	if err != nil {
		return err
	}
	grpSettings.WhoCanPostMessage = validTxt
	return nil
}

func mgsRepliesOnTopFlag(grpSettings *gset.Groups, flagVal bool) {
	lg.Debug("starting mgsRepliesOnTopFlag()")
	defer lg.Debug("finished mgsRepliesOnTopFlag()")

	if flagVal {
		grpSettings.FavoriteRepliesOnTop = "true"
	} else {
		grpSettings.FavoriteRepliesOnTop = "false"
		grpSettings.ForceSendFields = append(grpSettings.ForceSendFields, "FavoriteRepliesOnTop")
	}
}

func mgsReplyEmailFlag(grpSettings *gset.Groups, flagVal string) error {
	lg.Debug("starting mgsReplyEmailFlag()")
	defer lg.Debug("finished mgsReplyEmailFlag()")

	if flagVal == "" {
		grpSettings.CustomReplyTo = flagVal
		grpSettings.ForceSendFields = append(grpSettings.ForceSendFields, "CustomReplyTo")
		return nil
	}
	ok := valid.IsEmail(flagVal)
	if !ok {
		err := fmt.Errorf(gmess.ERR_INVALIDEMAILADDRESS, flagVal)
		return err
	}
	grpSettings.CustomReplyTo = flagVal
	return nil
}

func mgsReplyToFlag(grpSettings *gset.Groups, flagName string, flagVal string) error {
	lg.Debug("starting mgsReplyToFlag()")
	defer lg.Debug("finished mgsReplyToFlag()")

	validTxt, err := grpset.ValidateGroupSettingValue(grpset.ReplyToMap, flagName, flagVal)
	if err != nil {
		return err
	}
	grpSettings.ReplyTo = validTxt
	return nil
}

func mgsSpamModFlag(grpSettings *gset.Groups, flagName string, flagVal string) error {
	lg.Debug("starting mgsSpamModFlag()")
	defer lg.Debug("finished mgsSpamModFlag()")

	validTxt, err := grpset.ValidateGroupSettingValue(grpset.SpamModMap, flagName, flagVal)
	if err != nil {
		return err
	}
	grpSettings.SpamModerationLevel = validTxt
	return nil
}

func mgsViewGroupFlag(grpSettings *gset.Groups, flagName string, flagVal string) error {
	lg.Debug("starting mgsViewGroupFlag()")
	defer lg.Debug("finished mgsViewGroupFlag()")

	validTxt, err := grpset.ValidateGroupSettingValue(grpset.ViewGroupMap, flagName, flagVal)
	if err != nil {
		return err
	}
	grpSettings.WhoCanViewGroup = validTxt
	return nil
}

func mgsViewMembershipFlag(grpSettings *gset.Groups, flagName string, flagVal string) error {
	lg.Debug("starting mgsViewMembershipFlag()")
	defer lg.Debug("finished mgsViewMembershipFlag()")

	validTxt, err := grpset.ValidateGroupSettingValue(grpset.ViewMembershipMap, flagName, flagVal)
	if err != nil {
		return err
	}
	grpSettings.WhoCanViewMembership = validTxt
	return nil
}

func mgsWebPostingFlag(grpSettings *gset.Groups, flagVal bool) {
	lg.Debug("starting mgsWebPostingFlag()")
	defer lg.Debug("finished mgsWebPostingFlag()")

	if flagVal {
		grpSettings.AllowWebPosting = "true"
	} else {
		grpSettings.AllowWebPosting = "false"
		grpSettings.ForceSendFields = append(grpSettings.ForceSendFields, "AllowWebPosting")
	}
}
