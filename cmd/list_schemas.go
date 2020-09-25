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
	scs "github.com/plusworx/gmin/utils/schemas"
	"github.com/spf13/cobra"
	admin "google.golang.org/api/admin/directory/v1"
)

var listSchemasCmd = &cobra.Command{
	Use:     "schemas",
	Aliases: []string{"schema", "sc", "scs"},
	Args:    cobra.NoArgs,
	Short:   "Outputs a list of schemas",
	Long: `Outputs a list of schemas.
	
	Examples:	gmin list schemas -a displayname~schemaname
			gmin ls scs`,
	RunE: doListSchemas,
}

func doListSchemas(cmd *cobra.Command, args []string) error {
	logger.Debugw("starting doListSchemas()",
		"args", args)

	var (
		jsonData []byte
		schemas  *admin.Schemas
	)

	ds, err := cmn.CreateDirectoryService(admin.AdminDirectoryUserschemaReadonlyScope)
	if err != nil {
		logger.Error(err)
		return err
	}

	customerID, err := cfg.ReadConfigString("customerid")
	if err != nil {
		logger.Error(err)
		return err
	}

	sclc := ds.Schemas.List(customerID)

	if attrs != "" {
		listAttrs, err := cmn.ParseOutputAttrs(attrs, scs.SchemaAttrMap)
		if err != nil {
			logger.Error(err)
			return err
		}
		formattedAttrs := scs.StartSchemasField + listAttrs + scs.EndField

		listCall := scs.AddFields(sclc, formattedAttrs)
		sclc = listCall.(*admin.SchemasListCall)
	}

	schemas, err = scs.DoList(sclc)
	if err != nil {
		logger.Error(err)
		return err
	}

	jsonData, err = json.MarshalIndent(schemas, "", "    ")
	if err != nil {
		logger.Error(err)
		return err
	}

	if count {
		fmt.Println(len(schemas.Schemas))
	} else {
		fmt.Println(string(jsonData))
	}

	logger.Debug("finished doListSchemas()")
	return nil
}

func init() {
	listCmd.AddCommand(listSchemasCmd)

	listSchemasCmd.Flags().StringVarP(&attrs, "attributes", "a", "", "required schema attributes separated by (~)")
	listSchemasCmd.Flags().BoolVarP(&count, "count", "", false, "count number of entities returned")
}
