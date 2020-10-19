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
	Use:  "user <user email address, alias or id>",
	Args: cobra.ExactArgs(1),
	Example: `gmin update user another.user@mycompany.com -p strongpassword -s
gmin upd user finance.person@mycompany.com -l Newlastname`,
	Short: "Updates a user",
	Long:  `Updates a user.`,
	RunE:  doUpdateUser,
}

func doUpdateUser(cmd *cobra.Command, args []string) error {
	lg.Debugw("starting doUpdateUser()",
		"args", args)

	var (
		flagsPassed []string
		userKey     string
	)

	userKey = args[0]
	user := new(admin.User)
	name := new(admin.UserName)

	// Collect names of command flags passed in
	cmd.Flags().Visit(func(f *pflag.Flag) {
		flagsPassed = append(flagsPassed, f.Name)
	})

	// Process command flags
	err := processUpdUsrFlags(cmd, user, name, flagsPassed)
	if err != nil {
		lg.Error(err)
		return err
	}

	if name.FamilyName != "" || name.FullName != "" || name.GivenName != "" {
		user.Name = name
	}

	flgAttrsVal, err := cmd.Flags().GetString(flgnm.FLG_ATTRIBUTES)
	if err != nil {
		lg.Error(err)
		return err
	}
	if flgAttrsVal != "" {
		attrUser := new(admin.User)
		emptyVals := cmn.EmptyValues{}
		jsonBytes := []byte(flgAttrsVal)
		if !json.Valid(jsonBytes) {
			err = errors.New(gmess.ERR_INVALIDJSONATTR)
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
	}

	ds, err := cmn.CreateDirectoryService(admin.AdminDirectoryUserScope)
	if err != nil {
		lg.Error(err)
		return err
	}

	uuc := ds.Users.Update(userKey, user)
	_, err = uuc.Do()
	if err != nil {
		lg.Error(err)
		return err
	}

	lg.Infof(gmess.INFO_USERUPDATED, userKey)
	fmt.Println(cmn.GminMessage(fmt.Sprintf(gmess.INFO_USERUPDATED, userKey)))

	lg.Debug("finished doUpdateUser()")
	return nil
}

func init() {
	updateCmd.AddCommand(updateUserCmd)

	updateUserCmd.Flags().StringVarP(&attrs, flgnm.FLG_ATTRIBUTES, "a", "", "user's attributes as a JSON string")
	updateUserCmd.Flags().BoolVarP(&changePassword, flgnm.FLG_CHANGEPWD, "c", false, "user must change password on next login")
	updateUserCmd.Flags().StringVarP(&userEmail, flgnm.FLG_EMAIL, "e", "", "user's primary email address")
	updateUserCmd.Flags().StringVarP(&firstName, flgnm.FLG_FIRSTNAME, "f", "", "user's first name")
	updateUserCmd.Flags().StringVar(&forceSend, flgnm.FLG_FORCE, "", "field list for ForceSendFields separated by (~)")
	updateUserCmd.Flags().BoolVarP(&gal, flgnm.FLG_GAL, "g", false, "display user in Global Address List")
	updateUserCmd.Flags().StringVarP(&lastName, flgnm.FLG_LASTNAME, "l", "", "user's last name")
	updateUserCmd.Flags().StringVarP(&orgUnit, flgnm.FLG_ORGUNIT, "o", "", "user's orgunit")
	updateUserCmd.Flags().StringVarP(&password, flgnm.FLG_PASSWORD, "p", "", "user's password")
	updateUserCmd.Flags().StringVarP(&recoveryEmail, flgnm.FLG_RECEMAIL, "z", "", "user's recovery email address")
	updateUserCmd.Flags().StringVarP(&recoveryPhone, flgnm.FLG_RECPHONE, "k", "", "user's recovery phone")
	updateUserCmd.Flags().BoolVarP(&suspended, flgnm.FLG_SUSPENDED, "s", false, "user is suspended")
}

func processUpdUsrFlags(cmd *cobra.Command, user *admin.User, name *admin.UserName, flagNames []string) error {
	lg.Debugw("starting processUpdUsrFlags()",
		"flagNames", flagNames)
	for _, flName := range flagNames {
		if flName == flgnm.FLG_CHANGEPWD {
			flgChangePwdVal, err := cmd.Flags().GetBool(flgnm.FLG_CHANGEPWD)
			if err != nil {
				lg.Error(err)
				return err
			}
			uuChangePasswordFlag(user, flgChangePwdVal)
		}
		if flName == flgnm.FLG_EMAIL {
			flgEmailVal, err := cmd.Flags().GetString(flgnm.FLG_EMAIL)
			if err != nil {
				lg.Error(err)
				return err
			}
			err = uuEmailFlag(user, "--"+flName, flgEmailVal)
			if err != nil {
				return err
			}
		}
		if flName == flgnm.FLG_FIRSTNAME {
			flgFstNameVal, err := cmd.Flags().GetString(flgnm.FLG_FIRSTNAME)
			if err != nil {
				lg.Error(err)
				return err
			}
			err = uuFirstnameFlag(name, "--"+flName, flgFstNameVal)
			if err != nil {
				return err
			}
		}
		if flName == flgnm.FLG_FORCE {
			flgForceVal, err := cmd.Flags().GetString(flgnm.FLG_FORCE)
			if err != nil {
				lg.Error(err)
				return err
			}
			err = uuForceFlag(user, flgForceVal)
			if err != nil {
				return err
			}
		}
		if flName == flgnm.FLG_GAL {
			flgGalVal, err := cmd.Flags().GetBool(flgnm.FLG_GAL)
			if err != nil {
				lg.Error(err)
				return err
			}
			uuGalFlag(user, flgGalVal)
		}
		if flName == flgnm.FLG_LASTNAME {
			flgLstNameVal, err := cmd.Flags().GetString(flgnm.FLG_LASTNAME)
			if err != nil {
				lg.Error(err)
				return err
			}
			err = uuLastnameFlag(name, "--"+flName, flgLstNameVal)
			if err != nil {
				return err
			}
		}
		if flName == flgnm.FLG_ORGUNIT {
			flgOUVal, err := cmd.Flags().GetString(flgnm.FLG_ORGUNIT)
			if err != nil {
				lg.Error(err)
				return err
			}
			err = uuOrgunitFlag(user, "--"+flName, flgOUVal)
			if err != nil {
				return err
			}
		}
		if flName == flgnm.FLG_PASSWORD {
			flgPasswdVal, err := cmd.Flags().GetString(flgnm.FLG_PASSWORD)
			if err != nil {
				lg.Error(err)
				return err
			}
			err = uuPasswordFlag(user, "--"+flName, flgPasswdVal)
			if err != nil {
				return err
			}
		}
		if flName == flgnm.FLG_RECEMAIL {
			flgRecEmailVal, err := cmd.Flags().GetString(flgnm.FLG_RECEMAIL)
			if err != nil {
				lg.Error(err)
				return err
			}
			err = uuRecoveryEmailFlag(user, "--"+flName, flgRecEmailVal)
			if err != nil {
				return err
			}
		}
		if flName == flgnm.FLG_RECPHONE {
			flgRecPhoneVal, err := cmd.Flags().GetString(flgnm.FLG_RECPHONE)
			if err != nil {
				lg.Error(err)
				return err
			}
			err = uuRecoveryPhoneFlag(user, "--"+flName, flgRecPhoneVal)
			if err != nil {
				return err
			}
		}
		if flName == flgnm.FLG_SUSPENDED {
			flgSuspendedVal, err := cmd.Flags().GetBool(flgnm.FLG_SUSPENDED)
			if err != nil {
				lg.Error(err)
				return err
			}
			uuSuspendedFlag(user, flgSuspendedVal)
		}
	}
	lg.Debug("finished processUpdUsrFlags()")
	return nil
}

func uuChangePasswordFlag(user *admin.User, flgVal bool) {
	lg.Debugw("starting uuChangePasswordFlag()",
		"flgVal", flgVal)
	if flgVal {
		user.ChangePasswordAtNextLogin = true
	} else {
		user.ChangePasswordAtNextLogin = false
		user.ForceSendFields = append(user.ForceSendFields, "ChangePasswordAtNextLogin")
	}
	lg.Debug("finished uuChangePasswordFlag()")
}

func uuEmailFlag(user *admin.User, flagName string, flgVal string) error {
	lg.Debugw("starting uuEmailFlag()",
		"flagName", flagName,
		"flgVal", flgVal)
	if flgVal == "" {
		err := fmt.Errorf(gmess.ERR_EMPTYSTRING, flagName)
		return err
	}
	user.PrimaryEmail = flgVal
	lg.Debug("finished uuEmailFlag()")
	return nil
}

func uuFirstnameFlag(name *admin.UserName, flagName string, flgVal string) error {
	lg.Debugw("starting uuFirstnameFlag()",
		"flagName", flagName,
		"flgVal", flgVal)
	if flgVal == "" {
		err := fmt.Errorf(gmess.ERR_EMPTYSTRING, flagName)
		return err
	}
	name.GivenName = flgVal
	lg.Debug("finished uuFirstnameFlag()")
	return nil
}

func uuForceFlag(user *admin.User, flgVal string) error {
	lg.Debugw("starting uuForceFlag()",
		"flgVal", flgVal)
	fields, err := cmn.ParseForceSend(flgVal, usrs.UserAttrMap)
	if err != nil {
		lg.Error(err)
		return err
	}
	for _, fld := range fields {
		user.ForceSendFields = append(user.ForceSendFields, fld)
	}
	lg.Debug("finished uuForceFlag()")
	return nil
}

func uuGalFlag(user *admin.User, flgVal bool) {
	lg.Debugw("starting uuGalFlag()",
		"flgVal", flgVal)
	if flgVal {
		user.IncludeInGlobalAddressList = true
	} else {
		user.IncludeInGlobalAddressList = false
		user.ForceSendFields = append(user.ForceSendFields, "IncludeInGlobalAddressList")
	}
	lg.Debug("finished uuGalFlag()")
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
	if string(recoveryPhone[0]) != "+" {
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
