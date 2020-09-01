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

var batchDelMemberCmd = &cobra.Command{
	Use:     "group-members <group email address or id> [-i input file path]",
	Aliases: []string{"group-member", "grp-members", "grp-member", "gmembers", "gmember", "gmems", "gmem"},
	Args:    cobra.ExactArgs(1),
	Short:   "Deletes a batch of group members",
	Long: `Deletes a batch of group members where group member details are provided in a text input file or through a pipe.
	
	Examples:	gmin batch-delete group-members somegroup@mycompany.com -i inputfile.txt
			gmin bdel gmems somegroup@mycompany.com -i inputfile.txt
			gmin ls gmem mygroup@mycompany.co.uk -a email | jq '.members[] | .email' -r | ./gmin bdel gmem mygroup@mycompany.co.uk
			
	The input should have the group member email addresses, aliases or ids to be deleted on separate lines like this:
	
	frank.castle@mycompany.com
	bruce.wayne@mycompany.com
	peter.parker@mycompany.com`,
	RunE: doBatchDelMember,
}

func doBatchDelMember(cmd *cobra.Command, args []string) error {
	ds, err := cmn.CreateDirectoryService(admin.AdminDirectoryGroupMemberScope)
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

	group := args[0]
	wg := new(sync.WaitGroup)

	for scanner.Scan() {
		member := scanner.Text()
		mdc := ds.Members.Delete(group, member)

		wg.Add(1)

		go deleteMember(wg, mdc, member, group)
	}

	wg.Wait()

	return nil
}

func deleteMember(wg *sync.WaitGroup, mdc *admin.MembersDeleteCall, member string, group string) {
	defer wg.Done()

	b := backoff.NewExponentialBackOff()
	b.MaxElapsedTime = 32 * time.Second

	err := backoff.Retry(func() error {
		var err error
		err = mdc.Do()
		if err == nil {
			fmt.Println(cmn.GminMessage("**** gmin: member " + member + " of group " + group + " deleted ****"))
			return err
		}
		if !cmn.IsErrRetryable(err) {
			return backoff.Permanent(errors.New(cmn.GminMessage("gmin: error - " + err.Error() + " " + member)))
		}
		return errors.New(cmn.GminMessage("gmin: error - " + err.Error() + " " + member))
	}, b)
	if err != nil {
		fmt.Println(err)
	}
}

func init() {
	batchDelCmd.AddCommand(batchDelMemberCmd)

	batchDelMemberCmd.Flags().StringVarP(&inputFile, "inputfile", "i", "", "filepath to member data text file")
}
