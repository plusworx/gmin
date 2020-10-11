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

package common

import (
	"testing"

	tsts "github.com/plusworx/gmin/tests"
)

func TestHashPassword(t *testing.T) {
	pwd := "MySuperStrongPassword"

	hashedPwd, _ := HashPassword(pwd)

	if hashedPwd != "e1f7c050db42a86e4d358e8c1dcef57e3b4f2fc0" {
		t.Errorf("Expected user.Password to be %v but got %v", "e1f7c050db42a86e4d358e8c1dcef57e3b4f2fc0", hashedPwd)
	}
}

func TestIsValidAttr(t *testing.T) {
	cases := []struct {
		attr          string
		attrMap       map[string]string
		expectedErr   string
		expectedValue string
	}{
		{
			attr:          "admincreated",
			attrMap:       tsts.TestGroupAttrMap,
			expectedErr:   "",
			expectedValue: "adminCreated",
		},
		{
			attr:          "nonexistent",
			attrMap:       tsts.TestGroupAttrMap,
			expectedErr:   "gmin: error - attribute nonexistent is unrecognized",
			expectedValue: "",
		},
	}

	for _, c := range cases {

		output, err := IsValidAttr(c.attr, c.attrMap)

		if err != nil {
			if err.Error() != c.expectedErr {
				t.Errorf("Got error: %v - expected error: %v", err.Error(), c.expectedErr)
			}

			continue
		}

		if output != c.expectedValue {
			t.Errorf("Got output: %v - expected: %v", output, c.expectedValue)
		}

	}
}

func TestSliceContainsStr(t *testing.T) {
	cases := []struct {
		expectedResult bool
		input          string
		sl             []string
	}{
		{
			expectedResult: true,
			input:          "Hello",
			sl:             []string{"Hello", "brave", "new", "world"},
		},
		{
			expectedResult: false,
			input:          "Goodbye",
			sl:             []string{"Hello", "brave", "new", "world"},
		},
	}

	for _, c := range cases {
		res := SliceContainsStr(c.sl, c.input)
		if res != c.expectedResult {
			t.Errorf("Got result: %v - expected result: %v", res, c.expectedResult)
		}
	}
}
