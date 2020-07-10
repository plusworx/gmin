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
	admin "google.golang.org/api/admin/directory/v1"
)

func TestAddListCustomer(t *testing.T) {
	cases := []struct {
		customerID string
	}{
		{
			customerID: "my_customer",
		},
	}

	ds, err := tsts.DummyDirectoryService(admin.AdminDirectoryGroupReadonlyScope)
	if err != nil {
		t.Error("Error: failed to create dummy admin.Service")
	}

	glc := ds.Groups.List()

	for _, c := range cases {

		newGLC := AddListCustomer(glc, c.customerID)

		if newGLC == nil {
			t.Error("Error: failed to add Customer to GroupsListCall")
		}
	}
}

func TestAddListDomain(t *testing.T) {
	cases := []struct {
		domain string
	}{
		{
			domain: "my_company.org",
		},
	}

	ds, err := tsts.DummyDirectoryService(admin.AdminDirectoryGroupReadonlyScope)
	if err != nil {
		t.Error("Error: failed to create dummy admin.Service")
	}

	glc := ds.Groups.List()

	for _, c := range cases {

		newGLC := AddListDomain(glc, c.domain)

		if newGLC == nil {
			t.Error("Error: failed to add Domain to GroupsListCall")
		}
	}
}

func TestAddListFields(t *testing.T) {
	cases := []struct {
		fields string
	}{
		{
			fields: "name,email",
		},
	}

	ds, err := tsts.DummyDirectoryService(admin.AdminDirectoryGroupReadonlyScope)
	if err != nil {
		t.Error("Error: failed to create dummy admin.Service")
	}

	glc := ds.Groups.List()

	for _, c := range cases {

		newGLC := AddListFields(glc, c.fields)

		if newGLC == nil {
			t.Error("Error: failed to add Fields to GroupsListCall")
		}
	}
}

func TestAddListMaxResults(t *testing.T) {
	cases := []struct {
		maxResults int64
	}{
		{
			maxResults: 150,
		},
	}

	ds, err := tsts.DummyDirectoryService(admin.AdminDirectoryGroupReadonlyScope)
	if err != nil {
		t.Error("Error: failed to create dummy admin.Service")
	}

	glc := ds.Groups.List()

	for _, c := range cases {

		newGLC := AddListMaxResults(glc, c.maxResults)

		if newGLC == nil {
			t.Error("Error: failed to add MaxResults to GroupsListCall")
		}
	}
}

func TestAddListOrderBy(t *testing.T) {
	cases := []struct {
		orderBy string
	}{
		{
			orderBy: "email",
		},
	}

	ds, err := tsts.DummyDirectoryService(admin.AdminDirectoryGroupReadonlyScope)
	if err != nil {
		t.Error("Error: failed to create dummy admin.Service")
	}

	glc := ds.Groups.List()

	for _, c := range cases {

		newGLC := AddListOrderBy(glc, c.orderBy)

		if newGLC == nil {
			t.Error("Error: failed to add OrderBy to GroupsListCall")
		}
	}
}

func TestAddListQuery(t *testing.T) {
	cases := []struct {
		query string
	}{
		{
			query: "name=test",
		},
	}

	ds, err := tsts.DummyDirectoryService(admin.AdminDirectoryGroupReadonlyScope)
	if err != nil {
		t.Error("Error: failed to create dummy admin.Service")
	}

	glc := ds.Groups.List()

	for _, c := range cases {

		newGLC := AddListQuery(glc, c.query)

		if newGLC == nil {
			t.Error("Error: failed to add Query to GroupsListCall")
		}
	}
}

func TestAddListSortOrder(t *testing.T) {
	cases := []struct {
		sortOrder string
	}{
		{
			sortOrder: "descending",
		},
	}

	ds, err := tsts.DummyDirectoryService(admin.AdminDirectoryGroupReadonlyScope)
	if err != nil {
		t.Error("Error: failed to create dummy admin.Service")
	}

	glc := ds.Groups.List()

	for _, c := range cases {

		newGLC := AddListSortOrder(glc, c.sortOrder)

		if newGLC == nil {
			t.Error("Error: failed to add SortOrder to GroupsListCall")
		}
	}
}
