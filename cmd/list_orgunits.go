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
	flgnm "github.com/plusworx/gmin/utils/flagnames"
	gmess "github.com/plusworx/gmin/utils/gminmessages"
	gpars "github.com/plusworx/gmin/utils/gminparsers"
	lg "github.com/plusworx/gmin/utils/logging"
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
	lg.Debugw("starting doListOUs()",
		"args", args)

	var (
		jsonData []byte
		orgUnits *admin.OrgUnits
	)

	ds, err := cmn.CreateDirectoryService(admin.AdminDirectoryOrgunitReadonlyScope)
	if err != nil {
		lg.Error(err)
		return err
	}

	customerID, err := cfg.ReadConfigString(cfg.CONFIGCUSTID)
	if err != nil {
		lg.Error(err)
		return err
	}

	oulc := ds.Orgunits.List(customerID)

	flgAttrsVal, err := cmd.Flags().GetString(flgnm.FLG_ATTRIBUTES)
	if err != nil {
		lg.Error(err)
		return err
	}
	if flgAttrsVal != "" {
		listAttrs, err := gpars.ParseOutputAttrs(flgAttrsVal, ous.OrgUnitAttrMap)
		if err != nil {
			lg.Error(err)
			return err
		}
		formattedAttrs := ous.STARTORGUNITSFIELD + listAttrs + ous.ENDFIELD

		listCall := ous.AddFields(oulc, formattedAttrs)
		oulc = listCall.(*admin.OrgunitsListCall)
	}

	flgOUPathVal, err := cmd.Flags().GetString(flgnm.FLG_ORGUNITPATH)
	if err != nil {
		lg.Error(err)
		return err
	}
	if flgOUPathVal != "" {
		oulc = ous.AddOUPath(oulc, flgOUPathVal)
	}

	flgSearchTypeVal, err := cmd.Flags().GetString(flgnm.FLG_SEARCHTYPE)
	if err != nil {
		lg.Error(err)
		return err
	}
	lowerSearchType := strings.ToLower(flgSearchTypeVal)

	ok := cmn.SliceContainsStr(ous.ValidSearchTypes, lowerSearchType)
	if !ok {
		err := fmt.Errorf(gmess.ERR_INVALIDSEARCHTYPE, flgSearchTypeVal)
		lg.Error(err)
		return err
	}
	oulc = ous.AddType(oulc, lowerSearchType)

	orgUnits, err = ous.DoList(oulc)
	if err != nil {
		lg.Error(err)
		return err
	}

	jsonData, err = json.MarshalIndent(orgUnits, "", "    ")
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
		fmt.Println(len(orgUnits.OrganizationUnits))
	} else {
		fmt.Println(string(jsonData))
	}

	lg.Debug("finished doListOUs()")
	return nil
}

func init() {
	listCmd.AddCommand(listOUsCmd)

	listOUsCmd.Flags().StringVarP(&attrs, flgnm.FLG_ATTRIBUTES, "a", "", "required orgunit attributes separated by (~)")
	listOUsCmd.Flags().BoolVarP(&count, flgnm.FLG_COUNT, "", false, "count number of entities returned")
	listOUsCmd.Flags().StringVarP(&orgUnit, flgnm.FLG_ORGUNITPATH, "o", "", "orgunitpath or id of starting orgunit")
	listOUsCmd.Flags().StringVarP(&searchType, flgnm.FLG_SEARCHTYPE, "t", "children", "all sub-organizational units or only immediate children")
}
