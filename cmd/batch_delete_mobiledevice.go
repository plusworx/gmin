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

var batchDelMobDevCmd = &cobra.Command{
	Use:     "mobiledevices [-i input file path]",
	Aliases: []string{"mobiledevice", "mobdevices", "mobdevice", "mobdevs", "mobdev", "mdevs", "mdev"},
	Short:   "Deletes a batch of mobile devices",
	Long: `Deletes a batch of mobile devices where mobile device details are provided in a text input file or through a pipe.
	
	Examples:	gmin batch-delete mobiledevices -i inputfile.txt
			gmin bdel mdevs -i inputfile.txt
			gmin ls mdevs -q user:William* -a resourceId | jq '.mobiledevices[] | .resourceId' -r | gmin bdel mdevs
			
The input should have the mobile device resource ids to be deleted on separate lines like this:

4cx07eba348f09b3Yjklj93xjsol0kE30lkl
Hkj98764yKK4jw8yyoyq9987js07q1hs7y98
lkalkju9027ja98na65wqHaTBOOUgarTQKk9`,
	RunE: doBatchDelMobDev,
}

func doBatchDelMobDev(cmd *cobra.Command, args []string) error {
	ds, err := cmn.CreateDirectoryService(admin.AdminDirectoryDeviceMobileScope)
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
		mobResID := scanner.Text()
		mdc := ds.Mobiledevices.Delete(customerID, mobResID)

		wg.Add(1)

		go deleteMobDev(wg, mdc, mobResID)
	}

	wg.Wait()

	return nil
}

func deleteMobDev(wg *sync.WaitGroup, mdc *admin.MobiledevicesDeleteCall, resourceID string) {
	defer wg.Done()

	b := backoff.NewExponentialBackOff()
	b.MaxElapsedTime = 32 * time.Second

	err := backoff.Retry(func() error {
		var err error
		err = mdc.Do()
		if err == nil {
			logger.Infof(cmn.InfoMDevDeleted, resourceID)
			fmt.Println(cmn.GminMessage(fmt.Sprintf(cmn.InfoMDevDeleted, resourceID)))
			return err
		}
		if !cmn.IsErrRetryable(err) {
			return backoff.Permanent(fmt.Errorf(cmn.ErrBatchMobileDevice, err.Error(), resourceID))
		}
		// Log the retries
		logger.Errorw(err.Error(),
			"retrying", b.Clock.Now().String(),
			"mobile device", resourceID)
		return fmt.Errorf(cmn.ErrBatchMobileDevice, err.Error(), resourceID)
	}, b)
	if err != nil {
		// Log final error
		logger.Error(err)
		fmt.Println(cmn.GminMessage(err.Error()))
	}
}

func init() {
	batchDelCmd.AddCommand(batchDelMobDevCmd)

	batchDelMobDevCmd.Flags().StringVarP(&inputFile, "inputfile", "i", "", "filepath to mobile device data text file")
}
