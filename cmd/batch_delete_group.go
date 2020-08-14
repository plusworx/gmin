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

var batchDelGroupCmd = &cobra.Command{
	Use:     "groups [-i input file path]",
	Aliases: []string{"group", "grps", "grp"},
	Short:   "Deletes a batch of groups",
	Long: `Deletes a batch of groups where group details are provided in a text input file or through a pipe.
	
	Examples:	gmin batch-delete groups -i inputfile.txt
			gmin bdel grps -i inputfile.txt
			gmin ls grp -q name:Test1* -a email | jq '.groups[] | .email' -r | gmin bdel grp
			
The input should have the group email addresses, aliases or ids to be deleted on separate lines like this:

oldsales@company.com
oldaccounts@company.com
unused_group@company.com`,
	RunE: doBatchDelGroup,
}

func doBatchDelGroup(cmd *cobra.Command, args []string) error {
	ds, err := cmn.CreateDirectoryService(admin.AdminDirectoryGroupScope)
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
		group := scanner.Text()

		b := backoff.NewExponentialBackOff()
		b.MaxElapsedTime = 30 * time.Second

		err = backoff.Retry(func() error {
			var err error
			err = deleteGroup(ds, group)
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

func deleteGroup(ds *admin.Service, group string) error {
	gdc := ds.Groups.Delete(group)

	err := gdc.Do()
	if err != nil {
		return err
	}

	fmt.Printf("**** gmin: group %s deleted ****\n", group)

	return nil
}

func init() {
	batchDelCmd.AddCommand(batchDelGroupCmd)

	batchDelGroupCmd.Flags().StringVarP(&inputFile, "inputfile", "i", "", "filepath to group data text file")
}
