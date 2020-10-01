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
	logger.Debugw("starting doUpdateUser()",
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
		logger.Error(err)
		return err
	}

	if name.FamilyName != "" || name.FullName != "" || name.GivenName != "" {
		user.Name = name
	}

	if attrs != "" {
		attrUser := new(admin.User)
		emptyVals := cmn.EmptyValues{}
		jsonBytes := []byte(attrs)
		if !json.Valid(jsonBytes) {
			err = errors.New(cmn.ErrInvalidJSONAttr)
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

		err = mergo.Merge(user, attrUser)
		if err != nil {
			logger.Error(err)
			return err
		}
	}

	ds, err := cmn.CreateDirectoryService(admin.AdminDirectoryUserScope)
	if err != nil {
		logger.Error(err)
		return err
	}

	uuc := ds.Users.Update(userKey, user)
	_, err = uuc.Do()
	if err != nil {
		logger.Error(err)
		return err
	}

	logger.Infof(cmn.InfoUserUpdated, userKey)
	fmt.Println(cmn.GminMessage(fmt.Sprintf(cmn.InfoUserUpdated, userKey)))

	logger.Debug("finished doUpdateUser()")
	return nil
}

func init() {
	updateCmd.AddCommand(updateUserCmd)

	updateUserCmd.Flags().StringVarP(&attrs, "attributes", "a", "", "user's attributes as a JSON string")
	updateUserCmd.Flags().BoolVarP(&changePassword, "change-password", "c", false, "user must change password on next login")
	updateUserCmd.Flags().StringVarP(&userEmail, "email", "e", "", "user's primary email address")
	updateUserCmd.Flags().StringVarP(&firstName, "firstname", "f", "", "user's first name")
	updateUserCmd.Flags().StringVar(&forceSend, "force", "", "field list for ForceSendFields separated by (~)")
	updateUserCmd.Flags().BoolVarP(&gal, "gal", "g", false, "display user in Global Address List")
	updateUserCmd.Flags().StringVarP(&lastName, "lastname", "l", "", "user's last name")
	updateUserCmd.Flags().StringVarP(&orgUnit, "orgunit", "o", "", "user's orgunit")
	updateUserCmd.Flags().StringVarP(&password, "password", "p", "", "user's password")
	updateUserCmd.Flags().StringVarP(&recoveryEmail, "recovery-email", "z", "", "user's recovery email address")
	updateUserCmd.Flags().StringVarP(&recoveryPhone, "recovery-phone", "k", "", "user's recovery phone")
	updateUserCmd.Flags().BoolVarP(&suspended, "suspended", "s", false, "user is suspended")
}

func processUpdUsrFlags(cmd *cobra.Command, user *admin.User, name *admin.UserName, flagNames []string) error {
	logger.Debugw("starting processUpdUsrFlags()",
		"flagNames", flagNames)
	for _, flName := range flagNames {
		if flName == "change-password" {
			uuChangePasswordFlag(user)
		}
		if flName == "email" {
			err := uuEmailFlag(user, "--"+flName)
			if err != nil {
				return err
			}
		}
		if flName == "firstname" {
			err := uuFirstnameFlag(name, "--"+flName)
			if err != nil {
				return err
			}
		}
		if flName == "force" {
			err := uuForceFlag(user)
			if err != nil {
				return err
			}
		}
		if flName == "gal" {
			uuGalFlag(user)
		}
		if flName == "lastname" {
			err := uuLastnameFlag(name, "--"+flName)
			if err != nil {
				return err
			}
		}
		if flName == "orgunit" {
			err := uuOrgunitFlag(user, "--"+flName)
			if err != nil {
				return err
			}
		}
		if flName == "password" {
			err := uuPasswordFlag(user, "--"+flName)
			if err != nil {
				return err
			}
		}
		if flName == "recovery-email" {
			err := uuRecoveryEmailFlag(user, "--"+flName)
			if err != nil {
				return err
			}
		}
		if flName == "recovery-phone" {
			err := uuRecoveryPhoneFlag(user, "--"+flName)
			if err != nil {
				return err
			}
		}
		if flName == "suspended" {
			uuSuspendedFlag(user)
		}
	}
	logger.Debug("finished processUpdUsrFlags()")
	return nil
}

func uuChangePasswordFlag(user *admin.User) {
	logger.Debug("starting uuChangePasswordFlag()")
	if changePassword {
		user.ChangePasswordAtNextLogin = true
	} else {
		user.ChangePasswordAtNextLogin = false
		user.ForceSendFields = append(user.ForceSendFields, "ChangePasswordAtNextLogin")
	}
	logger.Debug("finished uuChangePasswordFlag()")
}

func uuEmailFlag(user *admin.User, flagName string) error {
	logger.Debugw("starting uuEmailFlag()",
		"flagName", flagName)
	if userEmail == "" {
		err := fmt.Errorf(cmn.ErrEmptyString, flagName)
		return err
	}
	user.PrimaryEmail = userEmail
	logger.Debug("finished uuEmailFlag()")
	return nil
}

func uuFirstnameFlag(name *admin.UserName, flagName string) error {
	logger.Debugw("starting uuFirstnameFlag()",
		"flagName", flagName)
	if firstName == "" {
		err := fmt.Errorf(cmn.ErrEmptyString, flagName)
		return err
	}
	name.GivenName = firstName
	logger.Debug("finished uuFirstnameFlag()")
	return nil
}

func uuForceFlag(user *admin.User) error {
	logger.Debug("starting uuForceFlag()")
	fields, err := cmn.ParseForceSend(forceSend, usrs.UserAttrMap)
	if err != nil {
		logger.Error(err)
		return err
	}
	for _, fld := range fields {
		user.ForceSendFields = append(user.ForceSendFields, fld)
	}
	logger.Debug("finished uuForceFlag()")
	return nil
}

func uuGalFlag(user *admin.User) {
	logger.Debug("starting uuGalFlag()")
	if gal {
		user.IncludeInGlobalAddressList = true
	} else {
		user.IncludeInGlobalAddressList = false
		user.ForceSendFields = append(user.ForceSendFields, "IncludeInGlobalAddressList")
	}
	logger.Debug("finished uuGalFlag()")
}

func uuLastnameFlag(name *admin.UserName, flagName string) error {
	logger.Debugw("starting uuLastnameFlag()",
		"flagName", flagName)
	if lastName == "" {
		err := fmt.Errorf(cmn.ErrEmptyString, flagName)
		return err
	}
	name.FamilyName = lastName
	logger.Debug("finished uuLastnameFlag()")
	return nil
}

func uuOrgunitFlag(user *admin.User, flagName string) error {
	logger.Debugw("starting uuOrgunitFlag()",
		"flagName", flagName)
	if orgUnit == "" {
		err := fmt.Errorf(cmn.ErrEmptyString, flagName)
		if err != nil {
			return err
		}
	}
	user.OrgUnitPath = orgUnit
	logger.Debug("finished uuOrgunitFlag()")
	return nil
}

func uuPasswordFlag(user *admin.User, flagName string) error {
	logger.Debugw("starting uuPasswordFlag()",
		"flagName", flagName)
	if password == "" {
		err := fmt.Errorf(cmn.ErrEmptyString, flagName)
		return err
	}
	pwd, err := cmn.HashPassword(password)
	if err != nil {
		return err
	}
	user.Password = pwd
	user.HashFunction = cmn.HashFunction
	logger.Debug("finished uuPasswordFlag()")
	return nil
}

func uuRecoveryEmailFlag(user *admin.User, flagName string) error {
	logger.Debugw("starting uuRecoveryEmailFlag()",
		"flagName", flagName)
	if recoveryEmail == "" {
		err := fmt.Errorf(cmn.ErrEmptyString, flagName)
		return err
	}
	user.RecoveryEmail = recoveryEmail
	logger.Debug("finished uuRecoveryEmailFlag()")
	return nil
}

func uuRecoveryPhoneFlag(user *admin.User, flagName string) error {
	logger.Debugw("starting uuRecoveryPhoneFlag()",
		"flagName", flagName)
	if recoveryPhone == "" {
		err := fmt.Errorf(cmn.ErrEmptyString, flagName)
		return err
	}
	if string(recoveryPhone[0]) != "+" {
		err := fmt.Errorf(cmn.ErrInvalidRecoveryPhone, recoveryPhone)
		return err
	}
	user.RecoveryPhone = recoveryPhone
	logger.Debug("finished uuRecoveryPhoneFlag()")
	return nil
}

func uuSuspendedFlag(user *admin.User) {
	logger.Debug("starting uuSuspendedFlag()")
	if suspended {
		user.Suspended = true
	} else {
		user.Suspended = false
		user.ForceSendFields = append(user.ForceSendFields, "Suspended")
	}
	logger.Debug("finished uuSuspendedFlag()")
}
