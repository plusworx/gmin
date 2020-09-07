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
	Use:   "user <user email address>",
	Args:  cobra.ExactArgs(1),
	Short: "Creates a user",
	Long: `Creates a user.
	
	Examples:	gmin create user another.user@mycompany.com  -f Another -l User -p strongpassword
			gmin crt user finance.person@mycompany.com -f Finance -l Person -p greatpassword -c`,
	RunE: doCreateUser,
}

func doCreateUser(cmd *cobra.Command, args []string) error {
	var flagsPassed []string

	user := new(admin.User)
	name := new(admin.UserName)

	ok := valid.IsEmail(args[0])
	if !ok {
		return errors.New("gmin: error - invalid email address")
	}

	user.PrimaryEmail = args[0]

	// Collect names of command flags passed in
	cmd.Flags().Visit(func(f *pflag.Flag) {
		flagsPassed = append(flagsPassed, f.Name)
	})

	// Process command flags
	err := processCrtUsrFlags(cmd, user, name, flagsPassed)
	if err != nil {
		return err
	}

	user.Name = name

	if attrs != "" {
		attrUser := new(admin.User)
		emptyVals := cmn.EmptyValues{}
		jsonBytes := []byte(attrs)
		if !json.Valid(jsonBytes) {
			return errors.New("gmin: error - attribute string is not valid JSON")
		}

		outStr, err := cmn.ParseInputAttrs(jsonBytes)
		if err != nil {
			return err
		}

		err = cmn.ValidateInputAttrs(outStr, usrs.UserAttrMap)
		if err != nil {
			return err
		}

		err = json.Unmarshal(jsonBytes, &attrUser)
		if err != nil {
			return err
		}

		err = json.Unmarshal(jsonBytes, &emptyVals)
		if err != nil {
			return err
		}
		if len(emptyVals.ForceSendFields) > 0 {
			attrUser.ForceSendFields = emptyVals.ForceSendFields
		}

		if user.Password == "" && attrUser.Password != "" {
			pwd, err := cmn.HashPassword(attrUser.Password)
			if err != nil {
				return err
			}
			attrUser.Password = pwd
			attrUser.HashFunction = cmn.HashFunction
		}

		err = mergo.Merge(user, attrUser)
		if err != nil {
			return err
		}
	}

	if user.Name.GivenName == "" || user.Name.FamilyName == "" || user.Password == "" {
		return errors.New("gmin: error - firstname, lastname and password must all be provided")
	}

	ds, err := cmn.CreateDirectoryService(admin.AdminDirectoryUserScope)
	if err != nil {
		return err
	}

	uic := ds.Users.Insert(user)
	newUser, err := uic.Do()
	if err != nil {
		return err
	}

	fmt.Println(cmn.GminMessage("**** gmin: user " + newUser.PrimaryEmail + " created ****"))

	return nil
}

func init() {
	createCmd.AddCommand(createUserCmd)

	createUserCmd.Flags().StringVarP(&attrs, "attributes", "a", "", "user's attributes as a JSON string")
	createUserCmd.Flags().BoolVarP(&changePassword, "changepassword", "c", false, "user must change password on next login")
	createUserCmd.Flags().StringVarP(&firstName, "firstname", "f", "", "user's first name")
	createUserCmd.Flags().StringVarP(&forceSend, "force", "", "", "field list for ForceSendFields separated by (~)")
	createUserCmd.Flags().StringVarP(&lastName, "lastname", "l", "", "user's last name")
	createUserCmd.Flags().BoolVarP(&noGAL, "nogal", "n", false, "do not display user in Global Address List")
	createUserCmd.Flags().StringVarP(&orgUnit, "orgunit", "o", "", "user's orgunit")
	createUserCmd.Flags().StringVarP(&password, "password", "p", "", "user's password")
	createUserCmd.Flags().StringVarP(&recoveryEmail, "recoveryemail", "z", "", "user's recovery email address")
	createUserCmd.Flags().StringVarP(&recoveryPhone, "recoveryphone", "k", "", "user's recovery phone")
	createUserCmd.Flags().BoolVarP(&suspended, "suspended", "s", false, "user is suspended")

	createUserCmd.MarkFlagRequired("firstname")
	createUserCmd.MarkFlagRequired("lastname")
	createUserCmd.MarkFlagRequired("password")
}

func processCrtUsrFlags(cmd *cobra.Command, user *admin.User, name *admin.UserName, flagNames []string) error {
	for _, flName := range flagNames {
		switch flName {
		case "changepassword":
			user.ChangePasswordAtNextLogin = true
		case "firstname":
			name.GivenName = firstName
		case "forceSend":
			fields, err := cmn.ParseForceSend(forceSend, usrs.UserAttrMap)
			if err != nil {
				return err
			}
			for _, fld := range fields {
				user.ForceSendFields = append(user.ForceSendFields, fld)
			}
		case "lastname":
			name.FamilyName = lastName
		case "nogal":
			user.IncludeInGlobalAddressList = false
			user.ForceSendFields = append(user.ForceSendFields, "IncludeInGlobalAddressList")
		case "orgunit":
			user.OrgUnitPath = orgUnit
		case "password":
			if password == "" {
				return errors.New("gmin: error - password cannot be empty string")
			}
			pwd, err := cmn.HashPassword(password)
			if err != nil {
				return err
			}
			user.Password = pwd
			user.HashFunction = cmn.HashFunction
		case "recoveryemail":
			if recoveryEmail == "" {
				return errors.New("gmin: error - recoveryemail cannot be empty string")
			}
			user.RecoveryEmail = recoveryEmail
		case "recoveryphone":
			if recoveryPhone == "" {
				return errors.New("gmin: error - recoveryphone cannot be empty string")
			}
			if string(recoveryPhone[0]) != "+" {
				err := fmt.Errorf("gmin: error - recovery phone number %v must start with '+'", recoveryPhone)
				return err
			}
			user.RecoveryPhone = recoveryPhone
		case "suspended":
			user.Suspended = true
		}
	}
	return nil
}
