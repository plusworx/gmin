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

func TestDoComposite(t *testing.T) {
	cases := []struct {
		attrStack     []string
		expectedErr   string
		expectedNSLen int
		expectedVals  map[string]string
		noElems       int
	}{
		{
			attrStack: []string{"address", "{streetaddress", "201 Arbour Avenue", "locality", "Leeds", "postalcode", "LS2 1ND", "type", "home}"},
			noElems:   1,
		},
		{
			attrStack: []string{"address", "{streetaddress", "201 Arbour Avenue", "locality", "Leeds", "postalcode", "LS2 1ND", "type", "home",
				"**** repeat ****", "streetaddress", "15 Stavely Gardens", "locality", "Leeds", "postalcode", "LS1 3BB", "type", "work}"},
			noElems: 2,
		},
		{
			attrStack: []string{"addresses", "{streetaddress", "201 Arbour Avenue", "locality", "Leeds", "postalcode", "LS2 1ND", "type", "home}"},
			noElems:   1,
		},
		{
			attrStack: []string{"addresses", "{streetaddress", "201 Arbour Avenue", "locality", "Leeds", "postalcode", "LS2 1ND", "type", "home",
				"**** repeat ****", "streetaddress", "15 Stavely Gardens", "locality", "Leeds", "postalcode", "LS1 3BB", "type", "work}"},
			noElems: 2,
		},
		{
			attrStack:   []string{"address", "{streetaddress", "201 Arbour Avenue", "locality", "Leeds", "postalcode", "LS2 1ND", "type}"},
			expectedErr: "gmin: error - malformed attribute string",
		},
		{
			attrStack: []string{"address", "streetaddress", "201 Arbour Avenue", "locality", "Leeds", "postalcode", "LS2 1ND", "type", "home}"},
			noElems:   1,
		},
		{
			attrStack:   []string{"address", "{streetaddress", "201 Arbour Avenue", "locality", "Leeds", "postalcode", "LS2 1ND", "type", "home"},
			expectedErr: "gmin: error - malformed attribute string",
		},
		{
			attrStack:   []string{"address", "{roadaddress", "201 Arbour Avenue", "locality", "Leeds", "postalcode", "LS2 1ND", "type", "home}"},
			expectedErr: "gmin: error - roadaddress is not a valid UserAddress attribute",
		},
		{
			attrStack: []string{"address", "{StreetAddress", "201 Arbour Avenue", "Locality", "Leeds", "PostalCode", "LS2 1ND", "TYPE", "home}"},
			noElems:   1,
		},
		{
			attrStack:   []string{"address", "{streetaddress", "201 Arbour Avenue", "locality", "Leeds", "postalcode", "LS2 1ND", "type", "wrong}"},
			expectedErr: "gmin: error - wrong is not a valid address type",
		},
		{
			attrStack: []string{"   address   ", "{   streetaddress   ", "201 Arbour Avenue", "   locality  ", "Leeds",
				"   postalcode   ", "LS2 1ND", "   type   ", "home}"},
			noElems: 1,
		},
		{
			attrStack: []string{"email", "{address", "chief.exec@mycorp.com", "primary", "true", "type", "work}"},
			noElems:   1,
		},
		{
			attrStack: []string{"emails", "{address", "chief.exec@mycorp.com", "primary", "true", "type", "work}"},
			noElems:   1,
		},
		{
			attrStack: []string{"externalid", "{customtype", "GitHubID", "type", "custom", "value", "1234567890}"},
			noElems:   1,
		},
		{
			attrStack: []string{"externalids", "{customtype", "GitHubID", "type", "custom", "value", "1234567890}"},
			noElems:   1,
		},
		{
			attrStack: []string{"gender", "{addressmeas", "they/them", "customgender", "non-binary", "type", "other}"},
		},
		{
			attrStack: []string{"im", "{customprotocol", "plusworx", "customtype", "experimental", "im", "@mistered",
				"primary", "true", "protocol", "custom_protocol", "type", "custom}"},
			noElems: 1,
		},
		{
			attrStack: []string{"ims", "{customprotocol", "plusworx", "customtype", "experimental", "im", "@mistered",
				"primary", "true", "protocol", "custom_protocol", "type", "custom}"},
			noElems: 1,
		},
		{
			attrStack: []string{"keyword", "{customtype", "workhours", "type", "custom", "value", "part-time}"},
			noElems:   1,
		},
		{
			attrStack: []string{"keywords", "{customtype", "workhours", "type", "custom", "value", "part-time}"},
			noElems:   1,
		},
		{
			attrStack: []string{"language", "{languagecode", "en-GB}"},
			noElems:   1,
		},
		{
			attrStack: []string{"languages", "{languagecode", "en-GB}"},
			noElems:   1,
		},
		{
			attrStack: []string{"location", "{area", "Shoreditch", "buildingid", "Grebe House", "deskcode", "D12", "floorname", "12",
				"floorsection", "Marketing", "type", "desk}"},
			noElems: 1,
		},
		{
			attrStack: []string{"locations", "{area", "Shoreditch", "buildingid", "Grebe House", "deskcode", "D12", "floorname", "12",
				"floorsection", "Marketing", "type", "desk}"},
			noElems: 1,
		},
		{
			attrStack: []string{"notes", "{contenttype", "text_plain", "value", "This user is one of the company founders.}"},
		},
		{
			attrStack: []string{"organisation", "{costcenter", "104", "department", "Finance", "description", "Head Office Finance Department",
				"domain", "majestic.co.uk", "fulltimeequivalent", "100000", "location", "Newcastle", "name", "Majestic Film Ltd",
				"primary", "true", "symbol", "MAJ", "title", "CFO", "type", "work}"},
			noElems: 1,
		},
		{
			attrStack: []string{"organisations", "{costcenter", "104", "department", "Finance", "description", "Head Office Finance Department",
				"domain", "majestic.co.uk", "fulltimeequivalent", "100000", "location", "Newcastle", "name", "Majestic Film Ltd",
				"primary", "true", "symbol", "MAJ", "title", "CFO", "type", "work}"},
			noElems: 1,
		},
		{
			attrStack: []string{"organization", "{costcenter", "105", "customtype", "acquisition", "department", "Sales", "description", "Head Office Sales Department",
				"domain", "majestic.co.uk", "fulltimeequivalent", "90000", "location", "Newcastle", "name", "Majestic Film Ltd",
				"primary", "false", "symbol", "MAJ", "title", "Head of Sales", "type", "custom}"},
			noElems: 1,
		},
		{
			attrStack: []string{"organizations", "{costcenter", "105", "customtype", "acquisition", "department", "Sales", "description", "Head Office Sales Department",
				"domain", "majestic.co.uk", "fulltimeequivalent", "90000", "location", "Newcastle", "name", "Majestic Film Ltd",
				"primary", "false", "symbol", "MAJ", "title", "Head of Sales", "type", "custom}"},
			noElems: 1,
		},
		{
			attrStack: []string{"phone", "{primary", "false", "type", "mobile", "value", "05467983211}"},
			noElems:   1,
		},
		{
			attrStack: []string{"phones", "{primary", "false", "type", "mobile", "value", "05467983211}"},
			noElems:   1,
		},
		{
			attrStack: []string{"posixaccount", "{accountid", "1000", "gecos", "Brian Phelps", "gid", "1000", "homedirectory", "/home/brian",
				"operatingsystemtype", "linux", "primary", "true", "shell", "/bin/bash", "systemid", "2000",
				"uid", "1000", "username", "brian}"},
			noElems: 1,
		},
		{
			attrStack: []string{"posixaccounts", "{accountid", "1000", "gecos", "Brian Phelps", "gid", "1000", "homedirectory", "/home/brian",
				"operatingsystemtype", "linux", "primary", "true", "shell", "/bin/bash", "systemid", "2000",
				"uid", "1000", "username", "brian}"},
			noElems: 1,
		},
		{
			attrStack: []string{"relation", "{type", "partner", "value", "David Letterman}"},
			noElems:   1,
		},
		{
			attrStack: []string{"relations", "{type", "partner", "value", "David Letterman}"},
			noElems:   1,
		},
		{
			attrStack: []string{"sshpublickey", "{expirationtimeusec", "1625123095000", "key", "id-rsaxxxxxxxxxxxxxxxxxxxxxxxxxx}"},
			noElems:   1,
		},
		{
			attrStack: []string{"sshpublickeys", "{expirationtimeusec", "1625123095000", "key", "id-rsaxxxxxxxxxxxxxxxxxxxxxxxxxx}"},
			noElems:   1,
		},
		{
			attrStack: []string{"website", "{primary", "true", "type", "blog", "value", "blm.org}"},
			noElems:   1,
		},
		{
			attrStack: []string{"websites", "{primary", "true", "type", "blog", "value", "blm.org}"},
			noElems:   1,
		},
	}

	for _, c := range cases {
		user := new(admin.User)

		attrStack := c.attrStack

		newStack, err := doComposite(user, attrStack)

		if err != nil {
			if err.Error() != c.expectedErr {
				t.Errorf("Got error: %v - expected error: %v", err.Error(), c.expectedErr)
			}
			continue
		}

		if newStack != nil && len(newStack) != c.expectedNSLen {
			t.Errorf("Expected newStack length %v - got %v", c.expectedNSLen, len(newStack))
			continue
		}

		switch true {
		case attrStack[0] == "address" || attrStack[0] == "addresses":
			addresses := user.Addresses.([]*admin.UserAddress)
			if len(addresses) != c.noElems {
				t.Errorf("Address error - expected %v address got %v", c.noElems, len(addresses))
			}
		case attrStack[0] == "email" || attrStack[0] == "emails":
			emails := user.Emails.([]*admin.UserEmail)
			if len(emails) != c.noElems {
				t.Errorf("Email error - expected %v email got %v", c.noElems, len(emails))
			}
		case attrStack[0] == "externalid" || attrStack[0] == "externalids":
			externalids := user.ExternalIds.([]*admin.UserExternalId)
			if len(externalids) != c.noElems {
				t.Errorf("ExternalID error - expected %v external id got %v", c.noElems, len(externalids))
			}
		case attrStack[0] == "gender":
			gender := user.Gender.(*admin.UserGender)
			if gender.AddressMeAs == "" || gender.CustomGender == "" || gender.Type == "" {
				t.Error("Gender error: Fields not populated")
			}
		case attrStack[0] == "im" || attrStack[0] == "ims":
			ims := user.Ims.([]*admin.UserIm)
			if len(ims) != c.noElems {
				t.Errorf("Im error - expected %v im got %v", c.noElems, len(ims))
			}
		case attrStack[0] == "keyword" || attrStack[0] == "keywords":
			keywords := user.Keywords.([]*admin.UserKeyword)
			if len(keywords) != c.noElems {
				t.Errorf("Keyword error - expected %v keyword got %v", c.noElems, len(keywords))
			}
		case attrStack[0] == "language" || attrStack[0] == "languages":
			languages := user.Languages.([]*admin.UserLanguage)
			if len(languages) != c.noElems {
				t.Errorf("Language error - expected %v language got %v", c.noElems, len(languages))
			}
		case attrStack[0] == "location" || attrStack[0] == "locations":
			locations := user.Locations.([]*admin.UserLocation)
			if len(locations) != c.noElems {
				t.Errorf("Location error - expected %v location got %v", c.noElems, len(locations))
			}
		case attrStack[0] == "notes":
			about := user.Notes.(*admin.UserAbout)
			if about.ContentType == "" || about.Value == "" {
				t.Error("Notes error: Fields not populated")
			}
		case attrStack[0] == "organisation" || attrStack[0] == "organisations" || attrStack[0] == "organizations" || attrStack[0] == "organization":
			organizations := user.Organizations.([]*admin.UserOrganization)
			if len(organizations) != c.noElems {
				t.Errorf("Organization error - expected %v organization got %v", c.noElems, len(organizations))
			}
		case attrStack[0] == "phone" || attrStack[0] == "phones":
			phones := user.Phones.([]*admin.UserPhone)
			if len(phones) != c.noElems {
				t.Errorf("Phone error - expected %v phone got %v", c.noElems, len(phones))
			}
		case attrStack[0] == "posixaccount" || attrStack[0] == "posixaccounts":
			posixaccounts := user.PosixAccounts.([]*admin.UserPosixAccount)
			if len(posixaccounts) != c.noElems {
				t.Errorf("PosixAccount error - expected %v posixaccount got %v", c.noElems, len(posixaccounts))
			}
		case attrStack[0] == "relation" || attrStack[0] == "relations":
			relations := user.Relations.([]*admin.UserRelation)
			if len(relations) != c.noElems {
				t.Errorf("Relation error - expected %v relation got %v", c.noElems, len(relations))
			}
		case attrStack[0] == "sshpublickey" || attrStack[0] == "sshpublickeys":
			sshpublickeys := user.SshPublicKeys.([]*admin.UserSshPublicKey)
			if len(sshpublickeys) != c.noElems {
				t.Errorf("SshPublicKey error - expected %v sshpublickey got %v", c.noElems, len(sshpublickeys))
			}
		case attrStack[0] == "website" || attrStack[0] == "websites":
			websites := user.Websites.([]*admin.UserWebsite)
			if len(websites) != c.noElems {
				t.Errorf("Website error - expected %v website got %v", c.noElems, len(websites))
			}
		}

	}
}
func TestDoName(t *testing.T) {
	cases := []struct {
		attrStack         []string
		expectedErr       string
		expectedFirstName string
		expectedFullName  string
		expectedLastName  string
		expectedNSLen     int
	}{
		{
			attrStack:         []string{"name", "{firstname", "Arthur", "lastname", "Dent}"},
			expectedErr:       "",
			expectedFirstName: "Arthur",
			expectedFullName:  "",
			expectedLastName:  "Dent",
		},
		{
			attrStack:         []string{"name", "{christianname", "Arthur", "surname", "Dent}"},
			expectedErr:       "",
			expectedFirstName: "Arthur",
			expectedFullName:  "",
			expectedLastName:  "Dent",
		},
		{
			attrStack:         []string{"name", "{firstname", "Arthur", "fullname", "Arthur Dent", "lastname", "Dent}"},
			expectedErr:       "",
			expectedFirstName: "Arthur",
			expectedFullName:  "Arthur Dent",
			expectedLastName:  "Dent",
		},
		{
			attrStack:         []string{"name", "firstname", "Arthur", "lastname", "Dent}"},
			expectedErr:       "",
			expectedFirstName: "Arthur",
			expectedFullName:  "",
			expectedLastName:  "Dent",
		},
		{
			attrStack:   []string{"name", "{firstname", "Arthur", "lastname", "Dent"},
			expectedErr: "gmin: error - malformed name attribute",
		},
		{
			attrStack:         []string{"name", "{FirstName", "Arthur", "FullName", "Arthur Dent", "LASTNAME", "Dent}"},
			expectedErr:       "",
			expectedFirstName: "Arthur",
			expectedFullName:  "Arthur Dent",
			expectedLastName:  "Dent",
		},
		{
			attrStack:         []string{"name", "{firstname", "Arthur", "lastname", "Dent}", "address", "{formatted", "10 Worlds End, Paignton, TQ2 6TF}"},
			expectedErr:       "",
			expectedFirstName: "Arthur",
			expectedFullName:  "",
			expectedLastName:  "Dent",
			expectedNSLen:     3,
		},
		{
			attrStack:   []string{"name", "{firstname", "Arthur", "lastname", "Dent", "address", "{formatted", "10 Worlds End, Paignton, TQ2 6TF}"},
			expectedErr: "gmin: error - malformed attribute string",
		},
	}

	for _, c := range cases {
		name := new(admin.UserName)

		attrStack := c.attrStack

		newStack, err := doName(name, attrStack)

		if err != nil {
			if err.Error() != c.expectedErr {
				t.Errorf("Got error: %v - expected error: %v", err.Error(), c.expectedErr)
			}

			continue
		}

		if len(newStack) != 0 && len(newStack) != c.expectedNSLen {
			t.Errorf("Expected newStack length %v - got %v", c.expectedNSLen, len(newStack))
			continue
		}

		if name.GivenName != c.expectedFirstName || name.FamilyName != c.expectedLastName || name.FullName != c.expectedFullName {
			t.Errorf("Name error - expected firstName: %v; fullName: %v; lastName: %v - got firstName: %v; fullName: %v; lastName: %v",
				c.expectedFirstName, c.expectedFullName, c.expectedLastName, name.GivenName, name.FullName, name.FamilyName)
		}
	}
}
func TestDoNonComposite(t *testing.T) {
	cases := []struct {
		attrStack    []string
		expectedErr  string
		expectedVals map[string]string
	}{
		{
			attrStack:    []string{"changepasswordatnextlogin", "true"},
			expectedVals: map[string]string{"ChangePasswordAtNextLogin": "true"},
		},
		{
			attrStack:    []string{"changepasswordatnextlogin", "false"},
			expectedVals: map[string]string{"ChangePasswordAtNextLogin": "false"},
		},
		{
			attrStack:    []string{"includeinglobaladdresslist", "true"},
			expectedVals: map[string]string{"IncludeInGlobalAddressList": "true"},
		},
		{
			attrStack:    []string{"includeinglobaladdresslist", "false"},
			expectedVals: map[string]string{"IncludeInGlobalAddressList": "false"},
		},
		{
			attrStack:    []string{"ipwhitelisted", "true"},
			expectedVals: map[string]string{"IpWhitelisted": "true"},
		},
		{
			attrStack:    []string{"ipwhitelisted", "false"},
			expectedVals: map[string]string{"IpWhitelisted": "false"},
		},
		{
			attrStack:    []string{"orgunitpath", "/Finance"},
			expectedVals: map[string]string{"OrgUnitPath": "/Finance"},
		},
		{
			attrStack:    []string{"password", "ExtraSecurePassword"},
			expectedVals: map[string]string{"Password": "f04b2e2e92336f5412d4c709749b26e29ea48e2f"},
		},
		{
			attrStack:    []string{"primaryemail", "dick.turpin@famoushighwaymen.com"},
			expectedVals: map[string]string{"PrimaryEmail": "dick.turpin@famoushighwaymen.com"},
		},
		{
			attrStack:    []string{"recoveryemail", "dick.turpin@alternative.com"},
			expectedVals: map[string]string{"RecoveryEmail": "dick.turpin@alternative.com"},
		},
		{
			attrStack:    []string{"recoveryphone", "+447880234167"},
			expectedVals: map[string]string{"RecoveryPhone": "+447880234167"},
		},
		{
			attrStack:   []string{"recoveryphone", "447880234167"},
			expectedErr: "gmin: error - recovery phone number 447880234167 must start with '+'",
		},
		{
			attrStack:    []string{"suspended", "true"},
			expectedVals: map[string]string{"Suspended": "true"},
		},
		{
			attrStack:    []string{"suspended", "false"},
			expectedVals: map[string]string{"Suspended": "false"},
		},
		{
			attrStack:   []string{"bogus", "false"},
			expectedErr: "gmin: error - attribute bogus not recognized",
		},
	}

	for _, c := range cases {
		user := new(admin.User)

		attrStack := c.attrStack

		_, err := doNonComposite(user, attrStack)
		if err != nil {
			if err.Error() != c.expectedErr {
				t.Errorf("Got error: %v - expected error: %v", err.Error(), c.expectedErr)
			}

			continue
		}

		err = tsts.UserCompActualExpected(user, c.expectedVals)
		if err != nil {
			t.Errorf("User NonComposite error: %v", err.Error())
		}
	}
}

func TestFormatAttrs(t *testing.T) {
	cases := []struct {
		attrs         []string
		expectedValue string
		get           bool
	}{
		{
			attrs:         []string{"name(givenName)", "name(familyName)", "primaryEmail"},
			expectedValue: "name(givenName),name(familyName),primaryEmail",
			get:           true,
		},
		{
			attrs:         []string{"name(givenName)", "name(familyName)", "primaryEmail"},
			expectedValue: "users(name(givenName),name(familyName),primaryEmail)",
			get:           false,
		},
		{
			attrs:         []string{"givenName", "familyName", "primaryEmail"},
			expectedValue: "primaryEmail,name(givenName,familyName)",
			get:           true,
		},
		{
			attrs:         []string{"givenName", "familyName", "primaryEmail"},
			expectedValue: "users(primaryEmail,name(givenName,familyName))",
			get:           false,
		},
	}

	for _, c := range cases {
		output := FormatAttrs(c.attrs, c.get)
		if output != c.expectedValue {
			t.Errorf("Expected output: %v - Got: %v", c.expectedValue, output)
		}
	}
}

func TestIsCompositeAttr(t *testing.T) {
	cases := []struct {
		attr           string
		expectedReturn bool
	}{
		{
			attr:           "address",
			expectedReturn: true,
		},
		{
			attr:           "addresses",
			expectedReturn: true,
		},
		{
			attr:           "email",
			expectedReturn: true,
		},
		{
			attr:           "emails",
			expectedReturn: true,
		},
		{
			attr:           "externalid",
			expectedReturn: true,
		},
		{
			attr:           "externalids",
			expectedReturn: true,
		},
		{
			attr:           "gender",
			expectedReturn: true,
		},
		{
			attr:           "im",
			expectedReturn: true,
		},
		{
			attr:           "ims",
			expectedReturn: true,
		},
		{
			attr:           "keyword",
			expectedReturn: true,
		},
		{
			attr:           "keywords",
			expectedReturn: true,
		},
		{
			attr:           "language",
			expectedReturn: true,
		},
		{
			attr:           "languages",
			expectedReturn: true,
		},
		{
			attr:           "location",
			expectedReturn: true,
		},
		{
			attr:           "locations",
			expectedReturn: true,
		},
		{
			attr:           "name",
			expectedReturn: true,
		},
		{
			attr:           "notes",
			expectedReturn: true,
		},
		{
			attr:           "organisation",
			expectedReturn: true,
		},
		{
			attr:           "organisations",
			expectedReturn: true,
		},
		{
			attr:           "organization",
			expectedReturn: true,
		},
		{
			attr:           "organizations",
			expectedReturn: true,
		},
		{
			attr:           "phone",
			expectedReturn: true,
		},
		{
			attr:           "phones",
			expectedReturn: true,
		},
		{
			attr:           "posixaccount",
			expectedReturn: true,
		},
		{
			attr:           "posixaccounts",
			expectedReturn: true,
		},
		{
			attr:           "relation",
			expectedReturn: true,
		},
		{
			attr:           "relations",
			expectedReturn: true,
		},
		{
			attr:           "sshpublickey",
			expectedReturn: true,
		},
		{
			attr:           "sshpublickeys",
			expectedReturn: true,
		},
		{
			attr:           "website",
			expectedReturn: true,
		},
		{
			attr:           "websites",
			expectedReturn: true,
		},
		{
			attr:           "primaryemail",
			expectedReturn: false,
		},
	}

	for _, c := range cases {
		output := isCompositeAttr(c.attr)
		if output != c.expectedReturn {
			t.Errorf("Composite attribute: %v - Expected output: %v - Got: %v", c.attr, c.expectedReturn, output)
		}
	}
}

func TestMakeAbout(t *testing.T) {
	cases := []struct {
		addrParts           []string
		expectedContentType string
		expectedErr         string
		expectedValue       string
	}{
		{
			addrParts:           []string{"contenttype", "text_plain", "value", "This is a test note."},
			expectedContentType: "text_plain",
			expectedValue:       "This is a test note.",
		},
		{
			addrParts:   []string{"contenttype", "text-plain", "value", "This is a test note."},
			expectedErr: "gmin: error - text-plain is not a valid notes content type",
		},
		{
			addrParts:   []string{"content", "text_html", "value", "This is a test note."},
			expectedErr: "gmin: error - content is not a valid UserAbout attribute",
		},
		{
			addrParts:   []string{"content", "text_html", "value"},
			expectedErr: "gmin: error - malformed attribute string",
		},
	}

	for _, c := range cases {
		var about *admin.UserAbout

		about = new(admin.UserAbout)

		about, err := makeAbout(c.addrParts)
		if err != nil && (err.Error() != c.expectedErr) {
			t.Errorf("Expected error: %v - Got: %v", c.expectedErr, err.Error())
			continue
		}

		if about == nil {
			continue
		}

		if about.ContentType != c.expectedContentType || about.Value != c.expectedValue {
			t.Errorf("Expected about.ContentType: %v; about.Value: %v; Got aboutContentType: %v; about.Value: %v",
				c.expectedContentType, c.expectedValue, about.ContentType, about.Value)
		}

	}
}

func TestMakeAddress(t *testing.T) {
	cases := []struct {
		addrParts    []string
		expectedVals map[string]string
		expectedErr  string
	}{
		{
			addrParts: []string{"country", "USA", "countrycode", "USA", "extendedaddress",
				"Pentagram Building, 2301 Lake Shore Drive, Chicago, IL 60616, USA", "formatted", "2301 Lake Shore Drive, Chicago, IL 60616",
				"locality", "Chicago", "postalcode", "IL 60616", "primary", "true", "streetaddress", "2301 Lake Shore Drive", "type", "work"},
			expectedVals: map[string]string{"Country": "USA",
				"CountryCode":     "USA",
				"ExtendedAddress": "Pentagram Building, 2301 Lake Shore Drive, Chicago, IL 60616, USA",
				"Formatted":       "2301 Lake Shore Drive, Chicago, IL 60616",
				"Locality":        "Chicago",
				"PostalCode":      "IL 60616",
				"Primary":         "true",
				"StreetAddress":   "2301 Lake Shore Drive",
				"Type":            "work"},
		},
		{
			addrParts: []string{"country", "USA", "countrycode", "USA", "extendedaddress",
				"Pentagram Building, 2301 Lake Shore Drive, Chicago, IL 60616, USA", "formatted", "2301 Lake Shore Drive, Chicago, IL 60616",
				"locality", "Chicago", "postalcode", "IL 60616", "primary", "True", "streetaddress", "2301 Lake Shore Drive", "type", "work"},
			expectedVals: map[string]string{"Country": "USA",
				"CountryCode":     "USA",
				"ExtendedAddress": "Pentagram Building, 2301 Lake Shore Drive, Chicago, IL 60616, USA",
				"Formatted":       "2301 Lake Shore Drive, Chicago, IL 60616",
				"Locality":        "Chicago",
				"PostalCode":      "IL 60616",
				"Primary":         "true",
				"StreetAddress":   "2301 Lake Shore Drive",
				"Type":            "work"},
		},
		{
			addrParts: []string{"country", "USA", "countrycode", "USA", "extendedaddress",
				"Pentagram Building, 2301 Lake Shore Drive, Chicago, IL 60616, USA", "formatted", "2301 Lake Shore Drive, Chicago, IL 60616",
				"locality", "Chicago", "postalcode", "IL 60616", "primary", "true", "streetaddress", "2301 Lake Shore Drive", "type", "badtype"},
			expectedErr: "gmin: error - badtype is not a valid address type",
		},
		{
			addrParts: []string{"country", "USA", "countrycode", "USA", "customtype", "satellite_office", "extendedaddress",
				"Pentagram Building, 2301 Lake Shore Drive, Chicago, IL 60616, USA", "formatted", "2301 Lake Shore Drive, Chicago, IL 60616",
				"locality", "Chicago", "postalcode", "IL 60616", "primary", "true", "streetaddress", "2301 Lake Shore Drive", "type", "custom"},
			expectedVals: map[string]string{"Country": "USA",
				"CountryCode":     "USA",
				"CustomType":      "satellite_office",
				"ExtendedAddress": "Pentagram Building, 2301 Lake Shore Drive, Chicago, IL 60616, USA",
				"Formatted":       "2301 Lake Shore Drive, Chicago, IL 60616",
				"Locality":        "Chicago",
				"PostalCode":      "IL 60616",
				"Primary":         "true",
				"StreetAddress":   "2301 Lake Shore Drive",
				"Type":            "custom"},
		},
		{
			addrParts: []string{"nation", "USA", "countrycode", "USA", "extendedaddress",
				"Pentagram Building, 2301 Lake Shore Drive, Chicago, IL 60616, USA", "formatted", "2301 Lake Shore Drive, Chicago, IL 60616",
				"locality", "Chicago", "postalcode", "IL 60616", "primary", "true", "streetaddress", "2301 Lake Shore Drive", "type", "badtype"},
			expectedErr: "gmin: error - nation is not a valid UserAddress attribute",
		},
		{
			addrParts: []string{"nation", "USA", "countrycode", "USA", "extendedaddress",
				"Pentagram Building, 2301 Lake Shore Drive, Chicago, IL 60616, USA", "formatted", "2301 Lake Shore Drive, Chicago, IL 60616",
				"locality", "Chicago", "postalcode", "IL 60616", "primary", "true", "streetaddress", "2301 Lake Shore Drive", "type"},
			expectedErr: "gmin: error - malformed attribute string",
		},
	}

	for _, c := range cases {
		var address *admin.UserAddress

		address = new(admin.UserAddress)

		address, err := makeAddress(c.addrParts)
		if err != nil && (err.Error() != c.expectedErr) {
			t.Errorf("Expected error: %v - Got: %v", c.expectedErr, err.Error())
			continue
		}

		err = tsts.AddressCompActualExpected(address, c.expectedVals)
		if err != nil {
			t.Errorf("MakeAddress error: %v", err.Error())
		}
	}
}

func TestMakeEmail(t *testing.T) {
	cases := []struct {
		emailParts   []string
		expectedVals map[string]string
		expectedErr  string
	}{
		{
			emailParts:   []string{"address", "chief.exec@mycorp.com", "primary", "true", "type", "work"},
			expectedVals: map[string]string{"Address": "chief.exec@mycorp.com", "Primary": "true", "Type": "work"},
		},
	}

	for _, c := range cases {
		var email *admin.UserEmail

		email = new(admin.UserEmail)

		email, err := makeEmail(c.emailParts)
		if err != nil && (err.Error() != c.expectedErr) {
			t.Errorf("Expected error: %v - Got: %v", c.expectedErr, err.Error())
			continue
		}

		err = tsts.EmailCompActualExpected(email, c.expectedVals)
		if err != nil {
			t.Errorf("MakeEmail error: %v", err.Error())
		}
	}
}

func TestMakeExtID(t *testing.T) {
	cases := []struct {
		extIDParts   []string
		expectedVals map[string]string
		expectedErr  string
	}{
		{
			extIDParts: []string{"customtype", "GitHubID", "type", "custom", "value", "1234567890"},
			expectedVals: map[string]string{"CustomType": "GitHubID",
				"Type":  "custom",
				"Value": "1234567890"},
		},
	}

	for _, c := range cases {
		var extID *admin.UserExternalId

		extID = new(admin.UserExternalId)

		extID, err := makeExtID(c.extIDParts)
		if err != nil && (err.Error() != c.expectedErr) {
			t.Errorf("Expected error: %v - Got: %v", c.expectedErr, err.Error())
			continue
		}

		err = tsts.ExtIDCompActualExpected(extID, c.expectedVals)
		if err != nil {
			t.Errorf("MakeExtID error: %v", err.Error())
		}
	}
}

func TestMakeGender(t *testing.T) {
	cases := []struct {
		genParts     []string
		expectedVals map[string]string
		expectedErr  string
	}{
		{
			genParts: []string{"addressmeas", "they/them", "customgender", "non-binary", "type", "other"},
			expectedVals: map[string]string{"AddressMeAs": "they/them",
				"CustomGender": "non-binary",
				"Type":         "other"},
		},
	}

	for _, c := range cases {
		var gender *admin.UserGender

		gender = new(admin.UserGender)

		gender, err := makeGender(c.genParts)
		if err != nil && (err.Error() != c.expectedErr) {
			t.Errorf("Expected error: %v - Got: %v", c.expectedErr, err.Error())
			continue
		}

		err = tsts.GenderCompActualExpected(gender, c.expectedVals)
		if err != nil {
			t.Errorf("MakeGender error: %v", err.Error())
		}
	}
}

func TestMakeIm(t *testing.T) {
	cases := []struct {
		imParts      []string
		expectedVals map[string]string
		expectedErr  string
	}{
		{
			imParts: []string{"customprotocol", "plusworx", "customtype", "experimental", "im", "@mistered",
				"primary", "true", "protocol", "custom_protocol", "type", "custom"},
			expectedVals: map[string]string{"CustomProtocol": "plusworx",
				"CustomType": "experimental",
				"Im":         "@mistered",
				"Primary":    "true",
				"Protocol":   "custom_protocol",
				"Type":       "custom"},
		},
	}

	for _, c := range cases {
		var im *admin.UserIm

		im = new(admin.UserIm)

		im, err := makeIm(c.imParts)
		if err != nil && (err.Error() != c.expectedErr) {
			t.Errorf("Expected error: %v - Got: %v", c.expectedErr, err.Error())
			continue
		}

		err = tsts.ImCompActualExpected(im, c.expectedVals)
		if err != nil {
			t.Errorf("MakeIm error: %v", err.Error())
		}
	}
}

func TestMakeKeyword(t *testing.T) {
	cases := []struct {
		keyParts     []string
		expectedVals map[string]string
		expectedErr  string
	}{
		{
			keyParts: []string{"customtype", "workhours", "type", "custom", "value", "part-time"},
			expectedVals: map[string]string{"CustomType": "workhours",
				"Type":  "custom",
				"Value": "part-time"},
		},
	}

	for _, c := range cases {
		var keyword *admin.UserKeyword

		keyword = new(admin.UserKeyword)

		keyword, err := makeKeyword(c.keyParts)
		if err != nil && (err.Error() != c.expectedErr) {
			t.Errorf("Expected error: %v - Got: %v", c.expectedErr, err.Error())
			continue
		}

		err = tsts.KeywordCompActualExpected(keyword, c.expectedVals)
		if err != nil {
			t.Errorf("MakeKeyword error: %v", err.Error())
		}
	}
}

func TestMakeLanguage(t *testing.T) {
	cases := []struct {
		langParts    []string
		expectedVals map[string]string
		expectedErr  string
	}{
		{
			langParts:    []string{"languagecode", "en-GB"},
			expectedVals: map[string]string{"LanguageCode": "en-GB"},
		},
	}

	for _, c := range cases {
		var language *admin.UserLanguage

		language = new(admin.UserLanguage)

		language, err := makeLanguage(c.langParts)
		if err != nil && (err.Error() != c.expectedErr) {
			t.Errorf("Expected error: %v - Got: %v", c.expectedErr, err.Error())
			continue
		}

		err = tsts.LanguageCompActualExpected(language, c.expectedVals)
		if err != nil {
			t.Errorf("MakeLanguage error: %v", err.Error())
		}
	}
}

func TestMakeLocation(t *testing.T) {
	cases := []struct {
		locParts     []string
		expectedVals map[string]string
		expectedErr  string
	}{
		{
			locParts: []string{"area", "Shoreditch", "buildingid", "Grebe House", "deskcode", "D12", "floorname", "12",
				"floorsection", "Marketing", "type", "desk"},
			expectedVals: map[string]string{"Area": "Shoreditch",
				"BuildingId":   "Grebe House",
				"DeskCode":     "D12",
				"FloorName":    "12",
				"FloorSection": "Marketing",
				"Type":         "desk"},
		},
	}

	for _, c := range cases {
		var location *admin.UserLocation

		location = new(admin.UserLocation)

		location, err := makeLocation(c.locParts)
		if err != nil && (err.Error() != c.expectedErr) {
			t.Errorf("Expected error: %v - Got: %v", c.expectedErr, err.Error())
			continue
		}

		err = tsts.LocationCompActualExpected(location, c.expectedVals)
		if err != nil {
			t.Errorf("MakeLocation error: %v", err.Error())
		}
	}
}

func TestMakeOrganization(t *testing.T) {
	cases := []struct {
		orgParts     []string
		expectedVals map[string]string
		expectedErr  string
	}{
		{
			orgParts: []string{"costcenter", "104", "department", "Finance", "description", "Head Office Finance Department",
				"domain", "majestic.co.uk", "fulltimeequivalent", "100000", "location", "Newcastle", "name", "Majestic Film Ltd",
				"primary", "true", "symbol", "MAJ", "title", "CFO", "type", "work"},
			expectedVals: map[string]string{"CostCenter": "104",
				"Department":         "Finance",
				"Description":        "Head Office Finance Department",
				"Domain":             "majestic.co.uk",
				"FullTimeEquivalent": "100000",
				"Location":           "Newcastle",
				"Name":               "Majestic Film Ltd",
				"Primary":            "true",
				"Symbol":             "MAJ",
				"Title":              "CFO",
				"Type":               "work"},
		},
	}

	for _, c := range cases {
		var org *admin.UserOrganization

		org = new(admin.UserOrganization)

		org, err := makeOrganization(c.orgParts)
		if err != nil && (err.Error() != c.expectedErr) {
			t.Errorf("Expected error: %v - Got: %v", c.expectedErr, err.Error())
			continue
		}

		err = tsts.OrgCompActualExpected(org, c.expectedVals)
		if err != nil {
			t.Errorf("MakeOrganization error: %v", err.Error())
		}
	}
}

func TestMakePhone(t *testing.T) {
	cases := []struct {
		phoneParts   []string
		expectedVals map[string]string
		expectedErr  string
	}{
		{
			phoneParts: []string{"primary", "false", "type", "mobile", "value", "05467983211"},
			expectedVals: map[string]string{"Primary": "false",
				"Type":  "mobile",
				"Value": "05467983211"},
		},
	}

	for _, c := range cases {
		var phone *admin.UserPhone

		phone = new(admin.UserPhone)

		phone, err := makePhone(c.phoneParts)
		if err != nil && (err.Error() != c.expectedErr) {
			t.Errorf("Expected error: %v - Got: %v", c.expectedErr, err.Error())
			continue
		}

		err = tsts.PhoneCompActualExpected(phone, c.expectedVals)
		if err != nil {
			t.Errorf("MakePhone error: %v", err.Error())
		}
	}
}

func TestMakePosixAccount(t *testing.T) {
	cases := []struct {
		posixParts   []string
		expectedVals map[string]string
		expectedErr  string
	}{
		{
			posixParts: []string{"accountid", "1000", "gecos", "Brian Phelps", "gid", "1000", "homedirectory", "/home/brian",
				"operatingsystemtype", "linux", "primary", "true", "shell", "/bin/bash", "systemid", "2000",
				"uid", "1000", "username", "brian"},
			expectedVals: map[string]string{"AccountId": "1000",
				"Gecos":               "Brian Phelps",
				"Gid":                 "1000",
				"HomeDirectory":       "/home/brian",
				"OperatingSystemType": "linux",
				"Primary":             "true",
				"Shell":               "/bin/bash",
				"SystemId":            "2000",
				"Uid":                 "1000",
				"Username":            "brian"},
		},
	}

	for _, c := range cases {
		var posix *admin.UserPosixAccount

		posix = new(admin.UserPosixAccount)

		posix, err := makePosAcct(c.posixParts)
		if err != nil && (err.Error() != c.expectedErr) {
			t.Errorf("Expected error: %v - Got: %v", c.expectedErr, err.Error())
			continue
		}

		err = tsts.PosixCompActualExpected(posix, c.expectedVals)
		if err != nil {
			t.Errorf("MakePosAcct error: %v", err.Error())
		}
	}
}

func TestMakeRelation(t *testing.T) {
	cases := []struct {
		relParts     []string
		expectedVals map[string]string
		expectedErr  string
	}{
		{
			relParts: []string{"type", "partner", "value", "David Letterman"},
			expectedVals: map[string]string{"Type": "partner",
				"Value": "David Letterman"},
		},
	}

	for _, c := range cases {
		var relation *admin.UserRelation

		relation = new(admin.UserRelation)

		relation, err := makeRelation(c.relParts)
		if err != nil && (err.Error() != c.expectedErr) {
			t.Errorf("Expected error: %v - Got: %v", c.expectedErr, err.Error())
			continue
		}

		err = tsts.RelationCompActualExpected(relation, c.expectedVals)
		if err != nil {
			t.Errorf("MakeRelation error: %v", err.Error())
		}
	}
}

func TestMakeSSHPublicKey(t *testing.T) {
	cases := []struct {
		sshParts     []string
		expectedVals map[string]string
		expectedErr  string
	}{
		{
			sshParts: []string{"expirationtimeusec", "1625123095000", "key", "id-rsaxxxxxxxxxxxxxxxxxxxxxxxxxx"},
			expectedVals: map[string]string{"ExpirationTimeUsec": "1625123095000",
				"Key": "id-rsaxxxxxxxxxxxxxxxxxxxxxxxxxx"},
		},
	}

	for _, c := range cases {
		var sshkey *admin.UserSshPublicKey

		sshkey = new(admin.UserSshPublicKey)

		sshkey, err := makeSSHPubKey(c.sshParts)
		if err != nil && (err.Error() != c.expectedErr) {
			t.Errorf("Expected error: %v - Got: %v", c.expectedErr, err.Error())
			continue
		}

		err = tsts.SSHKeyCompActualExpected(sshkey, c.expectedVals)
		if err != nil {
			t.Errorf("MakeSSHPubKey error: %v", err.Error())
		}
	}
}

func TestMakeWebsite(t *testing.T) {
	cases := []struct {
		webParts     []string
		expectedVals map[string]string
		expectedErr  string
	}{
		{
			webParts: []string{"primary", "true", "type", "blog", "value", "blm.org"},
			expectedVals: map[string]string{"Primary": "true",
				"Type":  "blog",
				"Value": "blm.org"},
		},
	}

	for _, c := range cases {
		var website *admin.UserWebsite

		website = new(admin.UserWebsite)

		website, err := makeWebsite(c.webParts)
		if err != nil && (err.Error() != c.expectedErr) {
			t.Errorf("Expected error: %v - Got: %v", c.expectedErr, err.Error())
			continue
		}

		err = tsts.WebsiteCompActualExpected(website, c.expectedVals)
		if err != nil {
			t.Errorf("MakeWebsite error: %v", err.Error())
		}
	}
}
