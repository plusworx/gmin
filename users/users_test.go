package users

import (
	"strconv"
	"testing"

	admin "google.golang.org/api/admin/directory/v1"
)

func TestDoComposite(t *testing.T) {
	cases := []struct {
		attrStack     []string
		expectedErr   string
		expectedNSLen int
		noElems       int
	}{
		{
			attrStack:   []string{"address", "{streetaddress", "201 Arbour Avenue", "locality", "Leeds", "postalcode", "LS2 1ND", "type", "home}"},
			expectedErr: "",
			noElems:     1,
		},
		{
			attrStack:   []string{"address", "{streetaddress", "201 Arbour Avenue", "locality", "Leeds", "postalcode", "LS2 1ND", "type}"},
			expectedErr: "gmin: error - malformed attribute string",
			noElems:     1,
		},
		{
			attrStack:   []string{"address", "streetaddress", "201 Arbour Avenue", "locality", "Leeds", "postalcode", "LS2 1ND", "type", "home}"},
			expectedErr: "",
			noElems:     1,
		},
		{
			attrStack:   []string{"address", "{streetaddress", "201 Arbour Avenue", "locality", "Leeds", "postalcode", "LS2 1ND", "type", "home"},
			expectedErr: "gmin: error - malformed attribute string",
		},
		{
			attrStack:   []string{"address", "{roadaddress", "201 Arbour Avenue", "locality", "Leeds", "postalcode", "LS2 1ND", "type", "home}"},
			expectedErr: "gmin: error - attribute roadaddress is unrecognized",
		},
		{
			attrStack:   []string{"address", "{StreetAddress", "201 Arbour Avenue", "Locality", "Leeds", "PostalCode", "LS2 1ND", "TYPE", "home}"},
			expectedErr: "",
			noElems:     1,
		},
		{
			attrStack:   []string{"address", "{streetaddress", "201 Arbour Avenue", "locality", "Leeds", "postalcode", "LS2 1ND", "type", "wrong}"},
			expectedErr: "gmin: error - wrong is not a valid address type",
		},
		{
			attrStack: []string{"   address   ", "{   streetaddress   ", "201 Arbour Avenue", "   locality  ", "Leeds",
				"   postalcode   ", "LS2 1ND", "   type   ", "home}"},
			expectedErr: "",
			noElems:     1,
		},
		{
			attrStack:   []string{"email", "{address", "chief.exec@mycorp.com", "primary", "true", "type", "work}"},
			expectedErr: "",
			noElems:     1,
		},
		{
			attrStack:   []string{"externalid", "{customtype", "GitHubID", "type", "custom", "value", "1234567890}"},
			expectedErr: "",
			noElems:     1,
		},
		{
			attrStack:   []string{"gender", "{addressmeas", "they/them", "customgender", "non-binary", "type", "other}"},
			expectedErr: "",
		},
		{
			attrStack: []string{"im", "{customprotocol", "plusworx", "customtype", "experimental", "im", "@mistered",
				"primary", "true", "protocol", "custom_protocol", "type", "custom}"},
			expectedErr: "",
			noElems:     1,
		},
		{
			attrStack:   []string{"keyword", "{customtype", "workhours", "type", "custom", "value", "part-time}"},
			expectedErr: "",
			noElems:     1,
		},
		{
			attrStack:   []string{"language", "{languagecode", "en-GB}"},
			expectedErr: "",
			noElems:     1,
		},
		{
			attrStack: []string{"location", "{area", "Shoreditch", "buildingid", "Grebe House", "deskcode", "D12", "floorname", "12",
				"floorsection", "Marketing", "type", "desk}"},
			expectedErr: "",
			noElems:     1,
		},
		{
			attrStack:   []string{"notes", "{contenttype", "text_plain", "value", "This user is one of the company founders.}"},
			expectedErr: "",
		},
		{
			attrStack: []string{"organisation", "{costcenter", "104", "department", "Finance", "description", "Head Office Finance Department",
				"domain", "majestic.co.uk", "fulltimeequivalent", "100000", "location", "Newcastle", "name", "Majestic Film Ltd",
				"primary", "true", "symbol", "MAJ", "title", "CFO", "type", "work}"},
			expectedErr: "",
			noElems:     1,
		},
		{
			attrStack: []string{"organization", "{costcenter", "105", "customtype", "acquisition", "department", "Sales", "description", "Head Office Sales Department",
				"domain", "majestic.co.uk", "fulltimeequivalent", "90000", "location", "Newcastle", "name", "Majestic Film Ltd",
				"primary", "false", "symbol", "MAJ", "title", "Head of Sales", "type", "custom}"},
			expectedErr: "",
			noElems:     1,
		},
		{
			attrStack:   []string{"phone", "{primary", "false", "type", "mobile", "value", "05467983211}"},
			expectedErr: "",
			noElems:     1,
		},
		{
			attrStack: []string{"posixaccount", "{accountid", "1000", "gecos", "Brian Phelps", "gid", "1000", "homedirectory", "/home/brian",
				"operatingsystemtype", "linux", "primary", "true", "shell", "/bin/bash", "systemid", "2000",
				"uid", "1000", "username", "brian}"},
			expectedErr: "",
			noElems:     1,
		},
		{
			attrStack:   []string{"relation", "{type", "partner", "value", "David Letterman}"},
			expectedErr: "",
			noElems:     1,
		},
		{
			attrStack:   []string{"sshpublickey", "{expirationtimeusec", "1625123095000", "key", "id-rsaxxxxxxxxxxxxxxxxxxxxxxxxxx}"},
			expectedErr: "",
			noElems:     1,
		},
		{
			attrStack:   []string{"website", "{primary", "true", "type", "blog", "value", "blm.org}"},
			expectedErr: "",
			noElems:     1,
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
		case attrStack[0] == "address":
			addresses := user.Addresses.([]*admin.UserAddress)
			if len(addresses) != c.noElems {
				t.Errorf("Address error - expected %v address got %v", c.noElems, len(addresses))
			}
		case attrStack[0] == "email":
			emails := user.Emails.([]*admin.UserEmail)
			if len(emails) != c.noElems {
				t.Errorf("Email error - expected %v email got %v", c.noElems, len(emails))
			}
		case attrStack[0] == "externalid":
			externalids := user.ExternalIds.([]*admin.UserExternalId)
			if len(externalids) != c.noElems {
				t.Errorf("ExternalID error - expected %v external id got %v", c.noElems, len(externalids))
			}
		case attrStack[0] == "gender":
			gender := user.Gender.(*admin.UserGender)
			if gender.AddressMeAs != "they/them" || gender.CustomGender != "non-binary" || gender.Type != "other" {
				t.Errorf("Gender error - expected: addressmeas = they/them; customgender = non-binary; type = other; got: addressmeas = %v; customgender = %v; type = %v",
					gender.AddressMeAs, gender.CustomGender, gender.Type)
			}
		case attrStack[0] == "im":
			ims := user.Ims.([]*admin.UserIm)
			if len(ims) != c.noElems {
				t.Errorf("Im error - expected %v im got %v", c.noElems, len(ims))
			}
		case attrStack[0] == "keyword":
			keywords := user.Keywords.([]*admin.UserKeyword)
			if len(keywords) != c.noElems {
				t.Errorf("Keyword error - expected %v keyword got %v", c.noElems, len(keywords))
			}
		case attrStack[0] == "language":
			languages := user.Languages.([]*admin.UserLanguage)
			if len(languages) != c.noElems {
				t.Errorf("Language error - expected %v language got %v", c.noElems, len(languages))
			}
		case attrStack[0] == "location":
			locations := user.Locations.([]*admin.UserLocation)
			if len(locations) != c.noElems {
				t.Errorf("Location error - expected %v location got %v", c.noElems, len(locations))
			}
		case attrStack[0] == "notes":
			about := user.Notes.(*admin.UserAbout)
			if about.ContentType != "text_plain" || about.Value != "This user is one of the company founders." {
				t.Errorf("Notes error - expected: contenttype = text_plain; value = This user is one of the company founders. got: contenttype = %v; value = %v",
					about.ContentType, about.Value)
			}
		case attrStack[0] == "organisation" || attrStack[0] == "organization":
			organizations := user.Organizations.([]*admin.UserOrganization)
			if len(organizations) != c.noElems {
				t.Errorf("Organization error - expected %v organization got %v", c.noElems, len(organizations))
			}
		case attrStack[0] == "phone":
			phones := user.Phones.([]*admin.UserPhone)
			if len(phones) != c.noElems {
				t.Errorf("Phone error - expected %v phone got %v", c.noElems, len(phones))
			}
		case attrStack[0] == "posixaccount":
			posixaccounts := user.PosixAccounts.([]*admin.UserPosixAccount)
			if len(posixaccounts) != c.noElems {
				t.Errorf("Posixaccount error - expected %v posixaccount got %v", c.noElems, len(posixaccounts))
			}
		case attrStack[0] == "relation":
			relations := user.Relations.([]*admin.UserRelation)
			if len(relations) != c.noElems {
				t.Errorf("Relation error - expected %v relation got %v", c.noElems, len(relations))
			}
		case attrStack[0] == "sshpublickey":
			sshpublickeys := user.SshPublicKeys.([]*admin.UserSshPublicKey)
			if len(sshpublickeys) != c.noElems {
				t.Errorf("SshPublicKey error - expected %v sshpublickey got %v", c.noElems, len(sshpublickeys))
			}
		case attrStack[0] == "website":
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
			attrStack:         []string{"name", "{firstname", "Arthur", "fullname", "Algernon", "lastname", "Dent}"},
			expectedErr:       "",
			expectedFirstName: "Arthur",
			expectedFullName:  "Algernon",
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
			attrStack:         []string{"name", "{firstname", "Arthur", "lastname", "Dent"},
			expectedErr:       "gmin: error - malformed name attribute",
			expectedFirstName: "Arthur",
			expectedFullName:  "",
			expectedLastName:  "Dent",
		},
		{
			attrStack:         []string{"name", "{FirstName", "Arthur", "FullName", "Algernon", "LASTNAME", "Dent}"},
			expectedErr:       "",
			expectedFirstName: "Arthur",
			expectedFullName:  "Algernon",
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
			attrStack:         []string{"name", "{firstname", "Arthur", "lastname", "Dent", "address", "{formatted", "10 Worlds End, Paignton, TQ2 6TF}"},
			expectedErr:       "gmin: error - malformed attribute string",
			expectedFirstName: "Arthur",
			expectedFullName:  "",
			expectedLastName:  "Dent",
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
		attrStack     []string
		expectedErr   string
		expectedValue string
		expectedNSLen int
	}{
		{
			attrStack:     []string{"changepasswordatnextlogin", "true"},
			expectedErr:   "",
			expectedValue: "true",
		},
		{
			attrStack:     []string{"changepasswordatnextlogin", "false"},
			expectedErr:   "",
			expectedValue: "false",
		},
		{
			attrStack:     []string{"includeinglobaladdresslist", "true"},
			expectedErr:   "",
			expectedValue: "true",
		},
		{
			attrStack:     []string{"includeinglobaladdresslist", "false"},
			expectedErr:   "",
			expectedValue: "false",
		},
		{
			attrStack:     []string{"ipwhitelisted", "true"},
			expectedErr:   "",
			expectedValue: "true",
		},
		{
			attrStack:     []string{"ipwhitelisted", "false"},
			expectedErr:   "",
			expectedValue: "false",
		},
		{
			attrStack:     []string{"orgunitpath", "/Finance"},
			expectedErr:   "",
			expectedValue: "/Finance",
		},
		{
			attrStack:     []string{"password", "ExtraSecurePassword"},
			expectedErr:   "",
			expectedValue: "f04b2e2e92336f5412d4c709749b26e29ea48e2f",
		},
		{
			attrStack:     []string{"primaryemail", "dick.turpin@famoushighwaymen.com"},
			expectedErr:   "",
			expectedValue: "dick.turpin@famoushighwaymen.com",
		},
		{
			attrStack:     []string{"recoveryemail", "dick.turpin@alternative.com"},
			expectedErr:   "",
			expectedValue: "dick.turpin@alternative.com",
		},
		{
			attrStack:     []string{"recoveryphone", "+447880234167"},
			expectedErr:   "",
			expectedValue: "+447880234167",
		},
		{
			attrStack:   []string{"recoveryphone", "447880234167"},
			expectedErr: "gmin: error - recovery phone number 447880234167 must start with '+'",
		},
		{
			attrStack:     []string{"suspended", "true"},
			expectedErr:   "",
			expectedValue: "true",
		},
		{
			attrStack:     []string{"suspended", "false"},
			expectedErr:   "",
			expectedValue: "false",
		},
		{
			attrStack:   []string{"bogus", "false"},
			expectedErr: "gmin: error - attribute bogus not recognized",
		},
	}

	for _, c := range cases {
		user := new(admin.User)

		attrStack := c.attrStack

		newStack, err := doNonComposite(user, attrStack)

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

		switch true {
		case attrStack[0] == "changepasswordatnextlogin":
			b, _ := strconv.ParseBool(c.expectedValue)

			if b != user.ChangePasswordAtNextLogin {
				t.Errorf("Expected user.ChangePasswordAtNextLogin to be %v but got %v", b, user.ChangePasswordAtNextLogin)
			}
		case attrStack[0] == "includeinglobaladdresslist":
			b, _ := strconv.ParseBool(c.expectedValue)

			if b != user.IncludeInGlobalAddressList {
				t.Errorf("Expected user.IncludeInGlobalAddressList to be %v but got %v", b, user.IncludeInGlobalAddressList)
			}
		case attrStack[0] == "ipwhitelisted":
			b, _ := strconv.ParseBool(c.expectedValue)

			if b != user.IpWhitelisted {
				t.Errorf("Expected user.IpWhitelisted to be %v but got %v", b, user.IpWhitelisted)
			}
		case attrStack[0] == "orgunitpath":
			if user.OrgUnitPath != c.expectedValue {
				t.Errorf("Expected user.OrgUnitPath to be %v but got %v", c.expectedValue, user.OrgUnitPath)
			}
		case attrStack[0] == "password":
			if user.Password != c.expectedValue {
				t.Errorf("Expected user.Password to be %v but got %v", c.expectedValue, user.Password)
			}
		case attrStack[0] == "primaryemail":
			if user.PrimaryEmail != c.expectedValue {
				t.Errorf("Expected user.PrimaryEmail to be %v but got %v", c.expectedValue, user.PrimaryEmail)
			}
		case attrStack[0] == "recoveryemail":
			if user.RecoveryEmail != c.expectedValue {
				t.Errorf("Expected user.RecoveryEmail to be %v but got %v", c.expectedValue, user.RecoveryEmail)
			}
		case attrStack[0] == "recoveryphone":
			if user.RecoveryPhone != c.expectedValue {
				t.Errorf("Expected user.RecoveryPhone to be %v but got %v", c.expectedValue, user.RecoveryPhone)
			}
		case attrStack[0] == "suspended":
			b, _ := strconv.ParseBool(c.expectedValue)

			if b != user.Suspended {
				t.Errorf("Expected user.Suspended to be %v but got %v", b, user.Suspended)
			}
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
			attr:           "email",
			expectedReturn: true,
		},
		{
			attr:           "externalid",
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
			attr:           "keyword",
			expectedReturn: true,
		},
		{
			attr:           "language",
			expectedReturn: true,
		},
		{
			attr:           "location",
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
			attr:           "organization",
			expectedReturn: true,
		},
		{
			attr:           "phone",
			expectedReturn: true,
		},
		{
			attr:           "posixaccount",
			expectedReturn: true,
		},
		{
			attr:           "relation",
			expectedReturn: true,
		},
		{
			attr:           "sshpublickey",
			expectedReturn: true,
		},
		{
			attr:           "website",
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
		aboutParts          []string
		expectedContentType string
		expectedErr         string
		expectedValue       string
	}{
		{
			aboutParts:          []string{"contenttype", "text_plain", "value", "This is a test note."},
			expectedContentType: "text_plain",
			expectedValue:       "This is a test note.",
		},
		{
			aboutParts:  []string{"contenttype", "text-plain", "value", "This is a test note."},
			expectedErr: "gmin: error - text-plain is not a valid notes content type",
		},
		{
			aboutParts:  []string{"content", "text_html", "value", "This is a test note."},
			expectedErr: "gmin: error - content is not a valid UserAbout attribute",
		},
	}

	for _, c := range cases {
		var about *admin.UserAbout

		about = new(admin.UserAbout)

		about, err := makeAbout(c.aboutParts)
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
