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
	Use:  "user <user email address>",
	Args: cobra.ExactArgs(1),
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
	createUserCmd.Flags().StringVarP(&firstName, "firstname", "f", "", "user's first name")
	createUserCmd.Flags().StringVar(&forceSend, "force", "", "field list for ForceSendFields separated by (~)")
	createUserCmd.Flags().StringVarP(&lastName, "lastname", "l", "", "user's last name")
	createUserCmd.Flags().StringVarP(&orgUnit, "orgunit", "o", "", "user's orgunit")
	createUserCmd.Flags().StringVarP(&password, "password", "p", "", "user's password")
	createUserCmd.Flags().StringVarP(&recoveryEmail, "recovery-email", "z", "", "user's recovery email address")
	createUserCmd.Flags().StringVarP(&recoveryPhone, "recovery-phone", "k", "", "user's recovery phone")
	createUserCmd.Flags().BoolVarP(&suspended, "suspended", "s", false, "user is suspended")

	createUserCmd.MarkFlagRequired("firstname")
	createUserCmd.MarkFlagRequired("lastname")
	createUserCmd.MarkFlagRequired("password")
}

func processCrtUsrFlags(cmd *cobra.Command, user *admin.User, name *admin.UserName, flagNames []string) error {
	logger.Debugw("starting processCrtUsrFlags()",
		"flagNames", flagNames)

	for _, flName := range flagNames {
		switch flName {
		case "change-password":
			user.ChangePasswordAtNextLogin = true
		case "firstname":
			name.GivenName = firstName
		case "force":
			fields, err := cmn.ParseForceSend(forceSend, usrs.UserAttrMap)
			if err != nil {
				logger.Error(err)
				return err
			}
			for _, fld := range fields {
				user.ForceSendFields = append(user.ForceSendFields, fld)
			}
		case "lastname":
			name.FamilyName = lastName
		case "gal":
			if !gal {
				user.IncludeInGlobalAddressList = false
				user.ForceSendFields = append(user.ForceSendFields, "IncludeInGlobalAddressList")
			}
		case "orgunit":
			user.OrgUnitPath = orgUnit
		case "password":
			if password == "" {
				err := fmt.Errorf(cmn.ErrEmptyString, "--password")
				logger.Error(err)
				return err
			}
			pwd, err := cmn.HashPassword(password)
			if err != nil {
				logger.Error(err)
				return err
			}
			user.Password = pwd
			user.HashFunction = cmn.HashFunction
		case "recovery-email":
			if recoveryEmail == "" {
				err := fmt.Errorf(cmn.ErrEmptyString, "--recovery-email")
				logger.Error(err)
				return err
			}
			user.RecoveryEmail = recoveryEmail
		case "recovery-phone":
			if recoveryPhone == "" {
				err := fmt.Errorf(cmn.ErrEmptyString, "--recovery-phone")
				logger.Error(err)
				return err
			}
			if string(recoveryPhone[0]) != "+" {
				err := fmt.Errorf(cmn.ErrInvalidRecoveryPhone, recoveryPhone)
				logger.Error(err)
				return err
			}
			user.RecoveryPhone = recoveryPhone
		case "suspended":
			user.Suspended = true
		}
	}
	logger.Debug("finished processCrtUsrFlags()")
	return nil
}
