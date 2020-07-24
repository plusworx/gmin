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

	valid "github.com/asaskevich/govalidator"
	"github.com/imdario/mergo"
	admin "google.golang.org/api/admin/directory/v1"
)

var createUserCmd = &cobra.Command{
	Use:   "user <user email address>",
	Args:  cobra.ExactArgs(1),
	Short: "Creates a user",
	Long: `Creates a user.
	
	Examples: gmin create user another.user@mycompany.com  -f Another -l User -p strongpassword
	          gmin crt user finance.person@mycompany.com -f Finance -l Person -p greatpassword -c`,
	RunE: doCreateUser,
}

func doCreateUser(cmd *cobra.Command, args []string) error {
	var bPasswordCreated bool

	user := new(admin.User)
	name := new(admin.UserName)

	ok := valid.IsEmail(args[0])
	if !ok {
		return errors.New("gmin: error - invalid email address")
	}

	user.PrimaryEmail = args[0]

	if changePassword {
		user.ChangePasswordAtNextLogin = true
	}

	if firstName != "" {
		name.GivenName = firstName
	}

	if forceSend != "" {
		fields, err := cmn.ParseForceSend(forceSend, usrs.UserAttrMap)
		if err != nil {
			return err
		}
		for _, f := range fields {
			user.ForceSendFields = append(user.ForceSendFields, f)
		}
	}

	if lastName != "" {
		name.FamilyName = lastName
	}

	user.HashFunction = cmn.HashFunction

	if noGAL {
		user.IncludeInGlobalAddressList = false
		user.ForceSendFields = append(user.ForceSendFields, "IncludeInGlobalAddressList")

	}

	if orgUnit != "" {
		user.OrgUnitPath = orgUnit
	}

	if password != "" {
		pwd, err := cmn.HashPassword(password)
		if err != nil {
			return err
		}

		user.Password = pwd
		bPasswordCreated = true
	}

	if recoveryEmail != "" {
		user.RecoveryEmail = recoveryEmail
	}

	if recoveryPhone != "" {
		if string(recoveryPhone[0]) != "+" {
			err := fmt.Errorf("gmin: error - recovery phone number %v must start with '+'", recoveryPhone)
			return err
		}
		user.RecoveryPhone = recoveryPhone
	}

	if suspended {
		user.Suspended = true
	}

	user.Name = name

	if attrs != "" {
		attrUser := new(admin.User)
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

		if !bPasswordCreated {
			if attrUser.Password != "" {
				pwd, err := cmn.HashPassword(attrUser.Password)
				if err != nil {
					return err
				}

				attrUser.Password = pwd
			}
		}

		err = mergo.Merge(user, attrUser)
		if err != nil {
			return err
		}
	}

	if user.Name.GivenName == "" || user.Name.FamilyName == "" || user.Password == "" {
		err := errors.New("gmin: error - firstname, lastname and password must all be provided")
		return err
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

	fmt.Println("**** user " + newUser.PrimaryEmail + " created ****")

	return nil
}

func init() {
	createCmd.AddCommand(createUserCmd)

	createUserCmd.Flags().StringVarP(&attrs, "attributes", "a", "", "user's attributes as a JSON string")
	createUserCmd.Flags().BoolVarP(&changePassword, "changepassword", "c", false, "user must change password on next login")
	createUserCmd.Flags().StringVarP(&firstName, "firstname", "f", "", "user's first name")
	createUserCmd.Flags().StringVarP(&forceSend, "force", "", "", "force send fields")
	createUserCmd.Flags().StringVarP(&lastName, "lastname", "l", "", "user's last name")
	createUserCmd.Flags().BoolVarP(&noGAL, "nogal", "n", false, "do not display user in Global Address List")
	createUserCmd.Flags().StringVarP(&orgUnit, "orgunit", "o", "", "user's orgunit")
	createUserCmd.Flags().StringVarP(&password, "password", "p", "", "user's password")
	createUserCmd.Flags().StringVarP(&recoveryEmail, "recoveryemail", "z", "", "user's recovery email address")
	createUserCmd.Flags().StringVarP(&recoveryPhone, "recoveryphone", "k", "", "user's recovery phone")
	createUserCmd.Flags().BoolVarP(&suspended, "suspended", "s", false, "user is suspended")
}
