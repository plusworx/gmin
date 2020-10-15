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
	logger.Debugw("starting doCreateUser()",
		"args", args)

	var flagsPassed []string

	user := new(admin.User)
	name := new(admin.UserName)

	ok := valid.IsEmail(args[0])
	if !ok {
		err := fmt.Errorf(gmess.ERR_INVALIDEMAILADDRESS, args[0])
		logger.Error(err)
		return err
	}

	user.PrimaryEmail = args[0]

	// Collect names of command flags passed in
	cmd.Flags().Visit(func(f *pflag.Flag) {
		flagsPassed = append(flagsPassed, f.Name)
	})

	// Process command flags
	err := processCrtUsrFlags(cmd, user, name, flagsPassed)
	if err != nil {
		logger.Error(err)
		return err
	}

	user.Name = name

	flgAttrsVal, err := cmd.Flags().GetString(flgnm.FLG_ATTRIBUTES)
	if err != nil {
		logger.Error(err)
		return err
	}

	if flgAttrsVal != "" {
		attrUser := new(admin.User)
		emptyVals := cmn.EmptyValues{}
		jsonBytes := []byte(flgAttrsVal)
		if !json.Valid(jsonBytes) {
			err = errors.New(gmess.ERR_INVALIDJSONATTR)
			logger.Error(err)
			return err
		}

		outStr, err := cmn.ParseInputAttrs(jsonBytes)
		if err != nil {
			logger.Error(err)
			return err
		}

		err = cmn.ValidateInputAttrs(outStr, usrs.UserAttrMap)
		if err != nil {
			logger.Error(err)
			return err
		}

		err = json.Unmarshal(jsonBytes, &attrUser)
		if err != nil {
			logger.Error(err)
			return err
		}

		err = json.Unmarshal(jsonBytes, &emptyVals)
		if err != nil {
			logger.Error(err)
			return err
		}
		if len(emptyVals.ForceSendFields) > 0 {
			attrUser.ForceSendFields = emptyVals.ForceSendFields
		}

		if user.Password == "" && attrUser.Password != "" {
			pwd, err := cmn.HashPassword(attrUser.Password)
			if err != nil {
				logger.Error(err)
				return err
			}
			attrUser.Password = pwd
			attrUser.HashFunction = cmn.HASHFUNCTION
		}

		err = mergo.Merge(user, attrUser)
		if err != nil {
			logger.Error(err)
			return err
		}
	}

	if user.Name.GivenName == "" || user.Name.FamilyName == "" || user.Password == "" {
		err = errors.New(gmess.ERR_MISSINGUSERDATA)
		logger.Error(err)
		return err
	}

	ds, err := cmn.CreateDirectoryService(admin.AdminDirectoryUserScope)
	if err != nil {
		logger.Error(err)
		return err
	}

	uic := ds.Users.Insert(user)
	newUser, err := uic.Do()
	if err != nil {
		logger.Error(err)
		return err
	}

	logger.Infof(gmess.INFO_USERCREATED, newUser.PrimaryEmail)
	fmt.Println(cmn.GminMessage(fmt.Sprintf(gmess.INFO_USERCREATED, newUser.PrimaryEmail)))

	logger.Debug("finished doCreateUser()")
	return nil
}

func init() {
	createCmd.AddCommand(createUserCmd)

	createUserCmd.Flags().StringVarP(&attrs, flgnm.FLG_ATTRIBUTES, "a", "", "user's attributes as a JSON string")
	createUserCmd.Flags().BoolVarP(&changePassword, flgnm.FLG_CHANGEPWD, "c", false, "user must change password on next login")
	createUserCmd.Flags().StringVarP(&firstName, flgnm.FLG_FIRSTNAME, "f", "", "user's first name")
	createUserCmd.Flags().StringVar(&forceSend, flgnm.FLG_FORCE, "", "field list for ForceSendFields separated by (~)")
	createUserCmd.Flags().BoolVarP(&gal, flgnm.FLG_GAL, "g", false, "user is included in Global Address List")
	createUserCmd.Flags().StringVarP(&lastName, flgnm.FLG_LASTNAME, "l", "", "user's last name")
	createUserCmd.Flags().StringVarP(&orgUnit, flgnm.FLG_ORGUNIT, "o", "", "user's orgunit")
	createUserCmd.Flags().StringVarP(&password, flgnm.FLG_PASSWORD, "p", "", "user's password")
	createUserCmd.Flags().StringVarP(&recoveryEmail, flgnm.FLG_RECEMAIL, "z", "", "user's recovery email address")
	createUserCmd.Flags().StringVarP(&recoveryPhone, flgnm.FLG_RECPHONE, "k", "", "user's recovery phone")
	createUserCmd.Flags().BoolVarP(&suspended, flgnm.FLG_SUSPENDED, "s", false, "user is suspended")

	createUserCmd.MarkFlagRequired(flgnm.FLG_FIRSTNAME)
	createUserCmd.MarkFlagRequired(flgnm.FLG_LASTNAME)
	createUserCmd.MarkFlagRequired(flgnm.FLG_PASSWORD)
}

func cuFirstnameFlag(name *admin.UserName, flagName string, flgVal string) error {
	logger.Debugw("starting cuFirstnameFlag()",
		"flagName", flagName)
	if flgVal == "" {
		err := fmt.Errorf(gmess.ERR_EMPTYSTRING, flagName)
		if err != nil {
			logger.Error(err)
			return err
		}
	}
	name.GivenName = flgVal
	logger.Debug("finished cuFirstnameFlag()")
	return nil
}

func cuForceFlag(forceSend string, user *admin.User) error {
	logger.Debugw("starting cuForceFlag()",
		"forceSend", forceSend)
	fields, err := cmn.ParseForceSend(forceSend, usrs.UserAttrMap)
	if err != nil {
		logger.Error(err)
		return err
	}
	for _, fld := range fields {
		user.ForceSendFields = append(user.ForceSendFields, fld)
	}
	logger.Debug("finished cuForceFlag()")
	return nil
}

func cuGalFlag(user *admin.User, flgVal bool) {
	if !flgVal {
		user.IncludeInGlobalAddressList = false
		user.ForceSendFields = append(user.ForceSendFields, "IncludeInGlobalAddressList")
	}
}

func cuLastnameFlag(name *admin.UserName, flagName string, flgVal string) error {
	logger.Debugw("starting cuLastnameFlag()",
		"flagName", flagName)
	if flgVal == "" {
		err := fmt.Errorf(gmess.ERR_EMPTYSTRING, flagName)
		if err != nil {
			logger.Error(err)
			return err
		}
	}
	name.FamilyName = flgVal
	logger.Debug("finished cuLastnameFlag()")
	return nil
}

func cuOrgunitFlag(user *admin.User, flagName string, flgVal string) error {
	logger.Debugw("starting cuOrgunitFlag()",
		"flagName", flagName)
	if flgVal == "" {
		err := fmt.Errorf(gmess.ERR_EMPTYSTRING, flagName)
		if err != nil {
			logger.Error(err)
			return err
		}
	}
	user.OrgUnitPath = flgVal
	logger.Debug("finished cuOrgunitFlag()")
	return nil
}

func cuPasswordFlag(user *admin.User, flagName string, flgVal string) error {
	logger.Debugw("starting cuPasswordFlag()",
		"flagName", flagName)
	if flgVal == "" {
		err := fmt.Errorf(gmess.ERR_EMPTYSTRING, flagName)
		logger.Error(err)
		return err
	}
	pwd, err := cmn.HashPassword(flgVal)
	if err != nil {
		logger.Error(err)
		return err
	}
	user.Password = pwd
	user.HashFunction = cmn.HASHFUNCTION
	logger.Debug("finished cuPasswordFlag()")
	return nil
}

func cuRecoveryEmailFlag(user *admin.User, flagName string, flgVal string) error {
	logger.Debugw("starting cuRecoveryEmailFlag()",
		"flagName", flagName)
	if flgVal == "" {
		err := fmt.Errorf(gmess.ERR_EMPTYSTRING, flagName)
		logger.Error(err)
		return err
	}
	ok := valid.IsEmail(flgVal)
	if !ok {
		err := fmt.Errorf(gmess.ERR_INVALIDEMAILADDRESS, flgVal)
		logger.Error(err)
		return err
	}
	user.RecoveryEmail = flgVal
	logger.Debug("finished cuRecoveryEmailFlag()")
	return nil
}

func cuRecoveryPhoneFlag(user *admin.User, flagName string, flgVal string) error {
	logger.Debugw("starting cuRecoveryPhoneFlag()",
		"flagName", flagName)
	if flgVal == "" {
		err := fmt.Errorf(gmess.ERR_EMPTYSTRING, flagName)
		logger.Error(err)
		return err
	}
	if string(recoveryPhone[0]) != "+" {
		err := fmt.Errorf(gmess.ERR_INVALIDRECOVERYPHONE, flgVal)
		logger.Error(err)
		return err
	}
	user.RecoveryPhone = flgVal
	logger.Debug("finished cuRecoveryPhoneFlag()")
	return nil
}

func processCrtUsrFlags(cmd *cobra.Command, user *admin.User, name *admin.UserName, flagNames []string) error {
	logger.Debugw("starting processCrtUsrFlags()",
		"flagNames", flagNames)
	for _, flName := range flagNames {
		if flName == flgnm.FLG_CHANGEPWD {
			flgChgPwdVal, err := cmd.Flags().GetBool(flName)
			if err != nil {
				logger.Error(err)
				return err
			}
			if flgChgPwdVal {
				user.ChangePasswordAtNextLogin = true
			}
		}
		if flName == flgnm.FLG_FIRSTNAME {
			flgFstNameVal, err := cmd.Flags().GetString(flName)
			if err != nil {
				logger.Error(err)
				return err
			}

			err = cuFirstnameFlag(name, "--"+flName, flgFstNameVal)
			if err != nil {
				return err
			}
		}
		if flName == flgnm.FLG_FORCE {
			flgForceVal, err := cmd.Flags().GetString(flName)
			if err != nil {
				logger.Error(err)
				return err
			}

			err = cuForceFlag(flgForceVal, user)
			if err != nil {
				return err
			}
		}
		if flName == flgnm.FLG_GAL {
			flgGalVal, err := cmd.Flags().GetBool(flName)
			if err != nil {
				logger.Error(err)
				return err
			}
			cuGalFlag(user, flgGalVal)
		}
		if flName == flgnm.FLG_LASTNAME {
			flgLstNameVal, err := cmd.Flags().GetString(flName)
			if err != nil {
				logger.Error(err)
				return err
			}

			err = cuLastnameFlag(name, "--"+flName, flgLstNameVal)
			if err != nil {
				return err
			}
		}
		if flName == flgnm.FLG_ORGUNIT {
			flgOUVal, err := cmd.Flags().GetString(flName)
			if err != nil {
				logger.Error(err)
				return err
			}

			err = cuOrgunitFlag(user, "--"+flName, flgOUVal)
			if err != nil {
				return err
			}
		}
		if flName == flgnm.FLG_PASSWORD {
			flgPwdVal, err := cmd.Flags().GetString(flName)
			if err != nil {
				logger.Error(err)
				return err
			}

			err = cuPasswordFlag(user, "--"+flName, flgPwdVal)
			if err != nil {
				return err
			}
		}
		if flName == flgnm.FLG_RECEMAIL {
			flgRecEmailVal, err := cmd.Flags().GetString(flName)
			if err != nil {
				logger.Error(err)
				return err
			}

			err = cuRecoveryEmailFlag(user, "--"+flName, flgRecEmailVal)
			if err != nil {
				return err
			}
		}
		if flName == flgnm.FLG_RECPHONE {
			flgRecPhoneVal, err := cmd.Flags().GetString(flName)
			if err != nil {
				logger.Error(err)
				return err
			}

			err = cuRecoveryPhoneFlag(user, "--"+flName, flgRecPhoneVal)
			if err != nil {
				return err
			}
		}
		if flName == flgnm.FLG_SUSPENDED {
			flgSuspVal, err := cmd.Flags().GetBool(flName)
			if err != nil {
				logger.Error(err)
				return err
			}
			if flgSuspVal {
				user.Suspended = true
			}
		}
	}
	logger.Debug("finished processCrtUsrFlags()")
	return nil
}
