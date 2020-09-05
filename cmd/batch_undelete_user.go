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
	"sync"
	"time"

	"github.com/cenkalti/backoff/v4"
	cmn "github.com/plusworx/gmin/utils/common"
	"github.com/spf13/cobra"
	admin "google.golang.org/api/admin/directory/v1"
)

var batchUndelUserCmd = &cobra.Command{
	Use:     "users -i <input file path>",
	Aliases: []string{"user"},
	Short:   "Undeletes a batch of users",
	Long: `Undeletes a batch of users where user details are provided in a text input file.
	
	Examples:	gmin batch-undelete users -i inputfile.txt
			gmin bund user -i inputfile.txt -o /Engineering
			
	The input file should contain a list of user ids like this:
	
	417578192529765228417
	308127142904731923463
	107967172367714327529`,
	RunE: doBatchUndelUser,
}

func doBatchUndelUser(cmd *cobra.Command, args []string) error {
	var userUndelete *admin.UserUndelete
	userUndelete = new(admin.UserUndelete)

	if orgUnit == "" {
		userUndelete.OrgUnitPath = "/"
	} else {
		userUndelete.OrgUnitPath = orgUnit
	}

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

	wg := new(sync.WaitGroup)

	for scanner.Scan() {
		user := scanner.Text()

		uuc := ds.Users.Undelete(user, userUndelete)

		wg.Add(1)

		go undeleteUser(wg, uuc, user)
	}

	wg.Wait()

	return nil
}

func undeleteUser(wg *sync.WaitGroup, uuc *admin.UsersUndeleteCall, user string) {
	defer wg.Done()

	b := backoff.NewExponentialBackOff()
	b.MaxElapsedTime = 32 * time.Second

	err := backoff.Retry(func() error {
		var err error

		err = uuc.Do()
		if err == nil {
			fmt.Println(cmn.GminMessage("**** gmin: user " + user + " undeleted ****"))
			return err
		}
		if !cmn.IsErrRetryable(err) {
			return backoff.Permanent(errors.New(cmn.GminMessage("gmin: error - " + err.Error() + " " + user)))
		}
		return errors.New(cmn.GminMessage("gmin: error - " + err.Error() + " " + user))
	}, b)
	if err != nil {
		fmt.Println(err)
	}
}

func init() {
	batchUndeleteCmd.AddCommand(batchUndelUserCmd)

	batchUndelUserCmd.Flags().StringVarP(&inputFile, "inputfile", "i", "", "filepath to user data text file")
	batchUndelUserCmd.Flags().StringVarP(&orgUnit, "orgunit", "o", "", "path of orgunit to restore user to")

}
