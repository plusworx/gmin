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
	cfg "github.com/plusworx/gmin/utils/config"
	"github.com/spf13/cobra"
	admin "google.golang.org/api/admin/directory/v1"
)

var batchDelOrgUnitCmd = &cobra.Command{
	Use:     "orgunits [-i input file path]",
	Aliases: []string{"orgunit", "ous", "ou"},
	Short:   "Deletes a batch of orgunits",
	Long: `Deletes a batch of orgunits where orgunit details are provided in a text input file or through a pipe.
	
	Examples:	gmin batch-delete orgunits -i inputfile.txt
			gmin bdel ous -i inputfile.txt
			gmin ls ous -o TestOU -a orgunitpath | jq '.organizationUnits[] | .orgUnitPath' -r | gmin bdel ou
			
	The input should have the orgunit paths or ids to be deleted on separate lines like this:
	
	Engineering/Skunkworx
	Engineering/SecretOps
	Engineering/Surplus`,
	RunE: doBatchDelOrgUnit,
}

func doBatchDelOrgUnit(cmd *cobra.Command, args []string) error {
	ds, err := cmn.CreateDirectoryService(admin.AdminDirectoryOrgunitScope)
	if err != nil {
		logger.Error(err)
		return err
	}

	customerID, err := cfg.ReadConfigString("customerid")
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
		text := scanner.Text()

		if text[0] == '/' {
			text = text[1:]
		}

		oudc := ds.Orgunits.Delete(customerID, text)

		// Sleep for 2 seconds because only 1 orgunit can be created per second but 1 second interval
		// still results in rate limit errors
		time.Sleep(2 * time.Second)

		wg.Add(1)

		go deleteOU(wg, oudc, text)
	}

	wg.Wait()

	return nil
}

func deleteOU(wg *sync.WaitGroup, oudc *admin.OrgunitsDeleteCall, ouPath string) {
	defer wg.Done()

	b := backoff.NewExponentialBackOff()
	b.MaxElapsedTime = 32 * time.Second

	err := backoff.Retry(func() error {
		var err error

		err = oudc.Do()
		if err == nil {
			logger.Infof(cmn.InfoOUDeleted, ouPath)
			fmt.Println(cmn.GminMessage(fmt.Sprintf(cmn.InfoOUDeleted, ouPath)))
			return err
		}
		if !cmn.IsErrRetryable(err) {
			return backoff.Permanent(errors.New(cmn.GminMessage(fmt.Sprintf(cmn.ErrBatchOU, err.Error(), ouPath))))
		}
		// Log the retries
		logger.Errorw(err.Error(),
			"retrying", b.Clock.Now().String(),
			"orgunit", ouPath)
		return errors.New(cmn.GminMessage(fmt.Sprintf(cmn.ErrBatchOU, err.Error(), ouPath)))
	}, b)
	if err != nil {
		// Log final error
		logger.Error(err)
		fmt.Println(err)
	}
}

func init() {
	batchDelCmd.AddCommand(batchDelOrgUnitCmd)

	batchDelOrgUnitCmd.Flags().StringVarP(&inputFile, "inputfile", "i", "", "filepath to orgunit data text file")
}
