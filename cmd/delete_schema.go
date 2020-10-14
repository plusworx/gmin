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
	"fmt"

	cmn "github.com/plusworx/gmin/utils/common"
	cfg "github.com/plusworx/gmin/utils/config"
	gmess "github.com/plusworx/gmin/utils/gminmessages"
	"github.com/spf13/cobra"
	admin "google.golang.org/api/admin/directory/v1"
)

var deleteSchemaCmd = &cobra.Command{
	Use:     "schema <name or id>",
	Aliases: []string{"sc"},
	Args:    cobra.ExactArgs(1),
	Example: `gmin delete schema TestSchema
gmin del sc TestSchema`,
	Short: "Deletes schema",
	Long:  `Deletes schema.`,
	RunE:  doDeleteSchema,
}

func doDeleteSchema(cmd *cobra.Command, args []string) error {
	logger.Debugw("starting doDeleteSchema()",
		"args", args)

	ds, err := cmn.CreateDirectoryService(admin.AdminDirectoryUserschemaScope)
	if err != nil {
		logger.Error(err)
		return err
	}

	customerID, err := cfg.ReadConfigString("customerid")
	if err != nil {
		logger.Error(err)
		return err
	}

	scdc := ds.Schemas.Delete(customerID, args[0])

	err = scdc.Do()
	if err != nil {
		logger.Error(err)
		return err
	}

	logger.Infof(gmess.INFO_SCHEMADELETED, args[0])
	fmt.Println(cmn.GminMessage(fmt.Sprintf(gmess.INFO_SCHEMADELETED, args[0])))

	logger.Debug("finished doDeleteSchema()")
	return nil
}

func init() {
	deleteCmd.AddCommand(deleteSchemaCmd)
}
