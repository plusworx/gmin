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
	"errors"
	"fmt"
	"io/ioutil"

	cmn "github.com/plusworx/gmin/utils/common"
	cfg "github.com/plusworx/gmin/utils/config"
	flgnm "github.com/plusworx/gmin/utils/flagnames"
	gmess "github.com/plusworx/gmin/utils/gminmessages"
	scs "github.com/plusworx/gmin/utils/schemas"
	"github.com/spf13/cobra"
	admin "google.golang.org/api/admin/directory/v1"
)

var createSchemaCmd = &cobra.Command{
	Use:     "schema -i <input file path>",
	Aliases: []string{"sc"},
	Example: `gmin create schema -i inputfile.json
gmin crt sc -i inputfile.json`,
	Short: "Creates a schema",
	Long: `Creates a schema where schema details are provided in a JSON input file.
			
The contents of the JSON file should look something like this:

{
	"displayName": "Test Schema",
	"fields": [
		{
			"displayName": "Projects",
			"fieldName": "projects",
			"fieldType": "STRING",
			"multiValued":true,
			"readAccessType": "ADMINS_AND_SELF"
		},
		{
			"displayName": "Location",
			"fieldName": "location",
			"fieldType": "STRING",
			"readAccessType": "ADMINS_AND_SELF"
		},
		{
			"displayName": "Start Date",
			"fieldName": "startDate",
			"fieldType": "DATE",
			"readAccessType": "ADMINS_AND_SELF"
		},
		{
			"displayName": "End Date",
			"fieldName": "endDate",
			"fieldType": "DATE",
			"readAccessType": "ADMINS_AND_SELF"
		},
		{
			"displayName": "Job Level",
			"fieldName": "jobLevel",
			"fieldType": "INT64",
			"indexed":true,
			"numericIndexingSpec":
				{
					"minValue":1,
					"maxValue":7
				},  
			"readAccessType": "ADMINS_AND_SELF"
		}
	],
	"schemaName": "TestSchema"
}`,
	RunE: doCreateSchema,
}

func doCreateSchema(cmd *cobra.Command, args []string) error {
	logger.Debugw("starting doCreateSchema()",
		"args", args)

	var schema *admin.Schema

	schema = new(admin.Schema)

	flgInFileVal, err := cmd.Flags().GetString(flgnm.FLG_INPUTFILE)
	if err != nil {
		logger.Error(err)
		return err
	}

	if flgInFileVal == "" {
		err := errors.New(gmess.ERR_NOINPUTFILE)
		logger.Error(err)
		return err
	}

	fileData, err := ioutil.ReadFile(flgInFileVal)
	if err != nil {
		logger.Error(err)
		return err
	}

	if !json.Valid(fileData) {
		err = errors.New(gmess.ERR_INVALIDJSONFILE)
		logger.Error(err)
		return err
	}

	outStr, err := cmn.ParseInputAttrs(fileData)
	if err != nil {
		logger.Error(err)
		return err
	}

	err = cmn.ValidateInputAttrs(outStr, scs.SchemaAttrMap)
	if err != nil {
		logger.Error(err)
		return err
	}

	err = json.Unmarshal(fileData, &schema)
	if err != nil {
		logger.Error(err)
		return err
	}

	customerID, err := cfg.ReadConfigString(cfg.CONFIGCUSTID)
	if err != nil {
		logger.Error(err)
		return err
	}

	ds, err := cmn.CreateDirectoryService(admin.AdminDirectoryUserschemaScope)
	if err != nil {
		logger.Error(err)
		return err
	}

	scic := ds.Schemas.Insert(customerID, schema)
	newSchema, err := scic.Do()
	if err != nil {
		logger.Error(err)
		return err
	}

	logger.Infof(gmess.INFO_SCHEMACREATED, newSchema.SchemaName)
	fmt.Println(cmn.GminMessage(fmt.Sprintf(gmess.INFO_SCHEMACREATED, newSchema.SchemaName)))

	logger.Debug("finished doCreateSchema()")
	return nil
}

func init() {
	createCmd.AddCommand(createSchemaCmd)

	createSchemaCmd.Flags().StringVarP(&inputFile, flgnm.FLG_INPUTFILE, "i", "", "filepath to schema data file")
	createSchemaCmd.MarkFlagRequired(flgnm.FLG_INPUTFILE)
}
