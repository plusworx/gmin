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

func TestValidateAttrs(t *testing.T) {
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
		attrs         string
		attrMap       map[string]string
		expectedErr   string
		expectedValue []string
	}{
		{
			attrs:         "name(firstname)~name(lastname)~primaryemail",
			attrMap:       userAttrMap,
			expectedErr:   "",
			expectedValue: []string{"name(givenName)", "name(familyName)", "primaryEmail"},
		},
		{
			attrs:       "name(firstname)~name(lastname)~primaryemail~addresses(streetaddress)~addresses(postalcode)",
			attrMap:     userAttrMap,
			expectedErr: "",
			expectedValue: []string{"name(givenName)", "name(familyName)", "primaryEmail", "addresses(streetAddress)",
				"addresses(postalCode)"},
		},
		{
			attrs:       "itain'tright~name(firstname)~name(lastname)~primaryemail",
			attrMap:     userAttrMap,
			expectedErr: "gmin: error - attribute itain'tright is unrecognized",
		},
		{
			attrs:         "firstname~lastname~primaryemail",
			attrMap:       userAttrMap,
			expectedErr:   "",
			expectedValue: []string{"givenName", "familyName", "primaryEmail"},
		},
	}

	for _, c := range cases {

		output, err := ValidateAttrs(c.attrs, c.attrMap)

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
