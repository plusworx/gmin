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

package members

import "testing"

func TestFormatAttrs(t *testing.T) {
	cases := []struct {
		attrs          []string
		getBool        bool
		expectedOutput string
	}{
		{
			attrs:          []string{"deliverySettings", "email", "role"},
			expectedOutput: "members(deliverySettings,email,role)",
			getBool:        false,
		},
		{
			attrs:          []string{"role", "status", "type"},
			expectedOutput: "role,status,type",
			getBool:        true,
		},
	}

	for _, c := range cases {
		fmtAttrs := FormatAttrs(c.attrs, c.getBool)

		if fmtAttrs != c.expectedOutput {
			t.Errorf("Expected output: %v  Got: %v", c.expectedOutput, fmtAttrs)
		}
	}
}
