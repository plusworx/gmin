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
	"testing"

	tsts "github.com/plusworx/gmin/tests"
)

func TestHashPassword(t *testing.T) {
	pwd := "MySuperStrongPassword"

	hashedPwd, _ := HashPassword(pwd)

	if hashedPwd != "e1f7c050db42a86e4d358e8c1dcef57e3b4f2fc0" {
		t.Errorf("Expected user.Password to be %v but got %v", "e1f7c050db42a86e4d358e8c1dcef57e3b4f2fc0", hashedPwd)
	}
}

func TestIsValidAttr(t *testing.T) {
	cases := []struct {
		attr          string
		attrMap       map[string]string
		expectedErr   string
		expectedValue string
	}{
		{
			attr:          "admincreated",
			attrMap:       tsts.TestGroupAttrMap,
			expectedErr:   "",
			expectedValue: "adminCreated",
		},
		{
			attr:          "nonexistent",
			attrMap:       tsts.TestGroupAttrMap,
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
		{
			attrs:          "name/givenname~primaryemail",
			expectedResult: "name/givenName,primaryEmail",
		},
		{
			attrs:          "customschemas/EmploymentData/startDate",
			expectedResult: "customSchemas/EmploymentData/startDate",
		},
	}

	for _, c := range cases {
		output, err := ParseOutputAttrs(c.attrs, tsts.TestUserAttrMap)
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
	cases := []struct {
		expectedErr    string
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
		{
			query:          "EmploymentData.projects:'GeneGnomes'",
			expectedResult: "EmploymentData.projects:'GeneGnomes'",
		},
		{
			query:          "EmploymentData.jobLevel>=7",
			expectedResult: "EmploymentData.jobLevel>=7",
		},
		{
			query:          "EmploymentData.jobLevel:[5,8]",
			expectedResult: "EmploymentData.jobLevel:[5,8]",
		},
		{
			query:       "wrongattr:Value",
			expectedErr: "gmin: error - attribute wrongattr is unrecognized",
		},
	}

	for _, c := range cases {
		output, err := ParseQuery(c.query, tsts.TestUserQueryAttrMap)
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
