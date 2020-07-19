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
	"errors"
	"fmt"

	cmn "github.com/plusworx/gmin/utils/common"
	"github.com/spf13/cobra"
	admin "google.golang.org/api/admin/directory/v1"
)

var deleteUserAliasCmd = &cobra.Command{
	Use:     "user-alias <alias email address> -u <user email address or id>",
	Aliases: []string{"ualias", "ua"},
	Args:    cobra.ExactArgs(1),
	Short:   "Deletes user alias",
	Long:    `Deletes user alias.`,
	RunE:    doDeleteUserAlias,
}

func doDeleteUserAlias(cmd *cobra.Command, args []string) error {
	if userKey == "" {
		err := errors.New("gmin: error - user email address or id must be provided")
		return err
	}

	ds, err := cmn.CreateDirectoryService(admin.AdminDirectoryUserAliasScope)
	if err != nil {
		return err
	}

	uadc := ds.Users.Aliases.Delete(userKey, args[0])

	err = uadc.Do()
	if err != nil {
		return err
	}

	fmt.Printf("**** gmin: user alias %s for user %s deleted ****\n", args[0], userKey)

	return nil
}

func init() {
	deleteCmd.AddCommand(deleteUserAliasCmd)

	deleteUserAliasCmd.Flags().StringVarP(&userKey, "user", "u", "", "email address or id of user")
}