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
			expected:   "gmin: error - attribute jklkjf not recognized",
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
