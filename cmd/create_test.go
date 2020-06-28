package cmd

import (
	"testing"
)

func TestDoCreateUser(t *testing.T) {
	cases := []struct {
		args          []string
		attributes    string
		expected      string
		firstname     string
		lastname      string
		password      string
		recoveryPhone string
	}{
		{
			args:          []string{"mickey.mouse@disney.com"},
			recoveryPhone: "988787686",
			expected:      "gmin: error - recovery phone number 988787686 must start with '+'",
		},
		{
			args:      []string{"mickey.mouse@disney.com"},
			firstname: "Mickey",
			lastname:  "Mouse",
			expected:  "gmin: error - firstname, lastname and password must all be provided",
		},
		{
			args:       []string{"mickey.mouse@disney.com"},
			attributes: "email",
			firstname:  "Mickey",
			lastname:   "Mouse",
			password:   "SuperStrongPassword",
			expected:   "gmin: error - malformed attribute string",
		},
		{
			args:       []string{"mickey.mouse@disney.com"},
			attributes: "jklkjf:kljkjf",
			firstname:  "Mickey",
			lastname:   "Mouse",
			password:   "SuperStrongPassword",
			expected:   "gmin: error - attribute jklkjf not recognised",
		},
	}

	for _, c := range cases {
		attrs = c.attributes
		firstName = c.firstname
		lastName = c.lastname
		password = c.password
		recoveryPhone = c.recoveryPhone

		got := doCreateUser(createUserCmd, c.args)

		if got.Error() != c.expected {
			t.Errorf("Expected error %v, got %v", c.expected, got.Error())
		}
	}
}
