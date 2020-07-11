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

func TestDoBatchCrtGroup(t *testing.T) {
	cases := []struct {
		expectedErr string
		infile      string
	}{
		{
			expectedErr: "gmin: error - must provide inputfile",
		},
		{
			infile:      "/home/me/nonexistentfile",
			expectedErr: "open /home/me/nonexistentfile: no such file or directory",
		},
	}

	for _, c := range cases {
		inputFile = ""

		if c.infile != "" {
			inputFile = c.infile
		}

		err := doBatchCrtGroup(nil, nil)
		if err.Error() != c.expectedErr {
			t.Errorf("Expected error %v, got %v", c.expectedErr, err)

		}
	}
}

func TestDoBatchCrtMember(t *testing.T) {
	cases := []struct {
		expectedErr string
		group       string
		infile      string
	}{
		{
			expectedErr: "gmin: error - group must be provided",
		},
		{
			infile:      "/home/me/nonexistentfile",
			expectedErr: "gmin: error - group must be provided",
		},
		{
			group:       "testgroup@mycompany.org",
			infile:      "/home/me/nonexistentfile",
			expectedErr: "open /home/me/nonexistentfile: no such file or directory",
		},
	}

	for _, c := range cases {
		inputFile = ""

		if c.infile != "" {
			group = c.group
			inputFile = c.infile
		}

		err := doBatchCrtMember(nil, nil)
		if err.Error() != c.expectedErr {
			t.Errorf("Expected error %v, got %v", c.expectedErr, err)

		}
	}
}

func TestDoBatchCrtOrgUnit(t *testing.T) {
	cases := []struct {
		expectedErr string
		infile      string
	}{
		{
			expectedErr: "gmin: error - must provide inputfile",
		},
		{
			infile:      "/home/me/nonexistentfile",
			expectedErr: "open /home/me/nonexistentfile: no such file or directory",
		},
	}

	for _, c := range cases {
		inputFile = ""

		if c.infile != "" {
			inputFile = c.infile
		}

		err := doBatchCrtOrgUnit(nil, nil)
		if err.Error() != c.expectedErr {
			t.Errorf("Expected error %v, got %v", c.expectedErr, err)

		}
	}
}
