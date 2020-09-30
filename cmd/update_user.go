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
	updateUserCmd.Flags().BoolVarP(&noChangePassword, "no-password-change", "d", false, "user doesn't have to change password on next login")
	updateUserCmd.Flags().StringVarP(&userEmail, "email", "e", "", "user's primary email address")
	updateUserCmd.Flags().StringVarP(&firstName, "firstname", "f", "", "user's first name")
	updateUserCmd.Flags().StringVar(&forceSend, "force", "", "field list for ForceSendFields separated by (~)")
	updateUserCmd.Flags().BoolVarP(&gal, "gal", "g", false, "display user in Global Address List")
	updateUserCmd.Flags().StringVarP(&lastName, "lastname", "l", "", "user's last name")
	updateUserCmd.Flags().BoolVarP(&noGAL, "nogal", "n", false, "do not display user in Global Address List")
	updateUserCmd.Flags().StringVarP(&orgUnit, "orgunit", "o", "", "user's orgunit")
	updateUserCmd.Flags().StringVarP(&password, "password", "p", "", "user's password")
	updateUserCmd.Flags().StringVarP(&recoveryEmail, "recovery-email", "z", "", "user's recovery email address")
	updateUserCmd.Flags().StringVarP(&recoveryPhone, "recovery-phone", "k", "", "user's recovery phone")
	updateUserCmd.Flags().BoolVarP(&suspended, "suspended", "s", false, "user is suspended")
	updateUserCmd.Flags().BoolVarP(&unSuspended, "unsuspended", "u", false, "user is unsuspended")
}

func processUpdUsrFlags(cmd *cobra.Command, user *admin.User, name *admin.UserName, flagNames []string) error {
	logger.Debugw("starting processUpdUsrFlags()",
		"flagNames", flagNames)
	for _, flName := range flagNames {
		switch flName {
		case "change-password":
			user.ChangePasswordAtNextLogin = true
		case "email":
			if userEmail == "" {
				err := errors.New(cmn.ErrEmptyEmail)
				logger.Error(err)
				return err
			}
			user.PrimaryEmail = userEmail
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
		case "gal":
			user.IncludeInGlobalAddressList = true
		case "lastname":
			name.FamilyName = lastName
		case "no-password-change":
			user.ChangePasswordAtNextLogin = false
			user.ForceSendFields = append(user.ForceSendFields, "ChangePasswordAtNextLogin")
		case "nogal":
			user.IncludeInGlobalAddressList = false
			user.ForceSendFields = append(user.ForceSendFields, "IncludeInGlobalAddressList")
		case "orgunit":
			user.OrgUnitPath = orgUnit
		case "password":
			if password == "" {
				err := errors.New(cmn.ErrEmptyPassword)
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
			user.RecoveryEmail = recoveryEmail
			user.ForceSendFields = append(user.ForceSendFields, "RecoveryEmail")
		case "recovery-phone":
			if recoveryPhone != "" && string(recoveryPhone[0]) != "+" {
				err := fmt.Errorf(cmn.ErrInvalidRecoveryPhone, recoveryPhone)
				logger.Error(err)
				return err
			}
			user.RecoveryPhone = recoveryPhone
			user.ForceSendFields = append(user.ForceSendFields, "RecoveryPhone")
		case "suspended":
			user.Suspended = true
		case "unsuspended":
			user.Suspended = false
			user.ForceSendFields = append(user.ForceSendFields, "Suspended")
		}
	}
	logger.Debug("finished processUpdUsrFlags()")
	return nil
}
