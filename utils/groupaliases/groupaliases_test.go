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

package groupaliases

import (
	"testing"

	tsts "github.com/plusworx/gmin/tests"
	lg "github.com/plusworx/gmin/utils/logging"
	admin "google.golang.org/api/admin/directory/v1"
)

func TestAddFields(t *testing.T) {
	cases := []struct {
		fields string
	}{
		{
			fields: "alias,id,primaryEmail",
		},
	}

	tsts.InitConfig()
	lg.InitLogging("info")

	ds, err := tsts.DummyDirectoryService(admin.AdminDirectoryGroupReadonlyScope)
	if err != nil {
		t.Error("Error: failed to create dummy admin.Service")
	}

	galc := ds.Groups.Aliases.List("testgroup@company.org")

	for _, c := range cases {

		newGALC := AddFields(galc, c.fields)

		if newGALC == nil {
			t.Error("Error: failed to add Fields to GroupsAliasesListCall")
		}
	}
}
