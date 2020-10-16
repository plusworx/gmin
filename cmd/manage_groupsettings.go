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
	lg.Debug("finished processMngGrpSettingFlags()")
	return nil
}

// Process command flag input

func mgsApproveMemberFlag(grpSettings *gset.Groups, flagName string, flagVal string) error {
	lg.Debug("starting mgsApproveMemberFlag()")
	validTxt, err := grpset.ValidateGroupSettingValue(grpset.ApproveMemberMap, flagName, flagVal)
	if err != nil {
		return err
	}
	grpSettings.WhoCanApproveMembers = validTxt
	lg.Debug("finished mgsApproveMemberFlag()")
	return nil
}

func mgsArchivedFlag(grpSettings *gset.Groups, flagVal bool) {
	lg.Debug("starting mgsArchivedFlag()")
	if flagVal {
		grpSettings.IsArchived = "true"
	} else {
		grpSettings.IsArchived = "false"
		grpSettings.ForceSendFields = append(grpSettings.ForceSendFields, "IsArchived")
	}
	lg.Debug("finished mgsArchivedFlag()")
}

func mgsArchiveOnlyFlag(grpSettings *gset.Groups, flagVal bool) {
	lg.Debug("starting mgsArchiveOnlyFlag()")
	if flagVal {
		grpSettings.ArchiveOnly = "true"
	} else {
		grpSettings.ArchiveOnly = "false"
		grpSettings.ForceSendFields = append(grpSettings.ForceSendFields, "ArchiveOnly")
	}
	lg.Debug("finished mgsArchiveOnlyFlag()")
}

func mgsAssistContentFlag(grpSettings *gset.Groups, flagName string, flagVal string) error {
	lg.Debug("starting mgsAssistContentFlag()")
	validTxt, err := grpset.ValidateGroupSettingValue(grpset.AssistContentMap, flagName, flagVal)
	if err != nil {
		return err
	}
	grpSettings.WhoCanAssistContent = validTxt
	lg.Debug("finished mgsAssistContentFlag()")
	return nil
}

func mgsBanUserFlag(grpSettings *gset.Groups, flagName string, flagVal string) error {
	lg.Debug("starting mgsBanUserFlag()")
	validTxt, err := grpset.ValidateGroupSettingValue(grpset.BanUserMap, flagName, flagVal)
	if err != nil {
		return err
	}
	grpSettings.WhoCanBanUsers = validTxt
	lg.Debug("finished mgsBanUserFlag()")
	return nil
}

func mgsCollabInboxFlag(grpSettings *gset.Groups, flagVal bool) {
	lg.Debug("starting mgsCollabInboxFlag()")
	if flagVal {
		grpSettings.EnableCollaborativeInbox = "true"
	} else {
		grpSettings.EnableCollaborativeInbox = "false"
		grpSettings.ForceSendFields = append(grpSettings.ForceSendFields, "EnableCollaborativeInbox")
	}
	lg.Debug("finished mgsCollabInboxFlag()")
}

func mgsContactOwnerFlag(grpSettings *gset.Groups, flagName string, flagVal string) error {
	lg.Debug("starting mgsContactOwnerFlag()")
	validTxt, err := grpset.ValidateGroupSettingValue(grpset.ContactOwnerMap, flagName, flagVal)
	if err != nil {
		return err
	}
	grpSettings.WhoCanContactOwner = validTxt
	lg.Debug("finished mgsContactOwnerFlag()")
	return nil
}

func mgsDenyTextFlag(grpSettings *gset.Groups, flagVal string) {
	lg.Debug("starting mgsDenyTextFlag()")
	if flagVal == "" {
		grpSettings.ForceSendFields = append(grpSettings.ForceSendFields, "DefaultMessageDenyNotificationText")
	}
	grpSettings.DefaultMessageDenyNotificationText = flagVal
	lg.Debug("finished mgsDenyTextFlag()")
}

func mgsDiscoverGroupFlag(grpSettings *gset.Groups, flagName string, flagVal string) error {
	lg.Debug("starting mgsDiscoverGroupFlag()")
	validTxt, err := grpset.ValidateGroupSettingValue(grpset.DiscoverGroupMap, flagName, flagVal)
	if err != nil {
		return err
	}
	grpSettings.WhoCanDiscoverGroup = validTxt
	lg.Debug("finished mgsDiscoverGroupFlag()")
	return nil
}

func mgsExtMemberFlag(grpSettings *gset.Groups, flagVal bool) {
	lg.Debug("starting mgsExtMemberFlag()")
	if flagVal {
		grpSettings.AllowExternalMembers = "true"
	} else {
		grpSettings.AllowExternalMembers = "false"
		grpSettings.ForceSendFields = append(grpSettings.ForceSendFields, "AllowExternalMembers")
	}
	lg.Debug("finished mgsExtMemberFlag()")
}

func mgsFooterOnFlag(grpSettings *gset.Groups, flagVal bool) {
	lg.Debug("starting mgsFooterOnFlag()")
	if flagVal {
		grpSettings.IncludeCustomFooter = "true"
	} else {
		grpSettings.IncludeCustomFooter = "false"
		grpSettings.ForceSendFields = append(grpSettings.ForceSendFields, "IncludeCustomFooter")
	}
	lg.Debug("finished mgsFooterOnFlag()")
}

func mgsFooterTextFlag(grpSettings *gset.Groups, flagVal string) {
	lg.Debug("starting mgsFooterTextFlag()")
	if flagVal == "" {
		grpSettings.ForceSendFields = append(grpSettings.ForceSendFields, "CustomFooterText")
	}
	grpSettings.CustomFooterText = flagVal
	lg.Debug("finished mgsFooterTextFlag()")
}

func mgsGalFlag(grpSettings *gset.Groups, flagVal bool) {
	lg.Debug("starting mgsGroupDirFlag()")
	if flagVal {
		grpSettings.IncludeInGlobalAddressList = "true"
	} else {
		grpSettings.IncludeInGlobalAddressList = "false"
		grpSettings.ForceSendFields = append(grpSettings.ForceSendFields, "IncludeInGlobalAddressList")
	}
	lg.Debug("finished mgsGroupDirFlag()")
}

func mgsJoinFlag(grpSettings *gset.Groups, flagName string, flagVal string) error {
	lg.Debug("starting mgsJoinFlag()")
	validTxt, err := grpset.ValidateGroupSettingValue(grpset.JoinMap, flagName, flagVal)
	if err != nil {
		return err
	}
	grpSettings.WhoCanJoin = validTxt
	lg.Debug("finished mgsJoinFlag()")
	return nil
}

func mgsLanguageFlag(grpSettings *gset.Groups, flagName string, flagVal string) error {
	lg.Debug("starting mgsLanguageFlag()")
	validTxt, err := grpset.ValidateGroupSettingValue(grpset.LanguageMap, flagName, flagVal)
	if err != nil {
		return err
	}
	grpSettings.PrimaryLanguage = validTxt
	lg.Debug("finished mgsLanguageFlag()")
	return nil
}

func mgsLeaveFlag(grpSettings *gset.Groups, flagName string, flagVal string) error {
	lg.Debug("starting mgsLeaveFlag()")
	validTxt, err := grpset.ValidateGroupSettingValue(grpset.LeaveMap, flagName, flagVal)
	if err != nil {
		return err
	}
	grpSettings.WhoCanLeaveGroup = validTxt
	lg.Debug("finished mgsLeaveFlag()")
	return nil
}

func mgsMessageModFlag(grpSettings *gset.Groups, flagName string, flagVal string) error {
	lg.Debug("starting mgsMessageModFlag()")
	validTxt, err := grpset.ValidateGroupSettingValue(grpset.MessageModMap, flagName, flagVal)
	if err != nil {
		return err
	}
	grpSettings.MessageModerationLevel = validTxt
	lg.Debug("finished mgsMessageModFlag()")
	return nil
}

func mgsModContentFlag(grpSettings *gset.Groups, flagName string, flagVal string) error {
	lg.Debug("starting mgsModContentFlag()")
	validTxt, err := grpset.ValidateGroupSettingValue(grpset.ModContentMap, flagName, flagVal)
	if err != nil {
		return err
	}
	grpSettings.WhoCanModerateContent = validTxt
	lg.Debug("finished mgsModContentFlag()")
	return nil
}

func mgsModMemberFlag(grpSettings *gset.Groups, flagName string, flagVal string) error {
	lg.Debug("starting mgsModMemberFlag()")
	validTxt, err := grpset.ValidateGroupSettingValue(grpset.ModMemberMap, flagName, flagVal)
	if err != nil {
		return err
	}
	grpSettings.WhoCanModerateMembers = validTxt
	lg.Debug("finished mgsModMemberFlag()")
	return nil
}

func mgsNotifyDenyFlag(grpSettings *gset.Groups, flagVal bool) {
	lg.Debug("starting mgsNotifyDenyFlag()")
	if flagVal {
		grpSettings.SendMessageDenyNotification = "true"
	} else {
		grpSettings.SendMessageDenyNotification = "false"
		grpSettings.ForceSendFields = append(grpSettings.ForceSendFields, "SendMessageDenyNotification")
	}
	lg.Debug("finished mgsNotifyDenyFlag()")
}

func mgsPostAsGroupFlag(grpSettings *gset.Groups, flagVal bool) {
	lg.Debug("starting mgsPostAsGroupFlag()")
	if flagVal {
		grpSettings.MembersCanPostAsTheGroup = "true"
	} else {
		grpSettings.MembersCanPostAsTheGroup = "false"
		grpSettings.ForceSendFields = append(grpSettings.ForceSendFields, "MembersCanPostAsTheGroup")
	}
	lg.Debug("finished mgsPostAsGroupFlag()")
}

func mgsPostMessageFlag(grpSettings *gset.Groups, flagName string, flagVal string) error {
	lg.Debug("starting mgsPostMessageFlag()")
	validTxt, err := grpset.ValidateGroupSettingValue(grpset.PostMessageMap, flagName, flagVal)
	if err != nil {
		return err
	}
	grpSettings.WhoCanPostMessage = validTxt
	lg.Debug("finished mgsPostMessageFlag()")
	return nil
}

func mgsRepliesOnTopFlag(grpSettings *gset.Groups, flagVal bool) {
	lg.Debug("starting mgsRepliesOnTopFlag()")
	if flagVal {
		grpSettings.FavoriteRepliesOnTop = "true"
	} else {
		grpSettings.FavoriteRepliesOnTop = "false"
		grpSettings.ForceSendFields = append(grpSettings.ForceSendFields, "FavoriteRepliesOnTop")
	}
	lg.Debug("finished mgsRepliesOnTopFlag()")
}

func mgsReplyEmailFlag(grpSettings *gset.Groups, flagVal string) error {
	lg.Debug("starting mgsReplyEmailFlag()")
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
	lg.Debug("finished mgsReplyEmailFlag()")
	return nil
}

func mgsReplyToFlag(grpSettings *gset.Groups, flagName string, flagVal string) error {
	lg.Debug("starting mgsReplyToFlag()")
	validTxt, err := grpset.ValidateGroupSettingValue(grpset.ReplyToMap, flagName, flagVal)
	if err != nil {
		return err
	}
	grpSettings.ReplyTo = validTxt
	lg.Debug("finished mgsReplyToFlag()")
	return nil
}

func mgsSpamModFlag(grpSettings *gset.Groups, flagName string, flagVal string) error {
	lg.Debug("starting mgsSpamModFlag()")
	validTxt, err := grpset.ValidateGroupSettingValue(grpset.SpamModMap, flagName, flagVal)
	if err != nil {
		return err
	}
	grpSettings.SpamModerationLevel = validTxt
	lg.Debug("finished mgsSpamModFlag()")
	return nil
}

func mgsViewGroupFlag(grpSettings *gset.Groups, flagName string, flagVal string) error {
	lg.Debug("starting mgsViewGroupFlag()")
	validTxt, err := grpset.ValidateGroupSettingValue(grpset.ViewGroupMap, flagName, flagVal)
	if err != nil {
		return err
	}
	grpSettings.WhoCanViewGroup = validTxt
	lg.Debug("finished mgsViewGroupFlag()")
	return nil
}

func mgsViewMembershipFlag(grpSettings *gset.Groups, flagName string, flagVal string) error {
	lg.Debug("starting mgsViewMembershipFlag()")
	validTxt, err := grpset.ValidateGroupSettingValue(grpset.ViewMembershipMap, flagName, flagVal)
	if err != nil {
		return err
	}
	grpSettings.WhoCanViewMembership = validTxt
	lg.Debug("finished mgsViewMembershipFlag()")
	return nil
}

func mgsWebPostingFlag(grpSettings *gset.Groups, flagVal bool) {
	lg.Debug("starting mgsWebPostingFlag()")
	if flagVal {
		grpSettings.AllowWebPosting = "true"
	} else {
		grpSettings.AllowWebPosting = "false"
		grpSettings.ForceSendFields = append(grpSettings.ForceSendFields, "AllowWebPosting")
	}
	lg.Debug("finished mgsWebPostingFlag()")
}
