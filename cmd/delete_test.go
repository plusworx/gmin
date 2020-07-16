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

func TestDoDeleteGroupAlias(t *testing.T) {
	cases := []struct {
		args        []string
		expectedErr string
		group       string
	}{
		{
			args:        []string{"group.alias@mycompany.org"},
			expectedErr: "gmin: error - group email address or id must be provided",
		},
	}

	for _, c := range cases {
		group = c.group

		got := doDeleteGroupAlias(deleteGroupAliasCmd, c.args)

		if got.Error() != c.expectedErr {
			t.Errorf("Expected error %v, got %v", c.expectedErr, got.Error())
		}
	}
}

func TestDoDeleteMember(t *testing.T) {
	cases := []struct {
		args        []string
		expectedErr string
		group       string
	}{
		{
			args:        []string{"mister.miyagi@mycompany.org"},
			expectedErr: "gmin: error - group email address or id must be provided",
		},
	}

	for _, c := range cases {
		group = c.group

		got := doDeleteMember(deleteMemberCmd, c.args)

		if got.Error() != c.expectedErr {
			t.Errorf("Expected error %v, got %v", c.expectedErr, got.Error())
		}
	}
}

func TestDoDeleteUserAlias(t *testing.T) {
	cases := []struct {
		args        []string
		expectedErr string
		userKey     string
	}{
		{
			args:        []string{"my.alias@mycompany.org"},
			expectedErr: "gmin: error - user email address or id must be provided",
		},
	}

	for _, c := range cases {
		userKey = c.userKey

		got := doDeleteUserAlias(deleteUserAliasCmd, c.args)

		if got.Error() != c.expectedErr {
			t.Errorf("Expected error %v, got %v", c.expectedErr, got.Error())
		}
	}
}
