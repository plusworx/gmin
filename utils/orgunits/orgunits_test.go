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

package orgunits

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
			fields: "name,orgUnitPath",
		},
	}

	lg.InitLogging("info")

	ds, err := tsts.DummyDirectoryService(admin.AdminDirectoryOrgunitReadonlyScope)
	if err != nil {
		t.Error("Error: failed to create dummy admin.Service")
	}

	oulc := ds.Orgunits.List("my_customer")

	for _, c := range cases {

		newOULC := AddFields(oulc, c.fields)

		if newOULC == nil {
			t.Error("Error: failed to add Fields to OrgunitsListCall")
		}
	}
}

func TestAddOUPath(t *testing.T) {
	cases := []struct {
		orgUnitPath string
	}{
		{
			orgUnitPath: "/Sales",
		},
	}

	lg.InitLogging("info")

	ds, err := tsts.DummyDirectoryService(admin.AdminDirectoryOrgunitReadonlyScope)
	if err != nil {
		t.Error("Error: failed to create dummy admin.Service")
	}

	oulc := ds.Orgunits.List("my_customer")

	for _, c := range cases {

		newOULC := AddOUPath(oulc, c.orgUnitPath)

		if newOULC == nil {
			t.Error("Error: failed to add OrgUnitPath to OrgunitsListCall")
		}
	}
}

func TestAddType(t *testing.T) {
	cases := []struct {
		searchType string
	}{
		{
			searchType: "all",
		},
	}

	lg.InitLogging("info")

	ds, err := tsts.DummyDirectoryService(admin.AdminDirectoryOrgunitReadonlyScope)
	if err != nil {
		t.Error("Error: failed to create dummy admin.Service")
	}

	oulc := ds.Orgunits.List("my_customer")

	for _, c := range cases {

		newOULC := AddType(oulc, c.searchType)

		if newOULC == nil {
			t.Error("Error: failed to add Type to OrgunitsListCall")
		}
	}
}
