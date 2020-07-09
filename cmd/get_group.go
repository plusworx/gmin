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
	grps "github.com/plusworx/gmin/utils/groups"
	"github.com/spf13/cobra"
	admin "google.golang.org/api/admin/directory/v1"
)

var getGroupCmd = &cobra.Command{
	Use:     "group <email address or id>",
	Aliases: []string{"grp"},
	Args:    cobra.ExactArgs(1),
	Short:   "Outputs information about a group",
	Long: `Outputs information about a group.
	
	Examples: gmin get group agroup@mydomain.org
	          gmin get group 12345678 -a email`,
	RunE: doGetGroup,
}

func doGetGroup(cmd *cobra.Command, args []string) error {
	var (
		group      *admin.Group
		validAttrs []string
	)

	ds, err := cmn.CreateDirectoryService(admin.AdminDirectoryGroupReadonlyScope)
	if err != nil {
		return err
	}

	ggc := ds.Groups.Get(args[0])

	if attrs != "" {
		validAttrs, err = cmn.ValidateAttrs(attrs, grps.GroupAttrMap)
		if err != nil {
			return err
		}

		formattedAttrs := grps.FormatAttrs(validAttrs, true)
		group, err = grps.GetAttrs(ggc, formattedAttrs)
		if err != nil {
			return err
		}
	} else {
		group, err = grps.Get(ggc)
		if err != nil {
			return err
		}
	}

	jsonData, err := json.MarshalIndent(group, "", "    ")
	if err != nil {
		return err
	}

	fmt.Println(string(jsonData))

	return nil
}

func init() {
	getCmd.AddCommand(getGroupCmd)

	getGroupCmd.Flags().StringVarP(&attrs, "attributes", "a", "", "required group attributes (separated by ~)")
}
