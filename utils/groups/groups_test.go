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

package groups

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

	tsts.InitConfig()
	lg.InitLogging("info")

	ds, err := tsts.DummyDirectoryService(admin.AdminDirectoryGroupReadonlyScope)
	if err != nil {
		t.Error("Error: failed to create dummy admin.Service")
	}

	glc := ds.Groups.List()

	for _, c := range cases {

		newGLC := AddCustomer(glc, c.customerID)

		if newGLC == nil {
			t.Error("Error: failed to add Customer to GroupsListCall")
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

	tsts.InitConfig()
	lg.InitLogging("info")

	ds, err := tsts.DummyDirectoryService(admin.AdminDirectoryGroupReadonlyScope)
	if err != nil {
		t.Error("Error: failed to create dummy admin.Service")
	}

	glc := ds.Groups.List()

	for _, c := range cases {

		newGLC := AddDomain(glc, c.domain)

		if newGLC == nil {
			t.Error("Error: failed to add Domain to GroupsListCall")
		}
	}
}

func TestAddFields(t *testing.T) {
	cases := []struct {
		fields string
	}{
		{
			fields: "name,email",
		},
	}

	tsts.InitConfig()
	lg.InitLogging("info")

	ds, err := tsts.DummyDirectoryService(admin.AdminDirectoryGroupReadonlyScope)
	if err != nil {
		t.Error("Error: failed to create dummy admin.Service")
	}

	glc := ds.Groups.List()

	for _, c := range cases {

		newGLC := AddFields(glc, c.fields)

		if newGLC == nil {
			t.Error("Error: failed to add Fields to GroupsListCall")
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

	ds, err := tsts.DummyDirectoryService(admin.AdminDirectoryGroupReadonlyScope)
	if err != nil {
		t.Error("Error: failed to create dummy admin.Service")
	}

	glc := ds.Groups.List()

	for _, c := range cases {

		newGLC := AddMaxResults(glc, c.maxResults)

		if newGLC == nil {
			t.Error("Error: failed to add MaxResults to GroupsListCall")
		}
	}
}

func TestAddOrderBy(t *testing.T) {
	cases := []struct {
		orderBy string
	}{
		{
			orderBy: "email",
		},
	}

	tsts.InitConfig()
	lg.InitLogging("info")

	ds, err := tsts.DummyDirectoryService(admin.AdminDirectoryGroupReadonlyScope)
	if err != nil {
		t.Error("Error: failed to create dummy admin.Service")
	}

	glc := ds.Groups.List()

	for _, c := range cases {

		newGLC := AddOrderBy(glc, c.orderBy)

		if newGLC == nil {
			t.Error("Error: failed to add OrderBy to GroupsListCall")
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

	ds, err := tsts.DummyDirectoryService(admin.AdminDirectoryGroupReadonlyScope)
	if err != nil {
		t.Error("Error: failed to create dummy admin.Service")
	}

	glc := ds.Groups.List()

	for _, c := range cases {

		newGLC := AddPageToken(glc, c.token)

		if newGLC == nil {
			t.Error("Error: failed to add PageToken to GroupsListCall")
		}
	}
}

func TestAddQuery(t *testing.T) {
	cases := []struct {
		query string
	}{
		{
			query: "name=test",
		},
	}

	tsts.InitConfig()
	lg.InitLogging("info")

	ds, err := tsts.DummyDirectoryService(admin.AdminDirectoryGroupReadonlyScope)
	if err != nil {
		t.Error("Error: failed to create dummy admin.Service")
	}

	glc := ds.Groups.List()

	for _, c := range cases {

		newGLC := AddQuery(glc, c.query)

		if newGLC == nil {
			t.Error("Error: failed to add Query to GroupsListCall")
		}
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

	tsts.InitConfig()
	lg.InitLogging("info")

	ds, err := tsts.DummyDirectoryService(admin.AdminDirectoryGroupReadonlyScope)
	if err != nil {
		t.Error("Error: failed to create dummy admin.Service")
	}

	glc := ds.Groups.List()

	for _, c := range cases {

		newGLC := AddSortOrder(glc, c.sortOrder)

		if newGLC == nil {
			t.Error("Error: failed to add SortOrder to GroupsListCall")
		}
	}
}

func TestAddUserKey(t *testing.T) {
	cases := []struct {
		key string
	}{
		{
			key: "a.user@company.org",
		},
	}

	tsts.InitConfig()
	lg.InitLogging("info")

	ds, err := tsts.DummyDirectoryService(admin.AdminDirectoryGroupReadonlyScope)
	if err != nil {
		t.Error("Error: failed to create dummy admin.Service")
	}

	glc := ds.Groups.List()

	for _, c := range cases {

		newGLC := AddUserKey(glc, c.key)

		if newGLC == nil {
			t.Error("Error: failed to add UserKey to GroupsListCall")
		}
	}
}
