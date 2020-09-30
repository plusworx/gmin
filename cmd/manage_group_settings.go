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
	grpset "github.com/plusworx/gmin/utils/groupsettings"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	gset "google.golang.org/api/groupssettings/v1"
)

var manageGroupSettingsCmd = &cobra.Command{
	Use:     "group-settings <group email address>",
	Aliases: []string{"grp-settings", "grp-set", "gsettings", "gset"},
	Args:    cobra.ExactArgs(1),
	Example: `gmin manage group-settings sales@mycompany.org
gmin mng gset finance@mycompany.org`,
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

	logger.Infof(cmn.InfoGroupSettingsChanged, newSettings.Email)
	fmt.Println(cmn.GminMessage(fmt.Sprintf(cmn.InfoGroupSettingsChanged, newSettings.Email)))

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
			err := approveMemberFlag(grpSettings, "--approve-member")
			if err != nil {
				return err
			}
			return nil
		}
		if flName == "archived" {
			archivedFlag(grpSettings)
			return nil
		}
		if flName == "archive-only" {
			archiveOnlyFlag(grpSettings)
			return nil
		}
		if flName == "assist-content" {
			err := assistContentFlag(grpSettings, "--assist-content")
			if err != nil {
				return err
			}
			return nil
		}
		if flName == "collab-inbox" {
			collabInboxFlag(grpSettings)
			return nil
		}
		if flName == "contact-owner" {
			err := contactOwnerFlag(grpSettings, "--contact-owner")
			if err != nil {
				return err
			}
			return nil
		}
		if flName == "deny-text" {
			denyTextFlag(grpSettings)
			return nil
		}
		if flName == "discover-group" {
			err := discoverGroupFlag(grpSettings, "--discover-group")
			if err != nil {
				return err
			}
			return nil
		}
		if flName == "ext-member" {
			extMemberFlag(grpSettings)
			return nil
		}
		if flName == "footer-on" {
			footerOnFlag(grpSettings)
			return nil
		}
		if flName == "footer-text" {
			footerTextFlag(grpSettings)
			return nil
		}
		if flName == "join" {
			err := joinFlag(grpSettings, "--join")
			if err != nil {
				return err
			}
			return nil
		}
		if flName == "language" {
			err := languageFlag(grpSettings, "--language")
			if err != nil {
				return err
			}
			return nil
		}
		if flName == "leave" {
			err := leaveFlag(grpSettings, "--leave")
			if err != nil {
				return err
			}
			return nil
		}
		if flName == "message-mod" {
			err := messageModFlag(grpSettings, "--message-mod")
			if err != nil {
				return err
			}
			return nil
		}
		if flName == "mod-content" {
			err := modContentFlag(grpSettings, "--mod-content")
			if err != nil {
				return err
			}
			return nil
		}
		if flName == "mod-member" {
			err := modMemberFlag(grpSettings, "--mod-member")
			if err != nil {
				return err
			}
			return nil
		}
		if flName == "notify-deny" {
			notifyDenyFlag(grpSettings)
			return nil
		}
		if flName == "post-as-group" {
			postAsGroupFlag(grpSettings)
			return nil
		}
		if flName == "post-message" {
			err := postMessageFlag(grpSettings, "--post-message")
			if err != nil {
				return err
			}
			return nil
		}
		if flName == "replies-on-top" {
			repliesOnTopFlag(grpSettings)
			return nil
		}
		if flName == "reply-email" {
			err := replyEmailFlag(grpSettings)
			if err != nil {
				return err
			}
			return nil
		}
		if flName == "reply-to" {
			err := replyToFlag(grpSettings, "--reply-to")
			if err != nil {
				return err
			}
			return nil
		}
		if flName == "spam-mod" {
			err := spamModFlag(grpSettings, "--spam-mod")
			if err != nil {
				return err
			}
			return nil
		}
		if flName == "view-group" {
			err := viewGroupFlag(grpSettings, "--view-group")
			if err != nil {
				return err
			}
			return nil
		}
		if flName == "view-membership" {
			err := viewMembershipFlag(grpSettings, "--view-membership")
			if err != nil {
				return err
			}
			return nil
		}
		if flName == "web-posting" {
			webPostingFlag(grpSettings)
			return nil
		}
	}
	logger.Debug("finished processMngGrpSettingFlags()")
	return nil
}

// Process command flag input

func approveMemberFlag(grpSettings *gset.Groups, flagName string) error {
	logger.Debug("starting approveMemberFlag()")
	validTxt, err := grpset.ValidateGroupSettingFlag(grpset.ApproveMemberMap, flagName, approveMems)
	if err != nil {
		return err
	}
	grpSettings.WhoCanApproveMembers = validTxt
	logger.Debug("finished approveMemberFlag()")
	return nil
}

func archivedFlag(grpSettings *gset.Groups) {
	logger.Debug("starting archivedFlag()")
	if isArchived {
		grpSettings.IsArchived = "true"
	} else {
		grpSettings.IsArchived = "false"
		grpSettings.ForceSendFields = append(grpSettings.ForceSendFields, "IsArchived")
	}
	logger.Debug("finished archivedFlag()")
}

func archiveOnlyFlag(grpSettings *gset.Groups) {
	logger.Debug("starting archiveOnlyFlag()")
	if archiveOnly {
		grpSettings.ArchiveOnly = "true"
	} else {
		grpSettings.ArchiveOnly = "false"
		grpSettings.ForceSendFields = append(grpSettings.ForceSendFields, "ArchiveOnly")
	}
	logger.Debug("finished archiveOnlyFlag()")
}

func assistContentFlag(grpSettings *gset.Groups, flagName string) error {
	logger.Debug("starting assistContentFlag()")
	validTxt, err := grpset.ValidateGroupSettingFlag(grpset.AssistContentMap, flagName, assistContent)
	if err != nil {
		return err
	}
	grpSettings.WhoCanAssistContent = validTxt
	logger.Debug("finished assistContentFlag()")
	return nil
}

func collabInboxFlag(grpSettings *gset.Groups) {
	logger.Debug("starting collabInboxFlag()")
	if collabInbox {
		grpSettings.EnableCollaborativeInbox = "true"
	} else {
		grpSettings.EnableCollaborativeInbox = "false"
		grpSettings.ForceSendFields = append(grpSettings.ForceSendFields, "EnableCollaborativeInbox")
	}
	logger.Debug("finished collabInboxFlag()")
}

func contactOwnerFlag(grpSettings *gset.Groups, flagName string) error {
	logger.Debug("starting contactOwnerFlag()")
	validTxt, err := grpset.ValidateGroupSettingFlag(grpset.ContactOwnerMap, flagName, contactOwner)
	if err != nil {
		return err
	}
	grpSettings.WhoCanContactOwner = validTxt
	logger.Debug("finished contactOwnerFlag()")
	return nil
}

func denyTextFlag(grpSettings *gset.Groups) {
	logger.Debug("starting denyTextFlag()")
	if denyText == "" {
		grpSettings.ForceSendFields = append(grpSettings.ForceSendFields, "DefaultMessageDenyNotificationText")
	}
	grpSettings.DefaultMessageDenyNotificationText = denyText
	logger.Debug("finished denyTextFlag()")
}

func discoverGroupFlag(grpSettings *gset.Groups, flagName string) error {
	logger.Debug("starting discoverGroupFlag()")
	validTxt, err := grpset.ValidateGroupSettingFlag(grpset.DiscoverGroupMap, flagName, discoverGroup)
	if err != nil {
		return err
	}
	grpSettings.WhoCanDiscoverGroup = validTxt
	logger.Debug("finished discoverGroupFlag()")
	return nil
}

func extMemberFlag(grpSettings *gset.Groups) {
	logger.Debug("starting extMemberFlag()")
	if extMems {
		grpSettings.AllowExternalMembers = "true"
	} else {
		grpSettings.AllowExternalMembers = "false"
		grpSettings.ForceSendFields = append(grpSettings.ForceSendFields, "AllowExternalMembers")
	}
	logger.Debug("finished extMemberFlag()")
}

func footerOnFlag(grpSettings *gset.Groups) {
	logger.Debug("starting footerOnFlag()")
	if incFooter {
		grpSettings.IncludeCustomFooter = "true"
	} else {
		grpSettings.IncludeCustomFooter = "false"
		grpSettings.ForceSendFields = append(grpSettings.ForceSendFields, "IncludeCustomFooter")
	}
	logger.Debug("finished footerOnFlag()")
}

func footerTextFlag(grpSettings *gset.Groups) {
	logger.Debug("starting footerTextFlag()")
	if footerText == "" {
		grpSettings.ForceSendFields = append(grpSettings.ForceSendFields, "CustomFooterText")
	}
	grpSettings.CustomFooterText = footerText
	logger.Debug("finished footerTextFlag()")
}

func joinFlag(grpSettings *gset.Groups, flagName string) error {
	logger.Debug("starting joinFlag()")
	validTxt, err := grpset.ValidateGroupSettingFlag(grpset.JoinMap, flagName, join)
	if err != nil {
		return err
	}
	grpSettings.WhoCanJoin = validTxt
	logger.Debug("finished joinFlag()")
	return nil
}

func languageFlag(grpSettings *gset.Groups, flagName string) error {
	logger.Debug("starting languageFlag()")
	validTxt, err := grpset.ValidateGroupSettingFlag(grpset.LanguageMap, flagName, language)
	if err != nil {
		return err
	}
	grpSettings.PrimaryLanguage = validTxt
	logger.Debug("finished languageFlag()")
	return nil
}

func leaveFlag(grpSettings *gset.Groups, flagName string) error {
	logger.Debug("starting leaveFlag()")
	validTxt, err := grpset.ValidateGroupSettingFlag(grpset.LeaveMap, flagName, leave)
	if err != nil {
		return err
	}
	grpSettings.WhoCanLeaveGroup = validTxt
	logger.Debug("finished leaveFlag()")
	return nil
}

func messageModFlag(grpSettings *gset.Groups, flagName string) error {
	logger.Debug("starting messageModFlag()")
	validTxt, err := grpset.ValidateGroupSettingFlag(grpset.MessageModMap, flagName, messageMod)
	if err != nil {
		return err
	}
	grpSettings.MessageModerationLevel = validTxt
	logger.Debug("finished messageModFlag()")
	return nil
}

func modContentFlag(grpSettings *gset.Groups, flagName string) error {
	logger.Debug("starting modContentFlag()")
	validTxt, err := grpset.ValidateGroupSettingFlag(grpset.ModContentMap, flagName, modContent)
	if err != nil {
		return err
	}
	grpSettings.WhoCanModerateContent = validTxt
	logger.Debug("finished modContentFlag()")
	return nil
}

func modMemberFlag(grpSettings *gset.Groups, flagName string) error {
	logger.Debug("starting modMemberFlag()")
	validTxt, err := grpset.ValidateGroupSettingFlag(grpset.ModMemberMap, flagName, modMems)
	if err != nil {
		return err
	}
	grpSettings.WhoCanModerateMembers = validTxt
	logger.Debug("finished modMemberFlag()")
	return nil
}

func notifyDenyFlag(grpSettings *gset.Groups) {
	logger.Debug("starting notifyDenyFlag()")
	if denyNotification {
		grpSettings.SendMessageDenyNotification = "true"
	} else {
		grpSettings.SendMessageDenyNotification = "false"
		grpSettings.ForceSendFields = append(grpSettings.ForceSendFields, "SendMessageDenyNotification")
	}
	logger.Debug("finished notifyDenyFlag()")
}

func postAsGroupFlag(grpSettings *gset.Groups) {
	logger.Debug("starting postAsGroupFlag()")
	if postAsGroup {
		grpSettings.MembersCanPostAsTheGroup = "true"
	} else {
		grpSettings.MembersCanPostAsTheGroup = "false"
		grpSettings.ForceSendFields = append(grpSettings.ForceSendFields, "MembersCanPostAsTheGroup")
	}
	logger.Debug("finished postAsGroupFlag()")
}

func postMessageFlag(grpSettings *gset.Groups, flagName string) error {
	logger.Debug("starting postMessageFlag()")
	validTxt, err := grpset.ValidateGroupSettingFlag(grpset.PostMessageMap, flagName, postMessage)
	if err != nil {
		return err
	}
	grpSettings.WhoCanPostMessage = validTxt
	logger.Debug("finished postMessageFlag()")
	return nil
}

func repliesOnTopFlag(grpSettings *gset.Groups) {
	logger.Debug("starting repliesOnTopFlag()")
	if repliesOnTop {
		grpSettings.FavoriteRepliesOnTop = "true"
	} else {
		grpSettings.FavoriteRepliesOnTop = "false"
		grpSettings.ForceSendFields = append(grpSettings.ForceSendFields, "FavoriteRepliesOnTop")
	}
	logger.Debug("finished repliesOnTopFlag()")
}

func replyEmailFlag(grpSettings *gset.Groups) error {
	logger.Debug("starting replyEmailFlag()")
	ok := valid.IsEmail(replyEmail)
	if !ok {
		err := fmt.Errorf(cmn.ErrInvalidEmailAddress, replyEmail)
		return err
	}
	if replyEmail == "" {
		grpSettings.ForceSendFields = append(grpSettings.ForceSendFields, "CustomReplyTo")
	}
	grpSettings.CustomReplyTo = replyEmail
	logger.Debug("finished replyEmailFlag()")
	return nil
}

func replyToFlag(grpSettings *gset.Groups, flagName string) error {
	logger.Debug("starting replyToFlag()")
	validTxt, err := grpset.ValidateGroupSettingFlag(grpset.ReplyToMap, flagName, replyTo)
	if err != nil {
		return err
	}
	grpSettings.ReplyTo = validTxt
	logger.Debug("finished replyToFlag()")
	return nil
}

func spamModFlag(grpSettings *gset.Groups, flagName string) error {
	logger.Debug("starting spamModFlag()")
	validTxt, err := grpset.ValidateGroupSettingFlag(grpset.SpamModMap, flagName, spamMod)
	if err != nil {
		return err
	}
	grpSettings.ReplyTo = validTxt
	logger.Debug("finished spamModFlag()")
	return nil
}

func viewGroupFlag(grpSettings *gset.Groups, flagName string) error {
	logger.Debug("starting viewGroupFlag()")
	validTxt, err := grpset.ValidateGroupSettingFlag(grpset.ViewGroupMap, flagName, viewGroup)
	if err != nil {
		return err
	}
	grpSettings.WhoCanViewGroup = validTxt
	logger.Debug("finished viewGroupFlag()")
	return nil
}

func viewMembershipFlag(grpSettings *gset.Groups, flagName string) error {
	logger.Debug("starting viewMembershipFlag()")
	validTxt, err := grpset.ValidateGroupSettingFlag(grpset.ViewMembershipMap, flagName, viewMems)
	if err != nil {
		return err
	}
	grpSettings.WhoCanViewMembership = validTxt
	logger.Debug("finished viewMembershipFlag()")
	return nil
}

func webPostingFlag(grpSettings *gset.Groups) {
	logger.Debug("starting webPostingFlag()")
	if webPosting {
		grpSettings.AllowWebPosting = "true"
	} else {
		grpSettings.AllowWebPosting = "false"
		grpSettings.ForceSendFields = append(grpSettings.ForceSendFields, "AllowWebPosting")
	}
	logger.Debug("finished webPostingFlag()")
}
