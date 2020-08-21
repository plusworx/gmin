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
	cfg "github.com/plusworx/gmin/utils/config"
	mdevs "github.com/plusworx/gmin/utils/mobiledevices"
	"github.com/spf13/cobra"
	admin "google.golang.org/api/admin/directory/v1"
)

var batchMngMobDevCmd = &cobra.Command{
	Use:     "mobiledevices <action> [-i input file]",
	Aliases: []string{"mobiledevice", "mobdevices", "mobdevice", "mobdevs", "mobdev", "mdevs", "mdev"},
	Args:    cobra.ExactArgs(1),
	Short:   "Manages a batch of mobile devices",
	Long: `Manages a batch of mobile devices where device details are provided in a text input file or 
	via a command line pipe.
	
	Examples:	gmin batch-manage mobiledevices admin_account_wipe -i inputfile.txt
			gmin bmng mdev approve -i inputfile.txt
			  
	The input file should contain a list of resource ids like this:
	
	4cx07eba348f09b3Yjklj93xjsol0kE30lkl
	Hkj98764yKK4jw8yyoyq9987js07q1hs7y98
	lkalkju9027ja98na65wqHaTBOOUgarTQKk9`,
	RunE: doBatchMngCrOSDev,
}

func doBatchMngMobDev(cmd *cobra.Command, args []string) error {
	ds, err := cmn.CreateDirectoryService(admin.AdminDirectoryDeviceMobileActionScope)
	if err != nil {
		return err
	}

	customerID, err := cfg.ReadConfigString("customerid")
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
		mobResID := scanner.Text()

		b := backoff.NewExponentialBackOff()
		b.MaxElapsedTime = 30 * time.Second

		err = backoff.Retry(func() error {
			var err error
			err = manageMobDev(ds, customerID, mobResID, args[0])
			if err == nil {
				return err
			}

			if strings.Contains(err.Error(), "Missing required field") ||
				strings.Contains(err.Error(), "not a valid") ||
				strings.Contains(err.Error(), "provide a reason") ||
				strings.Contains(err.Error(), "Resource Not Found") ||
				strings.Contains(err.Error(), "Illegal") {
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

func manageMobDev(ds *admin.Service, customerID string, resourceID string, action string) error {
	var devAction = admin.MobileDeviceAction{}

	lwrAction := strings.ToLower(action)
	ok := cmn.SliceContainsStr(mdevs.ValidActions, lwrAction)
	if !ok {
		return fmt.Errorf("gmin: error - %v is not a valid action type", action)
	}

	devAction.Action = lwrAction

	mdac := ds.Mobiledevices.Action(customerID, resourceID, &devAction)

	err := mdac.Do()
	if err != nil {
		return err
	}

	fmt.Println("**** gmin: " + action + " successfully performed on mobile device " + resourceID + " ****")

	return nil
}

func init() {
	batchManageCmd.AddCommand(batchMngMobDevCmd)

	batchMngMobDevCmd.Flags().StringVarP(&inputFile, "inputfile", "i", "", "filepath to device data file")
}
