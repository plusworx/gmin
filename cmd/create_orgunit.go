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
	flgnm "github.com/plusworx/gmin/utils/flagnames"
	gmess "github.com/plusworx/gmin/utils/gminmessages"
	lg "github.com/plusworx/gmin/utils/logging"
	"github.com/spf13/cobra"
	admin "google.golang.org/api/admin/directory/v1"
)

var createOUCmd = &cobra.Command{
	Use:     "orgunit <orgunit name>",
	Aliases: []string{"ou"},
	Args:    cobra.ExactArgs(1),
	Example: `gmin create orgunit Sales -d "Sales Department"
gmin crt ou Finance -d "Finance Department"`,
	Short: "Creates an orgunit",
	Long:  `Creates an orgunit.`,
	RunE:  doCreateOU,
}

func doCreateOU(cmd *cobra.Command, args []string) error {
	lg.Debugw("starting doCreateOU()",
		"args", args)
	defer lg.Debug("finished doCreateOU()")

	var orgunit *admin.OrgUnit

	orgunit = new(admin.OrgUnit)

	orgunit.Name = args[0]

	flgBlkInheritVal, err := cmd.Flags().GetBool(flgnm.FLG_BLOCKINHERIT)
	if err != nil {
		lg.Error(err)
		return err
	}

	if flgBlkInheritVal {
		orgunit.BlockInheritance = true
	}

	flgDescVal, err := cmd.Flags().GetString(flgnm.FLG_DESCRIPTION)
	if err != nil {
		lg.Error(err)
		return err
	}

	if flgDescVal != "" {
		orgunit.Description = flgDescVal
	}

	flgParPthVal, err := cmd.Flags().GetString(flgnm.FLG_PARENTPATH)
	if err != nil {
		lg.Error(err)
		return err
	}
	if flgParPthVal != "" {
		orgunit.ParentOrgUnitPath = flgParPthVal
	} else {
		orgunit.ParentOrgUnitPath = "/"
	}

	ds, err := cmn.CreateDirectoryService(admin.AdminDirectoryOrgunitScope)
	if err != nil {
		return err
	}

	customerID, err := cfg.ReadConfigString(cfg.CONFIGCUSTID)
	if err != nil {
		return err
	}

	ouic := ds.Orgunits.Insert(customerID, orgunit)
	newOrgUnit, err := ouic.Do()
	if err != nil {
		lg.Error(err)
		return err
	}

	lg.Infof(gmess.INFO_OUCREATED, newOrgUnit.OrgUnitPath)
	fmt.Println(cmn.GminMessage(fmt.Sprintf(gmess.INFO_OUCREATED, newOrgUnit.OrgUnitPath)))

	return nil
}

func init() {
	createCmd.AddCommand(createOUCmd)

	createOUCmd.Flags().BoolVarP(&blockInherit, flgnm.FLG_BLOCKINHERIT, "b", false, "block orgunit policy inheritance")
	createOUCmd.Flags().StringVarP(&orgUnitDesc, flgnm.FLG_DESCRIPTION, "d", "", "orgunit description")
	createOUCmd.Flags().StringVarP(&parentOUPath, flgnm.FLG_PARENTPATH, "p", "", "orgunit parent path")
}
