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
	"errors"
	"fmt"
	"strings"

	cmn "github.com/plusworx/gmin/utils/common"
	gmess "github.com/plusworx/gmin/utils/gminmessages"
	usrs "github.com/plusworx/gmin/utils/users"
	"github.com/spf13/cobra"
	admin "google.golang.org/api/admin/directory/v1"
)

var getUserCmd = &cobra.Command{
	Use:     "user <email address or id>",
	Aliases: []string{"usr"},
	Args:    cobra.ExactArgs(1),
	Example: `gmin get user auser@mydomain.org
gmin get user 114361578941906491576 -a primaryEmail~name`,
	Short: "Outputs information about a user",
	Long:  `Outputs information about a user.`,
	RunE:  doGetUser,
}

func doGetUser(cmd *cobra.Command, args []string) error {
	logger.Debugw("starting doGetUser()",
		"args", args)

	var (
		jsonData []byte
		user     *admin.User
	)

	ds, err := cmn.CreateDirectoryService(admin.AdminDirectoryUserReadonlyScope)
	if err != nil {
		logger.Error(err)
		return err
	}

	ugc := ds.Users.Get(args[0])

	if attrs != "" {
		formattedAttrs, err := cmn.ParseOutputAttrs(attrs, usrs.UserAttrMap)
		if err != nil {
			logger.Error(err)
			return err
		}

		getCall := usrs.AddFields(ugc, formattedAttrs)
		ugc = getCall.(*admin.UsersGetCall)
	}

	if projection != "" {
		proj := strings.ToLower(projection)
		ok := cmn.SliceContainsStr(usrs.ValidProjections, proj)
		if !ok {
			err = fmt.Errorf(gmess.ErrInvalidProjectionType, projection)
			logger.Error(err)
			return err
		}

		getCall := usrs.AddProjection(ugc, proj)
		ugc = getCall.(*admin.UsersGetCall)

		if proj == "custom" {
			if customField != "" {
				cFields := cmn.ParseTildeField(customField)
				mask := strings.Join(cFields, ",")
				getCall := usrs.AddCustomFieldMask(ugc, mask)
				ugc = getCall.(*admin.UsersGetCall)
			} else {
				err = errors.New(gmess.ErrNoCustomFieldMask)
				logger.Error(err)
				return err
			}
		}
	}

	if viewType != "" {
		vt := strings.ToLower(viewType)
		ok := cmn.SliceContainsStr(usrs.ValidViewTypes, vt)
		if !ok {
			err = fmt.Errorf(gmess.ErrInvalidViewType, viewType)
			logger.Error(err)
			return err
		}

		getCall := usrs.AddViewType(ugc, vt)
		ugc = getCall.(*admin.UsersGetCall)
	}

	user, err = usrs.DoGet(ugc)
	if err != nil {
		logger.Error(err)
		return err
	}

	jsonData, err = json.MarshalIndent(user, "", "    ")
	if err != nil {
		logger.Error(err)
		return err
	}

	fmt.Println(string(jsonData))

	logger.Debug("finished doGetUser()")
	return nil
}

func init() {
	getCmd.AddCommand(getUserCmd)

	getUserCmd.Flags().StringVarP(&attrs, "attributes", "a", "", "required user attributes (separated by ~)")
	getUserCmd.Flags().StringVarP(&customField, "custom-field-mask", "c", "", "custom field mask schemas (separated by ~)")
	getUserCmd.Flags().StringVarP(&projection, "projection", "j", "", "type of projection")
	getUserCmd.Flags().StringVarP(&viewType, "view-type", "v", "", "data view type")
}
