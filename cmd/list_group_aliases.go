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
	flgnm "github.com/plusworx/gmin/utils/flagnames"
	gpars "github.com/plusworx/gmin/utils/gminparsers"
	gas "github.com/plusworx/gmin/utils/groupaliases"
	"github.com/spf13/cobra"
	admin "google.golang.org/api/admin/directory/v1"
)

var listGroupAliasesCmd = &cobra.Command{
	Use:     "group-aliases <group email address or id>",
	Aliases: []string{"group-alias", "grp-aliases", "grp-alias", "galiases", "galias", "gas", "ga"},
	Args:    cobra.ExactArgs(1),
	Example: `gmin list group-aliases mygroup@mycompany.com
gmin ls gas mygroup@mycompany.com`,
	Short: "Outputs a list of group aliases",
	Long:  `Outputs a list of group aliases.`,
	RunE:  doListGroupAliases,
}

func doListGroupAliases(cmd *cobra.Command, args []string) error {
	logger.Debugw("starting doListGroupAliases()",
		"args", args)

	var aliases *admin.Aliases

	ds, err := cmn.CreateDirectoryService(admin.AdminDirectoryGroupReadonlyScope)
	if err != nil {
		logger.Error(err)
		return err
	}

	galc := ds.Groups.Aliases.List(args[0])

	flgAttrsVal, err := cmd.Flags().GetString(flgnm.FLG_ATTRIBUTES)
	if err != nil {
		logger.Error(err)
		return err
	}
	if flgAttrsVal != "" {
		listAttrs, err := gpars.ParseOutputAttrs(flgAttrsVal, gas.GroupAliasAttrMap)
		if err != nil {
			logger.Error(err)
			return err
		}
		formattedAttrs := gas.STARTALIASESFIELD + listAttrs + gas.ENDFIELD

		galc = gas.AddFields(galc, formattedAttrs)
	}

	aliases, err = gas.DoList(galc)
	if err != nil {
		logger.Error(err)
		return err
	}

	jsonData, err := json.MarshalIndent(aliases, "", "    ")
	if err != nil {
		logger.Error(err)
		return err
	}

	fmt.Println(string(jsonData))

	logger.Debug("finished doListGroupAliases()")
	return nil
}

func init() {
	listCmd.AddCommand(listGroupAliasesCmd)

	listGroupAliasesCmd.Flags().StringVarP(&attrs, flgnm.FLG_ATTRIBUTES, "a", "", "required group alias attributes (separated by ~)")
}
