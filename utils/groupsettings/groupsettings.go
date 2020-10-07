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

package groupsettings

import (
	"fmt"
	"sort"
	"strings"

	cmn "github.com/plusworx/gmin/utils/common"
	"google.golang.org/api/googleapi"
	gset "google.golang.org/api/groupssettings/v1"
)

// ApproveMemberMap holds valid approve-mem flag values
var ApproveMemberMap = map[string]string{
	"all_managers_can_approve": "ALL_MANAGERS_CAN_APPROVE",
	"all_members_can_approve":  "ALL_MEMBERS_CAN_APPROVE",
	"all_owners_can_approve":   "ALL_OWNERS_CAN_APPROVE",
	"none_can_approve":         "NONE_CAN_APPROVE",
}

// AssistContentMap holds valid assist-content flag values
var AssistContentMap = map[string]string{
	"all_members":         "ALL_MEMBERS",
	"managers_only":       "MANAGERS_ONLY",
	"none":                "NONE",
	"owners_and_managers": "OWNERS_AND_MANAGERS",
	"owners_only":         "OWNERS_ONLY",
}

var attrValues = []string{
	"messageModerationLevel",
	"primaryLanguage",
	"replyTo",
	"spamModerationLevel",
	"whoCanApproveMembers",
	"whoCanAssistContent",
	"whoCanBanUsers",
	"whoCanContactOwner",
	"whoCanDiscoverGroup",
	"whoCanJoin",
	"whoCanLeaveGroup",
	"whoCanModerateContent",
	"whoCanModerateMembers",
	"whoCanPostMessage",
	"whoCanViewGroup",
	"whoCanViewMembership",
}

// BanUserMap holds valid ban-user flag values
var BanUserMap = map[string]string{
	"all_members":         "ALL_MEMBERS",
	"none":                "NONE",
	"owners_and_managers": "OWNERS_AND_MANAGERS",
	"owners_only":         "OWNERS_ONLY",
}

// ContactOwnerMap holds valid contact-owner flag values
var ContactOwnerMap = map[string]string{
	"all_in_domain_can_contact": "ALL_IN_DOMAIN_CAN_CONTACT",
	"all_managers_can_contact":  "ALL_MANAGERS_CAN_CONTACT",
	"all_members_can_contact":   "ALL_MEMBERS_CAN_CONTACT",
	"anyone_can_contact":        "ANYONE_CAN_CONTACT",
}

// DiscoverGroupMap holds valid discover-group flag values
var DiscoverGroupMap = map[string]string{
	"anyone_can_discover":        "ANYONE_CAN_DISCOVER",
	"all_in_domain_can_discover": "ALL_IN_DOMAIN_CAN_DISCOVER",
	"all_members_can_discover":   "ALL_MEMBERS_CAN_DISCOVER",
}

var flagValues = []string{
	"approve-member",
	"assist-content",
	"ban-user",
	"contact-owner",
	"discover-group",
	"join",
	"language",
	"leave",
	"message-mod",
	"mod-content",
	"mod-member",
	"post-message",
	"reply-to",
	"spam-mod",
	"view-group",
	"view-membership",
}

// GroupSettingsAttrMap provides lowercase mappings to valid gset.Groups attributes
var GroupSettingsAttrMap = map[string]string{
	"allowexternalmembers":                    "allowExternalMembers",
	"allowwebposting":                         "allowWebPosting",
	"archiveonly":                             "archiveOnly",
	"customfootertext":                        "customFooterText",
	"customreplyto":                           "customReplyTo",
	"customrolesenabledforsettingstobemerged": "customRolesEnabledForSettingsToBeMerged",
	"defaultmessagedenynotificationtext":      "defaultMessageDenyNotificationText",
	"description":                             "description",
	"email":                                   "email",
	"enablecollaborativeinbox":                "enableCollaborativeInbox",
	"favoriterepliesontop":                    "favoriteRepliesOnTop",
	"forcesendfields":                         "forceSendFields",
	"groupkey":                                "groupKey", // Used in batch management
	"includecustomfooter":                     "includeCustomFooter",
	"includeinglobaladdresslist":              "includeInGlobalAddressList",
	"isarchived":                              "isArchived",
	"kind":                                    "kind",
	"memberscanpostasthegroup":                "membersCanPostAsTheGroup",
	"messagemoderationlevel":                  "messageModerationLevel",
	"name":                                    "name",
	"primarylanguage":                         "primaryLanguage",
	"replyto":                                 "replyTo",
	"sendmessagedenynotification":             "sendMessageDenyNotification",
	"spammoderationlevel":                     "spamModerationLevel",
	"whocanapprovemembers":                    "whoCanApproveMembers",
	"whocanassistcontent":                     "whoCanAssistContent",
	"whocanbanusers":                          "whoCanBanUsers",
	"whocancontactowner":                      "whoCanContactOwner",
	"whocandiscovergroup":                     "whoCanDiscoverGroup",
	"whocanjoin":                              "whoCanJoin",
	"whocanleavegroup":                        "whoCanLeaveGroup",
	"whocanmoderatecontent":                   "whoCanModerateContent",
	"whocanmoderatemembers":                   "whoCanModerateMembers",
	"whocanpostmessage":                       "whoCanPostMessage",
	"whocanviewgroup":                         "whoCanViewGroup",
	"whocanviewmembership":                    "whoCanViewMembership",
}

// JoinMap holds valid join flag values
var JoinMap = map[string]string{
	"all_in_domain_can_join": "ALL_IN_DOMAIN_CAN_JOIN",
	"anyone_can_join":        "ANYONE_CAN_JOIN",
	"can_request_to_join":    "CAN_REQUEST_TO_JOIN",
	"invited_can_join":       "INVITED_CAN_JOIN",
}

// LanguageMap holds valid language flag values
var LanguageMap = map[string]string{
	"af":     "af",
	"az":     "az",
	"id":     "id",
	"ms":     "ms",
	"ca":     "ca",
	"cs":     "cs",
	"cy":     "cy",
	"da":     "da",
	"de":     "de",
	"et":     "et",
	"en-gb":  "en-GB",
	"en":     "en",
	"es":     "es",
	"es-419": "es-419",
	"eu":     "eu",
	"fil":    "fil",
	"fr":     "fr",
	"fr-ca":  "fr-CA",
	"ga":     "ga",
	"gl":     "gl",
	"hr":     "hr",
	"it":     "it",
	"zu":     "zu",
	"is":     "is",
	"sw":     "sw",
	"lv":     "lv",
	"lt":     "lt",
	"hu":     "hu",
	"no":     "no",
	"nl":     "nl",
	"pl":     "pl",
	"pt-br":  "pt-BR",
	"pt-pt":  "pt-PT",
	"ro":     "ro",
	"sk":     "sk",
	"sl":     "sl",
	"fi":     "fi",
	"sv":     "sv",
	"vi":     "vi",
	"tr":     "tr",
	"el":     "el",
	"bg":     "bg",
	"mn":     "mn",
	"ru":     "ru",
	"sr":     "sr",
	"uk":     "uk",
	"hy":     "hy",
	"he":     "he",
	"ur":     "ur",
	"ar":     "ar",
	"fa":     "fa",
	"ne":     "ne",
	"mr":     "mr",
	"hi":     "hi",
	"bn":     "bn",
	"gu":     "gu",
	"ta":     "ta",
	"te":     "te",
	"kn":     "kn",
	"ml":     "ml",
	"si":     "si",
	"th":     "th",
	"lo":     "lo",
	"my":     "my",
	"ka":     "ka",
	"am":     "am",
	"chr":    "chr",
	"km":     "km",
	"zh-hk":  "zh-HK",
	"zh-cn":  "zh-CN",
	"zh-tw":  "zh-TW",
	"ja":     "ja",
	"ko":     "ko",
}

// LeaveMap holds valid leave flag values
var LeaveMap = map[string]string{
	"all_managers_can_leave": "ALL_MANAGERS_CAN_LEAVE",
	"all_members_can_leave":  "ALL_MEMBERS_CAN_LEAVE",
	"none_can_leave":         "NONE_CAN_LEAVE",
}

// MessageModMap holds valid message-mod flag values
var MessageModMap = map[string]string{
	"moderate_all_messages": "MODERATE_ALL_MESSAGES",
	"moderate_new_members":  "MODERATE_NEW_MEMBERS",
	"moderate_non_members":  "MODERATE_NON_MEMBERS",
	"moderate_none":         "MODERATE_NONE",
}

// ModContentMap holds valid mod-content flag values
var ModContentMap = map[string]string{
	"all_members":         "ALL_MEMBERS",
	"none":                "NONE",
	"owners_and_managers": "OWNERS_AND_MANAGERS",
	"owners_only":         "OWNERS_ONLY",
}

// ModMemberMap holds valid mod-member flag values
var ModMemberMap = map[string]string{
	"all_members":         "ALL_MEMBERS",
	"none":                "NONE",
	"owners_and_managers": "OWNERS_AND_MANAGERS",
	"owners_only":         "OWNERS_ONLY",
}

// PostMessageMap holds valid post-message flag values
var PostMessageMap = map[string]string{
	"all_in_domain_can_post": "ALL_IN_DOMAIN_CAN_POST",
	"all_managers_can_post":  "ALL_MANAGERS_CAN_POST",
	"all_members_can_post":   "ALL_MEMBERS_CAN_POST",
	"all_owners_can_post":    "ALL_OWNERS_CAN_POST",
	"anyone_can_post":        "ANYONE_CAN_POST",
	"none_can_post":          "NONE_CAN_POST",
}

// ReplyToMap holds valid reply-to flag values
var ReplyToMap = map[string]string{
	"reply_to_custom":   "REPLY_TO_CUSTOM",
	"reply_to_ignore":   "REPLY_TO_IGNORE",
	"reply_to_list":     "REPLY_TO_LIST",
	"reply_to_managers": "REPLY_TO_MANAGERS",
	"reply_to_owner":    "REPLY_TO_OWNER",
	"reply_to_sender":   "REPLY_TO_SENDER",
}

// SpamModMap holds valid spam-mod flag values
var SpamModMap = map[string]string{
	"allow":             "ALLOW",
	"moderate":          "MODERATE",
	"reject":            "REJECT",
	"silently_moderate": "SILENTLY_MODERATE",
}

// ViewGroupMap holds valid view-group flag values
var ViewGroupMap = map[string]string{
	"all_in_domain_can_view": "ALL_IN_DOMAIN_CAN_VIEW",
	"all_managers_can_view":  "ALL_MANAGERS_CAN_VIEW",
	"all_members_can_view":   "ALL_MEMBERS_CAN_VIEW",
	"anyone_can_view":        "ANYONE_CAN_VIEW",
}

// ViewMembershipMap holds valid view-membership flag values
var ViewMembershipMap = map[string]string{
	"all_in_domain_can_view": "ALL_IN_DOMAIN_CAN_VIEW",
	"all_managers_can_view":  "ALL_MANAGERS_CAN_VIEW",
	"all_members_can_view":   "ALL_MEMBERS_CAN_VIEW",
}

// Key is struct used to extract groupKey
type Key struct {
	GroupKey string
}

// AddFields adds fields to be returned from admin calls
func AddFields(gsgc *gset.GroupsGetCall, attrs string) interface{} {
	var (
		fields  googleapi.Field = googleapi.Field(attrs)
		newGSGC *gset.GroupsGetCall
	)

	newGSGC = gsgc.Fields(fields)
	return newGSGC
}

// DoGet calls the .Do() function on the gset.GroupsGetCall
func DoGet(gsgc *gset.GroupsGetCall) (*gset.Groups, error) {
	groups, err := gsgc.Do()
	if err != nil {
		return nil, err
	}
	return groups, nil
}

func printFlagValues(flagMap map[string]string, filter string) {
	values := []string{}
	for _, value := range flagMap {
		values = append(values, value)
	}
	sort.Strings(values)
	for _, v := range values {
		if filter == "" {
			fmt.Println(v)
			continue
		}
		ok := strings.Contains(strings.ToLower(v), strings.ToLower(filter))
		if ok {
			fmt.Println(v)
		}
	}
}

// ShowAttrs displays requested group attributes
func ShowAttrs(filter string) {
	keys := make([]string, 0, len(GroupSettingsAttrMap))
	for k := range GroupSettingsAttrMap {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		if filter == "" {
			fmt.Println(GroupSettingsAttrMap[k])
			continue
		}
		if strings.Contains(k, strings.ToLower(filter)) {
			fmt.Println(GroupSettingsAttrMap[k])
		}
	}
}

// ShowAttrValues displays enumerated attribute values
func ShowAttrValues(lenArgs int, args []string, filter string) error {
	values := []string{}

	if lenArgs > 2 {
		return fmt.Errorf(cmn.ErrTooManyArgsMax1, args[0])
	}

	if lenArgs == 1 {
		cmn.ShowAttrVals(attrValues, filter)
	}

	if lenArgs == 2 {
		attr := strings.ToLower(args[1])
		if attr == "messagemoderationlevel" {
			for _, val := range MessageModMap {
				values = append(values, val)
			}
		}
		if attr == "primarylanguage" {
			for _, val := range LanguageMap {
				values = append(values, val)
			}
		}
		if attr == "replyto" {
			for _, val := range ReplyToMap {
				values = append(values, val)
			}
		}
		if attr == "spammoderationlevel" {
			for _, val := range SpamModMap {
				values = append(values, val)
			}
		}
		if attr == "whocanapprovemembers" {
			for _, val := range ApproveMemberMap {
				values = append(values, val)
			}
		}
		if attr == "whocanassistcontent" {
			for _, val := range AssistContentMap {
				values = append(values, val)
			}
		}
		if attr == "whocanbanusers" {
			for _, val := range BanUserMap {
				values = append(values, val)
			}
		}
		if attr == "whocancontactowner" {
			for _, val := range ContactOwnerMap {
				values = append(values, val)
			}
		}
		if attr == "whocandiscovergroup" {
			for _, val := range DiscoverGroupMap {
				values = append(values, val)
			}
		}
		if attr == "whocanjoin" {
			for _, val := range JoinMap {
				values = append(values, val)
			}
		}
		if attr == "whocanleavegroup" {
			for _, val := range LeaveMap {
				values = append(values, val)
			}
		}
		if attr == "whocanmoderatecontent" {
			for _, val := range ModContentMap {
				values = append(values, val)
			}
		}
		if attr == "whocanmoderatemembers" {
			for _, val := range ModMemberMap {
				values = append(values, val)
			}
		}
		if attr == "whocanpostmessage" {
			for _, val := range PostMessageMap {
				values = append(values, val)
			}
		}
		if attr == "whocanviewgroup" {
			for _, val := range ViewGroupMap {
				values = append(values, val)
			}
		}
		if attr == "whocanviewmembership" {
			for _, val := range ViewMembershipMap {
				values = append(values, val)
			}
		}
		if len(values) < 1 {
			return fmt.Errorf(cmn.ErrAttrNotRecognized, args[1])
		}
	}
	sort.Strings(values)
	cmn.ShowAttrVals(values, filter)
	return nil
}

// ShowFlagValues displays enumerated flag values
func ShowFlagValues(lenArgs int, args []string, filter string) error {
	if lenArgs == 1 {
		for _, v := range flagValues {
			if filter == "" {
				fmt.Println(v)
				continue
			}
			ok := strings.Contains(v, filter)
			if ok {
				fmt.Println(v)
			}
		}
	}

	if lenArgs == 2 {
		flag := strings.ToLower(args[1])

		if flag == "approve-member" {
			printFlagValues(ApproveMemberMap, filter)
			return nil
		}
		if flag == "assist-content" {
			printFlagValues(AssistContentMap, filter)
			return nil
		}
		if flag == "ban-user" {
			printFlagValues(BanUserMap, filter)
			return nil
		}
		if flag == "contact-owner" {
			printFlagValues(ContactOwnerMap, filter)
			return nil
		}
		if flag == "discover-group" {
			printFlagValues(DiscoverGroupMap, filter)
			return nil
		}
		if flag == "join" {
			printFlagValues(JoinMap, filter)
			return nil
		}
		if flag == "language" {
			printFlagValues(LanguageMap, filter)
			return nil
		}
		if flag == "leave" {
			printFlagValues(LeaveMap, filter)
			return nil
		}
		if flag == "message-mod" {
			printFlagValues(MessageModMap, filter)
			return nil
		}
		if flag == "mod-content" {
			printFlagValues(ModContentMap, filter)
			return nil
		}
		if flag == "mod-member" {
			printFlagValues(ModMemberMap, filter)
			return nil
		}
		if flag == "post-message" {
			printFlagValues(PostMessageMap, filter)
			return nil
		}
		if flag == "reply-to" {
			printFlagValues(ReplyToMap, filter)
			return nil
		}
		if flag == "spam-mod" {
			printFlagValues(SpamModMap, filter)
			return nil
		}
		if flag == "view-group" {
			printFlagValues(ViewGroupMap, filter)
			return nil
		}
		if flag == "view-membership" {
			printFlagValues(ViewMembershipMap, filter)
			return nil
		}
		// Flag not recognized
		return fmt.Errorf(cmn.ErrFlagNotRecognized, args[1])
	}

	return nil
}

// ValidateGroupSettingValue checks that a valid value has been provided for flag or attribute
func ValidateGroupSettingValue(valueMap map[string]string, name string, value string) (string, error) {
	lowerVal := strings.ToLower(value)
	validStr := valueMap[lowerVal]
	if validStr == "" {
		return "", fmt.Errorf(cmn.ErrInvalidString, name, value)
	}
	return validStr, nil
}
