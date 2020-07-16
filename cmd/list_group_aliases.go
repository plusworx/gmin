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
	gas "github.com/plusworx/gmin/utils/groupaliases"
	"github.com/spf13/cobra"
	admin "google.golang.org/api/admin/directory/v1"
)

var listGroupAliasesCmd = &cobra.Command{
	Use:     "group-aliases <group email address or id>",
	Aliases: []string{"group-alias", "galiases", "galias", "gas", "ga"},
	Args:    cobra.ExactArgs(1),
	Short:   "Outputs a list of group aliases",
	Long:    `Outputs a list of group aliases.`,
	RunE:    doListGroupAliases,
}

func doListGroupAliases(cmd *cobra.Command, args []string) error {
	var (
		formattedAttrs string
		aliases        *admin.Aliases
		validAttrs     []string
	)

	ds, err := cmn.CreateDirectoryService(admin.AdminDirectoryGroupReadonlyScope)
	if err != nil {
		return err
	}

	galc := ds.Groups.Aliases.List(args[0])

	if attrs != "" {
		validAttrs, err = cmn.ValidateArgs(attrs, gas.GroupAliasAttrMap, cmn.AttrStr)
		if err != nil {
			return err
		}

		formattedAttrs = gas.FormatAttrs(validAttrs)
		galc = gas.AddFields(galc, formattedAttrs)
	}

	aliases, err = gas.DoList(galc)
	if err != nil {
		return err
	}

	jsonData, err := json.MarshalIndent(aliases, "", "    ")
	if err != nil {
		return err
	}

	fmt.Println(string(jsonData))

	return nil
}

func init() {
	listCmd.AddCommand(listGroupAliasesCmd)

	listGroupAliasesCmd.Flags().StringVarP(&attrs, "attributes", "a", "", "required group alias attributes (separated by ~)")
}
