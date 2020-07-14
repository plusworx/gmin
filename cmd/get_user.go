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
	usrs "github.com/plusworx/gmin/utils/users"
	"github.com/spf13/cobra"
	admin "google.golang.org/api/admin/directory/v1"
)

var getUserCmd = &cobra.Command{
	Use:   "user <email address or id>",
	Args:  cobra.ExactArgs(1),
	Short: "Outputs information about a user",
	Long: `Outputs information about a user.
	
	Examples: gmin get user auser@mydomain.org
	          gmin get user 12345678 -a primaryEmail`,
	RunE: doGetUser,
}

func doGetUser(cmd *cobra.Command, args []string) error {
	var (
		user       *admin.User
		validAttrs []string
	)

	ds, err := cmn.CreateDirectoryService(admin.AdminDirectoryUserReadonlyScope)
	if err != nil {
		return err
	}

	ugc := ds.Users.Get(args[0])

	if attrs != "" {
		validAttrs, err = cmn.ValidateArgs(attrs, usrs.UserAttrMap, cmn.AttrStr)
		if err != nil {
			return err
		}

		formattedAttrs := usrs.FormatAttrs(validAttrs, true)
		getCall := usrs.AddFields(ugc, formattedAttrs)
		ugc = getCall.(*admin.UsersGetCall)
	}

	if projection != "" {
		proj := strings.ToLower(projection)
		ok := cmn.SliceContainsStr(usrs.ValidProjections, proj)
		if !ok {
			return fmt.Errorf("gmin: error - %v is not a valid projection type", projection)
		}

		getCall := usrs.AddProjection(ugc, proj)
		ugc = getCall.(*admin.UsersGetCall)
	}

	if viewType != "" {
		vt := strings.ToLower(viewType)
		ok := cmn.SliceContainsStr(usrs.ValidViewTypes, vt)
		if !ok {
			return fmt.Errorf("gmin: error - %v is not a valid view type", viewType)
		}

		getCall := usrs.AddViewType(ugc, vt)
		ugc = getCall.(*admin.UsersGetCall)
	}

	user, err = usrs.DoGet(ugc)
	if err != nil {
		return err
	}

	jsonData, err := json.MarshalIndent(user, "", "    ")
	if err != nil {
		return err
	}

	fmt.Println(string(jsonData))

	return nil
}

func init() {
	getCmd.AddCommand(getUserCmd)

	getUserCmd.Flags().StringVarP(&attrs, "attributes", "a", "", "required user attributes (separated by ~)")
	getUserCmd.Flags().StringVarP(&projection, "projection", "p", "", "type of projection")
	getUserCmd.Flags().StringVarP(&viewType, "viewtype", "v", "", "data view type")
}
