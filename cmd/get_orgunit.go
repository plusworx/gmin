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
	ous "github.com/plusworx/gmin/utils/orgunits"
	"github.com/spf13/cobra"
	admin "google.golang.org/api/admin/directory/v1"
)

var getOrgUnitCmd = &cobra.Command{
	Use:     "orgunit <orgunit name>",
	Aliases: []string{"ou"},
	Args:    cobra.ExactArgs(1),
	Short:   "Outputs information about an orgunit",
	Long: `Outputs information about an orgunit.
	
	Examples: gmin get orgunit Accounts
	          gmin get orgunit Marketing -a "name orgUnitId"`,
	RunE: doGetOrgUnit,
}

func doGetOrgUnit(cmd *cobra.Command, args []string) error {
	var (
		orgUnit    *admin.OrgUnit
		validAttrs []string
	)

	ds, err := cmn.CreateDirectoryService(admin.AdminDirectoryOrgunitReadonlyScope)
	if err != nil {
		return err
	}

	customerID, err := cfg.ReadConfigString("customerid")
	if err != nil {
		return err
	}

	ougc := ds.Orgunits.Get(customerID, args)

	if attrs != "" {
		validAttrs, err = cmn.ValidateAttrs(attrs, ous.OrgUnitAttrMap)
		if err != nil {
			return err
		}

		formattedAttrs := ous.FormatAttrs(validAttrs)

		orgUnit, err = ous.SingleAttrs(ougc, formattedAttrs)
		if err != nil {
			return err
		}
	} else {
		orgUnit, err = ous.Single(ougc)
		if err != nil {
			return err
		}
	}

	jsonData, err := json.MarshalIndent(orgUnit, "", "    ")
	if err != nil {
		return err
	}

	fmt.Println(string(jsonData))

	return nil
}

func init() {
	getCmd.AddCommand(getOrgUnitCmd)

	getOrgUnitCmd.Flags().StringVarP(&attrs, "attributes", "a", "", "required orgunit attributes (separated by ~)")
}
