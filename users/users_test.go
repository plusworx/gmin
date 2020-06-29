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

		if newStack != nil {
			if len(newStack) != c.expectedNSLen {
				t.Errorf("Expected newStack length %v - got %v", c.expectedNSLen, len(newStack))
			}

			continue
		}

		switch true {
		case attrStack[0] == "address":
			addresses := user.Addresses.([]*admin.UserAddress)
			if len(addresses) != 1 {
				t.Errorf("Stack error - expected 1 address got %v", len(addresses))
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
			expectedErr:       "gmin: error - attribute address is unrecognized",
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
			expectedErr: "gmin: error - attribute bogus not recognised",
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
