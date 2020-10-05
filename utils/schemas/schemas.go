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

package schemas

import (
	"fmt"
	"sort"
	"strings"

	cmn "github.com/plusworx/gmin/utils/common"
	admin "google.golang.org/api/admin/directory/v1"
	"google.golang.org/api/googleapi"
)

const (
	// EndField is List call attribute string terminator
	EndField = ")"
	// StartSchemasField is List call attribute string prefix
	StartSchemasField = "schemas("
)

// SchemaAttrMap provides lowercase mappings to valid admin.Schema attributes
var SchemaAttrMap = map[string]string{
	"displayname":         "displayName",
	"etag":                "etag",
	"fieldid":             "fieldId",
	"fieldname":           "fieldName",
	"fields":              "fields",
	"fieldtype":           "fieldType",
	"forcesendfields":     "forceSendFields",
	"indexed":             "indexed",
	"kind":                "kind",
	"maxvalue":            "maxValue",
	"minvalue":            "minValue",
	"multivalued":         "multiValued",
	"numericindexingspec": "numericIndexingSpec",
	"readaccesstype":      "readAccessType",
	"schemaid":            "schemaId",
	"schemaname":          "schemaName",
}

var schemaAttrs = []string{
	"displayName",
	"etag",
	"fields",
	"forceSendFields",
	"kind",
	"schemaId",
	"schemaName",
}

var schemaCompAttrs = map[string]string{
	"fields": "fields",
}

var fieldSpecAttrs = []string{
	"displayname",
	"etag",
	"fieldid",
	"fieldname",
	"fieldtype",
	"indexed",
	"kind",
	"multivalued",
	"numericindexingspec",
	"readaccesstype",
}

var schemaFieldSpecCompAttrs = map[string]string{
	"numericindexingspec": "numericIndexingSpec",
}

var schemaFieldSpecNumIdxSpecAttrs = []string{
	"maxvalue",
	"minvalue",
}

// AddFields adds fields to be returned from admin calls
func AddFields(callObj interface{}, attrs string) interface{} {
	var fields googleapi.Field = googleapi.Field(attrs)

	switch callObj.(type) {
	case *admin.SchemasListCall:
		var newSLC *admin.SchemasListCall
		slc := callObj.(*admin.SchemasListCall)
		newSLC = slc.Fields(fields)

		return newSLC
	case *admin.SchemasGetCall:
		var newSGC *admin.SchemasGetCall
		sgc := callObj.(*admin.SchemasGetCall)
		newSGC = sgc.Fields(fields)

		return newSGC
	}

	return nil
}

// DoGet calls the .Do() function on the admin.SchemasGetCall
func DoGet(scgc *admin.SchemasGetCall) (*admin.Schema, error) {
	schema, err := scgc.Do()
	if err != nil {
		return nil, err
	}

	return schema, nil
}

// DoList calls the .Do() function on the admin.SchemasListCall
func DoList(sclc *admin.SchemasListCall) (*admin.Schemas, error) {
	schemas, err := sclc.Do()
	if err != nil {
		return nil, err
	}

	return schemas, nil
}

// ShowAttrs displays requested user attributes
func ShowAttrs(filter string) {
	for _, a := range schemaAttrs {
		lwrA := strings.ToLower(a)
		comp, _ := cmn.IsValidAttr(lwrA, schemaCompAttrs)
		if filter == "" {
			if comp != "" {
				fmt.Println("* ", a)
			} else {
				fmt.Println(a)
			}
			continue
		}

		if strings.Contains(lwrA, strings.ToLower(filter)) {
			if comp != "" {
				fmt.Println("* ", a)
			} else {
				fmt.Println(a)
			}
		}

	}
}

// ShowCompAttrs displays schema composite attributes
func ShowCompAttrs(filter string) {
	keys := make([]string, 0, len(schemaCompAttrs))
	for k := range schemaCompAttrs {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		if filter == "" {
			fmt.Println(schemaCompAttrs[k])
			continue
		}

		if strings.Contains(k, strings.ToLower(filter)) {
			fmt.Println(schemaCompAttrs[k])
		}

	}
}

// ShowSubCompAttrs displays schema field spec composite attributes
func ShowSubCompAttrs(subAttr string, filter string) error {
	if subAttr != "fields" {
		return fmt.Errorf(cmn.ErrInvalidSchemaCompAttr, subAttr)
	}

	keys := make([]string, 0, len(schemaFieldSpecCompAttrs))
	for k := range schemaFieldSpecCompAttrs {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		if filter == "" {
			fmt.Println(schemaFieldSpecCompAttrs[k])
			continue
		}

		if strings.Contains(k, strings.ToLower(filter)) {
			fmt.Println(schemaFieldSpecCompAttrs[k])
		}
	}
	return nil
}

// ShowSubAttrs displays attributes of composite attributes
func ShowSubAttrs(subAttr string, filter string) error {
	subAttrVal := strings.ToLower(subAttr)
	if subAttrVal != "fields" {
		return fmt.Errorf(cmn.ErrNotCompositeAttr, subAttr)
	}

	for _, a := range fieldSpecAttrs {
		lwrA := strings.ToLower(a)
		comp, _ := cmn.IsValidAttr(lwrA, schemaFieldSpecCompAttrs)
		if filter == "" {
			if comp != "" {
				fmt.Println("* ", comp)
			} else {
				fmt.Println(a)
			}
			continue
		}

		if strings.Contains(lwrA, strings.ToLower(filter)) {
			if comp != "" {
				fmt.Println("* ", a)
			} else {
				fmt.Println(a)
			}
		}

	}
	return nil
}

// ShowSubSubAttrs displays attributes of composite attributes
func ShowSubSubAttrs(subAttr string) error {
	if strings.ToLower(subAttr) != "numericindexingspec" {
		return fmt.Errorf(cmn.ErrNotCompositeAttr, subAttr)
	}

	for _, a := range schemaFieldSpecNumIdxSpecAttrs {
		attr, _ := cmn.IsValidAttr(a, SchemaAttrMap)
		fmt.Println(attr)
	}
	return nil
}
