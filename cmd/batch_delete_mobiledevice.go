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
	"strings"
	"sync"
	"time"

	"github.com/cenkalti/backoff/v4"
	btch "github.com/plusworx/gmin/utils/batch"
	cmn "github.com/plusworx/gmin/utils/common"
	cfg "github.com/plusworx/gmin/utils/config"
	flgnm "github.com/plusworx/gmin/utils/flagnames"
	gmess "github.com/plusworx/gmin/utils/gminmessages"
	lg "github.com/plusworx/gmin/utils/logging"
	mdevs "github.com/plusworx/gmin/utils/mobiledevices"
	"github.com/spf13/cobra"
	admin "google.golang.org/api/admin/directory/v1"
)

var batchDelMobDevCmd = &cobra.Command{
	Use:     "mobile-devices [-i input file path]",
	Aliases: []string{"mobile-device", "mob-devices", "mob-device", "mob-devs", "mob-dev", "mdevs", "mdev"},
	Example: `gmin batch-delete mobile-devices -i inputfile.txt
	gmin bdel mdevs -i inputfile.txt
	gmin ls mdevs -q user:William* -a resourceId | jq '.mobiledevices[] | .resourceId' -r | gmin bdel mdevs`,
	Short: "Deletes a batch of mobile devices",
	Long: `Deletes a batch of mobile devices where mobile device details are provided in a text input file or through a pipe.
			
The input file or piped in data should provide the mobile device resource ids to be deleted on separate lines like this:

4cx07eba348f09b3Yjklj93xjsol0kE30lkl
Hkj98764yKK4jw8yyoyq9987js07q1hs7y98
lkalkju9027ja98na65wqHaTBOOUgarTQKk9

An input Google sheet must have a header row with the following column names being the only ones that are valid:

resourceId [required]

The column name is case insensitive.`,
	RunE: doBatchDelMobDev,
}

func doBatchDelMobDev(cmd *cobra.Command, args []string) error {
	lg.Debugw("starting doBatchDelMobDev()",
		"args", args)
	defer lg.Debug("finished doBatchDelMobDev()")

	var mobdevs []string

	srv, err := cmn.CreateService(cmn.SRVTYPEADMIN, admin.AdminDirectoryDeviceMobileScope)
	if err != nil {
		return err
	}
	ds := srv.(*admin.Service)

	inputFlgVal, err := cmd.Flags().GetString(flgnm.FLG_INPUTFILE)
	if err != nil {
		lg.Error(err)
		return err
	}

	scanner, err := cmn.InputFromStdIn(inputFlgVal)
	if err != nil {
		return err
	}

	if inputFlgVal == "" && scanner == nil {
		err := errors.New(gmess.ERR_NOINPUTFILE)
		lg.Error(err)
		return err
	}

	formatFlgVal, err := cmd.Flags().GetString(flgnm.FLG_FORMAT)
	if err != nil {
		lg.Error(err)
		return err
	}
	lwrFmt := strings.ToLower(formatFlgVal)

	ok := cmn.SliceContainsStr(cmn.ValidFileFormats, lwrFmt)
	if !ok {
		err = fmt.Errorf(gmess.ERR_INVALIDFILEFORMAT, formatFlgVal)
		lg.Error(err)
		return err
	}

	switch {
	case lwrFmt == "text":
		mobdevs, err = btch.DeleteProcessTextFile(inputFlgVal, scanner)
		if err != nil {
			return err
		}
	case lwrFmt == "gsheet":
		rangeFlgVal, err := cmd.Flags().GetString(flgnm.FLG_SHEETRANGE)
		if err != nil {
			return err
		}

		mobdevs, err = btch.DeleteProcessGSheet(inputFlgVal, rangeFlgVal, mdevs.MobDevAttrMap, mdevs.KEYNAME)
		if err != nil {
			return err
		}
	default:
		err = fmt.Errorf(gmess.ERR_INVALIDFILEFORMAT, formatFlgVal)
		lg.Error(err)
		return err
	}

	err = bdmdProcessDeletion(ds, mobdevs)
	if err != nil {
		return err
	}

	return nil
}

func bdmdDelete(wg *sync.WaitGroup, mdc *admin.MobiledevicesDeleteCall, resourceID string) {
	lg.Debugw("starting bdmdDelete()",
		"resourceID", resourceID)
	defer lg.Debug("finished bdmdDelete()")

	defer wg.Done()

	b := backoff.NewExponentialBackOff()
	b.MaxElapsedTime = 32 * time.Second

	err := backoff.Retry(func() error {
		var err error
		err = mdc.Do()
		if err == nil {
			fmt.Println(cmn.GminMessage(fmt.Sprintf(gmess.INFO_MDEVDELETED, resourceID)))
			lg.Infof(gmess.INFO_MDEVDELETED, resourceID)
			return err
		}
		if !cmn.IsErrRetryable(err) {
			return backoff.Permanent(fmt.Errorf(gmess.ERR_BATCHMOBILEDEVICE, err.Error(), resourceID))
		}
		// Log the retries
		lg.Warnw(err.Error(),
			"retrying", b.GetElapsedTime().String(),
			"mobile device", resourceID)
		return fmt.Errorf(gmess.ERR_BATCHMOBILEDEVICE, err.Error(), resourceID)
	}, b)
	if err != nil {
		// Log final error
		lg.Error(err)
		fmt.Println(cmn.GminMessage(err.Error()))
	}
}

func bdmdProcessDeletion(ds *admin.Service, mobdevs []string) error {
	lg.Debug("starting bdmdProcessDeletion()")
	defer lg.Debug("finished bdmdProcessDeletion()")

	customerID, err := cfg.ReadConfigString(cfg.CONFIGCUSTID)
	if err != nil {
		lg.Error(err)
		return err
	}

	wg := new(sync.WaitGroup)

	for _, mobResID := range mobdevs {
		mdc := ds.Mobiledevices.Delete(customerID, mobResID)

		wg.Add(1)

		go bdmdDelete(wg, mdc, mobResID)
	}

	wg.Wait()

	return nil
}

func init() {
	batchDelCmd.AddCommand(batchDelMobDevCmd)

	batchDelMobDevCmd.Flags().StringVarP(&inputFile, flgnm.FLG_INPUTFILE, "i", "", "filepath to mobile device data text file")
	batchDelMobDevCmd.Flags().StringVarP(&delFormat, flgnm.FLG_FORMAT, "f", "text", "mobile device data file format (text or gsheet)")
	batchDelMobDevCmd.Flags().StringVarP(&sheetRange, flgnm.FLG_SHEETRANGE, "s", "", "mobile device data gsheet range")
}
