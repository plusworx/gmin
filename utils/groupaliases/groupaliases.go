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
	"fmt"
	"sort"
	"strings"

	admin "google.golang.org/api/admin/directory/v1"
	"google.golang.org/api/googleapi"
)

const (
	// EndField is List call attribute string terminator
	EndField string = ")"
	// StartAliasesField is List call attribute string prefix
	StartAliasesField string = "aliases("
)

// GroupAliasAttrMap provides lowercase mappings to valid admin.Alias attributes
var GroupAliasAttrMap = map[string]string{
	"alias":        "alias",
	"etag":         "etag",
	"id":           "id",
	"kind":         "kind",
	"primaryemail": "primaryEmail",
}

// AddFields adds Fields to admin calls
func AddFields(galc *admin.GroupsAliasesListCall, attrs string) *admin.GroupsAliasesListCall {
	var fields googleapi.Field = googleapi.Field(attrs)
	var newGALC *admin.GroupsAliasesListCall

	newGALC = galc.Fields(fields)

	return newGALC
}

// DoList calls the .Do() function on the admin.GroupsAliasesListCall
func DoList(galc *admin.GroupsAliasesListCall) (*admin.Aliases, error) {
	aliases, err := galc.Do()
	if err != nil {
		return nil, err
	}

	return aliases, nil
}

// ShowAttrs displays requested group alias attributes
func ShowAttrs(filter string) {
	keys := make([]string, 0, len(GroupAliasAttrMap))
	for k := range GroupAliasAttrMap {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		if filter == "" {
			fmt.Println(GroupAliasAttrMap[k])
			continue
		}

		if strings.Contains(k, strings.ToLower(filter)) {
			fmt.Println(GroupAliasAttrMap[k])
		}

	}
}
