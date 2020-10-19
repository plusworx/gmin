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

	gss, err := cmn.CreateGroupSettingService(gset.AppsGroupsSettingsScope)
	if err != nil {
		lg.Error(err)
		return err
	}

	gsuc := gss.Groups.Update(args[0], grpSettings)
	newSettings, err := gsuc.Do()
	if err != nil {
		lg.Error(err)
		return err
	}

	lg.Infof(gmess.INFO_GROUPSETTINGSCHANGED, newSettings.Email)
	fmt.Println(cmn.GminMessage(fmt.Sprintf(gmess.INFO_GROUPSETTINGSCHANGED, newSettings.Email)))

	lg.Debug("finished doManageGroupSettings()")
	return nil
}

func init() {
	manageCmd.AddCommand(manageGroupSettingsCmd)

	manageGroupSettingsCmd.Flags().StringVar(&approveMems, flgnm.FLG_APPROVEMEM, "", "who can approve members")
	manageGroupSettingsCmd.Flags().BoolVarP(&isArchived, flgnm.FLG_ARCHIVED, "r", false, "is archived")
	manageGroupSettingsCmd.Flags().BoolVarP(&archiveOnly, flgnm.FLG_ARCHIVEONLY, "a", false, "archive only")
	manageGroupSettingsCmd.Flags().StringVar(&assistContent, flgnm.FLG_ASSISTCONTENT, "", "who can moderate metadata")
	manageGroupSettingsCmd.Flags().StringVar(&banUser, flgnm.FLG_BANUSER, "", "who can ban user")
	manageGroupSettingsCmd.Flags().BoolVarP(&collabInbox, flgnm.FLG_COLLABINBOX, "c", false, "enable collaborative inbox")
	manageGroupSettingsCmd.Flags().StringVar(&contactOwner, flgnm.FLG_CONTACTOWNER, "", "who can contact owner")
	manageGroupSettingsCmd.Flags().StringVar(&denyText, flgnm.FLG_DENYTEXT, "", "default message deny notification text")
	manageGroupSettingsCmd.Flags().StringVar(&discoverGroup, flgnm.FLG_DISCGROUP, "", "who can discover group")
	manageGroupSettingsCmd.Flags().BoolVarP(&extMems, flgnm.FLG_EXTMEMBER, "e", false, "allow external members")
	manageGroupSettingsCmd.Flags().BoolVarP(&incFooter, flgnm.FLG_FOOTERON, "f", false, "include custom footer")
	manageGroupSettingsCmd.Flags().StringVar(&footerText, flgnm.FLG_FOOTERTEXT, "", "custom footer text")
	manageGroupSettingsCmd.Flags().BoolVarP(&gal, flgnm.FLG_GAL, "g", false, "include in Global Address List")
	manageGroupSettingsCmd.Flags().StringVar(&join, flgnm.FLG_JOIN, "", "who can join group")
	manageGroupSettingsCmd.Flags().StringVar(&language, flgnm.FLG_LANGUAGE, "", "primary language")
	manageGroupSettingsCmd.Flags().StringVar(&leave, flgnm.FLG_LEAVE, "", "who can leave group")
	manageGroupSettingsCmd.Flags().StringVar(&messageMod, flgnm.FLG_MESSAGEMOD, "", "message moderation level")
	manageGroupSettingsCmd.Flags().StringVar(&modContent, flgnm.FLG_MODCONTENT, "", "who can moderate content")
	manageGroupSettingsCmd.Flags().StringVar(&modMems, flgnm.FLG_MODMEMBER, "", "who can moderate members")
	manageGroupSettingsCmd.Flags().BoolVarP(&denyNotification, flgnm.FLG_NOTIFYDENY, "n", false, "send message deny notification")
	manageGroupSettingsCmd.Flags().BoolVarP(&postAsGroup, flgnm.FLG_POSTASGROUP, "p", false, "members can post as group")
	manageGroupSettingsCmd.Flags().StringVar(&postMessage, flgnm.FLG_POSTMESSAGE, "", "who can post messages")
	manageGroupSettingsCmd.Flags().BoolVarP(&repliesOnTop, flgnm.FLG_REPLIESONTOP, "t", false, "favourite replies on top")
	manageGroupSettingsCmd.Flags().StringVar(&replyEmail, flgnm.FLG_REPLYEMAIL, "", "custom reply to email address")
	manageGroupSettingsCmd.Flags().StringVar(&replyTo, flgnm.FLG_REPLYTO, "", "who receives the default reply")
	manageGroupSettingsCmd.Flags().StringVar(&spamMod, flgnm.FLG_SPAMMOD, "", "spam moderation level")
	manageGroupSettingsCmd.Flags().StringVar(&viewGroup, flgnm.FLG_VIEWGROUP, "", "who can view group")
	manageGroupSettingsCmd.Flags().StringVar(&viewMems, flgnm.FLG_VIEWMEMSHIP, "", "who can view membership")
	manageGroupSettingsCmd.Flags().BoolVarP(&webPosting, flgnm.FLG_WEBPOSTING, "w", false, "allow web posting")
}

func processMngGrpSettingFlags(cmd *cobra.Command, grpSettings *gset.Groups, flagNames []string) error {
	lg.Debugw("starting processMngGrpSettingFlags()",
		"flagNames", flagNames)
	defer lg.Debug("finished processMngGrpSettingFlags()")

	var (
		boolFlags = []string{
			flgnm.FLG_ARCHIVED,
			flgnm.FLG_ARCHIVEONLY,
			flgnm.FLG_COLLABINBOX,
			flgnm.FLG_EXTMEMBER,
			flgnm.FLG_FOOTERON,
			flgnm.FLG_GAL,
			flgnm.FLG_NOTIFYDENY,
			flgnm.FLG_POSTASGROUP,
			flgnm.FLG_REPLIESONTOP,
			flgnm.FLG_WEBPOSTING,
		}
		err     error
		flgBVal bool
		flgSVal string
	)

	for _, flName := range flagNames {
		ok := cmn.SliceContainsStr(boolFlags, flName)
		if ok {
			flgBVal, err = cmd.Flags().GetBool(flName)
			if err != nil {
				lg.Error(err)
				return err
			}
		} else {
			flgSVal, err = cmd.Flags().GetString(flName)
			if err != nil {
				lg.Error(err)
				return err
			}
		}
		if flName == flgnm.FLG_APPROVEMEM {
			err = mgsApproveMemberFlag(grpSettings, "--"+flName, flgSVal)
			if err != nil {
				return err
			}
			return nil
		}
		if flName == flgnm.FLG_ARCHIVED {
			mgsArchivedFlag(grpSettings, flgBVal)
			return nil
		}
		if flName == flgnm.FLG_ARCHIVEONLY {
			mgsArchiveOnlyFlag(grpSettings, flgBVal)
			return nil
		}
		if flName == flgnm.FLG_ASSISTCONTENT {
			err := mgsAssistContentFlag(grpSettings, "--"+flName, flgSVal)
			if err != nil {
				return err
			}
			return nil
		}
		if flName == flgnm.FLG_BANUSER {
			err := mgsBanUserFlag(grpSettings, "--"+flName, flgSVal)
			if err != nil {
				return err
			}
			return nil
		}
		if flName == flgnm.FLG_COLLABINBOX {
			mgsCollabInboxFlag(grpSettings, flgBVal)
			return nil
		}
		if flName == flgnm.FLG_CONTACTOWNER {
			err := mgsContactOwnerFlag(grpSettings, "--"+flName, flgSVal)
			if err != nil {
				return err
			}
			return nil
		}
		if flName == flgnm.FLG_DENYTEXT {
			mgsDenyTextFlag(grpSettings, flgSVal)
			return nil
		}
		if flName == flgnm.FLG_DISCGROUP {
			err := mgsDiscoverGroupFlag(grpSettings, "--"+flName, flgSVal)
			if err != nil {
				return err
			}
			return nil
		}
		if flName == flgnm.FLG_EXTMEMBER {
			mgsExtMemberFlag(grpSettings, flgBVal)
			return nil
		}
		if flName == flgnm.FLG_FOOTERON {
			mgsFooterOnFlag(grpSettings, flgBVal)
			return nil
		}
		if flName == flgnm.FLG_FOOTERTEXT {
			mgsFooterTextFlag(grpSettings, flgSVal)
			return nil
		}
		if flName == flgnm.FLG_GAL {
			mgsGalFlag(grpSettings, flgBVal)
			return nil
		}
		if flName == flgnm.FLG_JOIN {
			err := mgsJoinFlag(grpSettings, "--"+flName, flgSVal)
			if err != nil {
				return err
			}
			return nil
		}
		if flName == flgnm.FLG_LANGUAGE {
			err := mgsLanguageFlag(grpSettings, "--"+flName, flgSVal)
			if err != nil {
				return err
			}
			return nil
		}
		if flName == flgnm.FLG_LEAVE {
			err := mgsLeaveFlag(grpSettings, "--"+flName, flgSVal)
			if err != nil {
				return err
			}
			return nil
		}
		if flName == flgnm.FLG_MESSAGEMOD {
			err := mgsMessageModFlag(grpSettings, "--"+flName, flgSVal)
			if err != nil {
				return err
			}
			return nil
		}
		if flName == flgnm.FLG_MODCONTENT {
			err := mgsModContentFlag(grpSettings, "--"+flName, flgSVal)
			if err != nil {
				return err
			}
			return nil
		}
		if flName == flgnm.FLG_MODMEMBER {
			err := mgsModMemberFlag(grpSettings, "--"+flName, flgSVal)
			if err != nil {
				return err
			}
			return nil
		}
		if flName == flgnm.FLG_NOTIFYDENY {
			mgsNotifyDenyFlag(grpSettings, flgBVal)
			return nil
		}
		if flName == flgnm.FLG_POSTASGROUP {
			mgsPostAsGroupFlag(grpSettings, flgBVal)
			return nil
		}
		if flName == flgnm.FLG_POSTMESSAGE {
			err := mgsPostMessageFlag(grpSettings, "--"+flName, flgSVal)
			if err != nil {
				return err
			}
			return nil
		}
		if flName == flgnm.FLG_REPLIESONTOP {
			mgsRepliesOnTopFlag(grpSettings, flgBVal)
			return nil
		}
		if flName == flgnm.FLG_REPLYEMAIL {
			err := mgsReplyEmailFlag(grpSettings, flgSVal)
			if err != nil {
				return err
			}
			return nil
		}
		if flName == flgnm.FLG_REPLYTO {
			err := mgsReplyToFlag(grpSettings, "--"+flName, flgSVal)
			if err != nil {
				return err
			}
			return nil
		}
		if flName == flgnm.FLG_SPAMMOD {
			err := mgsSpamModFlag(grpSettings, "--"+flName, flgSVal)
			if err != nil {
				return err
			}
			return nil
		}
		if flName == flgnm.FLG_VIEWGROUP {
			err := mgsViewGroupFlag(grpSettings, "--"+flName, flgSVal)
			if err != nil {
				return err
			}
			return nil
		}
		if flName == flgnm.FLG_VIEWMEMSHIP {
			err := mgsViewMembershipFlag(grpSettings, "--"+flName, flgSVal)
			if err != nil {
				return err
			}
			return nil
		}
		if flName == flgnm.FLG_WEBPOSTING {
			mgsWebPostingFlag(grpSettings, flgBVal)
			return nil
		}
	}
	return nil
}

// Process command flag input

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

func mgsDenyTextFlag(grpSettings *gset.Groups, flagVal string) {
	lg.Debug("starting mgsDenyTextFlag()")
	defer lg.Debug("finished mgsDenyTextFlag()")

	if flagVal == "" {
		grpSettings.ForceSendFields = append(grpSettings.ForceSendFields, "DefaultMessageDenyNotificationText")
	}
	grpSettings.DefaultMessageDenyNotificationText = flagVal
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

func mgsFooterTextFlag(grpSettings *gset.Groups, flagVal string) {
	lg.Debug("starting mgsFooterTextFlag()")
	defer lg.Debug("finished mgsFooterTextFlag()")

	if flagVal == "" {
		grpSettings.ForceSendFields = append(grpSettings.ForceSendFields, "CustomFooterText")
	}
	grpSettings.CustomFooterText = flagVal
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
		grpSettings.CustomReplyTo = replyEmail
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
