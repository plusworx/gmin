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
	"fmt"
	"testing"

	cmn "github.com/plusworx/gmin/utils/common"
)

func TestDoShowAttrs(t *testing.T) {
	cases := []struct {
		args        []string
		composite   bool
		expectedErr string
		queryable   bool
	}{
		{
			args:        []string{"grp"},
			composite:   true,
			queryable:   true,
			expectedErr: cmn.ErrQueryAndCompositeFlags,
		},
		{
			args:        []string{"user-alias", "email"},
			queryable:   true,
			expectedErr: cmn.ErrQueryableFlag1Arg,
		},
		{
			args:        []string{"unrecognized"},
			expectedErr: fmt.Sprintf(cmn.ErrObjectNotFound, "unrecognized"),
		},
		{
			args:        []string{"schema", "fieldspec", "numericindexingspec"},
			composite:   true,
			expectedErr: fmt.Sprintf(cmn.ErrNoCompositeAttrs, "numericindexingspec"),
		},
		{
			args:        []string{"group", "email", "id"},
			expectedErr: fmt.Sprintf(cmn.ErrNoCompositeAttrs, "group"),
		},
		{
			args:        []string{"cdev", "recentusers"},
			composite:   true,
			expectedErr: fmt.Sprintf(cmn.ErrNoCompositeAttrs, "recentusers"),
		},
		{
			args:        []string{"ou", "name"},
			expectedErr: fmt.Sprintf(cmn.ErrNoCompositeAttrs, "ou"),
		},
		{
			args:        []string{"ga"},
			queryable:   true,
			expectedErr: fmt.Sprintf(cmn.ErrNoQueryableAttrs, "ga"),
		},
		{
			args:        []string{"gmem"},
			composite:   true,
			expectedErr: fmt.Sprintf(cmn.ErrNoCompositeAttrs, "gmem"),
		},
	}

	showCmd.AddCommand(showAttrsCmd)

	for _, c := range cases {
		composite = c.composite
		queryable = c.queryable

		got := doShowAttrs(showAttrsCmd, c.args)

		if got.Error() != c.expectedErr {
			t.Errorf("Expected error %v, got %v", c.expectedErr, got.Error())
		}
	}
}
