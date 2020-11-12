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
	gd "github.com/plusworx/gmin/utils/gendatastructs"
	gmess "github.com/plusworx/gmin/utils/gminmessages"
	gpars "github.com/plusworx/gmin/utils/gminparsers"
	lg "github.com/plusworx/gmin/utils/logging"
	usrs "github.com/plusworx/gmin/utils/users"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
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
	defer lg.Debug("finished doGetUser()")

	var (
		flagsPassed []string
		jsonData    []byte
		user        *admin.User
	)

	flagValueMap := map[string]interface{}{}

	cmd.Flags().Visit(func(f *pflag.Flag) {
		flagsPassed = append(flagsPassed, f.Name)
	})

	for _, flg := range flagsPassed {
		val, err := usrs.GetFlagVal(cmd, flg)
		if err != nil {
			return err
		}
		flagValueMap[flg] = val
	}

	srv, err := cmn.CreateService(cmn.SRVTYPEADMIN, admin.AdminDirectoryUserReadonlyScope)
	if err != nil {
		return err
	}
	ds := srv.(*admin.Service)

	ugc := ds.Users.Get(args[0])

	err = getUsrProcessFlags(ugc, flagValueMap)
	if err != nil {
		return err
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

	return nil
}

func getUsrAttributes(ugc *admin.UsersGetCall, flagVal interface{}) error {
	lg.Debug("starting getUsrAttributes()")
	defer lg.Debug("finished getUsrAttributes()")

	attrsVal := fmt.Sprintf("%v", flagVal)
	if attrsVal != "" {
		getAttrs, err := gpars.ParseOutputAttrs(attrsVal, usrs.UserAttrMap)
		if err != nil {
			lg.Error(err)
			return err
		}

		getCall := usrs.AddFields(ugc, getAttrs)
		ugc = getCall.(*admin.UsersGetCall)
	}
	return nil
}

func getUsrCustomProjection(flgValMap map[string]interface{}) (gd.TwoStrStruct, error) {
	lg.Debug("starting getUsrCustomProjection()")
	defer lg.Debug("finished getUsrCustomProjection()")

	var retMaskVal string

	projVal := flgValMap[flgnm.FLG_PROJECTION]
	custFldMaskVal, maskPresent := flgValMap[flgnm.FLG_CUSTFLDMASK]

	lowerProjVal := strings.ToLower(fmt.Sprintf("%v", projVal))

	if lowerProjVal == "custom" && !maskPresent {
		err := errors.New(gmess.ERR_NOCUSTOMFIELDMASK)
		lg.Error(err)
		return gd.TwoStrStruct{}, err
	}

	if maskPresent && lowerProjVal != "custom" {
		err := errors.New(gmess.ERR_PROJECTIONFLAGNOTCUSTOM)
		lg.Error(err)
		return gd.TwoStrStruct{}, err
	}

	if maskPresent {
		retMaskVal = fmt.Sprintf("%v", custFldMaskVal)
	} else {
		retMaskVal = ""
	}

	retStruct := gd.TwoStrStruct{Element1: lowerProjVal, Element2: retMaskVal}
	return retStruct, nil
}

func getUsrProcessFlags(ugc *admin.UsersGetCall, flgValMap map[string]interface{}) error {
	lg.Debug("starting getUsrProcessFlags()")
	defer lg.Debug("finished getUsrProcessFlags()")

	getUsrFuncMap := map[string]func(*admin.UsersGetCall, interface{}) error{
		flgnm.FLG_ATTRIBUTES: getUsrAttributes,
		flgnm.FLG_PROJECTION: getUsrProjection,
		flgnm.FLG_VIEWTYPE:   getUsrViewType,
	}

	// Cycle through flags that build the ugc
	for key, val := range flgValMap {
		// Projection has dependent custom field mask so deal with that
		if key == flgnm.FLG_PROJECTION {
			retStruct, err := getUsrCustomProjection(flgValMap)
			if err != nil {
				return err
			}
			val = retStruct
		}

		guf, ok := getUsrFuncMap[key]
		if !ok {
			continue
		}
		err := guf(ugc, val)
		if err != nil {
			return err
		}
	}
	return nil
}

func getUsrProjection(ugc *admin.UsersGetCall, inData interface{}) error {
	lg.Debug("starting getUsrProjection()")
	defer lg.Debug("finished getUsrProjection()")

	inStruct := inData.(gd.TwoStrStruct)
	projVal := inStruct.Element1

	if projVal != "" {
		ok := cmn.SliceContainsStr(usrs.ValidProjections, projVal)
		if !ok {
			err := fmt.Errorf(gmess.ERR_INVALIDPROJECTIONTYPE, projVal)
			lg.Error(err)
			return err
		}

		getCall := usrs.AddProjection(ugc, projVal)
		ugc = getCall.(*admin.UsersGetCall)

		if projVal == "custom" {
			custVal := inStruct.Element2
			if custVal != "" {
				cFields := strings.Split(custVal, "~")
				mask := strings.Join(cFields, ",")
				getCall := usrs.AddCustomFieldMask(ugc, mask)
				ugc = getCall.(*admin.UsersGetCall)
			} else {
				err := errors.New(gmess.ERR_NOCUSTOMFIELDMASK)
				lg.Error(err)
				return err
			}
		}
	}
	return nil
}

func getUsrViewType(ugc *admin.UsersGetCall, flagVal interface{}) error {
	lg.Debug("starting getUsrViewType()")
	defer lg.Debug("finished getUsrViewType()")

	vtVal := fmt.Sprintf("%v", flagVal)
	if vtVal != "" {
		lowerVt := strings.ToLower(vtVal)
		ok := cmn.SliceContainsStr(usrs.ValidViewTypes, lowerVt)
		if !ok {
			err := fmt.Errorf(gmess.ERR_INVALIDVIEWTYPE, vtVal)
			lg.Error(err)
			return err
		}

		getCall := usrs.AddViewType(ugc, lowerVt)
		ugc = getCall.(*admin.UsersGetCall)
	}
	return nil
}

func init() {
	getCmd.AddCommand(getUserCmd)

	getUserCmd.Flags().StringP(flgnm.FLG_ATTRIBUTES, "a", "", "required user attributes (separated by ~)")
	getUserCmd.Flags().StringP(flgnm.FLG_CUSTFLDMASK, "c", "", "custom field mask schemas (separated by ~)")
	getUserCmd.Flags().StringP(flgnm.FLG_PROJECTION, "j", "", "type of projection")
	getUserCmd.Flags().StringP(flgnm.FLG_VIEWTYPE, "v", "", "data view type")
}
