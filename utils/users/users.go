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
	"errors"
	"fmt"
	"strconv"
	"strings"

	cmn "github.com/plusworx/gmin/utils/common"
	admin "google.golang.org/api/admin/directory/v1"
	"google.golang.org/api/googleapi"
)

const (
	endField        = ")"
	startUsersField = "users("
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

// compositeAttrs contains names all admin.User attributes that are composite types
var compositeAttrs = []string{
	"address",
	"addresses",
	"email",
	"emails",
	"externalid",
	"externalids",
	"gender",
	"im",
	"ims",
	"keyword",
	"keywords",
	"language",
	"languages",
	"location",
	"locations",
	"name",
	"notes",
	"organisation",
	"organisations",
	"organization",
	"organizations",
	"phone",
	"phones",
	"posixaccount",
	"posixaccounts",
	"relation",
	"relations",
	"sshpublickey",
	"sshpublickeys",
	"website",
	"websites",
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
}

// UserAttrMap provides lowercase mappings to valid admin.User attributes
var UserAttrMap = map[string]string{
	"addresses":                          "addresses",
	"addresses(country)":                 "addresses(country)",
	"addresses(countrycode)":             "addresses(countryCode)",
	"addresses(customtype)":              "addresses(customType)",
	"addresses(extendedaddress)":         "addresses(extendedAddress)",
	"addresses(formatted)":               "addresses(formatted)",
	"addresses(locality)":                "addresses(locality)",
	"addresses(pobox)":                   "addresses(poBox)",
	"addresses(postalcode)":              "addresses(postalCode)",
	"addresses(primary)":                 "addresses(primary)",
	"addresses(region)":                  "addresses(region)",
	"addresses(streetaddress)":           "addresses(streetAddress)",
	"addresses(type)":                    "addresses(type)",
	"agreedtoterms":                      "agreedToTerms",
	"aliases":                            "aliases",
	"archived":                           "archived",
	"changepasswordatnextlogin":          "changePasswordAtNextLogin",
	"christianname":                      "givenName",
	"creationtime":                       "creationTime",
	"customschemas":                      "customSchemas",
	"customerid":                         "customerId",
	"deletiontime":                       "deletionTime",
	"emails":                             "emails",
	"emails(address)":                    "emails(address)",
	"emails(customtype)":                 "emails(customType)",
	"emails(primary)":                    "emails(primary)",
	"emails(type)":                       "emails(type)",
	"etag":                               "etag",
	"externalids":                        "externalIds",
	"externalids(customtype)":            "externalIds(customType)",
	"externalids(type)":                  "externalIds(type)",
	"externalids(value)":                 "externalIds(value)",
	"familyname":                         "familyName",
	"firstname":                          "givenName",
	"fullname":                           "fullName",
	"gender":                             "gender",
	"gender(addressmeas)":                "gender(addressMeAs)",
	"gender(customgender)":               "gender(customGender)",
	"gender(type)":                       "gender(type)",
	"givenname":                          "givenName",
	"hashfunction":                       "hashFunction",
	"id":                                 "id",
	"ims":                                "ims",
	"ims(customprotocol)":                "ims(customProtocol)",
	"ims(customtype)":                    "ims(customType)",
	"ims(im)":                            "ims(im)",
	"ims(primary)":                       "ims(primary)",
	"ims(protocol)":                      "ims(protocol)",
	"ims(type)":                          "ims(type)",
	"includeinglobaladdresslist":         "includeInGlobalAddressList",
	"ipwhitelisted":                      "ipWhiteListed",
	"isadmin":                            "isAdmin",
	"isdelegatedadmin":                   "isDelegatedAdmin",
	"isenforcedin2sv":                    "isEnforcedIn2Sv",
	"isenrolledin2sv":                    "isEnrolledIn2Sv",
	"ismailboxsetup":                     "isMailboxSetup",
	"keywords":                           "keywords",
	"keywords(customtype)":               "keywords(customType)",
	"keywords(type)":                     "keywords(type)",
	"keywords(value)":                    "keywords(value)",
	"kind":                               "kind",
	"languages":                          "languages",
	"languages(customlanguage)":          "languages(customLanguage)",
	"languages(languagecode)":            "languages(languageCode)",
	"lastlogintime":                      "lastLoginTime",
	"lastname":                           "familyName",
	"locations":                          "locations",
	"locations(area)":                    "locations(area)",
	"locations(buildingid)":              "locations(buildingId)",
	"locations(customtype)":              "locations(customType)",
	"locations(deskcode)":                "locations(deskCode)",
	"locations(floorname)":               "locations(floorName)",
	"locations(floorsection)":            "locations(floorSection)",
	"locations(type)":                    "locations(type)",
	"name":                               "name",
	"name(familyname)":                   "name(familyName)",
	"name(firstname)":                    "name(givenName)",
	"name(fullname)":                     "name(fullName)",
	"name(givenname)":                    "name(givenName)",
	"name(lastname)":                     "name(familyName)",
	"notes":                              "notes",
	"notes(contenttype)":                 "notes(contentType)",
	"notes(value)":                       "notes(value)",
	"noneditablealiases":                 "nonEditableAliases",
	"organisations":                      "organizations",
	"organisations(costcenter)":          "organizations(costCenter)",
	"organisations(customtype)":          "organizations(customType)",
	"organisations(department)":          "organizations(department)",
	"organisations(description)":         "organizations(description)",
	"organisations(domain)":              "organizations(domain)",
	"organisations(fulltimeequivalent)":  "organizations(fullTimeEquivalent)",
	"organisations(location)":            "organizations(location)",
	"organisations(name)":                "organizations(name)",
	"organisations(primary)":             "organizations(primary)",
	"organisations(symbol)":              "organizations(symbol)",
	"organisations(title)":               "organizations(title)",
	"organisations(type)":                "organizations(type)",
	"organizations":                      "organizations",
	"organizations(costcenter)":          "organizations(costCenter)",
	"organizations(customtype)":          "organizations(customType)",
	"organizations(department)":          "organizations(department)",
	"organizations(description)":         "organizations(description)",
	"organizations(domain)":              "organizations(domain)",
	"organizations(fulltimeequivalent)":  "organizations(fullTimeEquivalent)",
	"organizations(location)":            "organizations(location)",
	"organizations(name)":                "organizations(name)",
	"organizations(primary)":             "organizations(primary)",
	"organizations(symbol)":              "organizations(symbol)",
	"organizations(title)":               "organizations(title)",
	"organizations(type)":                "organizations(type)",
	"orgunitpath":                        "orgUnitPath",
	"password":                           "password",
	"phones":                             "phones",
	"phones(customtype)":                 "phones(customType)",
	"phones(primary)":                    "phones(primary)",
	"phones(type)":                       "phones(type)",
	"phones(value)":                      "phones(value)",
	"posixaccounts":                      "posixAccounts",
	"posixaccounts(accountid)":           "posixaccounts(accountId)",
	"posixaccounts(gecos)":               "posixaccounts(gecos)",
	"posixaccounts(gid)":                 "posixaccounts(gid)",
	"posixaccounts(homedirectory)":       "posixaccounts(homeDirectory)",
	"posixaccounts(operatingsystemtype)": "posixaccounts(operatingSystemType)",
	"posixaccounts(primary)":             "posixaccounts(primary)",
	"posixaccounts(shell)":               "posixaccounts(shell)",
	"posixaccounts(systemid)":            "posixaccounts(systemId)",
	"posixaccounts(uid)":                 "posixaccounts(uid)",
	"posixaccounts(username)":            "posixaccounts(username)",
	"primaryemail":                       "primaryEmail",
	"recoveryemail":                      "recoveryEmail",
	"recoveryphone":                      "recoveryPhone",
	"relations":                          "relations",
	"relations(customtype)":              "relations(customType)",
	"relations(type)":                    "relations(type)",
	"relations(value)":                   "relations(value)",
	"sshpublickeys":                      "sshPublicKeys",
	"sshpublickeys(expirationtimeusec)":  "sshPublicKeys(expirationTimeUsec)",
	"sshpublickeys(fingerprint)":         "sshPublicKeys(fingerprint)",
	"sshpublickeys(key)":                 "sshPublicKeys(key)",
	"surname":                            "familyName",
	"suspended":                          "suspended",
	"suspensionreason":                   "suspensionReason",
	"thumbnailphotoetag":                 "thumbnailPhotoEtag",
	"thumbnailphotourl":                  "thumbnailPhotoUrl",
	"type":                               "type",
	"websites":                           "websites",
	"websites(customtype)":               "websites(customType)",
	"websites(primary)":                  "websites(primary)",
	"websites(type)":                     "websites(type)",
	"websites(value)":                    "websites(value)",
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
	var newULC *admin.UsersListCall

	newULC = ulc.Customer(customerID)

	return newULC
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

// doComposite processes composite admin.UserName attributes
func doComposite(user *admin.User, attrStack []string) ([]string, error) {
	var (
		attrName     string
		compStack    []string
		newStack     []string
		processStack []string
		stop         bool
	)

	if len(attrStack) < 2 {
		err := errors.New("gmin: error - malformed attribute string")
		return nil, err
	}

	attrName = attrStack[0]
	processStack = attrStack[1:]

	for idx, elem := range processStack {
		if strings.Contains(elem, "{") {
			elem = strings.Replace(elem, "{", "", 1)
		}

		if strings.Contains(elem, "}") {
			elem = strings.Replace(elem, "}", "", 1)
			stop = true
		}

		compStack = append(compStack, elem)

		if stop {
			err := processCompStack(user, compStack, attrName)
			if err != nil {
				return nil, err
			}

			if idx == len(processStack)-1 {
				return nil, nil
			}

			startNewStack := idx + 1
			newStack = processStack[startNewStack:]

			return newStack, nil
		}
	}

	err := errors.New("gmin: error - malformed attribute string")
	return nil, err
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

// doName processes admin.UserName attributes
func doName(name *admin.UserName, attrStack []string) ([]string, error) {
	var (
		nameStack    []string
		newStack     []string
		processStack []string
		stop         bool
	)

	processStack = attrStack[1:]

	for idx, elem := range processStack {
		if strings.Contains(elem, "{") {
			elem = strings.Replace(elem, "{", "", 1)
		}

		if strings.Contains(elem, "}") {
			elem = strings.Replace(elem, "}", "", 1)
			stop = true
		}

		nameStack = append(nameStack, elem)

		if stop {
			newName, err := makeName(nameStack)
			if err != nil {
				return nil, err
			}

			if newName.FamilyName != "" {
				name.FamilyName = newName.FamilyName
			}

			if newName.FullName != "" {
				name.FullName = newName.FullName
			}

			if newName.GivenName != "" {
				name.GivenName = newName.GivenName
			}

			if idx == len(processStack)-1 {
				return nil, nil
			}

			startNewStack := idx + 1
			newStack = processStack[startNewStack:]

			return newStack, nil
		}
	}

	err := errors.New("gmin: error - malformed name attribute")
	return nil, err
}

// doNonComposite processes admin.User non-composite attributes
func doNonComposite(user *admin.User, attrStack []string) ([]string, error) {
	if len(attrStack) < 2 {
		err := errors.New("gmin: error - malformed attribute string")
		return nil, err
	}

	attrName := strings.ToLower(attrStack[0])
	attrValue := attrStack[1]
	newStack := []string{}

	if len(attrStack) > 2 {
		newStack = attrStack[2:]
	}

	switch true {
	case attrName == "changepasswordatnextlogin":
		if strings.ToLower(attrValue) == "true" {
			user.ChangePasswordAtNextLogin = true
		} else {
			user.ChangePasswordAtNextLogin = false
			user.ForceSendFields = append(user.ForceSendFields, "ChangePasswordAtNextLogin")
		}
	case attrName == "includeinglobaladdresslist":
		if strings.ToLower(attrValue) == "true" {
			user.IncludeInGlobalAddressList = true
		} else {
			user.IncludeInGlobalAddressList = false
			user.ForceSendFields = append(user.ForceSendFields, "IncludeGlobalAddressList")
		}
	case attrName == "ipwhitelisted":
		if strings.ToLower(attrValue) == "true" {
			user.IpWhitelisted = true
		} else {
			user.IpWhitelisted = false
			user.ForceSendFields = append(user.ForceSendFields, "IpWhitelisted")
		}
	case attrName == "orgunitpath":
		user.OrgUnitPath = attrValue
	case attrName == "password":
		pwd, err := cmn.HashPassword(attrValue)
		if err != nil {
			return nil, err
		}

		user.Password = pwd
	case attrName == "primaryemail":
		user.PrimaryEmail = attrValue
	case attrName == "recoveryemail":
		user.RecoveryEmail = attrValue
	case attrName == "recoveryphone":
		if string(attrValue[0]) != "+" {
			err := fmt.Errorf("gmin: error - recovery phone number %v must start with '+'", attrValue)
			return nil, err
		}
		user.RecoveryPhone = attrValue
	case attrName == "suspended":
		if strings.ToLower(attrValue) == "true" {
			user.Suspended = true
		} else {
			user.Suspended = false
			user.ForceSendFields = append(user.ForceSendFields, "Suspended")
		}
	default:
		err := fmt.Errorf("gmin: error - attribute %v not recognized", attrName)
		return nil, err
	}

	return newStack, nil
}

// FormatAttrs formats attributes for admin.UsersListCall.Fields or admin.UsersGetCall.Fields call
func FormatAttrs(attrs []string, get bool) string {
	var (
		nameRequired    bool
		outputName      string
		outputOtherFlds string
		outputStr       string
		name            []string
		userFields      []string
	)

	for _, a := range attrs {
		lowerA := strings.ToLower(a)
		if cmn.SliceContainsStr(nameAttrs, lowerA) {
			nameRequired = true
			name = append(name, a)
			continue
		}

		userFields = append(userFields, a)
	}

	if len(userFields) > 0 {
		outputOtherFlds = strings.Join(userFields, ",")
	}

	outputStr = outputOtherFlds

	if nameRequired {
		outputName = startNameField + strings.Join(name, ",") + endField
	}

	if outputName != "" && outputStr != "" {
		outputStr = outputStr + "," + outputName
	}

	if outputName != "" && outputStr == "" {
		outputStr = outputName
	}

	if !get {
		outputStr = startUsersField + outputStr + endField
	}

	return outputStr
}

// isCompositeAttr checks whether or not an attribute is a composite type
func isCompositeAttr(attr string) bool {

	if cmn.SliceContainsStr(compositeAttrs, attr) {
		return true
	}

	return false
}

func makeAbout(aboutParts []string) (*admin.UserAbout, error) {
	var (
		attrName string
		err      error
		newAbout *admin.UserAbout
	)

	if len(aboutParts)%2 != 0 {
		err := errors.New("gmin: error - malformed attribute string")
		return nil, err
	}

	newAbout = new(admin.UserAbout)

	for idx, part := range aboutParts {
		if idx%2 == 0 {
			attrName = strings.ToLower(part)
			ok := cmn.SliceContainsStr(notesAttrs, attrName)
			if !ok {
				err = fmt.Errorf("gmin: error - %v is not a valid UserAbout attribute", part)
				return nil, err
			}
		} else {
			switch true {
			case attrName == "contenttype":
				ok := cmn.SliceContainsStr(validNotesContentTypes, part)
				if !ok {
					err = fmt.Errorf("gmin: error - %v is not a valid notes content type", part)
					return nil, err
				}
				newAbout.ContentType = part
			case attrName == "value":
				newAbout.Value = part
			}
		}
	}

	return newAbout, nil
}

func makeAddress(addrParts []string) (*admin.UserAddress, error) {
	var (
		newAddress *admin.UserAddress
		attrName   string
	)

	if len(addrParts)%2 != 0 {
		err := errors.New("gmin: error - malformed attribute string")
		return nil, err
	}

	newAddress = new(admin.UserAddress)

	for idx, part := range addrParts {
		if idx%2 == 0 {
			attrName = strings.ToLower(part)
			ok := cmn.SliceContainsStr(addressAttrs, attrName)
			if !ok {
				err := fmt.Errorf("gmin: error - %v is not a valid UserAddress attribute", part)
				return nil, err
			}
		} else {
			switch true {
			case attrName == "country":
				newAddress.Country = part
			case attrName == "countrycode":
				newAddress.CountryCode = part
			case attrName == "customtype":
				newAddress.CustomType = part
			case attrName == "extendedaddress":
				newAddress.ExtendedAddress = part
			case attrName == "formatted":
				newAddress.Formatted = part
			case attrName == "locality":
				newAddress.Locality = part
			case attrName == "pobox":
				newAddress.PoBox = part
			case attrName == "postalcode":
				newAddress.PostalCode = part
			case attrName == "primary":
				if strings.ToLower(part) == "true" {
					newAddress.Primary = true
				} else {
					newAddress.Primary = false
					newAddress.ForceSendFields = append(newAddress.ForceSendFields, "Primary")
				}
			case attrName == "region":
				newAddress.Region = part
			case attrName == "streetaddress":
				newAddress.StreetAddress = part
			case attrName == "type":
				ok := cmn.SliceContainsStr(validAddressTypes, part)
				if !ok {
					err := fmt.Errorf("gmin: error - %v is not a valid address type", part)
					return nil, err
				}
				newAddress.Type = part
			}
		}
	}

	return newAddress, nil
}

func makeEmail(emailParts []string) (*admin.UserEmail, error) {
	var (
		attrName string
		err      error
		newEmail *admin.UserEmail
	)

	if len(emailParts)%2 != 0 {
		err := errors.New("gmin: error - malformed attribute string")
		return nil, err
	}

	newEmail = new(admin.UserEmail)

	for idx, part := range emailParts {
		if idx%2 == 0 {
			attrName = strings.ToLower(part)
			ok := cmn.SliceContainsStr(emailAttrs, attrName)
			if !ok {
				err = fmt.Errorf("gmin: error - %v is not a valid UserEmail attribute", part)
				return nil, err
			}
		} else {
			switch true {
			case attrName == "address":
				newEmail.Address = part
			case attrName == "customtype":
				newEmail.CustomType = part
			case attrName == "primary":
				if strings.ToLower(part) == "true" {
					newEmail.Primary = true
				} else {
					newEmail.Primary = false
					newEmail.ForceSendFields = append(newEmail.ForceSendFields, "Primary")
				}
			case attrName == "type":
				ok := cmn.SliceContainsStr(validEmailTypes, part)
				if !ok {
					err = fmt.Errorf("gmin: error - %v is not a valid email type", part)
					return nil, err
				}
				newEmail.Type = part
			}
		}
	}

	return newEmail, nil
}

func makeExtID(extIDParts []string) (*admin.UserExternalId, error) {
	var (
		attrName string
		err      error
		newExtID *admin.UserExternalId
	)

	if len(extIDParts)%2 != 0 {
		err := errors.New("gmin: error - malformed attribute string")
		return nil, err
	}

	newExtID = new(admin.UserExternalId)

	for idx, part := range extIDParts {
		if idx%2 == 0 {
			attrName = strings.ToLower(part)
			ok := cmn.SliceContainsStr(extIDAttrs, attrName)
			if !ok {
				err = fmt.Errorf("gmin: error - %v is not a valid UserExternalId attribute", part)
				return nil, err
			}
		} else {
			switch true {
			case attrName == "customtype":
				newExtID.CustomType = part
			case attrName == "type":
				ok := cmn.SliceContainsStr(validExtIDTypes, part)
				if !ok {
					err = fmt.Errorf("gmin: error - %v is not a valid external id type", part)
					return nil, err
				}
				newExtID.Type = part
			case attrName == "value":
				newExtID.Value = part
			}
		}
	}

	return newExtID, nil
}

func makeGender(genParts []string) (*admin.UserGender, error) {
	var (
		attrName  string
		err       error
		newGender *admin.UserGender
	)

	if len(genParts)%2 != 0 {
		err := errors.New("gmin: error - malformed attribute string")
		return nil, err
	}

	newGender = new(admin.UserGender)

	for idx, part := range genParts {
		if idx%2 == 0 {
			attrName = strings.ToLower(part)
			ok := cmn.SliceContainsStr(genderAttrs, attrName)
			if !ok {
				err = fmt.Errorf("gmin: error - %v is not a valid UserGender attribute", part)
				return nil, err
			}
		} else {
			switch true {
			case attrName == "addressmeas":
				newGender.AddressMeAs = part
			case attrName == "customgender":
				newGender.CustomGender = part
			case attrName == "type":
				ok := cmn.SliceContainsStr(validGenders, part)
				if !ok {
					err = fmt.Errorf("gmin: error - %v is not a valid gender type", part)
					return nil, err
				}
				newGender.Type = part
			}
		}
	}

	return newGender, nil
}

func makeIm(imParts []string) (*admin.UserIm, error) {
	var (
		attrName string
		err      error
		newIm    *admin.UserIm
	)

	if len(imParts)%2 != 0 {
		err := errors.New("gmin: error - malformed attribute string")
		return nil, err
	}

	newIm = new(admin.UserIm)

	for idx, part := range imParts {
		if idx%2 == 0 {
			attrName = strings.ToLower(part)
			ok := cmn.SliceContainsStr(imAttrs, attrName)
			if !ok {
				err = fmt.Errorf("gmin: error - %v is not a valid UserIm attribute", part)
				return nil, err
			}
		} else {
			switch true {
			case attrName == "customprotocol":
				newIm.CustomProtocol = part
			case attrName == "customtype":
				newIm.CustomType = part
			case attrName == "im":
				newIm.Im = part
			case attrName == "primary":
				if strings.ToLower(part) == "true" {
					newIm.Primary = true
				} else {
					newIm.Primary = false
					newIm.ForceSendFields = append(newIm.ForceSendFields, "Primary")
				}
			case attrName == "protocol":
				ok := cmn.SliceContainsStr(validImProtocols, part)
				if !ok {
					err = fmt.Errorf("gmin: error - %v is not a valid im protocol", part)
					return nil, err
				}
				newIm.Protocol = part
			case attrName == "type":
				ok := cmn.SliceContainsStr(validImTypes, part)
				if !ok {
					err = fmt.Errorf("gmin: error - %v is not a valid im type", part)
					return nil, err
				}
				newIm.Type = part
			}
		}
	}

	return newIm, nil
}

func makeKeyword(keyParts []string) (*admin.UserKeyword, error) {
	var (
		attrName   string
		err        error
		newKeyword *admin.UserKeyword
	)

	if len(keyParts)%2 != 0 {
		err := errors.New("gmin: error - malformed attribute string")
		return nil, err
	}

	newKeyword = new(admin.UserKeyword)

	for idx, part := range keyParts {
		if idx%2 == 0 {
			attrName = strings.ToLower(part)
			ok := cmn.SliceContainsStr(keywordAttrs, attrName)
			if !ok {
				err = fmt.Errorf("gmin: error - %v is not a valid UserKeyword attribute", part)
				return nil, err
			}
		} else {
			switch true {
			case attrName == "customtype":
				newKeyword.CustomType = part
			case attrName == "type":
				ok := cmn.SliceContainsStr(validKeywordTypes, part)
				if !ok {
					err = fmt.Errorf("gmin: error - %v is not a valid keyword type", part)
					return nil, err
				}
				newKeyword.Type = part
			case attrName == "value":
				newKeyword.Value = part
			}
		}
	}

	return newKeyword, nil
}

func makeLanguage(langParts []string) (*admin.UserLanguage, error) {
	var (
		attrName    string
		err         error
		newLanguage *admin.UserLanguage
	)

	if len(langParts)%2 != 0 {
		err := errors.New("gmin: error - malformed attribute string")
		return nil, err
	}

	newLanguage = new(admin.UserLanguage)

	for idx, part := range langParts {
		if idx%2 == 0 {
			attrName = strings.ToLower(part)
			ok := cmn.SliceContainsStr(languageAttrs, attrName)
			if !ok {
				err = fmt.Errorf("gmin: error - %v is not a valid UserLanguage attribute", part)
				return nil, err
			}
		} else {
			switch true {
			case attrName == "customlanguage":
				newLanguage.CustomLanguage = part
			case attrName == "languagecode":
				newLanguage.LanguageCode = part
			}
		}
	}

	return newLanguage, nil
}

func makeLocation(locParts []string) (*admin.UserLocation, error) {
	var (
		attrName    string
		err         error
		newLocation *admin.UserLocation
	)

	if len(locParts)%2 != 0 {
		err := errors.New("gmin: error - malformed attribute string")
		return nil, err
	}

	newLocation = new(admin.UserLocation)

	for idx, part := range locParts {
		if idx%2 == 0 {
			attrName = strings.ToLower(part)
			ok := cmn.SliceContainsStr(locationAttrs, attrName)
			if !ok {
				err = fmt.Errorf("gmin: error - %v is not a valid UserLocation attribute", part)
				return nil, err
			}
		} else {
			switch true {
			case attrName == "area":
				newLocation.Area = part
			case attrName == "buildingid":
				newLocation.BuildingId = part
			case attrName == "customtype":
				newLocation.CustomType = part
			case attrName == "deskcode":
				newLocation.DeskCode = part
			case attrName == "floorname":
				newLocation.FloorName = part
			case attrName == "floorsection":
				newLocation.FloorSection = part
			case attrName == "type":
				ok := cmn.SliceContainsStr(validLocationTypes, part)
				if !ok {
					err = fmt.Errorf("gmin: error - %v is not a valid location type", part)
					return nil, err
				}
				newLocation.Type = part
			}
		}
	}

	return newLocation, nil
}

func makeName(nameParts []string) (*admin.UserName, error) {
	var (
		attrName string
		err      error
		newName  *admin.UserName
	)

	if len(nameParts)%2 != 0 {
		err := errors.New("gmin: error - malformed attribute string")
		return nil, err
	}

	newName = new(admin.UserName)

	for idx, part := range nameParts {
		if idx%2 == 0 {
			attrName = strings.ToLower(part)
			ok := cmn.SliceContainsStr(nameAttrs, attrName)
			if !ok {
				err = fmt.Errorf("gmin: error - %v is not a valid UserName attribute", part)
				return nil, err
			}
		} else {
			switch true {
			case attrName == "familyname" || attrName == "lastname" || attrName == "surname":
				newName.FamilyName = part
			case attrName == "givenname" || attrName == "firstname" || attrName == "christianname":
				newName.GivenName = part
			case attrName == "fullname":
				newName.FullName = part
			}
		}
	}

	return newName, nil
}

func makeOrganization(orgParts []string) (*admin.UserOrganization, error) {
	var (
		attrName string
		err      error
		newOrg   *admin.UserOrganization
	)

	if len(orgParts)%2 != 0 {
		err := errors.New("gmin: error - malformed attribute string")
		return nil, err
	}

	newOrg = new(admin.UserOrganization)

	for idx, part := range orgParts {
		if idx%2 == 0 {
			attrName = strings.ToLower(part)
			ok := cmn.SliceContainsStr(organizationAttrs, attrName)
			if !ok {
				err = fmt.Errorf("gmin: error - %v is not a valid UserOrganization attribute", part)
				return nil, err
			}
		} else {
			switch true {
			case attrName == "costcenter":
				newOrg.CostCenter = part
			case attrName == "customtype":
				newOrg.CustomType = part
			case attrName == "department":
				newOrg.Department = part
			case attrName == "description":
				newOrg.Description = part
			case attrName == "domain":
				newOrg.Domain = part
			case attrName == "fulltimeequivalent":
				num, err := strconv.Atoi(part)
				if err != nil {
					err = errors.New("gmin: error - FullTimeEquivalent must be a number")
					return nil, err
				}
				newOrg.FullTimeEquivalent = int64(num)
			case attrName == "location":
				newOrg.Location = part
			case attrName == "name":
				newOrg.Name = part
			case attrName == "primary":
				if strings.ToLower(part) == "true" {
					newOrg.Primary = true
				} else {
					newOrg.Primary = false
					newOrg.ForceSendFields = append(newOrg.ForceSendFields, "Primary")
				}
			case attrName == "symbol":
				newOrg.Symbol = part
			case attrName == "title":
				newOrg.Title = part
			case attrName == "type":
				ok := cmn.SliceContainsStr(validOrgTypes, part)
				if !ok {
					err = fmt.Errorf("gmin: error - %v is not a valid organization type", part)
					return nil, err
				}
				newOrg.Type = part
			}
		}
	}

	return newOrg, nil
}

func makePhone(phoneParts []string) (*admin.UserPhone, error) {
	var (
		attrName string
		err      error
		newPhone *admin.UserPhone
	)

	if len(phoneParts)%2 != 0 {
		err := errors.New("gmin: error - malformed attribute string")
		return nil, err
	}

	newPhone = new(admin.UserPhone)

	for idx, part := range phoneParts {
		if idx%2 == 0 {
			attrName = strings.ToLower(part)
			ok := cmn.SliceContainsStr(phoneAttrs, attrName)
			if !ok {
				err = fmt.Errorf("gmin: error - %v is not a valid UserPhone attribute", part)
				return nil, err
			}
		} else {
			switch true {
			case attrName == "customtype":
				newPhone.CustomType = part
			case attrName == "primary":
				if strings.ToLower(part) == "true" {
					newPhone.Primary = true
				} else {
					newPhone.Primary = false
					newPhone.ForceSendFields = append(newPhone.ForceSendFields, "Primary")
				}
			case attrName == "type":
				ok := cmn.SliceContainsStr(validPhoneTypes, part)
				if !ok {
					err = fmt.Errorf("gmin: error - %v is not a valid phone type", part)
					return nil, err
				}
				newPhone.Type = part
			case attrName == "value":
				newPhone.Value = part
			}
		}
	}

	return newPhone, nil
}

func makePosAcct(posParts []string) (*admin.UserPosixAccount, error) {
	var (
		attrName   string
		err        error
		newPosAcct *admin.UserPosixAccount
	)

	if len(posParts)%2 != 0 {
		err := errors.New("gmin: error - malformed attribute string")
		return nil, err
	}

	newPosAcct = new(admin.UserPosixAccount)

	for idx, part := range posParts {
		if idx%2 == 0 {
			attrName = strings.ToLower(part)
			ok := cmn.SliceContainsStr(posAcctAttrs, attrName)
			if !ok {
				err = fmt.Errorf("gmin: error - %v is not a valid UserPosixAccount attribute", part)
				return nil, err
			}
		} else {
			switch true {
			case attrName == "accountid":
				newPosAcct.AccountId = part
			case attrName == "gecos":
				newPosAcct.Gecos = part
			case attrName == "gid":
				num, err := strconv.Atoi(part)
				if err != nil {
					err = errors.New("gmin: error - Gid must be a number")
					return nil, err
				}
				newPosAcct.Gid = uint64(num)
			case attrName == "homedirectory":
				newPosAcct.HomeDirectory = part
			case attrName == "operatingsystemtype":
				ok := cmn.SliceContainsStr(validOSTypes, part)
				if !ok {
					err = fmt.Errorf("gmin: error - %v is not a valid operating system type", part)
					return nil, err
				}
				newPosAcct.OperatingSystemType = part
			case attrName == "primary":
				if strings.ToLower(part) == "true" {
					newPosAcct.Primary = true
				} else {
					newPosAcct.Primary = false
					newPosAcct.ForceSendFields = append(newPosAcct.ForceSendFields, "Primary")
				}
			case attrName == "shell":
				newPosAcct.Shell = part
			case attrName == "systemid":
				newPosAcct.SystemId = part
			case attrName == "uid":
				num, err := strconv.Atoi(part)
				if err != nil {
					err = errors.New("gmin: error - Uid must be a number")
					return nil, err
				}
				newPosAcct.Uid = uint64(num)
			case attrName == "username":
				newPosAcct.Username = part
			}
		}
	}

	return newPosAcct, nil
}

func makeRelation(relParts []string) (*admin.UserRelation, error) {
	var (
		attrName    string
		err         error
		newRelation *admin.UserRelation
	)

	if len(relParts)%2 != 0 {
		err := errors.New("gmin: error - malformed attribute string")
		return nil, err
	}

	newRelation = new(admin.UserRelation)

	for idx, part := range relParts {
		if idx%2 == 0 {
			attrName = strings.ToLower(part)
			ok := cmn.SliceContainsStr(relationAttrs, attrName)
			if !ok {
				err = fmt.Errorf("gmin: error - %v is not a valid UserRelation attribute", part)
				return nil, err
			}
		} else {
			switch true {
			case attrName == "customtype":
				newRelation.CustomType = part
			case attrName == "type":
				ok := cmn.SliceContainsStr(validRelationTypes, part)
				if !ok {
					err = fmt.Errorf("gmin: error - %v is not a valid relation type", part)
					return nil, err
				}
				newRelation.Type = part
			case attrName == "value":
				newRelation.Value = part
			}
		}
	}

	return newRelation, nil
}

func makeSSHPubKey(pKeyParts []string) (*admin.UserSshPublicKey, error) {
	var (
		attrName  string
		err       error
		newPubKey *admin.UserSshPublicKey
	)

	if len(pKeyParts)%2 != 0 {
		err := errors.New("gmin: error - malformed attribute string")
		return nil, err
	}

	newPubKey = new(admin.UserSshPublicKey)

	for idx, part := range pKeyParts {
		if idx%2 == 0 {
			attrName = strings.ToLower(part)
			ok := cmn.SliceContainsStr(sshPubKeyAttrs, attrName)
			if !ok {
				err = fmt.Errorf("gmin: error - %v is not a valid UserSshPublicKey attribute", part)
				return nil, err
			}
		} else {
			switch true {
			case attrName == "expirationtimeusec":
				num, err := strconv.Atoi(part)
				if err != nil {
					err = errors.New("gmin: error - ExpirationTimeUsec must be a number")
					return nil, err
				}
				newPubKey.ExpirationTimeUsec = int64(num)
			case attrName == "key":
				newPubKey.Key = part
			}
		}
	}

	return newPubKey, nil
}

func makeWebsite(webParts []string) (*admin.UserWebsite, error) {
	var (
		attrName   string
		err        error
		newWebsite *admin.UserWebsite
	)

	if len(webParts)%2 != 0 {
		err := errors.New("gmin: error - malformed attribute string")
		return nil, err
	}

	newWebsite = new(admin.UserWebsite)

	for idx, part := range webParts {
		if idx%2 == 0 {
			attrName = strings.ToLower(part)
			ok := cmn.SliceContainsStr(websiteAttrs, attrName)
			if !ok {
				err = fmt.Errorf("gmin: error - %v is not a valid UserWebsite attribute", part)
				return nil, err
			}
		} else {
			switch true {
			case attrName == "customtype":
				newWebsite.CustomType = part
			case attrName == "primary":
				if strings.ToLower(part) == "true" {
					newWebsite.Primary = true
				} else {
					newWebsite.Primary = false
					newWebsite.ForceSendFields = append(newWebsite.ForceSendFields, "Primary")
				}
			case attrName == "type":
				ok := cmn.SliceContainsStr(validWebsiteTypes, part)
				if !ok {
					err = fmt.Errorf("gmin: error - %v is not a valid website type", part)
					return nil, err
				}
				newWebsite.Type = part
			case attrName == "value":
				newWebsite.Value = part
			}
		}
	}

	return newWebsite, nil
}

func processAttrStack(user *admin.User, name *admin.UserName, attrStack []string) error {
	var (
		attrName string
		err      error
		newStack []string
	)

	attrName = strings.ToLower(attrStack[0])
	isComp := isCompositeAttr(attrName)

	if attrName == "name" {
		newStack, err = doName(name, attrStack)
		if err != nil {
			return err
		}
	}

	if isComp && attrName != "name" {
		newStack, err = doComposite(user, attrStack)
		if err != nil {
			return err
		}
	}

	if !isComp {
		newStack, err = doNonComposite(user, attrStack)
		if err != nil {
			return err
		}
	}

	if len(newStack) > 0 {
		err = processAttrStack(user, name, newStack)
		if err != nil {
			return err
		}
	}

	return nil
}

// processCompStack processes stack of composite attribute elements
func processCompStack(user *admin.User, compStack []string, attrName string) error {
	var (
		addresses       []*admin.UserAddress
		elementStack    []string
		emails          []*admin.UserEmail
		externalids     []*admin.UserExternalId
		ims             []*admin.UserIm
		keywords        []*admin.UserKeyword
		languages       []*admin.UserLanguage
		locations       []*admin.UserLocation
		organizations   []*admin.UserOrganization
		phones          []*admin.UserPhone
		posixaccts      []*admin.UserPosixAccount
		relations       []*admin.UserRelation
		sshpubkeys      []*admin.UserSshPublicKey
		websites        []*admin.UserWebsite
		newAbout        *admin.UserAbout
		newAddr         *admin.UserAddress
		newEmail        *admin.UserEmail
		newExtID        *admin.UserExternalId
		newIm           *admin.UserIm
		newGender       *admin.UserGender
		newKeyword      *admin.UserKeyword
		newLanguage     *admin.UserLanguage
		newLocation     *admin.UserLocation
		newOrganization *admin.UserOrganization
		newPhone        *admin.UserPhone
		newPubKey       *admin.UserSshPublicKey
		newPosAcct      *admin.UserPosixAccount
		newRelation     *admin.UserRelation
		newWebsite      *admin.UserWebsite
		stop            bool
	)

	stackLen := len(compStack)

	for idx, elem := range compStack {
		if idx == stackLen-1 {
			elementStack = append(elementStack, elem)
			stop = true
		}

		if elem == cmn.RepeatTxt || stop {
			newAttr, err := processElementStack(user, elementStack, attrName)
			if err != nil {
				return err
			}

			switch true {
			case attrName == "address" || attrName == "addresses":
				newAddr = newAttr.(*admin.UserAddress)
				addresses = append(addresses, newAddr)
				user.Addresses = addresses
			case attrName == "email" || attrName == "emails":
				newEmail = newAttr.(*admin.UserEmail)
				emails = append(emails, newEmail)
				user.Emails = emails
			case attrName == "externalid" || attrName == "externalids":
				newExtID = newAttr.(*admin.UserExternalId)
				externalids = append(externalids, newExtID)
				user.ExternalIds = externalids
			case attrName == "gender":
				newGender = newAttr.(*admin.UserGender)
				user.Gender = newGender
			case attrName == "im" || attrName == "ims":
				newIm = newAttr.(*admin.UserIm)
				ims = append(ims, newIm)
				user.Ims = ims
			case attrName == "keyword" || attrName == "keywords":
				newKeyword = newAttr.(*admin.UserKeyword)
				keywords = append(keywords, newKeyword)
				user.Keywords = keywords
			case attrName == "language" || attrName == "languages":
				newLanguage = newAttr.(*admin.UserLanguage)
				languages = append(languages, newLanguage)
				user.Languages = languages
			case attrName == "location" || attrName == "locations":
				newLocation = newAttr.(*admin.UserLocation)
				locations = append(locations, newLocation)
				user.Locations = locations
			case attrName == "notes":
				newAbout = newAttr.(*admin.UserAbout)
				user.Notes = newAbout
			case attrName == "organization" || attrName == "organisation" || attrName == "organizations" || attrName == "organisations":
				newOrganization = newAttr.(*admin.UserOrganization)
				organizations = append(organizations, newOrganization)
				user.Organizations = organizations
			case attrName == "phone" || attrName == "phones":
				newPhone = newAttr.(*admin.UserPhone)
				phones = append(phones, newPhone)
				user.Phones = phones
			case attrName == "posixaccount" || attrName == "posixaccounts":
				newPosAcct = newAttr.(*admin.UserPosixAccount)
				posixaccts = append(posixaccts, newPosAcct)
				user.PosixAccounts = posixaccts
			case attrName == "relation" || attrName == "relations":
				newRelation = newAttr.(*admin.UserRelation)
				relations = append(relations, newRelation)
				user.Relations = relations
			case attrName == "sshpublickey" || attrName == "sshpublickeys":
				newPubKey = newAttr.(*admin.UserSshPublicKey)
				sshpubkeys = append(sshpubkeys, newPubKey)
				user.SshPublicKeys = sshpubkeys
			case attrName == "website" || attrName == "websites":
				newWebsite = newAttr.(*admin.UserWebsite)
				websites = append(websites, newWebsite)
				user.Websites = websites
			}

			elementStack = []string{}
			continue
		}

		elementStack = append(elementStack, elem)
	}

	return nil
}

// processElementStack processes elements of a composite attribute
func processElementStack(user *admin.User, elementStack []string, attrName string) (interface{}, error) {
	var (
		err     error
		newElem interface{}
	)

	switch true {
	case attrName == "address" || attrName == "addresses":
		newElem, err = makeAddress(elementStack)
		if err != nil {
			return nil, err
		}
	case attrName == "email" || attrName == "emails":
		newElem, err = makeEmail(elementStack)
		if err != nil {
			return nil, err
		}
	case attrName == "externalid" || attrName == "externalids":
		newElem, err = makeExtID(elementStack)
		if err != nil {
			return nil, err
		}
	case attrName == "gender":
		newElem, err = makeGender(elementStack)
		if err != nil {
			return nil, err
		}
	case attrName == "im" || attrName == "ims":
		newElem, err = makeIm(elementStack)
		if err != nil {
			return nil, err
		}
	case attrName == "keyword" || attrName == "keywords":
		newElem, err = makeKeyword(elementStack)
		if err != nil {
			return nil, err
		}
	case attrName == "language" || attrName == "languages":
		newElem, err = makeLanguage(elementStack)
		if err != nil {
			return nil, err
		}
	case attrName == "location" || attrName == "locations":
		newElem, err = makeLocation(elementStack)
		if err != nil {
			return nil, err
		}
	case attrName == "notes":
		newElem, err = makeAbout(elementStack)
		if err != nil {
			return nil, err
		}
	case attrName == "organization" || attrName == "organizations" || attrName == "organisation" || attrName == "organisations":
		newElem, err = makeOrganization(elementStack)
		if err != nil {
			return nil, err
		}
	case attrName == "phone" || attrName == "phones":
		newElem, err = makePhone(elementStack)
		if err != nil {
			return nil, err
		}
	case attrName == "posixaccount" || attrName == "posixaccounts":
		newElem, err = makePosAcct(elementStack)
		if err != nil {
			return nil, err
		}
	case attrName == "relation" || attrName == "relations":
		newElem, err = makeRelation(elementStack)
		if err != nil {
			return nil, err
		}
	case attrName == "sshpublickey" || attrName == "sshpublickeys":
		newElem, err = makeSSHPubKey(elementStack)
		if err != nil {
			return nil, err
		}
	case attrName == "website" || attrName == "websites":
		newElem, err = makeWebsite(elementStack)
		if err != nil {
			return nil, err
		}
	}

	return newElem, nil
}

// ProcessFreeformAttrs processes freeform admin.User attributes
func ProcessFreeformAttrs(user *admin.User, name *admin.UserName, ffAttrs string) error {
	var (
		attrStack []string
		subParts  []string
	)

	attrParts := strings.Split(ffAttrs, ":")

	for _, part := range attrParts {
		switch true {
		case strings.Contains(part, "~"):
			subParts = strings.Split(part, "~")
			for _, s := range subParts {
				attrStack = append(attrStack, strings.TrimSpace(s))
			}
		case strings.Contains(part, "}{"):
			subParts = strings.Split(part, "}{")
			attrStack = append(attrStack, strings.TrimSpace(subParts[0]))
			attrStack = append(attrStack, cmn.RepeatTxt)
			attrStack = append(attrStack, strings.TrimSpace(subParts[1]))
		default:
			attrStack = append(attrStack, strings.TrimSpace(part))
		}
	}

	err := processAttrStack(user, name, attrStack)
	if err != nil {
		return err
	}

	return nil
}
