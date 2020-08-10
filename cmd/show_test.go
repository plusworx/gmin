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

func TestDoShowAttrs(t *testing.T) {
	cases := []struct {
		args        []string
		composite   bool
		expectedErr string
		queryable   bool
	}{
		{
			args:        []string{"user", "name", "givenname", "anothername"},
			expectedErr: "gmin: error - maximum of 3 arguments exceeded",
		},
		{
			args:        []string{"grp"},
			composite:   true,
			queryable:   true,
			expectedErr: "gmin: error - cannot provide both --composite and --queryable flags",
		},
		{
			args:        []string{"user-alias", "email"},
			queryable:   true,
			expectedErr: "gmin: error - only one argument is allowed with --queryable flag",
		},
		{
			args:        []string{"unrecognized"},
			expectedErr: "gmin: error - unrecognized not found",
		},
		{
			args:        []string{"schema", "fieldspec", "numericindexingspec"},
			composite:   true,
			expectedErr: "gmin: error - numericindexingspec does not have any composite attributes",
		},
		{
			args:        []string{"group", "email", "id"},
			expectedErr: "gmin: error - email does not have any composite attributes",
		},
		{
			args:        []string{"cdev", "recentusers"},
			composite:   true,
			expectedErr: "gmin: error - recentusers does not have any composite attributes",
		},
		{
			args:        []string{"ou", "name"},
			expectedErr: "gmin: error - ou does not have any composite attributes",
		},
		{
			args:        []string{"ga"},
			queryable:   true,
			expectedErr: "gmin: error - ga does not have any queryable attributes",
		},
		{
			args:        []string{"gmem"},
			composite:   true,
			expectedErr: "gmin: error - group members do not have any composite attributes",
		},
	}

	for _, c := range cases {
		composite = c.composite
		queryable = c.queryable

		got := doShowAttrs(showAttrsCmd, c.args)

		if got.Error() != c.expectedErr {
			t.Errorf("Expected error %v, got %v", c.expectedErr, got.Error())
		}
	}
}
