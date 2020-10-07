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
	"strings"

	cmn "github.com/plusworx/gmin/utils/common"
	cfg "github.com/plusworx/gmin/utils/config"
	ous "github.com/plusworx/gmin/utils/orgunits"
	"github.com/spf13/cobra"
	admin "google.golang.org/api/admin/directory/v1"
)

var listOUsCmd = &cobra.Command{
	Use:     "orgunits",
	Aliases: []string{"orgunit", "ou", "ous"},
	Args:    cobra.NoArgs,
	Example: `gmin list orgunits -a description~orgunitpath
gmin ls ous -t all`,
	Short: "Outputs a list of orgunits",
	Long:  `Outputs a list of orgunits.`,
	RunE:  doListOUs,
}

func doListOUs(cmd *cobra.Command, args []string) error {
	logger.Debugw("starting doListOUs()",
		"args", args)

	var (
		jsonData []byte
		orgUnits *admin.OrgUnits
	)

	ds, err := cmn.CreateDirectoryService(admin.AdminDirectoryOrgunitReadonlyScope)
	if err != nil {
		logger.Error(err)
		return err
	}

	customerID, err := cfg.ReadConfigString("customerid")
	if err != nil {
		logger.Error(err)
		return err
	}

	oulc := ds.Orgunits.List(customerID)

	if attrs != "" {
		listAttrs, err := cmn.ParseOutputAttrs(attrs, ous.OrgUnitAttrMap)
		if err != nil {
			logger.Error(err)
			return err
		}
		formattedAttrs := ous.StartOrgUnitsField + listAttrs + ous.EndField

		listCall := ous.AddFields(oulc, formattedAttrs)
		oulc = listCall.(*admin.OrgunitsListCall)
	}

	if orgUnit != "" {
		oulc = ous.AddOUPath(oulc, orgUnit)
	}

	searchType = strings.ToLower(searchType)

	ok := cmn.SliceContainsStr(ous.ValidSearchTypes, searchType)
	if !ok {
		err := fmt.Errorf(cmn.ErrInvalidSearchType, searchType)
		logger.Error(err)
		return err
	}
	oulc = ous.AddType(oulc, searchType)

	orgUnits, err = ous.DoList(oulc)
	if err != nil {
		logger.Error(err)
		return err
	}

	jsonData, err = json.MarshalIndent(orgUnits, "", "    ")
	if err != nil {
		logger.Error(err)
		return err
	}

	if count {
		fmt.Println(len(orgUnits.OrganizationUnits))
	} else {
		fmt.Println(string(jsonData))
	}

	logger.Debug("finished doListOUs()")
	return nil
}

func init() {
	listCmd.AddCommand(listOUsCmd)

	listOUsCmd.Flags().StringVarP(&attrs, "attributes", "a", "", "required orgunit attributes separated by (~)")
	listOUsCmd.Flags().BoolVarP(&count, "count", "", false, "count number of entities returned")
	listOUsCmd.Flags().StringVarP(&orgUnit, "orgunit-path", "o", "", "orgunitpath or id of starting orgunit")
	listOUsCmd.Flags().StringVarP(&searchType, "type", "t", "children", "all sub-organizational units or only immediate children")
}
