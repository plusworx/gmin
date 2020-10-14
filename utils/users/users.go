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

package users

import (
	"fmt"
	"sort"
	"strings"

	cmn "github.com/plusworx/gmin/utils/common"
	gmess "github.com/plusworx/gmin/utils/gminmessages"
	admin "google.golang.org/api/admin/directory/v1"
	"google.golang.org/api/googleapi"
)

const (
	// ENDFIELD is List call attribute string terminator
	ENDFIELD = ")"
	// STARTUSERSFIELD is List call users attribute string prefix
	STARTUSERSFIELD = "users("
)

// addressAttrs contains names of all the addressable admin.UserAddress attributes
var addressAttrs = []string{
	"country",
	"countrycode",
	"customtype",
	"extendedaddress",
	"formatted",
	"locality",
	"pobox",
	"postalcode",
	"primary",
	"region",
	"streetaddress",
	"type",
}

var attrValues = []string{
	"address",
	"email",
	"externalId",
	"gender",
	"im",
	"keyword",
	"location",
	"notes",
	"organization",
	"phone",
	"posixAccount",
	"relation",
	"website",
}

// emailAttrs contains names of all the addressable admin.UserEmail attributes
var emailAttrs = []string{
	"address",
	"customtype",
	"primary",
	"type",
}

// extIDAttrs contains names of all the addressable admin.UserExternalId attributes
var extIDAttrs = []string{
	"customtype",
	"type",
	"value",
}

var flagValues = []string{
	"order-by",
	"projection",
	"sort-order",
	"view-type",
}

// genderAttrs contains names of all the addressable admin.UserGender attributes
var genderAttrs = []string{
	"addressmeas",
	"customgender",
	"type",
}

// imAttrs contains names of all the addressable admin.UserIm attributes
var imAttrs = []string{
	"customprotocol",
	"customtype",
	"im",
	"primary",
	"protocol",
	"type",
}

// keywordAttrs contains names of all the addressable admin.UserKeyword attributes
var keywordAttrs = []string{
	"customtype",
	"type",
	"value",
}

// languageAttrs contains names of all the addressable admin.UserLanguage attributes
var languageAttrs = []string{
	"customlanguage",
	"languagecode",
}

// locationAttrs contains names of all the addressable admin.UserLocation attributes
var locationAttrs = []string{
	"area",
	"buildingid",
	"customtype",
	"deskcode",
	"floorname",
	"floorsection",
	"type",
}

// nameAttrs contains names of all the addressable admin.UserName attributes
var nameAttrs = []string{
	"familyname",
	"fullname",
	"givenname",
}

// notesAttrs contains names of all the addressable admin.UserAbout attributes
var notesAttrs = []string{
	"contenttype",
	"value",
}

// organizationAttrs contains names of all the addressable admin.UserOrganization attributes
var organizationAttrs = []string{
	"costcenter",
	"customtype",
	"department",
	"description",
	"domain",
	"fulltimeequivalent",
	"location",
	"name",
	"primary",
	"symbol",
	"title",
	"type",
}

// phoneAttrs contains names of all the addressable admin.UserPhone attributes
var phoneAttrs = []string{
	"customtype",
	"primary",
	"type",
	"value",
}

// posAcctAttrs contains names of all the addressable admin.UserPosixAccount attributes
var posAcctAttrs = []string{
	"accountid",
	"gecos",
	"gid",
	"homedirectory",
	"operatingsystemtype",
	"primary",
	"shell",
	"systemid",
	"uid",
	"username",
}

// relationAttrs contains names of all the addressable admin.UserRelation attributes
var relationAttrs = []string{
	"customtype",
	"type",
	"value",
}

// sshPubKeyAttrs contains names of all the addressable admin.UserSshPublicKey attributes
var sshPubKeyAttrs = []string{
	"expirationtimeusec",
	"key",
}

// websiteAttrs contains names of all the addressable admin.UserWebsite attributes
var websiteAttrs = []string{
	"customtype",
	"primary",
	"type",
	"value",
}

// QueryAttrMap provides lowercase mappings to valid admin.User query attributes
var QueryAttrMap = map[string]string{
	"address":           "address",
	"addresspobox":      "addressPoBox",
	"addressextended":   "addressExtended",
	"addressstreet":     "addressStreet",
	"addresslocality":   "addressLocality",
	"addressregion":     "addressRegion",
	"addresspostalcode": "addressPostalCode",
	"addresscountry":    "addressCountry",
	"christianname":     "givenName",
	"directmanager":     "directManager",
	"directmanagerid":   "directManagerId",
	"email":             "email",
	"externalid":        "externalId",
	"familyname":        "familyName",
	"firstname":         "givenName",
	"givenname":         "givenName",
	"im":                "im",
	"isadmin":           "isAdmin",
	"isdelegatedadmin":  "isDelegatedAdmin",
	"isenrolledin2sv":   "isEnrolledIn2Sv",
	"isenforcedin2sv":   "isEnforcedIn2Sv",
	"issuspended":       "isSuspended",
	"lastname":          "familyName",
	"manager":           "manager",
	"managerid":         "managerId",
	"name":              "name",
	"orgcostcenter":     "orgCostCenter",
	"orgdepartment":     "orgDepartment",
	"orgdescription":    "orgDescription",
	"orgname":           "orgName",
	"orgtitle":          "orgTitle",
	"orgunitpath":       "orgUnitPath",
	"phone":             "phone",
	"surname":           "familyName",
}

// UserAttrMap provides lowercase mappings to valid admin.User attributes
var UserAttrMap = map[string]string{
	"accountid":                  "accountId",
	"address":                    "address",
	"addresses":                  "addresses",
	"addressmeas":                "addressMeAs",
	"agreedtoterms":              "agreedToTerms",
	"aliases":                    "aliases",
	"archived":                   "archived",
	"area":                       "area",
	"buildingid":                 "buildingId",
	"changepasswordatnextlogin":  "changePasswordAtNextLogin",
	"christianname":              "givenName",
	"contenttype":                "contentType",
	"costcenter":                 "costCenter",
	"country":                    "country",
	"countrycode":                "countryCode",
	"creationtime":               "creationTime",
	"customerid":                 "customerId",
	"customgender":               "customGender",
	"customlanguage":             "customLanguage",
	"customprotocol":             "customProtocol",
	"customschemas":              "customSchemas",
	"customtype":                 "customType",
	"deletiontime":               "deletionTime",
	"department":                 "department",
	"description":                "description",
	"deskcode":                   "deskCode",
	"domain":                     "domain",
	"emails":                     "emails",
	"etag":                       "etag",
	"expirationtimeusec":         "expirationTimeUsec",
	"externalids":                "externalIds",
	"extendedaddress":            "extendedAddress",
	"familyname":                 "familyName",
	"fingerprint":                "fingerprint",
	"firstname":                  "givenName",
	"floorname":                  "floorName",
	"floorsection":               "floorSection",
	"forcesendfields":            "forceSendFields",
	"formatted":                  "formatted",
	"fullname":                   "fullName",
	"fulltimeequivalent":         "fullTimeEquivalent",
	"gecos":                      "gecos",
	"gender":                     "gender",
	"gid":                        "gid",
	"givenname":                  "givenName",
	"hashfunction":               "hashFunction",
	"homedirectory":              "homeDirectory",
	"id":                         "id",
	"im":                         "im",
	"ims":                        "ims",
	"includeinglobaladdresslist": "includeInGlobalAddressList",
	"ipwhitelisted":              "ipWhitelisted",
	"isadmin":                    "isAdmin",
	"isdelegatedadmin":           "isDelegatedAdmin",
	"isenforcedin2sv":            "isEnforcedIn2Sv",
	"isenrolledin2sv":            "isEnrolledIn2Sv",
	"ismailboxsetup":             "isMailboxSetup",
	"key":                        "key",
	"keywords":                   "keywords",
	"kind":                       "kind",
	"languagecode":               "languageCode",
	"languages":                  "languages",
	"lastlogintime":              "lastLoginTime",
	"lastname":                   "familyName",
	"locality":                   "locality",
	"location":                   "location",
	"locations":                  "locations",
	"name":                       "name",
	"noneditablealiases":         "nonEditableAliases",
	"notes":                      "notes",
	"operatingsystemtype":        "operatingSystemType",
	"organisations":              "organizations",
	"organizations":              "organizations",
	"orgunitpath":                "orgUnitPath",
	"password":                   "password",
	"phones":                     "phones",
	"pobox":                      "poBox",
	"posixaccounts":              "posixAccounts",
	"postalcode":                 "postalCode",
	"primary":                    "primary",
	"primaryemail":               "primaryEmail",
	"protocol":                   "protocol",
	"recoveryemail":              "recoveryEmail",
	"recoveryphone":              "recoveryPhone",
	"region":                     "region",
	"relations":                  "relations",
	"shell":                      "shell",
	"sshpublickeys":              "sshPublicKeys",
	"streetaddress":              "streetAddress",
	"surname":                    "familyName",
	"suspended":                  "suspended",
	"suspensionreason":           "suspensionReason",
	"symbol":                     "symbol",
	"systemid":                   "systemId",
	"thumbnailphotoetag":         "thumbnailPhotoEtag",
	"thumbnailphotourl":          "thumbnailPhotoUrl",
	"title":                      "title",
	"type":                       "type",
	"uid":                        "uid",
	"userkey":                    "userKey", // Used in batch operations
	"username":                   "username",
	"value":                      "value",
	"websites":                   "websites",
}

var userAttrs = []string{
	"addresses",
	"agreedToTerms",
	"aliases",
	"archived",
	"changePasswordAtNextLogin",
	"creationTime",
	"customSchemas",
	"customerId",
	"deletionTime",
	"emails",
	"etag",
	"externalIds",
	"forceSendFields",
	"gender",
	"hashFunction",
	"id",
	"ims",
	"includeInGlobalAddressList",
	"ipWhitelisted",
	"isAdmin",
	"isDelegatedAdmin",
	"isEnforcedIn2Sv",
	"isEnrolledIn2Sv",
	"isMailboxSetup",
	"keywords",
	"kind",
	"languages",
	"lastLoginTime",
	"locations",
	"name",
	"nonEditableAliases",
	"notes",
	"organizations",
	"orgUnitPath",
	"password",
	"phones",
	"posixAccounts",
	"primaryEmail",
	"recoveryEmail",
	"recoveryPhone",
	"relations",
	"sshPublicKeys",
	"suspended",
	"suspensionReason",
	"thumbnailPhotoEtag",
	"thumbnailPhotoUrl",
	"userKey", // Used in batch update
	"websites",
}

var userCompAttrs = map[string]string{
	"addresses":     "address",
	"emails":        "email",
	"externalids":   "externalId",
	"gender":        "gender",
	"ims":           "im",
	"keywords":      "keyword",
	"languages":     "language",
	"locations":     "location",
	"name":          "name",
	"notes":         "notes",
	"organizations": "organization",
	"phones":        "phone",
	"posixaccounts": "posixAccount",
	"relations":     "relation",
	"sshpublickeys": "sshPublicKey",
	"websites":      "website",
}

var validAddressTypes = []string{
	"custom",
	"home",
	"other",
	"work",
}

var validEmailTypes = []string{
	"custom",
	"home",
	"other",
	"work",
}

var validExtIDTypes = []string{
	"account",
	"custom",
	"customer",
	"login_id",
	"network",
	"organization",
}

var validGenders = []string{
	"female",
	"male",
	"other",
	"unknown",
}

var validImProtocols = []string{
	"aim",
	"custom_protocol",
	"gtalk",
	"icq",
	"jabber",
	"msn",
	"net_meeting",
	"qq",
	"skype",
	"yahoo",
}

var validImTypes = []string{
	"custom",
	"home",
	"other",
	"work",
}

var validKeywordTypes = []string{
	"custom",
	"mission",
	"occupation",
	"outlook",
}

var validLocationTypes = []string{
	"custom",
	"default",
	"desk",
}

var validNotesContentTypes = []string{
	"text_html",
	"text_plain",
}

// ValidOrderByStrs provide valid strings to be used to set admin.UsersListCall OrderBy
var ValidOrderByStrs = []string{
	"email",
	"familyname",
	"firstname",
	"givenname",
	"lastname",
}

var validOrgTypes = []string{
	"custom",
	"domain_only",
	"school",
	"unknown",
	"work",
}

var validOSTypes = []string{
	"linux",
	"unspecified",
	"windows",
}

var validPhoneTypes = []string{
	"assistant",
	"callback",
	"car",
	"company_main",
	"custom",
	"grand_central",
	"home",
	"home_fax",
	"isdn",
	"main",
	"mobile",
	"other",
	"other_fax",
	"pager",
	"radio",
	"telex",
	"tty_tdd",
	"work",
	"work_fax",
	"work_mobile",
	"work_pager",
}

// ValidProjections provide valid strings to be used to set admin.UsersListCall Projection
var ValidProjections = []string{
	"basic",
	"custom",
	"full",
}

var validRelationTypes = []string{
	"admin_assistant",
	"assistant",
	"brother",
	"child",
	"custom",
	"domestic_partner",
	"dotted_line_manager",
	"exec_assistant",
	"father",
	"friend",
	"manager",
	"mother",
	"parent",
	"partner",
	"referred_by",
	"relative",
	"sister",
	"spouse",
}

// ValidViewTypes provide valid strings to be used to set admin.UsersListCall ViewType
var ValidViewTypes = []string{
	"admin_view",
	"domain_public",
}

var validWebsiteTypes = []string{
	"app_install_page",
	"blog",
	"custom",
	"ftp",
	"home",
	"home_page",
	"other",
	"profile",
	"reservations",
	"resume",
	"work",
}

// Key is struct used to extract userKey
type Key struct {
	UserKey string
}

// UndeleteUser is struct to extract undelete data
type UndeleteUser struct {
	UserKey     string
	OrgUnitPath string
}

// AddCustomer adds Customer to admin calls
func AddCustomer(ulc *admin.UsersListCall, customerID string) *admin.UsersListCall {
	var newULC *admin.UsersListCall

	newULC = ulc.Customer(customerID)

	return newULC
}

// AddCustomFieldMask adds CustomFieldMask to be used with get and list admin calls with custom projections
func AddCustomFieldMask(callObj interface{}, attrs string) interface{} {

	switch callObj.(type) {
	case *admin.UsersListCall:
		var newULC *admin.UsersListCall
		ulc := callObj.(*admin.UsersListCall)
		newULC = ulc.CustomFieldMask(attrs)

		return newULC
	case *admin.UsersGetCall:
		var newUGC *admin.UsersGetCall
		ugc := callObj.(*admin.UsersGetCall)
		newUGC = ugc.CustomFieldMask(attrs)

		return newUGC
	}

	return nil
}

// AddDomain adds domain to admin calls
func AddDomain(ulc *admin.UsersListCall, domain string) *admin.UsersListCall {
	var newULC *admin.UsersListCall

	newULC = ulc.Domain(domain)

	return newULC
}

// AddFields adds fields to be returned from admin calls
func AddFields(callObj interface{}, attrs string) interface{} {
	var fields googleapi.Field = googleapi.Field(attrs)

	switch callObj.(type) {
	case *admin.UsersListCall:
		var newULC *admin.UsersListCall
		ulc := callObj.(*admin.UsersListCall)
		newULC = ulc.Fields(fields)

		return newULC
	case *admin.UsersGetCall:
		var newUGC *admin.UsersGetCall
		ugc := callObj.(*admin.UsersGetCall)
		newUGC = ugc.Fields(fields)

		return newUGC
	}

	return nil
}

// AddMaxResults adds MaxResults to admin calls
func AddMaxResults(ulc *admin.UsersListCall, maxResults int64) *admin.UsersListCall {
	var newULC *admin.UsersListCall

	newULC = ulc.MaxResults(maxResults)

	return newULC
}

// AddOrderBy adds OrderBy to admin calls
func AddOrderBy(ulc *admin.UsersListCall, orderBy string) *admin.UsersListCall {
	var newULC *admin.UsersListCall

	newULC = ulc.OrderBy(orderBy)

	return newULC
}

// AddPageToken adds PageToken to admin calls
func AddPageToken(ulc *admin.UsersListCall, token string) *admin.UsersListCall {
	var newULC *admin.UsersListCall

	newULC = ulc.PageToken(token)

	return newULC
}

// AddProjection adds Projection to admin calls
func AddProjection(callObj interface{}, projection string) interface{} {
	switch callObj.(type) {
	case *admin.UsersListCall:
		var newULC *admin.UsersListCall
		ulc := callObj.(*admin.UsersListCall)
		newULC = ulc.Projection(projection)

		return newULC
	case *admin.UsersGetCall:
		var newUGC *admin.UsersGetCall
		ugc := callObj.(*admin.UsersGetCall)
		newUGC = ugc.Projection(projection)

		return newUGC
	}

	return nil
}

// AddQuery adds query to admin calls
func AddQuery(ulc *admin.UsersListCall, query string) *admin.UsersListCall {
	var newULC *admin.UsersListCall

	newULC = ulc.Query(query)

	return newULC
}

// AddShowDeleted adds ShowDeleted to admin calls
func AddShowDeleted(ulc *admin.UsersListCall) *admin.UsersListCall {
	var newULC *admin.UsersListCall

	newULC = ulc.ShowDeleted("true")

	return newULC
}

// AddSortOrder adds SortOrder to admin calls
func AddSortOrder(ulc *admin.UsersListCall, sortorder string) *admin.UsersListCall {
	var newULC *admin.UsersListCall

	newULC = ulc.SortOrder(sortorder)

	return newULC
}

// AddViewType adds ViewType to admin calls
func AddViewType(callObj interface{}, viewType string) interface{} {
	switch callObj.(type) {
	case *admin.UsersListCall:
		var newULC *admin.UsersListCall
		ulc := callObj.(*admin.UsersListCall)
		newULC = ulc.ViewType(viewType)

		return newULC
	case *admin.UsersGetCall:
		var newUGC *admin.UsersGetCall
		ugc := callObj.(*admin.UsersGetCall)
		newUGC = ugc.ViewType(viewType)

		return newUGC
	}

	return nil
}

// DoGet calls the .Do() function on the admin.UsersGetCall
func DoGet(ugc *admin.UsersGetCall) (*admin.User, error) {
	user, err := ugc.Do()
	if err != nil {
		return nil, err
	}

	return user, nil
}

// DoList calls the .Do() function on the admin.UsersListCall
func DoList(ulc *admin.UsersListCall) (*admin.Users, error) {
	users, err := ulc.Do()
	if err != nil {
		return nil, err
	}

	return users, nil
}

// ShowAttrs displays requested user attributes
func ShowAttrs(filter string) {
	for _, a := range userAttrs {
		lwrA := strings.ToLower(a)
		comp, _ := cmn.IsValidAttr(lwrA, userCompAttrs)
		if filter == "" {
			if comp != "" {
				fmt.Println("* ", a)
			} else {
				fmt.Println(a)
			}
			continue
		}

		if strings.Contains(lwrA, strings.ToLower(filter)) {
			if comp != "" {
				fmt.Println("* ", a)
			} else {
				fmt.Println(a)
			}
		}

	}
}

// ShowAttrValues displays enumerated attribute values
func ShowAttrValues(lenArgs int, args []string, filter string) error {
	if lenArgs == 1 {
		cmn.ShowAttrVals(attrValues, filter)
	}

	if lenArgs == 2 {
		attr := strings.ToLower(args[1])

		switch {
		case attr == "address" || attr == "email" || attr == "externalid" || attr == "gender" ||
			attr == "keyword" || attr == "location" || attr == "notes" || attr == "organization" || attr == "phone" ||
			attr == "relation" || attr == "website":
			fmt.Println("type")
		case attr == "im":
			fmt.Println("protocol")
			fmt.Println("type")
		case attr == "posixaccount":
			fmt.Println("operatingSystemType")
		default:
			return fmt.Errorf(gmess.ERR_ATTRNOTRECOGNIZED, args[1])
		}
	}

	if lenArgs == 3 {
		attr2 := strings.ToLower(args[1])
		attr3 := strings.ToLower(args[2])

		if attr2 == "address" {
			if attr3 == "type" {
				cmn.ShowAttrVals(validAddressTypes, filter)
				return nil
			}
			return fmt.Errorf(gmess.ERR_ATTRNOTRECOGNIZED, args[2])
		}
		if attr2 == "email" {
			if attr3 == "type" {
				cmn.ShowAttrVals(validEmailTypes, filter)
				return nil
			}
			return fmt.Errorf(gmess.ERR_ATTRNOTRECOGNIZED, args[2])
		}
		if attr2 == "externalid" {
			if attr3 == "type" {
				cmn.ShowAttrVals(validExtIDTypes, filter)
				return nil
			}
			return fmt.Errorf(gmess.ERR_ATTRNOTRECOGNIZED, args[2])
		}
		if attr2 == "gender" {
			if attr3 == "type" {
				cmn.ShowAttrVals(validGenders, filter)
				return nil
			}
			return fmt.Errorf(gmess.ERR_ATTRNOTRECOGNIZED, args[2])
		}
		if attr2 == "keyword" {
			if attr3 == "type" {
				cmn.ShowAttrVals(validKeywordTypes, filter)
				return nil
			}
			return fmt.Errorf(gmess.ERR_ATTRNOTRECOGNIZED, args[2])
		}
		if attr2 == "location" {
			if attr3 == "type" {
				cmn.ShowAttrVals(validLocationTypes, filter)
				return nil
			}
			return fmt.Errorf(gmess.ERR_ATTRNOTRECOGNIZED, args[2])
		}
		if attr2 == "notes" {
			if attr3 == "type" {
				cmn.ShowAttrVals(validNotesContentTypes, filter)
				return nil
			}
			return fmt.Errorf(gmess.ERR_ATTRNOTRECOGNIZED, args[2])
		}
		if attr2 == "organization" {
			if attr3 == "type" {
				cmn.ShowAttrVals(validOrgTypes, filter)
				return nil
			}
			return fmt.Errorf(gmess.ERR_ATTRNOTRECOGNIZED, args[2])
		}
		if attr2 == "phone" {
			if attr3 == "type" {
				cmn.ShowAttrVals(validPhoneTypes, filter)
				return nil
			}
			return fmt.Errorf(gmess.ERR_ATTRNOTRECOGNIZED, args[2])
		}
		if attr2 == "posixaccount" {
			if attr3 == "operatingsystemtype" {
				cmn.ShowAttrVals(validOSTypes, filter)
				return nil
			}
			return fmt.Errorf(gmess.ERR_ATTRNOTRECOGNIZED, args[2])
		}
		if attr2 == "relation" {
			if attr3 == "type" {
				cmn.ShowAttrVals(validRelationTypes, filter)
				return nil
			}
			return fmt.Errorf(gmess.ERR_ATTRNOTRECOGNIZED, args[2])
		}
		if attr2 == "website" {
			if attr3 == "type" {
				cmn.ShowAttrVals(validWebsiteTypes, filter)
				return nil
			}
			return fmt.Errorf(gmess.ERR_ATTRNOTRECOGNIZED, args[2])
		}
		if attr2 == "im" {
			if attr3 == "protocol" {
				cmn.ShowAttrVals(validImProtocols, filter)
				return nil
			}
			if attr3 == "type" {
				cmn.ShowAttrVals(validImTypes, filter)
				return nil
			}
			return fmt.Errorf(gmess.ERR_ATTRNOTRECOGNIZED, args[2])
		}
		// Attribute not recognized
		return fmt.Errorf(gmess.ERR_ATTRNOTRECOGNIZED, args[1])
	}

	return nil
}

// ShowCompAttrs displays user composite attributes
func ShowCompAttrs(filter string) {
	keys := make([]string, 0, len(userCompAttrs))
	for k := range userCompAttrs {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		if filter == "" {
			fmt.Println(userCompAttrs[k])
			continue
		}

		if strings.Contains(k, strings.ToLower(filter)) {
			fmt.Println(userCompAttrs[k])
		}

	}
}

// ShowFlagValues displays enumerated flag values
func ShowFlagValues(lenArgs int, args []string, filter string) error {
	if lenArgs == 1 {
		cmn.ShowFlagValues(flagValues, filter)
	}

	if lenArgs == 2 {
		flag := strings.ToLower(args[1])
		valSlice := []string{}

		switch {
		case flag == "order-by":
			for _, val := range ValidOrderByStrs {
				s, _ := cmn.IsValidAttr(val, UserAttrMap)
				if s == "" {
					s = val
				}
				valSlice = append(valSlice, s)
			}
			uniqueSlice := cmn.UniqueStrSlice(valSlice)
			cmn.ShowFlagValues(uniqueSlice, filter)
		case flag == "projection":
			cmn.ShowFlagValues(ValidProjections, filter)
		case flag == "sort-order":
			for _, v := range cmn.ValidSortOrders {
				valSlice = append(valSlice, v)
			}
			uniqueSlice := cmn.UniqueStrSlice(valSlice)
			cmn.ShowFlagValues(uniqueSlice, filter)
		case flag == "view-type":
			cmn.ShowFlagValues(ValidViewTypes, filter)
		default:
			return fmt.Errorf(gmess.ERR_FLAGNOTRECOGNIZED, args[1])
		}
	}

	return nil
}

// ShowSubAttrs displays attributes of composite attributes
func ShowSubAttrs(compAttr string, filter string) error {
	lwrCompAttr := strings.ToLower(compAttr)
	switch lwrCompAttr {
	case "address":
		cmn.ShowAttrs(addressAttrs, UserAttrMap, filter)
	case "email":
		cmn.ShowAttrs(emailAttrs, UserAttrMap, filter)
	case "externalid":
		cmn.ShowAttrs(extIDAttrs, UserAttrMap, filter)
	case "gender":
		cmn.ShowAttrs(genderAttrs, UserAttrMap, filter)
	case "im":
		cmn.ShowAttrs(imAttrs, UserAttrMap, filter)
	case "keyword":
		cmn.ShowAttrs(keywordAttrs, UserAttrMap, filter)
	case "language":
		cmn.ShowAttrs(languageAttrs, UserAttrMap, filter)
	case "location":
		cmn.ShowAttrs(locationAttrs, UserAttrMap, filter)
	case "name":
		cmn.ShowAttrs(nameAttrs, UserAttrMap, filter)
	case "notes":
		cmn.ShowAttrs(notesAttrs, UserAttrMap, filter)
	case "organization":
		cmn.ShowAttrs(organizationAttrs, UserAttrMap, filter)
	case "phone":
		cmn.ShowAttrs(phoneAttrs, UserAttrMap, filter)
	case "posixaccount":
		cmn.ShowAttrs(posAcctAttrs, UserAttrMap, filter)
	case "relation":
		cmn.ShowAttrs(relationAttrs, UserAttrMap, filter)
	case "sshpublickey":
		cmn.ShowAttrs(sshPubKeyAttrs, UserAttrMap, filter)
	case "website":
		cmn.ShowAttrs(websiteAttrs, UserAttrMap, filter)
	default:
		return fmt.Errorf(gmess.ERR_NOTCOMPOSITEATTR, compAttr)
	}

	return nil
}
