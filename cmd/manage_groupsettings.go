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
	gmess "github.com/plusworx/gmin/utils/gminmessages"
	grpset "github.com/plusworx/gmin/utils/groupsettings"
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
	logger.Debugw("starting doManageGroupSettings()",
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
		logger.Error(err)
		return err
	}

	grpSettings.Email = args[0]

	gss, err := cmn.CreateGroupSettingService(gset.AppsGroupsSettingsScope)
	if err != nil {
		logger.Error(err)
		return err
	}

	gsuc := gss.Groups.Update(args[0], grpSettings)
	newSettings, err := gsuc.Do()
	if err != nil {
		logger.Error(err)
		return err
	}

	logger.Infof(gmess.InfoGroupSettingsChanged, newSettings.Email)
	fmt.Println(cmn.GminMessage(fmt.Sprintf(gmess.InfoGroupSettingsChanged, newSettings.Email)))

	logger.Debug("finished doManageGroupSettings()")
	return nil
}

func init() {
	manageCmd.AddCommand(manageGroupSettingsCmd)

	manageGroupSettingsCmd.Flags().StringVar(&approveMems, "approve-member", "", "who can approve members")
	manageGroupSettingsCmd.Flags().BoolVarP(&isArchived, "archived", "r", false, "is archived")
	manageGroupSettingsCmd.Flags().BoolVarP(&archiveOnly, "archive-only", "a", false, "archive only")
	manageGroupSettingsCmd.Flags().StringVar(&assistContent, "assist-content", "", "who can moderate metadata")
	manageGroupSettingsCmd.Flags().StringVar(&banUser, "ban-user", "", "who can ban user")
	manageGroupSettingsCmd.Flags().BoolVarP(&collabInbox, "collab-inbox", "c", false, "enable collaborative inbox")
	manageGroupSettingsCmd.Flags().StringVar(&contactOwner, "contact-owner", "", "who can contact owner")
	manageGroupSettingsCmd.Flags().StringVar(&denyText, "deny-text", "", "default message deny notification text")
	manageGroupSettingsCmd.Flags().StringVar(&discoverGroup, "discover-group", "", "who can discover group")
	manageGroupSettingsCmd.Flags().BoolVarP(&extMems, "ext-member", "e", false, "allow external members")
	manageGroupSettingsCmd.Flags().BoolVarP(&incFooter, "footer-on", "f", false, "include custom footer")
	manageGroupSettingsCmd.Flags().StringVar(&footerText, "footer-text", "", "custom footer text")
	manageGroupSettingsCmd.Flags().BoolVarP(&gal, "gal", "g", false, "include in Global Address List")
	manageGroupSettingsCmd.Flags().StringVar(&join, "join", "", "who can join group")
	manageGroupSettingsCmd.Flags().StringVar(&language, "language", "", "primary language")
	manageGroupSettingsCmd.Flags().StringVar(&leave, "leave", "", "who can leave group")
	manageGroupSettingsCmd.Flags().StringVar(&messageMod, "message-mod", "", "message moderation level")
	manageGroupSettingsCmd.Flags().StringVar(&modContent, "mod-content", "", "who can moderate content")
	manageGroupSettingsCmd.Flags().StringVar(&modMems, "mod-member", "", "who can moderate members")
	manageGroupSettingsCmd.Flags().BoolVarP(&denyNotification, "notify-deny", "n", false, "send message deny notification")
	manageGroupSettingsCmd.Flags().BoolVarP(&postAsGroup, "post-as-group", "p", false, "members can post as group")
	manageGroupSettingsCmd.Flags().StringVar(&postMessage, "post-message", "", "who can post messages")
	manageGroupSettingsCmd.Flags().BoolVarP(&repliesOnTop, "replies-on-top", "t", false, "favourite replies on top")
	manageGroupSettingsCmd.Flags().StringVar(&replyEmail, "reply-email", "", "custom reply to email address")
	manageGroupSettingsCmd.Flags().StringVar(&replyTo, "reply-to", "", "who receives the default reply")
	manageGroupSettingsCmd.Flags().StringVar(&spamMod, "spam-mod", "", "spam moderation level")
	manageGroupSettingsCmd.Flags().StringVar(&viewGroup, "view-group", "", "who can view group")
	manageGroupSettingsCmd.Flags().StringVar(&viewMems, "view-membership", "", "who can view membership")
	manageGroupSettingsCmd.Flags().BoolVarP(&webPosting, "web-posting", "w", false, "allow web posting")
}

func processMngGrpSettingFlags(cmd *cobra.Command, grpSettings *gset.Groups, flagNames []string) error {
	logger.Debugw("starting processMngGrpSettingFlags()",
		"flagNames", flagNames)
	for _, flName := range flagNames {
		if flName == "approve-member" {
			err := mgsApproveMemberFlag(grpSettings, "--"+flName)
			if err != nil {
				return err
			}
			return nil
		}
		if flName == "archived" {
			mgsArchivedFlag(grpSettings)
			return nil
		}
		if flName == "archive-only" {
			mgsArchiveOnlyFlag(grpSettings)
			return nil
		}
		if flName == "assist-content" {
			err := mgsAssistContentFlag(grpSettings, "--"+flName)
			if err != nil {
				return err
			}
			return nil
		}
		if flName == "ban-user" {
			err := mgsBanUserFlag(grpSettings, "--"+flName)
			if err != nil {
				return err
			}
			return nil
		}
		if flName == "collab-inbox" {
			mgsCollabInboxFlag(grpSettings)
			return nil
		}
		if flName == "contact-owner" {
			err := mgsContactOwnerFlag(grpSettings, "--"+flName)
			if err != nil {
				return err
			}
			return nil
		}
		if flName == "deny-text" {
			mgsDenyTextFlag(grpSettings)
			return nil
		}
		if flName == "discover-group" {
			err := mgsDiscoverGroupFlag(grpSettings, "--"+flName)
			if err != nil {
				return err
			}
			return nil
		}
		if flName == "ext-member" {
			mgsExtMemberFlag(grpSettings)
			return nil
		}
		if flName == "footer-on" {
			mgsFooterOnFlag(grpSettings)
			return nil
		}
		if flName == "footer-text" {
			mgsFooterTextFlag(grpSettings)
			return nil
		}
		if flName == "join" {
			err := mgsJoinFlag(grpSettings, "--"+flName)
			if err != nil {
				return err
			}
			return nil
		}
		if flName == "language" {
			err := mgsLanguageFlag(grpSettings, "--"+flName)
			if err != nil {
				return err
			}
			return nil
		}
		if flName == "leave" {
			err := mgsLeaveFlag(grpSettings, "--"+flName)
			if err != nil {
				return err
			}
			return nil
		}
		if flName == "message-mod" {
			err := mgsMessageModFlag(grpSettings, "--"+flName)
			if err != nil {
				return err
			}
			return nil
		}
		if flName == "mod-content" {
			err := mgsModContentFlag(grpSettings, "--"+flName)
			if err != nil {
				return err
			}
			return nil
		}
		if flName == "mod-member" {
			err := mgsModMemberFlag(grpSettings, "--"+flName)
			if err != nil {
				return err
			}
			return nil
		}
		if flName == "notify-deny" {
			mgsNotifyDenyFlag(grpSettings)
			return nil
		}
		if flName == "post-as-group" {
			mgsPostAsGroupFlag(grpSettings)
			return nil
		}
		if flName == "post-message" {
			err := mgsPostMessageFlag(grpSettings, "--"+flName)
			if err != nil {
				return err
			}
			return nil
		}
		if flName == "replies-on-top" {
			mgsRepliesOnTopFlag(grpSettings)
			return nil
		}
		if flName == "reply-email" {
			err := mgsReplyEmailFlag(grpSettings)
			if err != nil {
				return err
			}
			return nil
		}
		if flName == "reply-to" {
			err := mgsReplyToFlag(grpSettings, "--"+flName)
			if err != nil {
				return err
			}
			return nil
		}
		if flName == "spam-mod" {
			err := mgsSpamModFlag(grpSettings, "--"+flName)
			if err != nil {
				return err
			}
			return nil
		}
		if flName == "view-group" {
			err := mgsViewGroupFlag(grpSettings, "--"+flName)
			if err != nil {
				return err
			}
			return nil
		}
		if flName == "view-membership" {
			err := mgsViewMembershipFlag(grpSettings, "--"+flName)
			if err != nil {
				return err
			}
			return nil
		}
		if flName == "web-posting" {
			mgsWebPostingFlag(grpSettings)
			return nil
		}
	}
	logger.Debug("finished processMngGrpSettingFlags()")
	return nil
}

// Process command flag input

func mgsApproveMemberFlag(grpSettings *gset.Groups, flagName string) error {
	logger.Debug("starting mgsApproveMemberFlag()")
	validTxt, err := grpset.ValidateGroupSettingValue(grpset.ApproveMemberMap, flagName, approveMems)
	if err != nil {
		return err
	}
	grpSettings.WhoCanApproveMembers = validTxt
	logger.Debug("finished mgsApproveMemberFlag()")
	return nil
}

func mgsArchivedFlag(grpSettings *gset.Groups) {
	logger.Debug("starting mgsArchivedFlag()")
	if isArchived {
		grpSettings.IsArchived = "true"
	} else {
		grpSettings.IsArchived = "false"
		grpSettings.ForceSendFields = append(grpSettings.ForceSendFields, "IsArchived")
	}
	logger.Debug("finished mgsArchivedFlag()")
}

func mgsArchiveOnlyFlag(grpSettings *gset.Groups) {
	logger.Debug("starting mgsArchiveOnlyFlag()")
	if archiveOnly {
		grpSettings.ArchiveOnly = "true"
	} else {
		grpSettings.ArchiveOnly = "false"
		grpSettings.ForceSendFields = append(grpSettings.ForceSendFields, "ArchiveOnly")
	}
	logger.Debug("finished mgsArchiveOnlyFlag()")
}

func mgsAssistContentFlag(grpSettings *gset.Groups, flagName string) error {
	logger.Debug("starting mgsAssistContentFlag()")
	validTxt, err := grpset.ValidateGroupSettingValue(grpset.AssistContentMap, flagName, assistContent)
	if err != nil {
		return err
	}
	grpSettings.WhoCanAssistContent = validTxt
	logger.Debug("finished mgsAssistContentFlag()")
	return nil
}

func mgsBanUserFlag(grpSettings *gset.Groups, flagName string) error {
	logger.Debug("starting mgsBanUserFlag()")
	validTxt, err := grpset.ValidateGroupSettingValue(grpset.BanUserMap, flagName, banUser)
	if err != nil {
		return err
	}
	grpSettings.WhoCanBanUsers = validTxt
	logger.Debug("finished mgsBanUserFlag()")
	return nil
}

func mgsCollabInboxFlag(grpSettings *gset.Groups) {
	logger.Debug("starting mgsCollabInboxFlag()")
	if collabInbox {
		grpSettings.EnableCollaborativeInbox = "true"
	} else {
		grpSettings.EnableCollaborativeInbox = "false"
		grpSettings.ForceSendFields = append(grpSettings.ForceSendFields, "EnableCollaborativeInbox")
	}
	logger.Debug("finished mgsCollabInboxFlag()")
}

func mgsContactOwnerFlag(grpSettings *gset.Groups, flagName string) error {
	logger.Debug("starting mgsContactOwnerFlag()")
	validTxt, err := grpset.ValidateGroupSettingValue(grpset.ContactOwnerMap, flagName, contactOwner)
	if err != nil {
		return err
	}
	grpSettings.WhoCanContactOwner = validTxt
	logger.Debug("finished mgsContactOwnerFlag()")
	return nil
}

func mgsDenyTextFlag(grpSettings *gset.Groups) {
	logger.Debug("starting mgsDenyTextFlag()")
	if denyText == "" {
		grpSettings.ForceSendFields = append(grpSettings.ForceSendFields, "DefaultMessageDenyNotificationText")
	}
	grpSettings.DefaultMessageDenyNotificationText = denyText
	logger.Debug("finished mgsDenyTextFlag()")
}

func mgsDiscoverGroupFlag(grpSettings *gset.Groups, flagName string) error {
	logger.Debug("starting mgsDiscoverGroupFlag()")
	validTxt, err := grpset.ValidateGroupSettingValue(grpset.DiscoverGroupMap, flagName, discoverGroup)
	if err != nil {
		return err
	}
	grpSettings.WhoCanDiscoverGroup = validTxt
	logger.Debug("finished mgsDiscoverGroupFlag()")
	return nil
}

func mgsExtMemberFlag(grpSettings *gset.Groups) {
	logger.Debug("starting mgsExtMemberFlag()")
	if extMems {
		grpSettings.AllowExternalMembers = "true"
	} else {
		grpSettings.AllowExternalMembers = "false"
		grpSettings.ForceSendFields = append(grpSettings.ForceSendFields, "AllowExternalMembers")
	}
	logger.Debug("finished mgsExtMemberFlag()")
}

func mgsFooterOnFlag(grpSettings *gset.Groups) {
	logger.Debug("starting mgsFooterOnFlag()")
	if incFooter {
		grpSettings.IncludeCustomFooter = "true"
	} else {
		grpSettings.IncludeCustomFooter = "false"
		grpSettings.ForceSendFields = append(grpSettings.ForceSendFields, "IncludeCustomFooter")
	}
	logger.Debug("finished mgsFooterOnFlag()")
}

func mgsFooterTextFlag(grpSettings *gset.Groups) {
	logger.Debug("starting mgsFooterTextFlag()")
	if footerText == "" {
		grpSettings.ForceSendFields = append(grpSettings.ForceSendFields, "CustomFooterText")
	}
	grpSettings.CustomFooterText = footerText
	logger.Debug("finished mgsFooterTextFlag()")
}

func mgsJoinFlag(grpSettings *gset.Groups, flagName string) error {
	logger.Debug("starting mgsJoinFlag()")
	validTxt, err := grpset.ValidateGroupSettingValue(grpset.JoinMap, flagName, join)
	if err != nil {
		return err
	}
	grpSettings.WhoCanJoin = validTxt
	logger.Debug("finished mgsJoinFlag()")
	return nil
}

func mgsLanguageFlag(grpSettings *gset.Groups, flagName string) error {
	logger.Debug("starting mgsLanguageFlag()")
	validTxt, err := grpset.ValidateGroupSettingValue(grpset.LanguageMap, flagName, language)
	if err != nil {
		return err
	}
	grpSettings.PrimaryLanguage = validTxt
	logger.Debug("finished mgsLanguageFlag()")
	return nil
}

func mgsLeaveFlag(grpSettings *gset.Groups, flagName string) error {
	logger.Debug("starting mgsLeaveFlag()")
	validTxt, err := grpset.ValidateGroupSettingValue(grpset.LeaveMap, flagName, leave)
	if err != nil {
		return err
	}
	grpSettings.WhoCanLeaveGroup = validTxt
	logger.Debug("finished mgsLeaveFlag()")
	return nil
}

func mgsMessageModFlag(grpSettings *gset.Groups, flagName string) error {
	logger.Debug("starting mgsMessageModFlag()")
	validTxt, err := grpset.ValidateGroupSettingValue(grpset.MessageModMap, flagName, messageMod)
	if err != nil {
		return err
	}
	grpSettings.MessageModerationLevel = validTxt
	logger.Debug("finished mgsMessageModFlag()")
	return nil
}

func mgsModContentFlag(grpSettings *gset.Groups, flagName string) error {
	logger.Debug("starting mgsModContentFlag()")
	validTxt, err := grpset.ValidateGroupSettingValue(grpset.ModContentMap, flagName, modContent)
	if err != nil {
		return err
	}
	grpSettings.WhoCanModerateContent = validTxt
	logger.Debug("finished mgsModContentFlag()")
	return nil
}

func mgsModMemberFlag(grpSettings *gset.Groups, flagName string) error {
	logger.Debug("starting mgsModMemberFlag()")
	validTxt, err := grpset.ValidateGroupSettingValue(grpset.ModMemberMap, flagName, modMems)
	if err != nil {
		return err
	}
	grpSettings.WhoCanModerateMembers = validTxt
	logger.Debug("finished mgsModMemberFlag()")
	return nil
}

func mgsNotifyDenyFlag(grpSettings *gset.Groups) {
	logger.Debug("starting mgsNotifyDenyFlag()")
	if denyNotification {
		grpSettings.SendMessageDenyNotification = "true"
	} else {
		grpSettings.SendMessageDenyNotification = "false"
		grpSettings.ForceSendFields = append(grpSettings.ForceSendFields, "SendMessageDenyNotification")
	}
	logger.Debug("finished mgsNotifyDenyFlag()")
}

func mgsPostAsGroupFlag(grpSettings *gset.Groups) {
	logger.Debug("starting mgsPostAsGroupFlag()")
	if postAsGroup {
		grpSettings.MembersCanPostAsTheGroup = "true"
	} else {
		grpSettings.MembersCanPostAsTheGroup = "false"
		grpSettings.ForceSendFields = append(grpSettings.ForceSendFields, "MembersCanPostAsTheGroup")
	}
	logger.Debug("finished mgsPostAsGroupFlag()")
}

func mgsPostMessageFlag(grpSettings *gset.Groups, flagName string) error {
	logger.Debug("starting mgsPostMessageFlag()")
	validTxt, err := grpset.ValidateGroupSettingValue(grpset.PostMessageMap, flagName, postMessage)
	if err != nil {
		return err
	}
	grpSettings.WhoCanPostMessage = validTxt
	logger.Debug("finished mgsPostMessageFlag()")
	return nil
}

func mgsRepliesOnTopFlag(grpSettings *gset.Groups) {
	logger.Debug("starting mgsRepliesOnTopFlag()")
	if repliesOnTop {
		grpSettings.FavoriteRepliesOnTop = "true"
	} else {
		grpSettings.FavoriteRepliesOnTop = "false"
		grpSettings.ForceSendFields = append(grpSettings.ForceSendFields, "FavoriteRepliesOnTop")
	}
	logger.Debug("finished mgsRepliesOnTopFlag()")
}

func mgsReplyEmailFlag(grpSettings *gset.Groups) error {
	logger.Debug("starting mgsReplyEmailFlag()")
	if replyEmail == "" {
		grpSettings.CustomReplyTo = replyEmail
		grpSettings.ForceSendFields = append(grpSettings.ForceSendFields, "CustomReplyTo")
		return nil
	}
	ok := valid.IsEmail(replyEmail)
	if !ok {
		err := fmt.Errorf(gmess.ErrInvalidEmailAddress, replyEmail)
		return err
	}
	grpSettings.CustomReplyTo = replyEmail
	logger.Debug("finished mgsReplyEmailFlag()")
	return nil
}

func mgsReplyToFlag(grpSettings *gset.Groups, flagName string) error {
	logger.Debug("starting mgsReplyToFlag()")
	validTxt, err := grpset.ValidateGroupSettingValue(grpset.ReplyToMap, flagName, replyTo)
	if err != nil {
		return err
	}
	grpSettings.ReplyTo = validTxt
	logger.Debug("finished mgsReplyToFlag()")
	return nil
}

func mgsSpamModFlag(grpSettings *gset.Groups, flagName string) error {
	logger.Debug("starting mgsSpamModFlag()")
	validTxt, err := grpset.ValidateGroupSettingValue(grpset.SpamModMap, flagName, spamMod)
	if err != nil {
		return err
	}
	grpSettings.SpamModerationLevel = validTxt
	logger.Debug("finished mgsSpamModFlag()")
	return nil
}

func mgsViewGroupFlag(grpSettings *gset.Groups, flagName string) error {
	logger.Debug("starting mgsViewGroupFlag()")
	validTxt, err := grpset.ValidateGroupSettingValue(grpset.ViewGroupMap, flagName, viewGroup)
	if err != nil {
		return err
	}
	grpSettings.WhoCanViewGroup = validTxt
	logger.Debug("finished mgsViewGroupFlag()")
	return nil
}

func mgsViewMembershipFlag(grpSettings *gset.Groups, flagName string) error {
	logger.Debug("starting mgsViewMembershipFlag()")
	validTxt, err := grpset.ValidateGroupSettingValue(grpset.ViewMembershipMap, flagName, viewMems)
	if err != nil {
		return err
	}
	grpSettings.WhoCanViewMembership = validTxt
	logger.Debug("finished mgsViewMembershipFlag()")
	return nil
}

func mgsWebPostingFlag(grpSettings *gset.Groups) {
	logger.Debug("starting mgsWebPostingFlag()")
	if webPosting {
		grpSettings.AllowWebPosting = "true"
	} else {
		grpSettings.AllowWebPosting = "false"
		grpSettings.ForceSendFields = append(grpSettings.ForceSendFields, "AllowWebPosting")
	}
	logger.Debug("finished mgsWebPostingFlag()")
}
