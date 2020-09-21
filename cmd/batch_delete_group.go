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
		logger.Error(err)
		return err
	}

	scanner, err := cmn.InputFromStdIn(inputFile)
	if err != nil {
		logger.Error(err)
		return err
	}

	if inputFile == "" && scanner == nil {
		err := errors.New(cmn.ErrNoInputFile)
		logger.Error(err)
		return err
	}

	if scanner == nil {
		file, err := os.Open(inputFile)
		if err != nil {
			logger.Error(err)
			return err
		}
		defer file.Close()

		scanner = bufio.NewScanner(file)
	}

	wg := new(sync.WaitGroup)

	for scanner.Scan() {
		group := scanner.Text()
		gdc := ds.Groups.Delete(group)

		wg.Add(1)

		go deleteGroup(wg, gdc, group)
	}

	wg.Wait()

	return nil
}

func deleteGroup(wg *sync.WaitGroup, gdc *admin.GroupsDeleteCall, group string) {
	defer wg.Done()

	b := backoff.NewExponentialBackOff()
	b.MaxElapsedTime = 32 * time.Second

	err := backoff.Retry(func() error {
		var err error

		err = gdc.Do()
		if err == nil {
			logger.Infof(cmn.InfoGroupDeleted, group)
			fmt.Println(cmn.GminMessage(fmt.Sprintf(cmn.InfoGroupDeleted, group)))
			return err
		}
		if !cmn.IsErrRetryable(err) {
			return backoff.Permanent(errors.New(cmn.GminMessage(fmt.Sprintf(cmn.ErrBatchGroup, err.Error(), group))))
		}
		// Log the retries
		logger.Errorw(err.Error(),
			"retrying", b.Clock.Now().String(),
			"group", group)
		return errors.New(cmn.GminMessage(fmt.Sprintf(cmn.ErrBatchGroup, err.Error(), group)))
	}, b)
	if err != nil {
		// Log final error
		logger.Error(err)
		fmt.Println(err)
	}
}

func init() {
	batchDelCmd.AddCommand(batchDelGroupCmd)

	batchDelGroupCmd.Flags().StringVarP(&inputFile, "inputfile", "i", "", "filepath to group data text file")
}
