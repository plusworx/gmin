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
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"sort"
	"strings"

	cmn "github.com/plusworx/gmin/utils/common"
	flgnm "github.com/plusworx/gmin/utils/flagnames"
	gmess "github.com/plusworx/gmin/utils/gminmessages"
	lg "github.com/plusworx/gmin/utils/logging"
	"github.com/spf13/cobra"
	admin "google.golang.org/api/admin/directory/v1"
	"google.golang.org/api/googleapi"
)

const (
	// ENDFIELD is List call attribute string terminator
	ENDFIELD = ")"
	// HASHFUNCTION specifies password hash function
	HASHFUNCTION string = "SHA-1"
	// KEYNAME is name of key for processing
	KEYNAME string = "userKey"
	// STARTUSERSFIELD is List call users attribute string prefix
	STARTUSERSFIELD = "users("
)

// Key is struct used to extract userKey
type Key struct {
	UserKey string
}

// UndeleteUser is struct to extract undelete data
type UndeleteUser struct {
	UserKey     string
	OrgUnitPath string
}

// UserParams is used in batch processing
type UserParams struct {
	UserKey string
	User    *admin.User
}

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

// AddCustomer adds Customer to admin calls
func AddCustomer(ulc *admin.UsersListCall, customerID string) *admin.UsersListCall {
	lg.Debugw("starting AddCustomer()",
		"customerID", customerID)
	defer lg.Debug("finished AddCustomer()")

	var newULC *admin.UsersListCall

	newULC = ulc.Customer(customerID)

	return newULC
}

// AddCustomFieldMask adds CustomFieldMask to be used with get and list admin calls with custom projections
func AddCustomFieldMask(callObj interface{}, attrs string) interface{} {
	lg.Debugw("starting AddCustomFieldMask()",
		"attrs", attrs)
	defer lg.Debug("finished AddCustomFieldMask()")

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
	lg.Debugw("starting AddDomain()",
		"domain", domain)
	defer lg.Debug("finished AddDomain()")

	var newULC *admin.UsersListCall

	newULC = ulc.Domain(domain)

	return newULC
}

// AddFields adds fields to be returned from admin calls
func AddFields(callObj interface{}, attrs string) interface{} {
	lg.Debugw("starting AddFields()",
		"attrs", attrs)
	defer lg.Debug("finished AddFields()")

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
	lg.Debugw("starting AddMaxResults()",
		"maxResults", maxResults)
	defer lg.Debug("finished AddMaxResults()")

	var newULC *admin.UsersListCall

	newULC = ulc.MaxResults(maxResults)

	return newULC
}

// AddOrderBy adds OrderBy to admin calls
func AddOrderBy(ulc *admin.UsersListCall, orderBy string) *admin.UsersListCall {
	lg.Debugw("starting AddOrderBy()",
		"orderBy", orderBy)
	defer lg.Debug("finished AddOrderBy()")

	var newULC *admin.UsersListCall

	newULC = ulc.OrderBy(orderBy)

	return newULC
}

// AddPageToken adds PageToken to admin calls
func AddPageToken(ulc *admin.UsersListCall, token string) *admin.UsersListCall {
	lg.Debugw("starting AddPageToken()",
		"token", token)
	defer lg.Debug("finished AddPageToken()")

	var newULC *admin.UsersListCall

	newULC = ulc.PageToken(token)

	return newULC
}

// AddProjection adds Projection to admin calls
func AddProjection(callObj interface{}, projection string) interface{} {
	lg.Debugw("starting AddProjection()",
		"projection", projection)
	defer lg.Debug("finished AddProjection()")

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
	lg.Debugw("starting AddQuery()",
		"query", query)
	defer lg.Debug("finished AddQuery()")

	var newULC *admin.UsersListCall

	newULC = ulc.Query(query)

	return newULC
}

// AddShowDeleted adds ShowDeleted to admin calls
func AddShowDeleted(ulc *admin.UsersListCall) *admin.UsersListCall {
	lg.Debug("starting AddShowDeleted()")
	defer lg.Debug("finished AddShowDeleted()")

	var newULC *admin.UsersListCall

	newULC = ulc.ShowDeleted("true")

	return newULC
}

// AddSortOrder adds SortOrder to admin calls
func AddSortOrder(ulc *admin.UsersListCall, sortorder string) *admin.UsersListCall {
	lg.Debugw("starting AddSortOrder()",
		"sortorder", sortorder)
	defer lg.Debug("finished AddSortOrder()")

	var newULC *admin.UsersListCall

	newULC = ulc.SortOrder(sortorder)

	return newULC
}

// AddViewType adds ViewType to admin calls
func AddViewType(callObj interface{}, viewType string) interface{} {
	lg.Debugw("starting AddViewType()",
		"viewType", viewType)
	defer lg.Debug("finished AddViewType()")

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
	lg.Debug("starting DoGet()")
	defer lg.Debug("finished DoGet()")

	user, err := ugc.Do()
	if err != nil {
		lg.Error(err)
		return nil, err
	}

	return user, nil
}

// DoList calls the .Do() function on the admin.UsersListCall
func DoList(ulc *admin.UsersListCall) (*admin.Users, error) {
	lg.Debug("starting DoList()")
	defer lg.Debug("finished DoList()")

	users, err := ulc.Do()
	if err != nil {
		return nil, err
	}

	return users, nil
}

// GetFlagVal returns user command flag values
func GetFlagVal(cmd *cobra.Command, flagName string) (interface{}, error) {
	lg.Debugw("starting GetFlagVal()",
		"flagName", flagName)
	defer lg.Debug("finished GetFlagVal()")

	boolFlags := []string{
		flgnm.FLG_CHANGEPWD,
		flgnm.FLG_COUNT,
		flgnm.FLG_DELETED,
		flgnm.FLG_GAL,
		flgnm.FLG_SUSPENDED,
	}

	if flagName == flgnm.FLG_MAXRESULTS {
		iVal, err := cmd.Flags().GetInt64(flagName)
		if err != nil {
			lg.Error(err)
			return nil, err
		}
		return iVal, nil
	}

	ok := cmn.SliceContainsStr(boolFlags, flagName)
	if ok {
		bVal, err := cmd.Flags().GetBool(flagName)
		if err != nil {
			lg.Error(err)
			return nil, err
		}
		return bVal, nil
	}

	sVal, err := cmd.Flags().GetString(flagName)
	if err != nil {
		lg.Error(err)
		return nil, err
	}
	return sVal, nil
}

// HashPassword creates a password hash
func HashPassword(password string) (string, error) {
	lg.Debugw("starting HashPassword()",
		"password", password)
	defer lg.Debug("finished HashPassword()")

	hasher := sha1.New()

	_, err := hasher.Write([]byte(password))
	if err != nil {
		lg.Error(err)
		return "", err
	}

	hashedBytes := hasher.Sum(nil)
	hexSha1 := hex.EncodeToString(hashedBytes)

	return hexSha1, nil
}

// PopulateUndeleteUser is used in batch processing
func PopulateUndeleteUser(undelUser *UndeleteUser, hdrMap map[int]string, objData []interface{}) error {
	lg.Debugw("starting PopulateUndeleteUser()",
		"hdrMap", hdrMap)
	defer lg.Debug("finished PopulateUndeleteUser()")

	for idx, attr := range objData {
		attrName := hdrMap[idx]

		switch {
		case attrName == "userKey":
			undelUser.UserKey = fmt.Sprintf("%v", attr)
		case attrName == "orgUnitPath":
			undelUser.OrgUnitPath = fmt.Sprintf("%v", attr)
		default:
			err := fmt.Errorf(gmess.ERR_ATTRNOTRECOGNIZED, attrName)
			return err
		}
	}
	return nil
}

// PopulateUser is used in batch processing
func PopulateUser(user *admin.User, hdrMap map[int]string, objData []interface{}) error {
	lg.Debugw("starting PopulateUser()",
		"hdrMap", hdrMap)
	defer lg.Debug("finished PopulateUser()")

	name := new(admin.UserName)

	attrUsrFuncMap := map[string]func(*admin.User, string, string) error{
		"changePasswordAtNextLogin":  puChangePwd,
		"includeInGlobalAddressList": puIncludeInGAL,
		"ipWhitelisted":              puIPWhitelisted,
		"orgUnitPath":                puOrgUnitPath,
		"password":                   puPassword,
		"primaryEmail":               puPrimaryEmail,
		"recoveryEmail":              puRecoveryEmail,
		"recoveryPhone":              puRecoveryPhone,
		"suspended":                  puSuspended,
	}

	attrNameFuncMap := map[string]func(*admin.UserName, string, string) error{
		"familyName": puFamilyName,
		"givenName":  puGivenName,
	}

	for idx, attr := range objData {
		attrName := hdrMap[idx]
		attrVal := fmt.Sprintf("%v", attr)

		nameFunc, naExists := attrNameFuncMap[attrName]
		if naExists {
			err := nameFunc(name, attrName, attrVal)
			if err != nil {
				return err
			}
		}

		usrFunc, uaExists := attrUsrFuncMap[attrName]
		if uaExists {
			err := usrFunc(user, attrName, attrVal)
			if err != nil {
				return err
			}
		}
	}
	user.Name = name
	return nil
}

// PopulateUserForUpdate is used in batch processing
func PopulateUserForUpdate(userParams *UserParams, hdrMap map[int]string, objData []interface{}) error {
	lg.Debugw("starting PopulateUserForUpdate()",
		"hdrMap", hdrMap)
	defer lg.Debug("finished PopulateUserForUpdate()")

	name := new(admin.UserName)

	attrUsrFuncMap := map[string]func(*UserParams, string, string) error{
		"changePasswordAtNextLogin":  pufuChangePwd,
		"includeInGlobalAddressList": pufuIncludeInGAL,
		"ipWhitelisted":              pufuIPWhitelisted,
		"orgUnitPath":                pufuOrgUnitPath,
		"password":                   pufuPassword,
		"primaryEmail":               pufuPrimaryEmail,
		"recoveryEmail":              pufuRecoveryEmail,
		"recoveryPhone":              pufuRecoveryPhone,
		"suspended":                  pufuSuspended,
		"userKey":                    pufuUserKey,
	}

	attrNameFuncMap := map[string]func(*admin.UserName, string, string) error{
		"familyName": puFamilyName,
		"givenName":  puGivenName,
	}

	for idx, attr := range objData {
		attrName := hdrMap[idx]
		attrVal := fmt.Sprintf("%v", attr)

		nameFunc, naExists := attrNameFuncMap[attrName]
		if naExists {
			err := nameFunc(name, attrName, attrVal)
			if err != nil {
				return err
			}
		}

		usrFunc, uaExists := attrUsrFuncMap[attrName]
		if uaExists {
			err := usrFunc(userParams, attrName, attrVal)
			if err != nil {
				return err
			}
		}
	}

	if name.FamilyName != "" || name.GivenName != "" || name.FullName != "" {
		userParams.User.Name = name
	}
	return nil
}

func puChangePwd(user *admin.User, attrName string, attrVal string) error {
	lg.Debug("starting puChangePwd()")
	defer lg.Debug("finished puChangePwd()")

	if strings.ToLower(attrVal) == "true" {
		user.ChangePasswordAtNextLogin = true
	}
	return nil
}

func puFamilyName(name *admin.UserName, attrName string, attrVal string) error {
	lg.Debug("starting puFamilyName()")
	defer lg.Debug("finished puFamilyName()")

	name.FamilyName = attrVal
	return nil
}

func puGivenName(name *admin.UserName, attrName string, attrVal string) error {
	lg.Debug("starting puGivenName()")
	defer lg.Debug("finished puGivenName()")

	name.GivenName = attrVal
	return nil
}

func puIncludeInGAL(user *admin.User, attrName string, attrVal string) error {
	lg.Debug("starting puIncludeInGAL()")
	defer lg.Debug("finished puIncludeInGAL()")

	if strings.ToLower(attrVal) == "false" {
		user.IncludeInGlobalAddressList = false
		user.ForceSendFields = append(user.ForceSendFields, "IncludeInGlobalAddressList")
	}
	return nil
}

func puIPWhitelisted(user *admin.User, attrName string, attrVal string) error {
	lg.Debug("starting puIpWhitelisted()")
	defer lg.Debug("finished puIpWhitelisted()")

	if strings.ToLower(attrVal) == "true" {
		user.IpWhitelisted = true
	}
	return nil
}

func puOrgUnitPath(user *admin.User, attrName string, attrVal string) error {
	lg.Debug("starting puOrgUnitPath()")
	defer lg.Debug("finished puOrgUnitPath()")

	user.OrgUnitPath = attrVal
	return nil
}

func puPassword(user *admin.User, attrName string, attrVal string) error {
	lg.Debug("starting puPassword()")
	defer lg.Debug("finished puPassword()")

	if attrVal == "" {
		err := fmt.Errorf(gmess.ERR_EMPTYSTRING, attrName)
		lg.Error(err)
		return err
	}
	pwd, err := HashPassword(attrVal)
	if err != nil {
		return err
	}
	user.Password = pwd
	user.HashFunction = HASHFUNCTION

	return nil
}

func puPrimaryEmail(user *admin.User, attrName string, attrVal string) error {
	lg.Debug("starting puPrimaryEmail()")
	defer lg.Debug("finished puPrimaryEmail()")

	if attrVal == "" {
		err := fmt.Errorf(gmess.ERR_EMPTYSTRING, attrName)
		lg.Error(err)
		return err
	}
	user.PrimaryEmail = attrVal

	return nil
}

func puRecoveryEmail(user *admin.User, attrName string, attrVal string) error {
	lg.Debug("starting puRecoveryEmail()")
	defer lg.Debug("finished puRecoveryEmail()")

	user.RecoveryEmail = attrVal
	return nil
}

func puRecoveryPhone(user *admin.User, attrName string, attrVal string) error {
	lg.Debug("starting puRecoveryPhone()")
	defer lg.Debug("finished puRecoveryPhone()")

	if attrVal != "" {
		err := cmn.ValidateRecoveryPhone(attrVal)
		if err != nil {
			return err
		}
		user.RecoveryPhone = attrVal
	}
	return nil
}

func puSuspended(user *admin.User, attrName string, attrVal string) error {
	lg.Debug("starting puSuspended()")
	defer lg.Debug("finished puSuspended()")

	if strings.ToLower(attrVal) == "true" {
		user.Suspended = true
	}
	return nil
}

func pufuChangePwd(userParams *UserParams, attrName string, attrVal string) error {
	lg.Debug("starting pufuChangePwd()")
	defer lg.Debug("finished pufuChangePwd()")

	if strings.ToLower(attrVal) == "true" {
		userParams.User.ChangePasswordAtNextLogin = true
	} else {
		userParams.User.ChangePasswordAtNextLogin = false
		userParams.User.ForceSendFields = append(userParams.User.ForceSendFields, "ChangePasswordAtNextLogin")
	}
	return nil
}

func pufuFamilyName(name *admin.UserName, attrName string, attrVal string) error {
	lg.Debug("starting pufuFamilyName()")
	defer lg.Debug("finished pufuFamilyName()")

	name.FamilyName = attrVal
	return nil
}

func pufuGivenName(name *admin.UserName, attrName string, attrVal string) error {
	lg.Debug("starting pufuGivenName()")
	defer lg.Debug("finished pufuGivenName()")

	name.GivenName = attrVal
	return nil
}

func pufuIncludeInGAL(userParams *UserParams, attrName string, attrVal string) error {
	lg.Debug("starting pufuIncludeInGAL()")
	defer lg.Debug("finished pufuIncludeInGAL()")

	if strings.ToLower(attrVal) == "true" {
		userParams.User.IncludeInGlobalAddressList = true
	} else {
		userParams.User.IncludeInGlobalAddressList = false
		userParams.User.ForceSendFields = append(userParams.User.ForceSendFields, "IncludeInGlobalAddressList")
	}
	return nil
}

func pufuIPWhitelisted(userParams *UserParams, attrName string, attrVal string) error {
	lg.Debug("starting pufuIPWhitelisted()")
	defer lg.Debug("finished pufuIPWhitelisted()")

	if strings.ToLower(attrVal) == "true" {
		userParams.User.IpWhitelisted = true
	} else {
		userParams.User.IpWhitelisted = false
		userParams.User.ForceSendFields = append(userParams.User.ForceSendFields, "IpWhitelisted")
	}
	return nil
}

func pufuOrgUnitPath(userParams *UserParams, attrName string, attrVal string) error {
	lg.Debug("starting pufuOrgUnitPath()")
	defer lg.Debug("finished pufuOrgUnitPath()")

	if attrVal == "" {
		err := fmt.Errorf(gmess.ERR_EMPTYSTRING, attrName)
		return err
	}
	userParams.User.OrgUnitPath = attrVal
	return nil
}

func pufuPassword(userParams *UserParams, attrName string, attrVal string) error {
	lg.Debug("starting pufuPassword()")
	defer lg.Debug("finished pufuPassword()")

	if attrVal != "" {
		pwd, err := HashPassword(attrVal)
		if err != nil {
			return err
		}
		userParams.User.Password = pwd
		userParams.User.HashFunction = HASHFUNCTION
	}
	return nil
}

func pufuPrimaryEmail(userParams *UserParams, attrName string, attrVal string) error {
	lg.Debug("starting pufuPrimaryEmail()")
	defer lg.Debug("finished pufuPrimaryEmail()")

	if attrVal == "" {
		err := fmt.Errorf(gmess.ERR_EMPTYSTRING, attrName)
		return err
	}
	userParams.User.PrimaryEmail = attrVal

	return nil
}

func pufuRecoveryEmail(userParams *UserParams, attrName string, attrVal string) error {
	lg.Debug("starting pufuRecoveryEmail()")
	defer lg.Debug("finished pufuRecoveryEmail()")

	userParams.User.RecoveryEmail = attrVal
	if attrVal == "" {
		userParams.User.ForceSendFields = append(userParams.User.ForceSendFields, "RecoveryEmail")
	}
	return nil
}

func pufuRecoveryPhone(userParams *UserParams, attrName string, attrVal string) error {
	lg.Debug("starting pufuRecoveryPhone()")
	defer lg.Debug("finished pufuRecoveryPhone()")

	if attrVal != "" {
		err := cmn.ValidateRecoveryPhone(attrVal)
		if err != nil {
			lg.Error(err)
			return err
		}
	}
	if attrVal == "" {
		userParams.User.ForceSendFields = append(userParams.User.ForceSendFields, "RecoveryPhone")
	}
	userParams.User.RecoveryPhone = attrVal

	return nil
}

func pufuSuspended(userParams *UserParams, attrName string, attrVal string) error {
	lg.Debug("starting pufuSuspended()")
	defer lg.Debug("finished pufuSuspended()")

	if strings.ToLower(attrVal) == "true" {
		userParams.User.Suspended = true
	} else {
		userParams.User.Suspended = false
		userParams.User.ForceSendFields = append(userParams.User.ForceSendFields, "Suspended")
	}
	return nil
}

func pufuUserKey(userParams *UserParams, attrName string, attrVal string) error {
	lg.Debug("starting pufuUserKey()")
	defer lg.Debug("finished pufuUserKey()")

	if attrVal == "" {
		err := fmt.Errorf(gmess.ERR_EMPTYSTRING, attrName)
		return err
	}
	userParams.UserKey = attrVal

	return nil
}

// ShowAttrs displays requested user attributes
func ShowAttrs(filter string) {
	lg.Debugw("starting ShowAttrs()",
		"filter", filter)
	defer lg.Debug("finished ShowAttrs()")

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

func showAttrAddress(attr string, filter string) error {
	lg.Debugw("starting showAttrAddress()",
		"attr", attr,
		"filter", filter)
	lg.Debug("finished showAttrAddress()")

	if attr == "type" {
		cmn.ShowAttrVals(validAddressTypes, filter)
		return nil
	}
	err := fmt.Errorf(gmess.ERR_ATTRNOTRECOGNIZED, attr)
	lg.Error(err)
	return err
}

func showAttrEmail(attr string, filter string) error {
	lg.Debugw("starting showAttrEmail()",
		"attr", attr,
		"filter", filter)
	lg.Debug("finished showAttrEmail()")

	if attr == "type" {
		cmn.ShowAttrVals(validEmailTypes, filter)
		return nil
	}
	err := fmt.Errorf(gmess.ERR_ATTRNOTRECOGNIZED, attr)
	lg.Error(err)
	return err
}

func showAttrExtID(attr string, filter string) error {
	lg.Debugw("starting showAttrExtID()",
		"attr", attr,
		"filter", filter)
	lg.Debug("finished showAttrExtID()")

	if attr == "type" {
		cmn.ShowAttrVals(validExtIDTypes, filter)
		return nil
	}
	err := fmt.Errorf(gmess.ERR_ATTRNOTRECOGNIZED, attr)
	lg.Error(err)
	return err
}

func showAttrGender(attr string, filter string) error {
	lg.Debugw("starting showAttrGender()",
		"attr", attr,
		"filter", filter)
	lg.Debug("finished showAttrGender()")

	if attr == "type" {
		cmn.ShowAttrVals(validGenders, filter)
		return nil
	}
	err := fmt.Errorf(gmess.ERR_ATTRNOTRECOGNIZED, attr)
	lg.Error(err)
	return err
}

func showAttrKeyword(attr string, filter string) error {
	lg.Debugw("starting showAttrKeyword()",
		"attr", attr,
		"filter", filter)
	lg.Debug("finished showAttrKeyword()")

	if attr == "type" {
		cmn.ShowAttrVals(validKeywordTypes, filter)
		return nil
	}
	err := fmt.Errorf(gmess.ERR_ATTRNOTRECOGNIZED, attr)
	lg.Error(err)
	return err
}

func showAttrLocation(attr string, filter string) error {
	lg.Debugw("starting showAttrLocation()",
		"attr", attr,
		"filter", filter)
	lg.Debug("finished showAttrLocation()")
	if attr == "type" {
		cmn.ShowAttrVals(validLocationTypes, filter)
		return nil
	}
	err := fmt.Errorf(gmess.ERR_ATTRNOTRECOGNIZED, attr)
	lg.Error(err)
	return err
}

func showAttrNotes(attr string, filter string) error {
	lg.Debugw("starting showAttrNotes()",
		"attr", attr,
		"filter", filter)
	lg.Debug("finished showAttrNotes()")

	if attr == "type" {
		cmn.ShowAttrVals(validNotesContentTypes, filter)
		return nil
	}
	err := fmt.Errorf(gmess.ERR_ATTRNOTRECOGNIZED, attr)
	lg.Error(err)
	return err
}

func showAttrOrg(attr string, filter string) error {
	lg.Debugw("starting showAttrOrg()",
		"attr", attr,
		"filter", filter)
	lg.Debug("finished showAttrOrg()")

	if attr == "type" {
		cmn.ShowAttrVals(validOrgTypes, filter)
		return nil
	}
	err := fmt.Errorf(gmess.ERR_ATTRNOTRECOGNIZED, attr)
	lg.Error(err)
	return err
}

func showAttrPhone(attr string, filter string) error {
	lg.Debugw("starting showAttrPhone()",
		"attr", attr,
		"filter", filter)
	lg.Debug("finished showAttrPhone()")

	if attr == "type" {
		cmn.ShowAttrVals(validPhoneTypes, filter)
		return nil
	}
	err := fmt.Errorf(gmess.ERR_ATTRNOTRECOGNIZED, attr)
	lg.Error(err)
	return err
}

func showAttrPxAcct(attr string, filter string) error {
	lg.Debugw("starting showAttrPxAcct()",
		"attr", attr,
		"filter", filter)
	lg.Debug("finished showAttrPxAcct()")

	if attr == "operatingsystemtype" {
		cmn.ShowAttrVals(validOSTypes, filter)
		return nil
	}
	err := fmt.Errorf(gmess.ERR_ATTRNOTRECOGNIZED, attr)
	lg.Error(err)
	return err
}

func showAttrRelation(attr string, filter string) error {
	lg.Debugw("starting showAttrRelation()",
		"attr", attr,
		"filter", filter)
	lg.Debug("finished showAttrRelation()")

	if attr == "type" {
		cmn.ShowAttrVals(validRelationTypes, filter)
		return nil
	}
	err := fmt.Errorf(gmess.ERR_ATTRNOTRECOGNIZED, attr)
	lg.Error(err)
	return err
}

func showAttrWebsite(attr string, filter string) error {
	lg.Debugw("starting showAttrWebsite()",
		"attr", attr,
		"filter", filter)
	lg.Debug("finished showAttrWebsite()")

	if attr == "type" {
		cmn.ShowAttrVals(validWebsiteTypes, filter)
		return nil
	}
	err := fmt.Errorf(gmess.ERR_ATTRNOTRECOGNIZED, attr)
	lg.Error(err)
	return err
}

func showAttrIm(attr string, filter string) error {
	lg.Debugw("starting showAttrIm()",
		"attr", attr,
		"filter", filter)
	lg.Debug("finished showAttrIm()")

	if attr == "protocol" {
		cmn.ShowAttrVals(validImProtocols, filter)
		return nil
	}
	if attr == "type" {
		cmn.ShowAttrVals(validImTypes, filter)
		return nil
	}
	err := fmt.Errorf(gmess.ERR_ATTRNOTRECOGNIZED, attr)
	lg.Error(err)
	return err

}

// ShowAttrValues displays enumerated attribute values
func ShowAttrValues(lenArgs int, args []string, filter string) error {
	lg.Debugw("starting ShowAttrValues()",
		"lenArgs", lenArgs,
		"args", args,
		"filter", filter)
	defer lg.Debug("finished ShowAttrValues()")

	attrNames := []string{
		"address",
		"email",
		"externalid",
		"gender",
		"keyword",
		"location",
		"notes",
		"organization",
		"phone",
		"relation",
		"website",
	}

	showFuncMap := map[string]func(string, string) error{
		"address":      showAttrAddress,
		"email":        showAttrEmail,
		"externalid":   showAttrExtID,
		"gender":       showAttrGender,
		"im":           showAttrIm,
		"keyword":      showAttrKeyword,
		"location":     showAttrLocation,
		"notes":        showAttrNotes,
		"organization": showAttrOrg,
		"phone":        showAttrPhone,
		"posixaccount": showAttrPxAcct,
		"relation":     showAttrRelation,
		"website":      showAttrWebsite,
	}

	if lenArgs == 1 {
		cmn.ShowAttrVals(attrValues, filter)
		return nil
	}

	if lenArgs == 2 {
		attr := strings.ToLower(args[1])

		switch {
		case cmn.SliceContainsStr(attrNames, attr):
			fmt.Println("type")
		case attr == "im":
			fmt.Println("protocol")
			fmt.Println("type")
		case attr == "posixaccount":
			fmt.Println("operatingSystemType")
		default:
			err := fmt.Errorf(gmess.ERR_ATTRNOTRECOGNIZED, args[1])
			lg.Error(err)
			return err
		}
		return nil
	}

	attr2 := strings.ToLower(args[1])
	attr3 := strings.ToLower(args[2])

	fAttr, ok := showFuncMap[attr2]
	if !ok {
		err := fmt.Errorf(gmess.ERR_ATTRNOTRECOGNIZED, attr2)
		lg.Error(err)
		return err
	}

	err := fAttr(attr3, filter)
	if err != nil {
		return err
	}
	return nil
}

// ShowCompAttrs displays user composite attributes
func ShowCompAttrs(filter string) {
	lg.Debugw("starting ShowCompAttrs()",
		"filter", filter)
	defer lg.Debug("finished ShowCompAttrs()")

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
	lg.Debugw("starting ShowFlagValues()",
		"lenArgs", lenArgs,
		"args", args,
		"filter", filter)
	defer lg.Debug("finished ShowFlagValues()")

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
			err := fmt.Errorf(gmess.ERR_FLAGNOTRECOGNIZED, args[1])
			lg.Error(err)
			return err
		}
	}

	return nil
}

// ShowSubAttrs displays attributes of composite attributes
func ShowSubAttrs(compAttr string, filter string) error {
	lg.Debugw("starting ShowSubAttrs()",
		"compAttr", compAttr,
		"filter", filter)
	defer lg.Debug("finished ShowSubAttrs()")

	lwrCompAttr := strings.ToLower(compAttr)

	attrMatchMap := map[string][]string{
		"address":      addressAttrs,
		"email":        emailAttrs,
		"externalid":   extIDAttrs,
		"gender":       genderAttrs,
		"im":           imAttrs,
		"keyword":      keywordAttrs,
		"language":     languageAttrs,
		"location":     locationAttrs,
		"name":         nameAttrs,
		"notes":        notesAttrs,
		"organization": organizationAttrs,
		"phone":        phoneAttrs,
		"posixaccount": posAcctAttrs,
		"relation":     relationAttrs,
		"sshpublickey": sshPubKeyAttrs,
		"website":      websiteAttrs,
	}

	attrMap, isCompAttr := attrMatchMap[lwrCompAttr]
	if isCompAttr {
		cmn.ShowAttrs(attrMap, UserAttrMap, filter)
		return nil
	}

	err := fmt.Errorf(gmess.ERR_NOTCOMPOSITEATTR, compAttr)
	lg.Error(err)
	return err
}
