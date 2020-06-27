package cmd

import (
	"testing"
)

func TestDoCreateUser(t *testing.T) {
	cases := []struct {
		args     []string
		expected string
	}{
		{

			args:     []string{"mickey.mouse@dev.plusworx.uk", "-f", "Mickey"},
			expected: "gmin: error - firstname, lastname and password must all be provided",
		},
	}

	for _, c := range cases {
		got := doCreateUser(createUserCmd, c.args)

		if got.Error() != c.expected {
			t.Errorf("Expected error %v, got %v", c.expected, got.Error())
		}
	}

}
