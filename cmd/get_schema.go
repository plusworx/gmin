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

	cmn "github.com/plusworx/gmin/utils/common"
	cfg "github.com/plusworx/gmin/utils/config"
	flgnm "github.com/plusworx/gmin/utils/flagnames"
	gpars "github.com/plusworx/gmin/utils/gminparsers"
	lg "github.com/plusworx/gmin/utils/logging"
	scs "github.com/plusworx/gmin/utils/schemas"
	"github.com/spf13/cobra"
	admin "google.golang.org/api/admin/directory/v1"
)

var getSchemaCmd = &cobra.Command{
	Use:     "schema <schema name>",
	Aliases: []string{"sc"},
	Args:    cobra.ExactArgs(1),
	Example: `gmin get schema EmployeeInfo
gmin get sc EmployeeInfo -a displayName~schemaName`,
	Short: "Outputs information about a schema",
	Long:  `Outputs information about a schema.`,
	RunE:  doGetSchema,
}

func doGetSchema(cmd *cobra.Command, args []string) error {
	lg.Debugw("starting doGetSchema()",
		"args", args)
	defer lg.Debug("finished doGetSchema()")

	var (
		jsonData []byte
		schema   *admin.Schema
	)

	ds, err := cmn.CreateDirectoryService(admin.AdminDirectoryUserschemaReadonlyScope)
	if err != nil {
		return err
	}

	customerID, err := cfg.ReadConfigString(cfg.CONFIGCUSTID)
	if err != nil {
		return err
	}

	scgc := ds.Schemas.Get(customerID, args[0])

	flgAttrsVal, err := cmd.Flags().GetString(flgnm.FLG_ATTRIBUTES)
	if err != nil {
		lg.Error(err)
		return err
	}
	if flgAttrsVal != "" {
		formattedAttrs, err := gpars.ParseOutputAttrs(flgAttrsVal, scs.SchemaAttrMap)
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

	jsonData, err = json.MarshalIndent(schema, "", "    ")
	if err != nil {
		lg.Error(err)
		return err
	}

	fmt.Println(string(jsonData))

	return nil
}

func init() {
	getCmd.AddCommand(getSchemaCmd)

	getSchemaCmd.Flags().StringVarP(&attrs, flgnm.FLG_ATTRIBUTES, "a", "", "required schema attributes (separated by ~)")
}
