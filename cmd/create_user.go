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

	cmn "github.com/plusworx/gmin/utils/common"
	flgnm "github.com/plusworx/gmin/utils/flagnames"
	gmess "github.com/plusworx/gmin/utils/gminmessages"
	lg "github.com/plusworx/gmin/utils/logging"
	usrs "github.com/plusworx/gmin/utils/users"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	valid "github.com/asaskevich/govalidator"
	"github.com/imdario/mergo"
	admin "google.golang.org/api/admin/directory/v1"
)

var createUserCmd = &cobra.Command{
	Use:     "user <user email address>",
	Aliases: []string{"usr"},
	Args:    cobra.ExactArgs(1),
	Example: `gmin create user another.user@mycompany.com  -f Another -l User -p strongpassword
gmin crt user finance.person@mycompany.com -f Finance -l Person -p greatpassword -c`,
	Short: "Creates a user",
	Long:  `Creates a user.`,
	RunE:  doCreateUser,
}

func doCreateUser(cmd *cobra.Command, args []string) error {
	lg.Debugw("starting doCreateUser()",
		"args", args)
	defer lg.Debug("finished doCreateUser()")

	var flagsPassed []string

	user := new(admin.User)
	name := new(admin.UserName)
	flagValueMap := map[string]interface{}{}

	ok := valid.IsEmail(args[0])
	if !ok {
		err := fmt.Errorf(gmess.ERR_INVALIDEMAILADDRESS, args[0])
		lg.Error(err)
		return err
	}

	user.PrimaryEmail = args[0]

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
	err := processCrtUsrFlags(cmd, user, name, flagValueMap)
	if err != nil {
		lg.Error(err)
		return err
	}

	user.Name = name

	if user.Name.GivenName == "" || user.Name.FamilyName == "" || user.Password == "" {
		err = errors.New(gmess.ERR_MISSINGUSERDATA)
		lg.Error(err)
		return err
	}

	// Process attrs last to guarantee the order that flag processing happens
	flgAttrsVal, attrsPresent := flagValueMap[flgnm.FLG_ATTRIBUTES]
	if attrsPresent {
		err = cuAttributeFlag(user, "--"+flgnm.FLG_ATTRIBUTES, fmt.Sprintf("%v", flgAttrsVal))
		if err != nil {
			return err
		}
	}

	srv, err := cmn.CreateService(cmn.SRVTYPEADMIN, admin.AdminDirectoryUserScope)
	if err != nil {
		return err
	}
	ds := srv.(*admin.Service)

	uic := ds.Users.Insert(user)
	newUser, err := uic.Do()
	if err != nil {
		lg.Error(err)
		return err
	}

	fmt.Println(cmn.GminMessage(fmt.Sprintf(gmess.INFO_USERCREATED, newUser.PrimaryEmail)))
	lg.Infof(gmess.INFO_USERCREATED, newUser.PrimaryEmail)

	return nil
}

func init() {
	createCmd.AddCommand(createUserCmd)

	createUserCmd.Flags().StringP(flgnm.FLG_ATTRIBUTES, "a", "", "user's attributes as a JSON string")
	createUserCmd.Flags().BoolP(flgnm.FLG_CHANGEPWD, "c", false, "user must change password on next login")
	createUserCmd.Flags().StringP(flgnm.FLG_FIRSTNAME, "f", "", "user's first name")
	createUserCmd.Flags().String(flgnm.FLG_FORCE, "", "field list for ForceSendFields separated by (~)")
	createUserCmd.Flags().BoolP(flgnm.FLG_GAL, "g", false, "user is included in Global Address List")
	createUserCmd.Flags().StringP(flgnm.FLG_LASTNAME, "l", "", "user's last name")
	createUserCmd.Flags().StringP(flgnm.FLG_ORGUNIT, "o", "", "user's orgunit")
	createUserCmd.Flags().StringP(flgnm.FLG_PASSWORD, "p", "", "user's password")
	createUserCmd.Flags().StringP(flgnm.FLG_RECEMAIL, "z", "", "user's recovery email address")
	createUserCmd.Flags().StringP(flgnm.FLG_RECPHONE, "k", "", "user's recovery phone")
	createUserCmd.Flags().BoolP(flgnm.FLG_SUSPENDED, "s", false, "user is suspended")
}

func cuAttributeFlag(user *admin.User, flagName string, flgVal string) error {
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

	if user.Password == "" && attrUser.Password != "" {
		pwd, err := usrs.HashPassword(attrUser.Password)
		if err != nil {
			lg.Error(err)
			return err
		}
		attrUser.Password = pwd
		attrUser.HashFunction = usrs.HASHFUNCTION
	}

	err = mergo.Merge(user, attrUser)
	if err != nil {
		lg.Error(err)
		return err
	}

	return nil
}

func cuChangePasswordFlag(user *admin.User, flgVal bool) {
	lg.Debugw("starting cuChangePasswordFlag()",
		"flgVal", flgVal)
	defer lg.Debug("finished cuChangePasswordFlag()")

	if flgVal {
		user.ChangePasswordAtNextLogin = true
	} else {
		user.ChangePasswordAtNextLogin = false
		user.ForceSendFields = append(user.ForceSendFields, "ChangePasswordAtNextLogin")
	}
}

func cuFirstnameFlag(name *admin.UserName, flagName string, flgVal string) error {
	lg.Debugw("starting cuFirstnameFlag()",
		"flagName", flagName,
		"flgVal", flgVal)
	defer lg.Debug("finished cuFirstnameFlag()")

	if flgVal == "" {
		err := fmt.Errorf(gmess.ERR_EMPTYSTRING, flagName)
		if err != nil {
			lg.Error(err)
			return err
		}
	}
	name.GivenName = flgVal
	return nil
}

func cuForceFlag(user *admin.User, forceSend string) error {
	lg.Debugw("starting cuForceFlag()",
		"forceSend", forceSend)
	fields, err := cmn.ParseForceSend(forceSend, usrs.UserAttrMap)
	if err != nil {
		lg.Error(err)
		return err
	}
	for _, fld := range fields {
		user.ForceSendFields = append(user.ForceSendFields, fld)
	}
	lg.Debug("finished cuForceFlag()")
	return nil
}

func cuGalFlag(user *admin.User, flgVal bool) {
	if flgVal {
		user.IncludeInGlobalAddressList = true
	} else {
		user.IncludeInGlobalAddressList = false
		user.ForceSendFields = append(user.ForceSendFields, "IncludeInGlobalAddressList")
	}
}

func cuLastnameFlag(name *admin.UserName, flagName string, flgVal string) error {
	lg.Debugw("starting cuLastnameFlag()",
		"flagName", flagName)
	if flgVal == "" {
		err := fmt.Errorf(gmess.ERR_EMPTYSTRING, flagName)
		if err != nil {
			lg.Error(err)
			return err
		}
	}
	name.FamilyName = flgVal
	lg.Debug("finished cuLastnameFlag()")
	return nil
}

func cuOrgunitFlag(user *admin.User, flagName string, flgVal string) error {
	lg.Debugw("starting cuOrgunitFlag()",
		"flagName", flagName)
	if flgVal == "" {
		err := fmt.Errorf(gmess.ERR_EMPTYSTRING, flagName)
		if err != nil {
			lg.Error(err)
			return err
		}
	}
	user.OrgUnitPath = flgVal
	lg.Debug("finished cuOrgunitFlag()")
	return nil
}

func cuPasswordFlag(user *admin.User, flagName string, flgVal string) error {
	lg.Debugw("starting cuPasswordFlag()",
		"flagName", flagName)
	if flgVal == "" {
		err := fmt.Errorf(gmess.ERR_EMPTYSTRING, flagName)
		lg.Error(err)
		return err
	}
	pwd, err := usrs.HashPassword(flgVal)
	if err != nil {
		lg.Error(err)
		return err
	}
	user.Password = pwd
	user.HashFunction = usrs.HASHFUNCTION
	lg.Debug("finished cuPasswordFlag()")
	return nil
}

func cuRecoveryEmailFlag(user *admin.User, flagName string, flgVal string) error {
	lg.Debugw("starting cuRecoveryEmailFlag()",
		"flagName", flagName)
	if flgVal == "" {
		err := fmt.Errorf(gmess.ERR_EMPTYSTRING, flagName)
		lg.Error(err)
		return err
	}
	ok := valid.IsEmail(flgVal)
	if !ok {
		err := fmt.Errorf(gmess.ERR_INVALIDEMAILADDRESS, flgVal)
		lg.Error(err)
		return err
	}
	user.RecoveryEmail = flgVal
	lg.Debug("finished cuRecoveryEmailFlag()")
	return nil
}

func cuRecoveryPhoneFlag(user *admin.User, flagName string, flgVal string) error {
	lg.Debugw("starting cuRecoveryPhoneFlag()",
		"flagName", flagName)
	if flgVal == "" {
		err := fmt.Errorf(gmess.ERR_EMPTYSTRING, flagName)
		lg.Error(err)
		return err
	}
	if string(flgVal[0]) != "+" {
		err := fmt.Errorf(gmess.ERR_INVALIDRECOVERYPHONE, flgVal)
		lg.Error(err)
		return err
	}
	user.RecoveryPhone = flgVal
	lg.Debug("finished cuRecoveryPhoneFlag()")
	return nil
}

func cuSuspendedFlag(user *admin.User, flgVal bool) {
	lg.Debugw("starting cuSuspendedFlag()",
		"flgVal", flgVal)
	defer lg.Debug("finished cuSuspendedFlag()")

	if flgVal {
		user.Suspended = true
	} else {
		user.Suspended = false
		user.ForceSendFields = append(user.ForceSendFields, "Suspended")
	}
}

func processCrtUsrFlags(cmd *cobra.Command, user *admin.User, name *admin.UserName, flagValueMap map[string]interface{}) error {
	lg.Debugw("starting processCrtUsrFlags()")
	defer lg.Debug("finished processCrtUsrFlags()")

	usrCrtBoolFuncMap := map[string]func(*admin.User, bool){
		flgnm.FLG_CHANGEPWD: cuChangePasswordFlag,
		flgnm.FLG_GAL:       cuGalFlag,
		flgnm.FLG_SUSPENDED: cuSuspendedFlag,
	}

	usrCrtNameFuncMap := map[string]func(*admin.UserName, string, string) error{
		flgnm.FLG_FIRSTNAME: cuFirstnameFlag,
		flgnm.FLG_LASTNAME:  cuLastnameFlag,
	}

	usrCrtOneStrFuncMap := map[string]func(*admin.User, string) error{
		flgnm.FLG_FORCE: cuForceFlag,
	}

	usrCrtTwoStrFuncMap := map[string]func(*admin.User, string, string) error{
		flgnm.FLG_ORGUNIT:  cuOrgunitFlag,
		flgnm.FLG_PASSWORD: cuPasswordFlag,
		flgnm.FLG_RECEMAIL: cuRecoveryEmailFlag,
		flgnm.FLG_RECPHONE: cuRecoveryPhoneFlag,
	}

	for flName, flgVal := range flagValueMap {
		nameFunc, nfExists := usrCrtNameFuncMap[flName]
		if nfExists {
			err := nameFunc(name, "--"+flName, fmt.Sprintf("%v", flgVal))
			if err != nil {
				return err
			}
			continue
		}

		boolFunc, bfExists := usrCrtBoolFuncMap[flName]
		if bfExists {
			boolFunc(user, flgVal.(bool))
			continue
		}

		strOneFunc, sf1Exists := usrCrtOneStrFuncMap[flName]
		if sf1Exists {
			err := strOneFunc(user, fmt.Sprintf("%v", flgVal))
			if err != nil {
				return err
			}
			continue
		}

		strTwoFunc, sf2Exists := usrCrtTwoStrFuncMap[flName]
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
