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

var listSchemasCmd = &cobra.Command{
	Use:     "schemas",
	Aliases: []string{"schema", "sc", "scs"},
	Args:    cobra.NoArgs,
	Example: `gmin list schemas -a displayname~schemaname
gmin ls scs`,
	Short: "Outputs a list of schemas",
	Long:  `Outputs a list of schemas.`,
	RunE:  doListSchemas,
}

func doListSchemas(cmd *cobra.Command, args []string) error {
	lg.Debugw("starting doListSchemas()",
		"args", args)
	defer lg.Debug("finished doListSchemas()")

	var (
		jsonData []byte
		schemas  *admin.Schemas
	)

	ds, err := cmn.CreateDirectoryService(admin.AdminDirectoryUserschemaReadonlyScope)
	if err != nil {
		return err
	}

	customerID, err := cfg.ReadConfigString(cfg.CONFIGCUSTID)
	if err != nil {
		return err
	}

	sclc := ds.Schemas.List(customerID)

	flgAttrsVal, err := cmd.Flags().GetString(flgnm.FLG_ATTRIBUTES)
	if err != nil {
		lg.Error(err)
		return err
	}
	if flgAttrsVal != "" {
		listAttrs, err := gpars.ParseOutputAttrs(flgAttrsVal, scs.SchemaAttrMap)
		if err != nil {
			return err
		}
		formattedAttrs := scs.STARTSCHEMASFIELD + listAttrs + scs.ENDFIELD

		listCall := scs.AddFields(sclc, formattedAttrs)
		sclc = listCall.(*admin.SchemasListCall)
	}

	schemas, err = scs.DoList(sclc)
	if err != nil {
		return err
	}

	jsonData, err = json.MarshalIndent(schemas, "", "    ")
	if err != nil {
		lg.Error(err)
		return err
	}

	flgCountVal, err := cmd.Flags().GetBool(flgnm.FLG_COUNT)
	if err != nil {
		lg.Error(err)
		return err
	}
	if flgCountVal {
		fmt.Println(len(schemas.Schemas))
	} else {
		fmt.Println(string(jsonData))
	}

	return nil
}

func init() {
	listCmd.AddCommand(listSchemasCmd)

	listSchemasCmd.Flags().StringVarP(&attrs, flgnm.FLG_ATTRIBUTES, "a", "", "required schema attributes separated by (~)")
	listSchemasCmd.Flags().BoolVarP(&count, flgnm.FLG_COUNT, "", false, "count number of entities returned")
}
