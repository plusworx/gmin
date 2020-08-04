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
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/cenkalti/backoff/v4"
	cmn "github.com/plusworx/gmin/utils/common"
	usrs "github.com/plusworx/gmin/utils/users"
	"github.com/spf13/cobra"
	admin "google.golang.org/api/admin/directory/v1"
)

var batchUpdUserCmd = &cobra.Command{
	Use:     "users -i <input file path>",
	Aliases: []string{"user"},
	Short:   "Updates a batch of users",
	Long: `Updates a batch of users where user details are provided in a JSON input file.
	
	Examples:	gmin batch-update users -i inputfile.json
			gmin bupd user -i inputfile.json
			  
	The contents of the JSON file should look something like this:
	
	{"userKey":"stan.laurel@myorg.org","name":{"givenName":"Stanislav","familyName":"Laurelius"},"primaryEmail":"stanislav.laurelius@myorg.org","password":"SuperSuperSecretPassword","changePasswordAtNextLogin":true}
	{"userKey":"oliver.hardy@myorg.org","name":{"givenName":"Oliviatus","familyName":"Hardium"},"primaryEmail":"oliviatus.hardium@myorg.org","password":"StealthySuperSecretPassword","changePasswordAtNextLogin":true}
	{"userKey":"harold.lloyd@myorg.org","name":{"givenName":"Haroldus","familyName":"Lloydius"},"primaryEmail":"haroldus.lloydius@myorg.org","password":"MightySuperSecretPassword","changePasswordAtNextLogin":true}
	
	N.B. userKey (user email address, alias or id) must be provided.`,
	RunE: doBatchUpdUser,
}

func doBatchUpdUser(cmd *cobra.Command, args []string) error {
	ds, err := cmn.CreateDirectoryService(admin.AdminDirectoryUserScope)
	if err != nil {
		return err
	}

	if inputFile == "" {
		err := errors.New("gmin: error - must provide inputfile")
		return err
	}

	file, err := os.Open(inputFile)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		jsonData := scanner.Text()

		b := backoff.NewExponentialBackOff()
		b.MaxElapsedTime = 30 * time.Second

		err = backoff.Retry(func() error {
			var err error
			err = updateUser(ds, jsonData)
			if err == nil {
				return err
			}

			if strings.Contains(err.Error(), "Missing required field") ||
				strings.Contains(err.Error(), "not valid") ||
				strings.Contains(err.Error(), "unrecognized") ||
				strings.Contains(err.Error(), "should be") ||
				strings.Contains(err.Error(), "must be included") {
				return backoff.Permanent(err)
			}

			return err
		}, b)
		if err != nil {
			return err
		}
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	return nil
}

func updateUser(ds *admin.Service, jsonData string) error {
	var (
		user   *admin.User
		usrKey = usrs.Key{}
	)

	user = new(admin.User)
	jsonBytes := []byte(jsonData)

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

	err = json.Unmarshal(jsonBytes, &usrKey)
	if err != nil {
		return err
	}

	if usrKey.UserKey == "" {
		return errors.New("gmin: error - userKey must be included in the JSON input string")
	}

	err = json.Unmarshal(jsonBytes, &user)
	if err != nil {
		return err
	}

	if user.Password != "" {
		user.HashFunction = cmn.HashFunction
		pwd, err := cmn.HashPassword(user.Password)
		if err != nil {
			return err
		}
		user.Password = pwd
	}

	uic := ds.Users.Update(usrKey.UserKey, user)
	_, err = uic.Do()
	if err != nil {
		return err
	}

	fmt.Println("**** gmin: user " + usrKey.UserKey + " updated ****")

	return nil
}

func init() {
	batchUpdateCmd.AddCommand(batchUpdUserCmd)

	batchUpdUserCmd.Flags().StringVarP(&inputFile, "inputfile", "i", "", "filepath to user data file")
}
