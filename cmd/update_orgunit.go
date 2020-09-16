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
	"github.com/spf13/cobra"
	admin "google.golang.org/api/admin/directory/v1"
)

var updateOUCmd = &cobra.Command{
	Use:     "orgunit <orgunit path or id>",
	Aliases: []string{"ou"},
	Args:    cobra.ExactArgs(1),
	Short:   "Updates an orgunit",
	Long: `Updates an orgunit .
	
	Examples:	gmin update orgunit Sales -n "New Name" -d "New description"
			gmin upd ou Engineering/Aerodynamics -p Engineering/Aeronautics`,
	RunE: doUpdateOU,
}

func doUpdateOU(cmd *cobra.Command, args []string) error {
	var orgunit *admin.OrgUnit

	orgunit = new(admin.OrgUnit)

	if orgUnitName != "" {
		orgunit.Name = orgUnitName
	}

	if blockInherit {
		orgunit.BlockInheritance = true
	}

	if unblockInherit {
		orgunit.BlockInheritance = false
		orgunit.ForceSendFields = append(orgunit.ForceSendFields, "BlockInheritance")
	}

	if orgUnitDesc != "" {
		orgunit.Description = orgUnitDesc
	}

	if parentOUPath != "" {
		orgunit.ParentOrgUnitPath = parentOUPath
	}

	ds, err := cmn.CreateDirectoryService(admin.AdminDirectoryOrgunitScope)
	if err != nil {
		return err
	}

	customerID, err := cfg.ReadConfigString("customerid")
	if err != nil {
		return err
	}

	ouuc := ds.Orgunits.Update(customerID, args[0], orgunit)
	_, err = ouuc.Do()
	if err != nil {
		return err
	}

	fmt.Println(cmn.GminMessage("**** gmin: orgunit " + args[0] + " updated ****"))

	return nil
}

func init() {
	updateCmd.AddCommand(updateOUCmd)

	updateOUCmd.Flags().BoolVarP(&blockInherit, "blockinherit", "b", false, "block orgunit policy inheritance")
	updateOUCmd.Flags().StringVarP(&orgUnitDesc, "description", "d", "", "orgunit description")
	updateOUCmd.Flags().StringVarP(&orgUnitName, "name", "n", "", "orgunit name")
	updateOUCmd.Flags().StringVarP(&parentOUPath, "parentpath", "p", "", "orgunit parent path")
	updateOUCmd.Flags().BoolVarP(&unblockInherit, "unblockinherit", "u", false, "unblock orgunit policy inheritance")

}
