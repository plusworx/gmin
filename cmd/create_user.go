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
		err := fmt.Errorf(cmn.ErrInvalidEmailAddress, args[0])
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

		if user.Password == "" && attrUser.Password != "" {
			pwd, err := cmn.HashPassword(attrUser.Password)
			if err != nil {
				logger.Error(err)
				return err
			}
			attrUser.Password = pwd
			attrUser.HashFunction = cmn.HashFunction
		}

		err = mergo.Merge(user, attrUser)
		if err != nil {
			logger.Error(err)
			return err
		}
	}

	if user.Name.GivenName == "" || user.Name.FamilyName == "" || user.Password == "" {
		err = errors.New(cmn.ErrMissingUserData)
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

	logger.Infof(cmn.InfoUserCreated, newUser.PrimaryEmail)
	fmt.Println(cmn.GminMessage(fmt.Sprintf(cmn.InfoUserCreated, newUser.PrimaryEmail)))

	logger.Debug("finished doCreateUser()")
	return nil
}

func init() {
	createCmd.AddCommand(createUserCmd)

	createUserCmd.Flags().StringVarP(&attrs, "attributes", "a", "", "user's attributes as a JSON string")
	createUserCmd.Flags().BoolVarP(&changePassword, "change-password", "c", false, "user must change password on next login")
	createUserCmd.Flags().StringVarP(&firstName, "first-name", "f", "", "user's first name")
	createUserCmd.Flags().StringVar(&forceSend, "force", "", "field list for ForceSendFields separated by (~)")
	createUserCmd.Flags().StringVarP(&lastName, "last-name", "l", "", "user's last name")
	createUserCmd.Flags().StringVarP(&orgUnit, "orgunit", "o", "", "user's orgunit")
	createUserCmd.Flags().StringVarP(&password, "password", "p", "", "user's password")
	createUserCmd.Flags().StringVarP(&recoveryEmail, "recovery-email", "z", "", "user's recovery email address")
	createUserCmd.Flags().StringVarP(&recoveryPhone, "recovery-phone", "k", "", "user's recovery phone")
	createUserCmd.Flags().BoolVarP(&suspended, "suspended", "s", false, "user is suspended")

	createUserCmd.MarkFlagRequired("first-name")
	createUserCmd.MarkFlagRequired("last-name")
	createUserCmd.MarkFlagRequired("password")
}

func cuFirstnameFlag(name *admin.UserName, flagName string) error {
	logger.Debugw("starting cuFirstnameFlag()",
		"flagName", flagName)
	if firstName == "" {
		err := fmt.Errorf(cmn.ErrEmptyString, flagName)
		if err != nil {
			return err
		}
	}
	name.GivenName = firstName
	logger.Debug("finished cuFirstnameFlag()")
	return nil
}

func cuForceFlag(forceSend string, user *admin.User) error {
	logger.Debugw("starting cuForceFlag()",
		"forceSend", forceSend)
	fields, err := cmn.ParseForceSend(forceSend, usrs.UserAttrMap)
	if err != nil {
		return err
	}
	for _, fld := range fields {
		user.ForceSendFields = append(user.ForceSendFields, fld)
	}
	logger.Debug("finished cuForceFlag()")
	return nil
}

func cuGalFlag(user *admin.User) {
	if !gal {
		user.IncludeInGlobalAddressList = false
		user.ForceSendFields = append(user.ForceSendFields, "IncludeInGlobalAddressList")
	}
}

func cuLastnameFlag(name *admin.UserName, flagName string) error {
	logger.Debugw("starting cuLastnameFlag()",
		"flagName", flagName)
	if lastName == "" {
		err := fmt.Errorf(cmn.ErrEmptyString, flagName)
		if err != nil {
			return err
		}
	}
	name.FamilyName = lastName
	logger.Debug("finished cuLastnameFlag()")
	return nil
}

func cuOrgunitFlag(user *admin.User, flagName string) error {
	logger.Debugw("starting cuOrgunitFlag()",
		"flagName", flagName)
	if orgUnit == "" {
		err := fmt.Errorf(cmn.ErrEmptyString, flagName)
		if err != nil {
			return err
		}
	}
	user.OrgUnitPath = orgUnit
	logger.Debug("finished cuOrgunitFlag()")
	return nil
}

func cuPasswordFlag(user *admin.User, flagName string) error {
	logger.Debugw("starting cuPasswordFlag()",
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
	logger.Debug("finished cuPasswordFlag()")
	return nil
}

func cuRecoveryEmailFlag(user *admin.User, flagName string) error {
	logger.Debugw("starting cuRecoveryEmailFlag()",
		"flagName", flagName)
	if recoveryEmail == "" {
		err := fmt.Errorf(cmn.ErrEmptyString, flagName)
		return err
	}
	user.RecoveryEmail = recoveryEmail
	logger.Debug("finished cuRecoveryEmailFlag()")
	return nil
}

func cuRecoveryPhoneFlag(user *admin.User, flagName string) error {
	logger.Debugw("starting cuRecoveryPhoneFlag()",
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
	logger.Debug("finished cuRecoveryPhoneFlag()")
	return nil
}

func processCrtUsrFlags(cmd *cobra.Command, user *admin.User, name *admin.UserName, flagNames []string) error {
	logger.Debugw("starting processCrtUsrFlags()",
		"flagNames", flagNames)
	for _, flName := range flagNames {
		if flName == "change-password" {
			user.ChangePasswordAtNextLogin = true
		}
		if flName == "first-name" {
			err := cuFirstnameFlag(name, "--"+flName)
			if err != nil {
				return err
			}
		}
		if flName == "force" {
			err := cuForceFlag(forceSend, user)
			if err != nil {
				return err
			}
		}
		if flName == "last-name" {
			err := cuLastnameFlag(name, "--"+flName)
			if err != nil {
				return err
			}
		}
		if flName == "gal" {
			cuGalFlag(user)
		}
		if flName == "orgunit" {
			err := cuOrgunitFlag(user, "--"+flName)
			if err != nil {
				return err
			}
		}
		if flName == "password" {
			err := cuPasswordFlag(user, "--"+flName)
			if err != nil {
				return err
			}
		}
		if flName == "recovery-email" {
			err := cuRecoveryEmailFlag(user, "--"+flName)
			if err != nil {
				return err
			}
		}
		if flName == "recovery-phone" {
			err := cuRecoveryPhoneFlag(user, "--"+flName)
			if err != nil {
				return err
			}
		}
		if flName == "suspended" {
			user.Suspended = true
		}
	}
	logger.Debug("finished processCrtUsrFlags()")
	return nil
}
