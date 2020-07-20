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

package common

import (
	"reflect"
	"testing"
)

func TestHashPassword(t *testing.T) {
	pwd := "MySuperStrongPassword"

	hashedPwd, _ := HashPassword(pwd)

	if hashedPwd != "e1f7c050db42a86e4d358e8c1dcef57e3b4f2fc0" {
		t.Errorf("Expected user.Password to be %v but got %v", "e1f7c050db42a86e4d358e8c1dcef57e3b4f2fc0", hashedPwd)
	}
}

func TestIsValidAttr(t *testing.T) {
	var groupAttrMap = map[string]string{
		"admincreated":       "adminCreated",
		"description":        "description",
		"directmemberscount": "directMembersCount",
		"email":              "email",
		"etag":               "etag",
		"id":                 "id",
		"kind":               "kind",
		"name":               "name",
		"noneditablealiases": "nonEditableAliases",
	}

	cases := []struct {
		attr          string
		attrMap       map[string]string
		expectedErr   string
		expectedValue string
	}{
		{
			attr:          "admincreated",
			attrMap:       groupAttrMap,
			expectedErr:   "",
			expectedValue: "adminCreated",
		},
		{
			attr:          "nonexistent",
			attrMap:       groupAttrMap,
			expectedErr:   "gmin: error - attribute nonexistent is unrecognized",
			expectedValue: "",
		},
	}

	for _, c := range cases {

		output, err := IsValidAttr(c.attr, c.attrMap)

		if err != nil {
			if err.Error() != c.expectedErr {
				t.Errorf("Got error: %v - expected error: %v", err.Error(), c.expectedErr)
			}

			continue
		}

		if output != c.expectedValue {
			t.Errorf("Got output: %v - expected: %v", output, c.expectedValue)
		}

	}
}

func TestParseOutputAttrs(t *testing.T) {
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
		"ipwhitelisted":              "ipWhiteListed",
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
		"username":                   "username",
		"value":                      "value",
		"websites":                   "websites",
	}

	cases := []struct {
		attrs          string
		expectedErr    string
		expectedResult string
	}{
		{
			attrs:          "givenname",
			expectedResult: "givenName",
		},
		{
			attrs:          "primaryEmail",
			expectedResult: "primaryEmail",
		},
		{
			attrs:          "EmaiLs",
			expectedResult: "emails",
		},
		{
			attrs:          "isadmin",
			expectedResult: "isAdmin",
		},
		{
			attrs:          "addresses(region)",
			expectedResult: "addresses(region)",
		},
		{
			attrs:          "name(firstname,lastname)",
			expectedResult: "name(givenName,familyName)",
		},
		{
			attrs:          "primaryEMail~emails~name(christianname)",
			expectedResult: "primaryEmail,emails,name(givenName)",
		},
		{
			attrs:          "directManager",
			expectedResult: "",
			expectedErr:    "gmin: error - attribute directManager is unrecognized",
		},
	}

	for _, c := range cases {
		output, err := ParseOutputAttrs(c.attrs, UserAttrMap)
		if err != nil {
			if err.Error() != c.expectedErr {
				t.Errorf("Got error: %v - expected error: %v", err.Error(), c.expectedErr)
				continue
			}
		}

		if output != c.expectedResult {
			t.Errorf("Got result: %v - expected result: %v", output, c.expectedResult)
		}

	}
}

func TestParseQuery(t *testing.T) {
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

	cases := []struct {
		expectedErr    error
		expectedResult string
		query          string
	}{
		{
			query:          "givenname=Frank",
			expectedResult: "givenName=Frank",
		},
		{
			query:          "name='Jane Smith'",
			expectedResult: "name='Jane Smith'",
		},
		{
			query:          "email:admin*",
			expectedResult: "email:admin*",
		},
		{
			query:          "isadmin=true",
			expectedResult: "isAdmin=true",
		},
		{
			query:          "isadmin=True",
			expectedResult: "isAdmin=true",
		},
		{
			query:          "orgtitle:Manager",
			expectedResult: "orgTitle:Manager",
		},
		{
			query:          "Orgtitle:Manager",
			expectedResult: "orgTitle:Manager",
		},
		{
			query:          "directManager='bobjones@example.com'",
			expectedResult: "directManager='bobjones@example.com'",
		},
		{
			query:          "orgName=Engineering~orgTitle:Manager",
			expectedResult: "orgName=Engineering orgTitle:Manager",
		},
		{
			query:          "orgdescription:'Some description text.'",
			expectedResult: "orgDescription:'Some description text.'",
		},
	}

	for _, c := range cases {
		output, err := ParseQuery(c.query, QueryAttrMap)
		if err != c.expectedErr {
			t.Errorf("Got error: %v - expected error: %v", err.Error(), c.expectedErr)
			continue
		}

		if output != c.expectedResult {
			t.Errorf("Got result: %v - expected result: %v", output, c.expectedResult)
		}
	}
}

func TestSliceContainsStr(t *testing.T) {
	cases := []struct {
		expectedResult bool
		input          string
		sl             []string
	}{
		{
			expectedResult: true,
			input:          "Hello",
			sl:             []string{"Hello", "brave", "new", "world"},
		},
		{
			expectedResult: false,
			input:          "Goodbye",
			sl:             []string{"Hello", "brave", "new", "world"},
		},
	}

	for _, c := range cases {
		res := SliceContainsStr(c.sl, c.input)
		if res != c.expectedResult {
			t.Errorf("Got result: %v - expected result: %v", res, c.expectedResult)
		}
	}
}

func TestValidateQuery(t *testing.T) {
	var queryAttrMap = map[string]string{
		"christianname": "givenName",
		"email":         "email",
		"firstname":     "givenName",
		"lastname":      "familyName",
		"name":          "name",
		"memberkey":     "memberKey",
		"surname":       "familyName",
	}

	cases := []struct {
		attrMap       map[string]string
		expectedErr   string
		expectedValue []string
		query         string
	}{
		{
			query:         "email=finance@mycompany.org",
			attrMap:       queryAttrMap,
			expectedErr:   "",
			expectedValue: []string{"email=finance@mycompany.org"},
		},
		{
			query:         "EmaIl=marketing@mycompany.org",
			attrMap:       queryAttrMap,
			expectedErr:   "",
			expectedValue: []string{"email=marketing@mycompany.org"},
		},
		{
			query:         "name:Fin*",
			attrMap:       queryAttrMap,
			expectedErr:   "",
			expectedValue: []string{"name:Fin*"},
		},
		{
			query:         "christianname:Bri*",
			attrMap:       queryAttrMap,
			expectedErr:   "",
			expectedValue: []string{"givenName:Bri*"},
		},
		{
			query:         "email:fin*~name:Finance*",
			attrMap:       queryAttrMap,
			expectedErr:   "",
			expectedValue: []string{"email:fin*", "name:Finance*"},
		},
		{
			query:         "groupemail=engineering@mycompany.org",
			attrMap:       queryAttrMap,
			expectedErr:   "gmin: error - query attribute groupemail is unrecognized",
			expectedValue: []string{"email=malcolmx@mycompany.org"},
		},
	}

	for _, c := range cases {

		output, err := ValidateQuery(c.query, c.attrMap)

		if err != nil {
			if err.Error() != c.expectedErr {
				t.Errorf("Got error: %v - expected error: %v", err.Error(), c.expectedErr)
			}

			continue
		}

		ok := reflect.DeepEqual(output, c.expectedValue)

		if !ok {
			t.Errorf("Expected output: %v got: %v", c.expectedValue, output)
		}
	}
}
