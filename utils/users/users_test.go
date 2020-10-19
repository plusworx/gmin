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
	"testing"

	tsts "github.com/plusworx/gmin/tests"
	lg "github.com/plusworx/gmin/utils/logging"
	admin "google.golang.org/api/admin/directory/v1"
)

func TestAddCustomer(t *testing.T) {
	cases := []struct {
		customerID string
	}{
		{
			customerID: "my_customer",
		},
	}

	lg.InitLogging("info")

	ds, err := tsts.DummyDirectoryService(admin.AdminDirectoryUserReadonlyScope)
	if err != nil {
		t.Error("Error: failed to create dummy admin.Service")
	}

	ulc := ds.Users.List()

	for _, c := range cases {

		newULC := AddCustomer(ulc, c.customerID)

		if newULC == nil {
			t.Error("Error: failed to add Customer to UsersListCall")
		}
	}
}

func TestAddDomain(t *testing.T) {
	cases := []struct {
		domain string
	}{
		{
			domain: "my_company.org",
		},
	}

	lg.InitLogging("info")

	ds, err := tsts.DummyDirectoryService(admin.AdminDirectoryUserReadonlyScope)
	if err != nil {
		t.Error("Error: failed to create dummy admin.Service")
	}

	ulc := ds.Users.List()

	for _, c := range cases {

		newULC := AddDomain(ulc, c.domain)

		if newULC == nil {
			t.Error("Error: failed to add Domain to UsersListCall")
		}
	}
}

func TestAddFields(t *testing.T) {
	cases := []struct {
		fields string
	}{
		{
			fields: "name,primaryEmail,id",
		},
	}

	lg.InitLogging("info")

	ds, err := tsts.DummyDirectoryService(admin.AdminDirectoryUserReadonlyScope)
	if err != nil {
		t.Error("Error: failed to create dummy admin.Service")
	}

	ulc := ds.Users.List()

	for _, c := range cases {

		newULC := AddFields(ulc, c.fields)

		if newULC == nil {
			t.Error("Error: failed to add Fields to UsersListCall")
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

	lg.InitLogging("info")

	ds, err := tsts.DummyDirectoryService(admin.AdminDirectoryUserReadonlyScope)
	if err != nil {
		t.Error("Error: failed to create dummy admin.Service")
	}

	ulc := ds.Users.List()

	for _, c := range cases {

		newULC := AddMaxResults(ulc, c.maxResults)

		if newULC == nil {
			t.Error("Error: failed to add MaxResults to UsersListCall")
		}
	}
}

func TestAddOrderBy(t *testing.T) {
	cases := []struct {
		orderBy string
	}{
		{
			orderBy: "givenName",
		},
	}

	lg.InitLogging("info")

	ds, err := tsts.DummyDirectoryService(admin.AdminDirectoryUserReadonlyScope)
	if err != nil {
		t.Error("Error: failed to create dummy admin.Service")
	}

	ulc := ds.Users.List()

	for _, c := range cases {

		newULC := AddOrderBy(ulc, c.orderBy)

		if newULC == nil {
			t.Error("Error: failed to add OrderBy to UsersListCall")
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

	lg.InitLogging("info")

	ds, err := tsts.DummyDirectoryService(admin.AdminDirectoryUserReadonlyScope)
	if err != nil {
		t.Error("Error: failed to create dummy admin.Service")
	}

	ulc := ds.Users.List()

	for _, c := range cases {

		newULC := AddPageToken(ulc, c.token)

		if newULC == nil {
			t.Error("Error: failed to add PageToken to UsersListCall")
		}
	}
}

func TestAddProjection(t *testing.T) {
	cases := []struct {
		projection string
	}{
		{
			projection: "basic",
		},
	}

	lg.InitLogging("info")

	ds, err := tsts.DummyDirectoryService(admin.AdminDirectoryUserReadonlyScope)
	if err != nil {
		t.Error("Error: failed to create dummy admin.Service")
	}

	ulc := ds.Users.List()

	for _, c := range cases {

		newULC := AddProjection(ulc, c.projection)

		if newULC == nil {
			t.Error("Error: failed to add Projection to UsersListCall")
		}
	}
}

func TestAddQuery(t *testing.T) {
	cases := []struct {
		query string
	}{
		{
			query: "name:Martin",
		},
	}

	lg.InitLogging("info")

	ds, err := tsts.DummyDirectoryService(admin.AdminDirectoryUserReadonlyScope)
	if err != nil {
		t.Error("Error: failed to create dummy admin.Service")
	}

	ulc := ds.Users.List()

	for _, c := range cases {

		newULC := AddQuery(ulc, c.query)

		if newULC == nil {
			t.Error("Error: failed to add Fields to UsersListCall")
		}
	}
}

func TestAddShowDeleted(t *testing.T) {
	lg.InitLogging("info")

	ds, err := tsts.DummyDirectoryService(admin.AdminDirectoryUserReadonlyScope)
	if err != nil {
		t.Error("Error: failed to create dummy admin.Service")
	}

	ulc := ds.Users.List()

	newULC := AddShowDeleted(ulc)

	if newULC == nil {
		t.Error("Error: failed to add ShowDeleted to UsersListCall")
	}
}

func TestAddSortOrder(t *testing.T) {
	cases := []struct {
		sortOrder string
	}{
		{
			sortOrder: "descending",
		},
	}

	lg.InitLogging("info")

	ds, err := tsts.DummyDirectoryService(admin.AdminDirectoryUserReadonlyScope)
	if err != nil {
		t.Error("Error: failed to create dummy admin.Service")
	}

	ulc := ds.Users.List()

	for _, c := range cases {

		newULC := AddSortOrder(ulc, c.sortOrder)

		if newULC == nil {
			t.Error("Error: failed to add SortOrder to UsersListCall")
		}
	}
}

func TestAddViewType(t *testing.T) {
	cases := []struct {
		viewType string
	}{
		{
			viewType: "domain_public",
		},
	}

	lg.InitLogging("info")

	ds, err := tsts.DummyDirectoryService(admin.AdminDirectoryUserReadonlyScope)
	if err != nil {
		t.Error("Error: failed to create dummy admin.Service")
	}

	ulc := ds.Users.List()

	for _, c := range cases {

		newULC := AddViewType(ulc, c.viewType)

		if newULC == nil {
			t.Error("Error: failed to add ViewType to UsersListCall")
		}
	}
}

func TestHashPassword(t *testing.T) {
	lg.InitLogging("info")

	pwd := "MySuperStrongPassword"

	hashedPwd, _ := HashPassword(pwd)

	if hashedPwd != "e1f7c050db42a86e4d358e8c1dcef57e3b4f2fc0" {
		t.Errorf("Expected user.Password to be %v but got %v", "e1f7c050db42a86e4d358e8c1dcef57e3b4f2fc0", hashedPwd)
	}
}
