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

func TestValidateArgs(t *testing.T) {
	var userAttrMap = map[string]string{
		"addresses":                  "addresses",
		"addresses(country)":         "addresses(country)",
		"addresses(countrycode)":     "addresses(countryCode)",
		"addresses(customtype)":      "addresses(customType)",
		"addresses(extendedaddress)": "addresses(extendedAddress)",
		"addresses(formatted)":       "addresses(formatted)",
		"addresses(pobox)":           "addresses(poBox)",
		"addresses(postalcode)":      "addresses(postalCode)",
		"addresses(primary)":         "addresses(primary)",
		"addresses(region)":          "addresses(region)",
		"addresses(streetaddress)":   "addresses(streetAddress)",
		"addresses(type)":            "addresses(type)",
		"agreedtoterms":              "agreedToTerms",
		"aliases":                    "aliases",
		"archived":                   "archived",
		"changepasswordatnextlogin":  "changePasswordAtNextLogin",
		"christianname":              "givenName",
		"creationtime":               "creationTime",
		"customschemas":              "customSchemas",
		"customerid":                 "customerId",
		"deletiontime":               "deletionTime",
		"emails":                     "emails",
		"emails(address)":            "emails(address)",
		"emails(customtype)":         "emails(customType)",
		"emails(primary)":            "emails(primary)",
		"emails(type)":               "emails(type)",
		"etag":                       "etag",
		"externalids":                "externalIds",
		"familyname":                 "familyName",
		"firstname":                  "givenName",
		"fullname":                   "fullName",
		"gender":                     "gender",
		"givenname":                  "givenName",
		"hashfunction":               "hashFunction",
		"id":                         "id",
		"ims":                        "ims",
		"includeinglobaladdresslist": "includeInGlobalAddressList",
		"ipwhitelisted":              "ipWhiteListed",
		"isadmin":                    "isAdmin",
		"isdelegatedadmin":           "isDelegatedAdmin",
		"isenforcedin2sv":            "isEnforcedIn2Sv",
		"isenrolledin2sv":            "isEnrolledIn2Sv",
		"ismailboxsetup":             "isMailboxSetup",
		"keywords":                   "keywords",
		"kind":                       "kind",
		"languages":                  "languages",
		"lastlogintime":              "lastLoginTime",
		"lastname":                   "familyName",
		"locations":                  "locations",
		"name":                       "name",
		"name(familyname)":           "name(familyName)",
		"name(firstname)":            "name(givenName)",
		"name(fullname)":             "name(fullName)",
		"name(givenname)":            "name(givenName)",
		"name(lastname)":             "name(familyName)",
		"notes":                      "notes",
		"noneditablealiases":         "nonEditableAliases",
		"orgunitpath":                "orgUnitPath",
		"password":                   "password",
		"phones":                     "phones",
		"posixaccounts":              "posixAccounts",
		"primaryemail":               "primaryEmail",
		"recoveryemail":              "recoveryEmail",
		"recoveryphone":              "recoveryPhone",
		"relations":                  "relations",
		"sshpublickeys":              "sshPublicKeys",
		"surname":                    "familyName",
		"suspended":                  "suspended",
		"suspensionreason":           "suspensionReason",
		"thumbnailphotoetag":         "thumbnailPhotoEtag",
		"thumbnailphotourl":          "thumbnailPhotoUrl",
		"type":                       "type",
		"websites":                   "websites",
	}

	cases := []struct {
		args          string
		argMap        map[string]string
		expectedErr   string
		expectedValue []string
	}{
		{
			args:          "name(firstname)~name(lastname)~primaryemail",
			argMap:        userAttrMap,
			expectedErr:   "",
			expectedValue: []string{"name(givenName)", "name(familyName)", "primaryEmail"},
		},
		{
			args:        "name(firstname)~name(lastname)~primaryemail~addresses(streetaddress)~addresses(postalcode)",
			argMap:      userAttrMap,
			expectedErr: "",
			expectedValue: []string{"name(givenName)", "name(familyName)", "primaryEmail", "addresses(streetAddress)",
				"addresses(postalCode)"},
		},
		{
			args:        "itain'tright~name(firstname)~name(lastname)~primaryemail",
			argMap:      userAttrMap,
			expectedErr: "gmin: error - attribute itain'tright is unrecognized",
		},
		{
			args:          "firstname~lastname~primaryemail",
			argMap:        userAttrMap,
			expectedErr:   "",
			expectedValue: []string{"givenName", "familyName", "primaryEmail"},
		},
	}

	for _, c := range cases {

		output, err := ValidateArgs(c.args, c.argMap, AttrStr)

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
