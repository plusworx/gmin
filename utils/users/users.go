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
	admin "google.golang.org/api/admin/directory/v1"
	"google.golang.org/api/googleapi"
)

const (
	// EndField is List call attribute string terminator
	EndField = ")"
	// StartUsersField is List call attribute string prefix
	StartUsersField = "users("
	startNameField  = "name("
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
	"christianname",
	"familyname",
	"firstname",
	"fullname",
	"givenname",
	"lastname",
	"surname",
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
	"userkey":                    "userKey", // Used in batch update
	"username":                   "username",
	"value":                      "value",
	"websites":                   "websites",
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

// GminUser is custom admin.User struct with no omitempty tags
type GminUser struct {
	Addresses interface{} `json:"addresses"`

	// AgreedToTerms: Indicates if user has agreed to terms (Read-only)
	AgreedToTerms bool `json:"agreedToTerms"`

	// Aliases: List of aliases (Read-only)
	Aliases []string `json:"aliases"`

	// Archived: Indicates if user is archived.
	Archived bool `json:"archived"`

	// ChangePasswordAtNextLogin: Boolean indicating if the user should
	// change password in next login
	ChangePasswordAtNextLogin bool `json:"changePasswordAtNextLogin"`

	// CreationTime: User's G Suite account creation time. (Read-only)
	CreationTime string `json:"creationTime"`

	// CustomSchemas: Custom fields of the user.
	CustomSchemas map[string]googleapi.RawMessage `json:"customSchemas"`

	// CustomerId: CustomerId of User (Read-only)
	CustomerId string `json:"customerId"`

	DeletionTime string `json:"deletionTime"`

	Emails interface{} `json:"emails"`

	// Etag: ETag of the resource.
	Etag string `json:"etag"`

	ExternalIds interface{} `json:"externalIds"`

	Gender interface{} `json:"gender"`

	// HashFunction: Hash function name for password. Supported are MD5,
	// SHA-1 and crypt
	HashFunction string `json:"hashFunction"`

	// Id: Unique identifier of User (Read-only)
	Id string `json:"id"`

	Ims interface{} `json:"ims"`

	// IncludeInGlobalAddressList: Boolean indicating if user is included in
	// Global Address List
	IncludeInGlobalAddressList bool `json:"includeInGlobalAddressList"`

	// IpWhitelisted: Boolean indicating if ip is whitelisted
	IpWhitelisted bool `json:"ipWhitelisted"`

	// IsAdmin: Boolean indicating if the user is admin (Read-only)
	IsAdmin bool `json:"isAdmin"`

	// IsDelegatedAdmin: Boolean indicating if the user is delegated admin
	// (Read-only)
	IsDelegatedAdmin bool `json:"isDelegatedAdmin"`

	// IsEnforcedIn2Sv: Is 2-step verification enforced (Read-only)
	IsEnforcedIn2Sv bool `json:"isEnforcedIn2Sv"`

	// IsEnrolledIn2Sv: Is enrolled in 2-step verification (Read-only)
	IsEnrolledIn2Sv bool `json:"isEnrolledIn2Sv"`

	// IsMailboxSetup: Is mailbox setup (Read-only)
	IsMailboxSetup bool `json:"isMailboxSetup"`

	Keywords interface{} `json:"keywords"`

	// Kind: Kind of resource this is.
	Kind string `json:"kind"`

	Languages interface{} `json:"languages"`

	// LastLoginTime: User's last login time. (Read-only)
	LastLoginTime string `json:"lastLoginTime"`

	Locations interface{} `json:"locations"`

	// Name: User's name
	Name *GminUserName `json:"name"`

	// NonEditableAliases: List of non editable aliases (Read-only)
	NonEditableAliases []string `json:"nonEditableAliases"`

	Notes interface{} `json:"notes"`

	// OrgUnitPath: OrgUnit of User
	OrgUnitPath string `json:"orgUnitPath"`

	Organizations interface{} `json:"organizations"`

	// Password: User's password
	Password string `json:"password"`

	Phones interface{} `json:"phones"`

	PosixAccounts interface{} `json:"posixAccounts"`

	// PrimaryEmail: username of User
	PrimaryEmail string `json:"primaryEmail"`

	// RecoveryEmail: Recovery email of the user.
	RecoveryEmail string `json:"recoveryEmail"`

	// RecoveryPhone: Recovery phone of the user. The phone number must be
	// in the E.164 format, starting with the plus sign (+). Example:
	// +16506661212.
	RecoveryPhone string `json:"recoveryPhone"`

	Relations interface{} `json:"relations"`

	SshPublicKeys interface{} `json:"sshPublicKeys"`

	// Suspended: Indicates if user is suspended.
	Suspended bool `json:"suspended"`

	// SuspensionReason: Suspension reason if user is suspended (Read-only)
	SuspensionReason string `json:"suspensionReason"`

	// ThumbnailPhotoEtag: ETag of the user's photo (Read-only)
	ThumbnailPhotoEtag string `json:"thumbnailPhotoEtag"`

	// ThumbnailPhotoUrl: Photo Url of the user (Read-only)
	ThumbnailPhotoUrl string `json:"thumbnailPhotoUrl"`

	Websites interface{} `json:"websites"`

	// ServerResponse contains the HTTP response code and headers from the
	// server.
	googleapi.ServerResponse `json:"-"`

	// ForceSendFields is a list of field names (e.g. "Addresses") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`

	// NullFields is a list of field names (e.g. "Addresses") to include in
	// API requests with the JSON null value. By default, fields with empty
	// values are omitted from API requests. However, any field with an
	// empty value appearing in NullFields will be sent to the server as
	// null. It is an error if a field in this list has a non-empty value.
	// This may be used to include null fields in Patch requests.
	NullFields []string `json:"-"`
}

// GminUsers is custom admin.Users struct containing GminUser
type GminUsers struct {
	// Etag: ETag of the resource.
	Etag string `json:"etag,omitempty"`

	// Kind: Kind of resource this is.
	Kind string `json:"kind,omitempty"`

	// NextPageToken: Token used to access next page of this result.
	NextPageToken string `json:"nextPageToken,omitempty"`

	// TriggerEvent: Event that triggered this response (only used in case
	// of Push Response)
	TriggerEvent string `json:"trigger_event,omitempty"`

	// Users: List of user objects.
	Users []*GminUser `json:"users,omitempty"`

	// ServerResponse contains the HTTP response code and headers from the
	// server.
	googleapi.ServerResponse `json:"-"`

	// ForceSendFields is a list of field names (e.g. "Etag") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`

	// NullFields is a list of field names (e.g. "Etag") to include in API
	// requests with the JSON null value. By default, fields with empty
	// values are omitted from API requests. However, any field with an
	// empty value appearing in NullFields will be sent to the server as
	// null. It is an error if a field in this list has a non-empty value.
	// This may be used to include null fields in Patch requests.
	NullFields []string `json:"-"`
}

// GminUserName is custom admin.UserName struct with no omitempty tags
type GminUserName struct {
	// FamilyName: Last Name
	FamilyName string `json:"familyName"`

	// FullName: Full Name
	FullName string `json:"fullName"`

	// GivenName: First Name
	GivenName string `json:"givenName"`

	// ForceSendFields is a list of field names (e.g. "FamilyName") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`

	// NullFields is a list of field names (e.g. "FamilyName") to include in
	// API requests with the JSON null value. By default, fields with empty
	// values are omitted from API requests. However, any field with an
	// empty value appearing in NullFields will be sent to the server as
	// null. It is an error if a field in this list has a non-empty value.
	// This may be used to include null fields in Patch requests.
	NullFields []string `json:"-"`
}

// Key is struct used to extract userKey
type Key struct {
	UserKey string
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
