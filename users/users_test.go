package users

import (
	"testing"

	admin "google.golang.org/api/admin/directory/v1"
)

func TestDoComposite(t *testing.T) {
	cases := []struct {
		attrStack   []string
		expectedErr string
		noElems     int
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
	}

	for _, c := range cases {
		user := new(admin.User)

		attrStack := c.attrStack

		_, err := doComposite(user, attrStack)

		if err != nil {
			if err.Error() != c.expectedErr {
				t.Errorf("Got error: %v - expected error: %v", err.Error(), c.expectedErr)
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
