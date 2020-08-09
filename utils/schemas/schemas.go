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
	admin "google.golang.org/api/admin/directory/v1"
	"google.golang.org/api/googleapi"
)

const (
	// EndField is List call attribute string terminator
	EndField = ")"
	// StartSchemasField is List call attribute string prefix
	StartSchemasField = "schemas("
)

// GminSchema is custom admin.Schema struct with no omitempty tags
type GminSchema struct {
	// DisplayName: Display name for the schema.
	DisplayName string `json:"displayName"`

	// Etag: ETag of the resource.
	Etag string `json:"etag"`

	// Fields: Fields of Schema
	Fields []*GminSchemaFieldSpec `json:"fields"`

	// Kind: Kind of resource this is.
	Kind string `json:"kind"`

	// SchemaId: Unique identifier of Schema (Read-only)
	SchemaId string `json:"schemaId"`

	// SchemaName: Schema name
	SchemaName string `json:"schemaName"`

	// ServerResponse contains the HTTP response code and headers from the
	// server.
	googleapi.ServerResponse `json:"-"`

	// ForceSendFields is a list of field names (e.g. "DisplayName") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`

	// NullFields is a list of field names (e.g. "DisplayName") to include
	// in API requests with the JSON null value. By default, fields with
	// empty values are omitted from API requests. However, any field with
	// an empty value appearing in NullFields will be sent to the server as
	// null. It is an error if a field in this list has a non-empty value.
	// This may be used to include null fields in Patch requests.
	NullFields []string `json:"-"`
}

// GminSchemaFieldSpec is custom admin.SchemaFieldSpec struct with no omitempty tags
type GminSchemaFieldSpec struct {
	// DisplayName: Display Name of the field.
	DisplayName string `json:"displayName"`

	// Etag: ETag of the resource.
	Etag string `json:"etag"`

	// FieldId: Unique identifier of Field (Read-only)
	FieldId string `json:"fieldId"`

	// FieldName: Name of the field.
	FieldName string `json:"fieldName"`

	// FieldType: Type of the field.
	FieldType string `json:"fieldType"`

	// Indexed: Boolean specifying whether the field is indexed or not.
	//
	// Default: true
	Indexed *bool `json:"indexed"`

	// Kind: Kind of resource this is.
	Kind string `json:"kind"`

	// MultiValued: Boolean specifying whether this is a multi-valued field
	// or not.
	MultiValued bool `json:"multiValued"`

	// NumericIndexingSpec: Indexing spec for a numeric field. By default,
	// only exact match queries will be supported for numeric fields.
	// Setting the numericIndexingSpec allows range queries to be supported.
	NumericIndexingSpec *GminSchemaFieldSpecNumericIndexingSpec `json:"numericIndexingSpec"`

	// ReadAccessType: Read ACLs on the field specifying who can view values
	// of this field. Valid values are "ALL_DOMAIN_USERS" and
	// "ADMINS_AND_SELF".
	ReadAccessType string `json:"readAccessType"`

	// ForceSendFields is a list of field names (e.g. "DisplayName") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`

	// NullFields is a list of field names (e.g. "DisplayName") to include
	// in API requests with the JSON null value. By default, fields with
	// empty values are omitted from API requests. However, any field with
	// an empty value appearing in NullFields will be sent to the server as
	// null. It is an error if a field in this list has a non-empty value.
	// This may be used to include null fields in Patch requests.
	NullFields []string `json:"-"`
}

// GminSchemaFieldSpecNumericIndexingSpec is a custom admin.GminSchemaFieldSpecNumericIndexingSpec struct with no omitempty tags
type GminSchemaFieldSpecNumericIndexingSpec struct {
	// MaxValue: Maximum value of this field. This is meant to be indicative
	// rather than enforced. Values outside this range will still be
	// indexed, but search may not be as performant.
	MaxValue float64 `json:"maxValue"`

	// MinValue: Minimum value of this field. This is meant to be indicative
	// rather than enforced. Values outside this range will still be
	// indexed, but search may not be as performant.
	MinValue float64 `json:"minValue"`

	// ForceSendFields is a list of field names (e.g. "MaxValue") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`

	// NullFields is a list of field names (e.g. "MaxValue") to include in
	// API requests with the JSON null value. By default, fields with empty
	// values are omitted from API requests. However, any field with an
	// empty value appearing in NullFields will be sent to the server as
	// null. It is an error if a field in this list has a non-empty value.
	// This may be used to include null fields in Patch requests.
	NullFields []string `json:"-"`
}

// GminSchemas is custom admin.Schemas struct containing GminSchema
type GminSchemas struct {
	// Etag: ETag of the resource.
	Etag string `json:"etag,omitempty"`

	// Kind: Kind of resource this is.
	Kind string `json:"kind,omitempty"`

	// Schemas: List of UserSchema objects.
	Schemas []*GminSchema `json:"schemas,omitempty"`

	// ServerResponse contains the HTTP response code and headers from the
	// server.
	googleapi.ServerResponse `json:"-"`

	// ForceSendFields is a list of field names (e.g. "Etag") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`

	// NullFields is a list of field names (e.g. "Etag") to include in API
	// requests with the JSON null value. By default, fields with empty
	// values are omitted from API requests. However, any field with an
	// empty value appearing in NullFields will be sent to the server as
	// null. It is an error if a field in this list has a non-empty value.
	// This may be used to include null fields in Patch requests.
	NullFields []string `json:"-"`
}

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
