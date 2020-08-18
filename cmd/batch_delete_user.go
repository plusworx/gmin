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
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/cenkalti/backoff/v4"
	cmn "github.com/plusworx/gmin/utils/common"
	"github.com/spf13/cobra"
	admin "google.golang.org/api/admin/directory/v1"
)

var batchDelUserCmd = &cobra.Command{
	Use:     "users [-i input file path]",
	Aliases: []string{"user"},
	Short:   "Deletes a batch of users",
	Long: `Deletes a batch of users where user details are provided in a text input file or from a pipe.
	
	Examples:	gmin batch-delete users -i inputfile.txt
			gmin bdel user -i inputfile.txt
			gmin ls user -a primaryemail -q orgunitpath=/TestOU | jq '.users[] | .primaryEmail' -r | gmin bdel user
			
	The input should provide the user email addresses, aliases or ids to be deleted on separate lines like this:
	
	frank.castle@mycompany.com
	bruce.wayne@mycompany.com
	peter.parker@mycompany.com`,
	RunE: doBatchDelUser,
}

func doBatchDelUser(cmd *cobra.Command, args []string) error {
	ds, err := cmn.CreateDirectoryService(admin.AdminDirectoryUserScope)
	if err != nil {
		return err
	}

	scanner, err := cmn.InputFromStdIn(inputFile)
	if err != nil {
		return err
	}

	if inputFile == "" && scanner == nil {
		err := errors.New("gmin: error - must provide inputfile")
		return err
	}

	if scanner == nil {
		file, err := os.Open(inputFile)
		if err != nil {
			return err
		}
		defer file.Close()

		scanner = bufio.NewScanner(file)
	}

	for scanner.Scan() {
		user := scanner.Text()

		b := backoff.NewExponentialBackOff()
		b.MaxElapsedTime = 30 * time.Second

		err = backoff.Retry(func() error {
			var err error
			err = deleteUser(ds, user)
			if err == nil {
				return err
			}

			if strings.Contains(err.Error(), "Resource Not Found") ||
				strings.Contains(err.Error(), "Bad Request") {
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

func deleteUser(ds *admin.Service, user string) error {
	udc := ds.Users.Delete(user)

	err := udc.Do()
	if err != nil {
		return err
	}

	fmt.Printf("**** gmin: user %s deleted ****\n", user)

	return nil
}

func init() {
	batchDelCmd.AddCommand(batchDelUserCmd)

	batchDelUserCmd.Flags().StringVarP(&inputFile, "inputfile", "i", "", "filepath to user data text file")
}
