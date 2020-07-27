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
	admin "google.golang.org/api/admin/directory/v1"
)

var updateUserCmd = &cobra.Command{
	Use:   "user <user email address, alias or id>",
	Args:  cobra.ExactArgs(1),
	Short: "Updates a user",
	Long: `Updates a user.
	
	Examples: gmin update user another.user@mycompany.com -p strongpassword -s
	          gmin upd user finance.person@mycompany.com -l Newlastname`,
	RunE: doUpdateUser,
}

func doUpdateUser(cmd *cobra.Command, args []string) error {
	var (
		name    *admin.UserName
		user    *admin.User
		userKey string
	)

	userKey = args[0]
	user = new(admin.User)
	name = new(admin.UserName)

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

	if name.FamilyName != "" || name.FullName != "" || name.GivenName != "" {
		user.Name = name
	}

	if password != "" {
		pwd, err := cmn.HashPassword(password)
		if err != nil {
			return err
		}

		user.Password = pwd
		user.HashFunction = cmn.HashFunction
	}

	if gal {
		user.IncludeInGlobalAddressList = true
	}

	if noGAL {
		user.IncludeInGlobalAddressList = false
		user.ForceSendFields = append(user.ForceSendFields, "IncludeInGlobalAddressList")
	}

	if orgUnit != "" {
		user.OrgUnitPath = orgUnit
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

	if unSuspended {
		user.Suspended = false
		user.ForceSendFields = append(user.ForceSendFields, "Suspended")
	}

	if userEmail != "" {
		user.PrimaryEmail = userEmail
	}

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

		err = mergo.Merge(user, attrUser)
		if err != nil {
			return err
		}
	}

	ds, err := cmn.CreateDirectoryService(admin.AdminDirectoryUserScope)
	if err != nil {
		return err
	}

	uuc := ds.Users.Update(userKey, user)
	_, err = uuc.Do()
	if err != nil {
		return err
	}

	fmt.Println("**** gmin: user " + userKey + " updated ****")

	return nil
}

func init() {
	updateCmd.AddCommand(updateUserCmd)

	updateUserCmd.Flags().StringVarP(&attrs, "attributes", "a", "", "user's attributes as a JSON string")
	updateUserCmd.Flags().BoolVarP(&changePassword, "changepassword", "c", false, "user must change password on next login")
	updateUserCmd.Flags().BoolVarP(&noChangePassword, "nochangepassword", "d", false, "user doesn't have to change password on next login")
	updateUserCmd.Flags().StringVarP(&userEmail, "email", "e", "", "user's primary email address")
	updateUserCmd.Flags().StringVarP(&firstName, "firstname", "f", "", "user's first name")
	updateUserCmd.Flags().StringVarP(&forceSend, "force", "", "", "field list for ForceSendFields separated by (~)")
	updateUserCmd.Flags().BoolVarP(&gal, "gal", "g", false, "display user in Global Address List")
	updateUserCmd.Flags().StringVarP(&lastName, "lastname", "l", "", "user's last name")
	updateUserCmd.Flags().BoolVarP(&noGAL, "nogal", "n", false, "do not display user in Global Address List")
	updateUserCmd.Flags().StringVarP(&orgUnit, "orgunit", "o", "", "user's orgunit")
	updateUserCmd.Flags().StringVarP(&password, "password", "p", "", "user's password")
	updateUserCmd.Flags().StringVarP(&recoveryEmail, "recoveryemail", "z", "", "user's recovery email address")
	updateUserCmd.Flags().StringVarP(&recoveryPhone, "recoveryphone", "k", "", "user's recovery phone")
	updateUserCmd.Flags().BoolVarP(&suspended, "suspended", "s", false, "user is suspended")
	updateUserCmd.Flags().BoolVarP(&unSuspended, "unsuspended", "u", false, "user is unsuspended")
}
