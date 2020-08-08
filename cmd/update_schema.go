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
	scs "github.com/plusworx/gmin/utils/schemas"
	"github.com/spf13/cobra"
	admin "google.golang.org/api/admin/directory/v1"
)

var updateSchemaCmd = &cobra.Command{
	Use:     "schema <schema name or id> -i <input file path>",
	Aliases: []string{"sc"},
	Args:    cobra.ExactArgs(1),
	Short:   "Updates a schema",
	Long: `Updates a schema where schema details are provided in a JSON input file.
	
	Examples:	gmin update schema TestSchema -i inputfile.json
			gmin upd sc "TVV0m_kySIOf7bBpcfma8A==" -i inputfile.json
			
	The contents of the JSON file should look something like this:
	
	{
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
                                "displayName": "Employment Start Date",
                                "fieldName": "empStartDate",
                                "fieldType": "DATE",
                                "readAccessType": "ADMINS_AND_SELF"
                        },
                        {
                                "displayName": "Employment End Date",
                                "fieldName": "empEndDate",
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
		"schemaName":"TestSchema"
    }`,
	RunE: doUpdateSchema,
}

func doUpdateSchema(cmd *cobra.Command, args []string) error {
	var schema *admin.Schema

	schema = new(admin.Schema)

	if inputFile == "" {
		err := errors.New("gmin: error - must provide inputfile")
		return err
	}

	fileData, err := ioutil.ReadFile(inputFile)
	if err != nil {
		return err
	}

	if !json.Valid(fileData) {
		return errors.New("gmin: error - input file is not valid JSON")
	}

	outStr, err := cmn.ParseInputAttrs(fileData)
	if err != nil {
		return err
	}

	err = cmn.ValidateInputAttrs(outStr, scs.SchemaAttrMap)
	if err != nil {
		return err
	}

	err = json.Unmarshal(fileData, &schema)
	if err != nil {
		return err
	}

	customerID, err := cfg.ReadConfigString("customerid")
	if err != nil {
		return err
	}

	ds, err := cmn.CreateDirectoryService(admin.AdminDirectoryUserschemaScope)
	if err != nil {
		return err
	}

	scuc := ds.Schemas.Update(customerID, args[0], schema)
	_, err = scuc.Do()
	if err != nil {
		return err
	}

	fmt.Println("**** gmin: schema " + args[0] + " updated ****")

	return nil
}

func init() {
	updateCmd.AddCommand(updateSchemaCmd)

	updateSchemaCmd.Flags().StringVarP(&inputFile, "inputfile", "i", "", "filepath to schema data file")
}
