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
	flgnm "github.com/plusworx/gmin/utils/flagnames"
	gmess "github.com/plusworx/gmin/utils/gminmessages"
	gpars "github.com/plusworx/gmin/utils/gminparsers"
	lg "github.com/plusworx/gmin/utils/logging"
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
	lg.Debugw("starting doGetUser()",
		"args", args)

	var (
		jsonData []byte
		user     *admin.User
	)

	srv, err := cmn.CreateService(cmn.SRVTYPEADMIN, admin.AdminDirectoryUserReadonlyScope)
	if err != nil {
		return err
	}
	ds := srv.(*admin.Service)

	ugc := ds.Users.Get(args[0])

	flgAttrsVal, err := cmd.Flags().GetString(flgnm.FLG_ATTRIBUTES)
	if err != nil {
		lg.Error(err)
		return err
	}
	if flgAttrsVal != "" {
		formattedAttrs, err := gpars.ParseOutputAttrs(flgAttrsVal, usrs.UserAttrMap)
		if err != nil {
			lg.Error(err)
			return err
		}

		getCall := usrs.AddFields(ugc, formattedAttrs)
		ugc = getCall.(*admin.UsersGetCall)
	}

	flgProjectionVal, err := cmd.Flags().GetString(flgnm.FLG_PROJECTION)
	if err != nil {
		lg.Error(err)
		return err
	}
	if flgProjectionVal != "" {
		proj := strings.ToLower(flgProjectionVal)
		ok := cmn.SliceContainsStr(usrs.ValidProjections, proj)
		if !ok {
			err = fmt.Errorf(gmess.ERR_INVALIDPROJECTIONTYPE, flgProjectionVal)
			lg.Error(err)
			return err
		}

		getCall := usrs.AddProjection(ugc, proj)
		ugc = getCall.(*admin.UsersGetCall)

		if proj == "custom" {
			flgCustFldVal, err := cmd.Flags().GetString(flgnm.FLG_CUSTFLDMASK)
			if err != nil {
				lg.Error(err)
				return err
			}
			if flgCustFldVal != "" {
				cFields := strings.Split(flgCustFldVal, "~")
				mask := strings.Join(cFields, ",")
				getCall := usrs.AddCustomFieldMask(ugc, mask)
				ugc = getCall.(*admin.UsersGetCall)
			} else {
				err = errors.New(gmess.ERR_NOCUSTOMFIELDMASK)
				lg.Error(err)
				return err
			}
		}
	}

	flgViewTypeVal, err := cmd.Flags().GetString(flgnm.FLG_VIEWTYPE)
	if err != nil {
		lg.Error(err)
		return err
	}
	if flgViewTypeVal != "" {
		vt := strings.ToLower(flgViewTypeVal)
		ok := cmn.SliceContainsStr(usrs.ValidViewTypes, vt)
		if !ok {
			err = fmt.Errorf(gmess.ERR_INVALIDVIEWTYPE, flgViewTypeVal)
			lg.Error(err)
			return err
		}

		getCall := usrs.AddViewType(ugc, vt)
		ugc = getCall.(*admin.UsersGetCall)
	}

	user, err = usrs.DoGet(ugc)
	if err != nil {
		return err
	}

	jsonData, err = json.MarshalIndent(user, "", "    ")
	if err != nil {
		lg.Error(err)
		return err
	}

	fmt.Println(string(jsonData))

	lg.Debug("finished doGetUser()")
	return nil
}

func init() {
	getCmd.AddCommand(getUserCmd)

	getUserCmd.Flags().StringP(flgnm.FLG_ATTRIBUTES, "a", "", "required user attributes (separated by ~)")
	getUserCmd.Flags().StringP(flgnm.FLG_CUSTFLDMASK, "c", "", "custom field mask schemas (separated by ~)")
	getUserCmd.Flags().StringP(flgnm.FLG_PROJECTION, "j", "", "type of projection")
	getUserCmd.Flags().StringP(flgnm.FLG_VIEWTYPE, "v", "", "data view type")
}
