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
	"encoding/json"
	"fmt"

	"github.com/jinzhu/copier"
	cmn "github.com/plusworx/gmin/utils/common"
	cfg "github.com/plusworx/gmin/utils/config"
	scs "github.com/plusworx/gmin/utils/schemas"
	"github.com/spf13/cobra"
	admin "google.golang.org/api/admin/directory/v1"
)

var getSchemaCmd = &cobra.Command{
	Use:     "schema <schema name>",
	Aliases: []string{"sc"},
	Args:    cobra.ExactArgs(1),
	Short:   "Outputs information about a schema",
	Long: `Outputs information about a schema.
	
	Examples:	gmin get schema EmployeeInfo
			gmin get sc EmployeeInfo -a displayName~schemaName`,
	RunE: doGetSchema,
}

func doGetSchema(cmd *cobra.Command, args []string) error {
	var (
		jsonData  []byte
		newSchema = scs.GminSchema{}
		schema    *admin.Schema
	)

	ds, err := cmn.CreateDirectoryService(admin.AdminDirectoryUserschemaReadonlyScope)
	if err != nil {
		return err
	}

	customerID, err := cfg.ReadConfigString("customerid")
	if err != nil {
		return err
	}

	scgc := ds.Schemas.Get(customerID, args[0])

	if attrs != "" {
		formattedAttrs, err := cmn.ParseOutputAttrs(attrs, scs.SchemaAttrMap)
		if err != nil {
			return err
		}
		getCall := scs.AddFields(scgc, formattedAttrs)
		scgc = getCall.(*admin.SchemasGetCall)
	}

	schema, err = scs.DoGet(scgc)
	if err != nil {
		return err
	}

	if attrs == "" {
		copier.Copy(&newSchema, schema)

		jsonData, err = json.MarshalIndent(newSchema, "", "    ")
		if err != nil {
			return err
		}
	} else {
		jsonData, err = json.MarshalIndent(schema, "", "    ")
		if err != nil {
			return err
		}
	}

	fmt.Println(string(jsonData))

	return nil
}

func init() {
	getCmd.AddCommand(getSchemaCmd)

	getSchemaCmd.Flags().StringVarP(&attrs, "attributes", "a", "", "required schema attributes (separated by ~)")
}
