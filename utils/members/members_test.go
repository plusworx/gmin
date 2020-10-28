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

package members

import (
	"testing"

	tsts "github.com/plusworx/gmin/tests"
	lg "github.com/plusworx/gmin/utils/logging"
	admin "google.golang.org/api/admin/directory/v1"
)

func TestAddFields(t *testing.T) {
	cases := []struct {
		fields string
	}{
		{
			fields: "email,role,status",
		},
	}

	tsts.InitConfig()
	lg.InitLogging("info")

	ds, err := tsts.DummyDirectoryService(admin.AdminDirectoryGroupMemberReadonlyScope)
	if err != nil {
		t.Error("Error: failed to create dummy admin.Service")
	}

	mlc := ds.Members.List("test@company.org")

	for _, c := range cases {

		newMLC := AddFields(mlc, c.fields)

		if newMLC == nil {
			t.Error("Error: failed to add Fields to MembersListCall")
		}
	}
}

func TestAddPageToken(t *testing.T) {
	cases := []struct {
		token string
	}{
		{
			token: "token_string",
		},
	}

	tsts.InitConfig()
	lg.InitLogging("info")

	ds, err := tsts.DummyDirectoryService(admin.AdminDirectoryGroupMemberReadonlyScope)
	if err != nil {
		t.Error("Error: failed to create dummy admin.Service")
	}

	mlc := ds.Members.List("test@company.org")

	for _, c := range cases {

		newMLC := AddPageToken(mlc, c.token)

		if newMLC == nil {
			t.Error("Error: failed to add PageToken to MembersListCall")
		}
	}
}

func TestAddMaxResults(t *testing.T) {
	cases := []struct {
		maxResults int64
	}{
		{
			maxResults: 150,
		},
	}

	tsts.InitConfig()
	lg.InitLogging("info")

	ds, err := tsts.DummyDirectoryService(admin.AdminDirectoryGroupMemberReadonlyScope)
	if err != nil {
		t.Error("Error: failed to create dummy admin.Service")
	}

	mlc := ds.Members.List("test@company.org")

	for _, c := range cases {

		newMLC := AddMaxResults(mlc, c.maxResults)

		if newMLC == nil {
			t.Error("Error: failed to add MaxResults to MembersListCall")
		}
	}
}

func TestValidateDeliverySetting(t *testing.T) {
	cases := []struct {
		delSetting      string
		expectedErr     string
		expectedSetting string
	}{
		{
			delSetting:      "All_Mail",
			expectedErr:     "",
			expectedSetting: "ALL_MAIL",
		},
		{
			delSetting:      "Unknown",
			expectedErr:     "invalid delivery setting: Unknown",
			expectedSetting: "",
		},
	}

	tsts.InitConfig()
	lg.InitLogging("info")

	for _, c := range cases {
		output, err := ValidateDeliverySetting(c.delSetting)

		if err != nil {
			if err.Error() != c.expectedErr {
				t.Errorf("Expected error: %v  Got: %v", c.expectedErr, err.Error())
			}
			continue
		}

		if output != c.expectedSetting {
			t.Errorf("Expected output: %v  Got: %v", c.expectedSetting, output)
		}
	}
}

func TestValidateRole(t *testing.T) {
	cases := []struct {
		role         string
		expectedErr  string
		expectedRole string
	}{
		{
			role:         "Owner",
			expectedErr:  "",
			expectedRole: "OWNER",
		},
		{
			role:         "Unknown",
			expectedErr:  "invalid role: Unknown",
			expectedRole: "",
		},
	}

	tsts.InitConfig()
	lg.InitLogging("info")

	for _, c := range cases {
		output, err := ValidateRole(c.role)

		if err != nil {
			if err.Error() != c.expectedErr {
				t.Errorf("Expected error: %v  Got: %v", c.expectedErr, err.Error())
			}
			continue
		}

		if output != c.expectedRole {
			t.Errorf("Expected output: %v  Got: %v", c.expectedRole, output)
		}
	}
}
