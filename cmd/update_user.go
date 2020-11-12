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

	"github.com/imdario/mergo"
	cmn "github.com/plusworx/gmin/utils/common"
	flgnm "github.com/plusworx/gmin/utils/flagnames"
	gmess "github.com/plusworx/gmin/utils/gminmessages"
	lg "github.com/plusworx/gmin/utils/logging"
	usrs "github.com/plusworx/gmin/utils/users"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	admin "google.golang.org/api/admin/directory/v1"
)

var updateUserCmd = &cobra.Command{
	Use:     "user <user email address, alias or id>",
	Aliases: []string{"usr"},
	Args:    cobra.ExactArgs(1),
	Example: `gmin update user another.user@mycompany.com -p strongpassword -s
gmin upd user finance.person@mycompany.com -l Newlastname`,
	Short: "Updates a user",
	Long:  `Updates a user.`,
	RunE:  doUpdateUser,
}

func doUpdateUser(cmd *cobra.Command, args []string) error {
	lg.Debugw("starting doUpdateUser()",
		"args", args)
	defer lg.Debug("finished doUpdateUser()")

	var (
		flagsPassed []string
		userKey     string
	)

	userKey = args[0]
	user := new(admin.User)
	name := new(admin.UserName)

	flagValueMap := map[string]interface{}{}

	// Collect names of command flags passed in
	cmd.Flags().Visit(func(f *pflag.Flag) {
		flagsPassed = append(flagsPassed, f.Name)
	})

	// Populate flag value map
	for _, flg := range flagsPassed {
		val, err := usrs.GetFlagVal(cmd, flg)
		if err != nil {
			return err
		}
		flagValueMap[flg] = val
	}

	// Process command flags
	err := processUpdUsrFlags(cmd, user, name, flagValueMap)
	if err != nil {
		lg.Error(err)
		return err
	}

	if name.FamilyName != "" || name.FullName != "" || name.GivenName != "" {
		user.Name = name
	}

	// Process attrs last to guarantee the order that flag processing happens
	flgAttrsVal, attrsPresent := flagValueMap[flgnm.FLG_ATTRIBUTES]
	if attrsPresent {
		err = uuAttributeFlag(user, "--"+flgnm.FLG_ATTRIBUTES, fmt.Sprintf("%v", flgAttrsVal))
		if err != nil {
			return err
		}
	}

	srv, err := cmn.CreateService(cmn.SRVTYPEADMIN, admin.AdminDirectoryUserScope)
	if err != nil {
		return err
	}
	ds := srv.(*admin.Service)

	uuc := ds.Users.Update(userKey, user)
	_, err = uuc.Do()
	if err != nil {
		lg.Error(err)
		return err
	}

	fmt.Println(cmn.GminMessage(fmt.Sprintf(gmess.INFO_USERUPDATED, userKey)))
	lg.Infof(gmess.INFO_USERUPDATED, userKey)

	return nil
}

func init() {
	updateCmd.AddCommand(updateUserCmd)

	updateUserCmd.Flags().StringP(flgnm.FLG_ATTRIBUTES, "a", "", "user's attributes as a JSON string")
	updateUserCmd.Flags().BoolP(flgnm.FLG_CHANGEPWD, "c", false, "user must change password on next login")
	updateUserCmd.Flags().StringP(flgnm.FLG_EMAIL, "e", "", "user's primary email address")
	updateUserCmd.Flags().StringP(flgnm.FLG_FIRSTNAME, "f", "", "user's first name")
	updateUserCmd.Flags().String(flgnm.FLG_FORCE, "", "field list for ForceSendFields separated by (~)")
	updateUserCmd.Flags().BoolP(flgnm.FLG_GAL, "g", false, "display user in Global Address List")
	updateUserCmd.Flags().StringP(flgnm.FLG_LASTNAME, "l", "", "user's last name")
	updateUserCmd.Flags().StringP(flgnm.FLG_ORGUNIT, "o", "", "user's orgunit")
	updateUserCmd.Flags().StringP(flgnm.FLG_PASSWORD, "p", "", "user's password")
	updateUserCmd.Flags().StringP(flgnm.FLG_RECEMAIL, "z", "", "user's recovery email address")
	updateUserCmd.Flags().StringP(flgnm.FLG_RECPHONE, "k", "", "user's recovery phone")
	updateUserCmd.Flags().BoolP(flgnm.FLG_SUSPENDED, "s", false, "user is suspended")
}

func processUpdUsrFlags(cmd *cobra.Command, user *admin.User, name *admin.UserName, flagValueMap map[string]interface{}) error {
	lg.Debug("starting processUpdUsrFlags()")
	defer lg.Debug("finished processUpdUsrFlags()")

	usrUpdBoolFuncMap := map[string]func(*admin.User, bool){
		flgnm.FLG_CHANGEPWD: uuChangePasswordFlag,
		flgnm.FLG_GAL:       uuGalFlag,
		flgnm.FLG_SUSPENDED: uuSuspendedFlag,
	}

	usrUpdNameFuncMap := map[string]func(*admin.UserName, string, string) error{
		flgnm.FLG_FIRSTNAME: uuFirstnameFlag,
		flgnm.FLG_LASTNAME:  uuLastnameFlag,
	}

	usrUpdOneStrFuncMap := map[string]func(*admin.User, string) error{
		flgnm.FLG_FORCE: uuForceFlag,
	}

	usrUpdTwoStrFuncMap := map[string]func(*admin.User, string, string) error{
		flgnm.FLG_EMAIL:    uuEmailFlag,
		flgnm.FLG_ORGUNIT:  uuOrgunitFlag,
		flgnm.FLG_PASSWORD: uuPasswordFlag,
		flgnm.FLG_RECEMAIL: uuRecoveryEmailFlag,
		flgnm.FLG_RECPHONE: uuRecoveryPhoneFlag,
	}

	for flName, flgVal := range flagValueMap {
		nameFunc, nfExists := usrUpdNameFuncMap[flName]
		if nfExists {
			err := nameFunc(name, "--"+flName, fmt.Sprintf("%v", flgVal))
			if err != nil {
				return err
			}
			continue
		}

		boolFunc, bfExists := usrUpdBoolFuncMap[flName]
		if bfExists {
			boolFunc(user, flgVal.(bool))
			continue
		}

		strOneFunc, sf1Exists := usrUpdOneStrFuncMap[flName]
		if sf1Exists {
			err := strOneFunc(user, fmt.Sprintf("%v", flgVal))
			if err != nil {
				return err
			}
			continue
		}

		strTwoFunc, sf2Exists := usrUpdTwoStrFuncMap[flName]
		if sf2Exists {
			err := strTwoFunc(user, "--"+flName, fmt.Sprintf("%v", flgVal))
			if err != nil {
				return err
			}
			continue
		}
	}
	return nil
}

func uuAttributeFlag(user *admin.User, flagName string, flgVal string) error {
	lg.Debugw("starting uuAttributeFlag()",
		"flgVal", flgVal)
	defer lg.Debug("finished uuAttributeFlag()")

	if flgVal == "" {
		return fmt.Errorf(gmess.ERR_EMPTYSTRING, flagName)
	}

	attrUser := new(admin.User)
	emptyVals := cmn.EmptyValues{}
	jsonBytes := []byte(flgVal)
	if !json.Valid(jsonBytes) {
		err := errors.New(gmess.ERR_INVALIDJSONATTR)
		lg.Error(err)
		return err
	}

	outStr, err := cmn.ParseInputAttrs(jsonBytes)
	if err != nil {
		lg.Error(err)
		return err
	}

	err = cmn.ValidateInputAttrs(outStr, usrs.UserAttrMap)
	if err != nil {
		lg.Error(err)
		return err
	}

	err = json.Unmarshal(jsonBytes, &attrUser)
	if err != nil {
		lg.Error(err)
		return err
	}

	err = json.Unmarshal(jsonBytes, &emptyVals)
	if err != nil {
		lg.Error(err)
		return err
	}
	if len(emptyVals.ForceSendFields) > 0 {
		attrUser.ForceSendFields = emptyVals.ForceSendFields
	}

	err = mergo.Merge(user, attrUser)
	if err != nil {
		lg.Error(err)
		return err
	}

	return nil
}

func uuChangePasswordFlag(user *admin.User, flgVal bool) {
	lg.Debugw("starting uuChangePasswordFlag()",
		"flgVal", flgVal)
	defer lg.Debug("finished uuChangePasswordFlag()")

	if flgVal {
		user.ChangePasswordAtNextLogin = true
	} else {
		user.ChangePasswordAtNextLogin = false
		user.ForceSendFields = append(user.ForceSendFields, "ChangePasswordAtNextLogin")
	}
}

func uuEmailFlag(user *admin.User, flagName string, flgVal string) error {
	lg.Debugw("starting uuEmailFlag()",
		"flagName", flagName,
		"flgVal", flgVal)
	defer lg.Debug("finished uuEmailFlag()")

	if flgVal == "" {
		err := fmt.Errorf(gmess.ERR_EMPTYSTRING, flagName)
		lg.Error(err)
		return err
	}

	user.PrimaryEmail = flgVal
	return nil
}

func uuFirstnameFlag(name *admin.UserName, flagName string, flgVal string) error {
	lg.Debugw("starting uuFirstnameFlag()",
		"flagName", flagName,
		"flgVal", flgVal)
	defer lg.Debug("finished uuFirstnameFlag()")

	if flgVal == "" {
		err := fmt.Errorf(gmess.ERR_EMPTYSTRING, flagName)
		return err
	}
	name.GivenName = flgVal
	return nil
}

func uuForceFlag(user *admin.User, flgVal string) error {
	lg.Debugw("starting uuForceFlag()",
		"flgVal", flgVal)
	defer lg.Debug("finished uuForceFlag()")

	fields, err := cmn.ParseForceSend(flgVal, usrs.UserAttrMap)
	if err != nil {
		lg.Error(err)
		return err
	}
	for _, fld := range fields {
		user.ForceSendFields = append(user.ForceSendFields, fld)
	}
	return nil
}

func uuGalFlag(user *admin.User, flgVal bool) {
	lg.Debugw("starting uuGalFlag()",
		"flgVal", flgVal)
	defer lg.Debug("finished uuGalFlag()")

	if flgVal {
		user.IncludeInGlobalAddressList = true
	} else {
		user.IncludeInGlobalAddressList = false
		user.ForceSendFields = append(user.ForceSendFields, "IncludeInGlobalAddressList")
	}
}

func uuLastnameFlag(name *admin.UserName, flagName string, flgVal string) error {
	lg.Debugw("starting uuLastnameFlag()",
		"flagName", flagName,
		"flgVal", flgVal)
	if flgVal == "" {
		err := fmt.Errorf(gmess.ERR_EMPTYSTRING, flagName)
		return err
	}
	name.FamilyName = flgVal
	lg.Debug("finished uuLastnameFlag()")
	return nil
}

func uuOrgunitFlag(user *admin.User, flagName string, flgVal string) error {
	lg.Debugw("starting uuOrgunitFlag()",
		"flagName", flagName,
		"flgVal", flgVal)
	if flgVal == "" {
		err := fmt.Errorf(gmess.ERR_EMPTYSTRING, flagName)
		if err != nil {
			return err
		}
	}
	user.OrgUnitPath = flgVal
	lg.Debug("finished uuOrgunitFlag()")
	return nil
}

func uuPasswordFlag(user *admin.User, flagName string, flgVal string) error {
	lg.Debugw("starting uuPasswordFlag()",
		"flagName", flagName,
		"flgVal", flgVal)
	if flgVal == "" {
		err := fmt.Errorf(gmess.ERR_EMPTYSTRING, flagName)
		return err
	}
	pwd, err := usrs.HashPassword(flgVal)
	if err != nil {
		return err
	}
	user.Password = pwd
	user.HashFunction = usrs.HASHFUNCTION
	lg.Debug("finished uuPasswordFlag()")
	return nil
}

func uuRecoveryEmailFlag(user *admin.User, flagName string, flgVal string) error {
	lg.Debugw("starting uuRecoveryEmailFlag()",
		"flagName", flagName,
		"flgVal", flgVal)
	if flgVal == "" {
		err := fmt.Errorf(gmess.ERR_EMPTYSTRING, flagName)
		return err
	}
	user.RecoveryEmail = flgVal
	lg.Debug("finished uuRecoveryEmailFlag()")
	return nil
}

func uuRecoveryPhoneFlag(user *admin.User, flagName string, flgVal string) error {
	lg.Debugw("starting uuRecoveryPhoneFlag()",
		"flagName", flagName,
		"flgVal", flgVal)
	if flgVal == "" {
		err := fmt.Errorf(gmess.ERR_EMPTYSTRING, flagName)
		return err
	}
	if string(flgVal[0]) != "+" {
		err := fmt.Errorf(gmess.ERR_INVALIDRECOVERYPHONE, flgVal)
		return err
	}
	user.RecoveryPhone = flgVal
	lg.Debug("finished uuRecoveryPhoneFlag()")
	return nil
}

func uuSuspendedFlag(user *admin.User, flgVal bool) {
	lg.Debug("starting uuSuspendedFlag()")
	if flgVal {
		user.Suspended = true
	} else {
		user.Suspended = false
		user.ForceSendFields = append(user.ForceSendFields, "Suspended")
	}
	lg.Debug("finished uuSuspendedFlag()")
}
