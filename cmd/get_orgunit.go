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
	ous "github.com/plusworx/gmin/utils/orgunits"
	"github.com/spf13/cobra"
	admin "google.golang.org/api/admin/directory/v1"
)

var getOrgUnitCmd = &cobra.Command{
	Use:     "orgunit <orgunit name>",
	Aliases: []string{"ou"},
	Args:    cobra.ExactArgs(1),
	Example: `gmin get orgunit Accounts
gmin get ou Marketing -a name~orgUnitId`,
	Short: "Outputs information about an orgunit",
	Long:  `Outputs information about an orgunit.`,
	RunE:  doGetOrgUnit,
}

func doGetOrgUnit(cmd *cobra.Command, args []string) error {
	lg.Debugw("starting doGetOrgUnit()",
		"args", args)
	defer lg.Debug("finished doGetOrgUnit()")

	var (
		jsonData []byte
		orgUnit  *admin.OrgUnit
	)

	srv, err := cmn.CreateService(cmn.SRVTYPEADMIN, admin.AdminDirectoryOrgunitReadonlyScope)
	if err != nil {
		return err
	}
	ds := srv.(*admin.Service)

	customerID, err := cfg.ReadConfigString(cfg.CONFIGCUSTID)
	if err != nil {
		return err
	}

	ou := args[0]
	if ou[0] == '/' {
		ou = ou[1:]
		args[0] = ou
	}

	ougc := ds.Orgunits.Get(customerID, args[0])

	flgAttrsVal, err := cmd.Flags().GetString(flgnm.FLG_ATTRIBUTES)
	if err != nil {
		lg.Error(err)
		return err
	}
	if flgAttrsVal != "" {
		formattedAttrs, err := gpars.ParseOutputAttrs(flgAttrsVal, ous.OrgUnitAttrMap)
		if err != nil {
			return err
		}
		getCall := ous.AddFields(ougc, formattedAttrs)
		ougc = getCall.(*admin.OrgunitsGetCall)
	}

	orgUnit, err = ous.DoGet(ougc)
	if err != nil {
		return err
	}

	jsonData, err = json.MarshalIndent(orgUnit, "", "    ")
	if err != nil {
		lg.Error(err)
		return err
	}

	fmt.Println(string(jsonData))

	return nil
}

func init() {
	getCmd.AddCommand(getOrgUnitCmd)

	getOrgUnitCmd.Flags().StringVarP(&attrs, flgnm.FLG_ATTRIBUTES, "a", "", "required orgunit attributes (separated by ~)")
}
