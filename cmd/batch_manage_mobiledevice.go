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

var batchMngMobDevCmd = &cobra.Command{
	Use:     "mobile-devices -i <input file>",
	Aliases: []string{"mobile-device", "mob-devices", "mob-device", "mob-devs", "mob-dev", "mdevs", "mdev"},
	Example: `gmin batch-manage mobile-devices -i inputfile.json
gmin bmng mdevs -i inputfile.csv -f csv
gmin bmng mdev -i 1odyAIp3jGspd3M4xeepxWD6aeQIUuHBgrZB2OHSu8MI -s 'Sheet1!A1:B25' -f gsheet`,
	Short: "Manages a batch of mobile devices",
	Long: `Manages a batch of mobile devices where device details are provided in a Google Sheet, CSV/JSON input file or piped JSON.
			  
The JSON file or piped input should contain mobile device management details like this:

{"resourceId":"4cx07eba348f09b3Yjklj93xjsol0kE30lkl","action":"admin_account_wipe"}
{"resourceId":"Hkj98764yKK4jw8yyoyq9987js07q1hs7y98","action":"approve"}
{"resourceId":"lkalkju9027ja98na65wqHaTBOOUgarTQKk9","action":"admin_remote_wipe"}

N.B. resourceId must be used NOT deviceId.

CSV and Google sheets must have a header row with the following column names being the only ones that are valid:

action [required]
resourceId [required]

The column names are case insensitive and can be in any order.

Valid actions are:
admin_account_wipe
admin_remote_wipe
approve
block
cancel_remote_wipe_then_activate
cancel_remote_wipe_then_block`,
	RunE: doBatchMngMobDev,
}

func doBatchMngMobDev(cmd *cobra.Command, args []string) error {
	lg.Debugw("starting doBatchMngMobDev()",
		"args", args)
	defer lg.Debug("finished doBatchMngMobDev()")

	var (
		managedDevs []mdevs.ManagedDevice
		objs        []interface{}
	)

	srv, err := cmn.CreateService(cmn.SRVTYPEADMIN, admin.AdminDirectoryDeviceMobileActionScope)
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

	callParams := btch.CallParams{CallType: cmn.CALLTYPEMANAGE, ObjectType: cmn.OBJTYPEMOBDEV}

	switch {
	case lwrFmt == "csv":
		objs, err = btch.ProcessCSVFile(callParams, inputFlgVal, mdevs.MobDevAttrMap)
		if err != nil {
			return err
		}
	case lwrFmt == "json":
		objs, err = btch.ProcessJSON(callParams, inputFlgVal, scanner, mdevs.MobDevAttrMap)
		if err != nil {
			return err
		}
	case lwrFmt == "gsheet":
		rangeFlgVal, err := cmd.Flags().GetString(flgnm.FLG_SHEETRANGE)
		if err != nil {
			lg.Error(err)
			return err
		}

		objs, err = btch.ProcessGSheet(callParams, inputFlgVal, rangeFlgVal, mdevs.MobDevAttrMap)
		if err != nil {
			return err
		}
	default:
		err = fmt.Errorf(gmess.ERR_INVALIDFILEFORMAT, formatFlgVal)
		lg.Error(err)
		return err
	}

	for _, mdevObj := range objs {
		managedDevs = append(managedDevs, mdevObj.(mdevs.ManagedDevice))
	}

	err = bmngmProcessObjects(ds, managedDevs)
	if err != nil {
		return err
	}

	return nil
}

func bmngmPerformAction(resourceID string, action string, wg *sync.WaitGroup, mdac *admin.MobiledevicesActionCall) {
	lg.Debugw("starting bmngmPerformAction()",
		"action", action,
		"resourceID", resourceID)
	defer lg.Debug("finished bmngmPerformAction()")

	defer wg.Done()

	b := backoff.NewExponentialBackOff()
	b.MaxElapsedTime = 32 * time.Second

	err := backoff.Retry(func() error {
		var err error
		err = mdac.Do()
		if err == nil {
			fmt.Println(cmn.GminMessage(fmt.Sprintf(gmess.INFO_MDEVACTIONPERFORMED, action, resourceID)))
			lg.Infof(gmess.INFO_MDEVACTIONPERFORMED, action, resourceID)
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

func bmngmProcessObjects(ds *admin.Service, managedDevs []mdevs.ManagedDevice) error {
	lg.Debug("starting bmngmProcessObjects()")
	defer lg.Debug("finished bmngmProcessObjects()")

	customerID, err := cfg.ReadConfigString(cfg.CONFIGCUSTID)
	if err != nil {
		lg.Error(err)
		return err
	}

	wg := new(sync.WaitGroup)

	for _, md := range managedDevs {
		devAction := admin.MobileDeviceAction{}

		devAction.Action = md.Action

		mdac := ds.Mobiledevices.Action(customerID, md.ResourceId, &devAction)

		wg.Add(1)

		go bmngmPerformAction(md.ResourceId, md.Action, wg, mdac)
	}

	wg.Wait()

	return nil
}

func init() {
	batchManageCmd.AddCommand(batchMngMobDevCmd)

	batchMngMobDevCmd.Flags().StringVarP(&inputFile, flgnm.FLG_INPUTFILE, "i", "", "filepath to device data file")
	batchMngMobDevCmd.Flags().StringVarP(&format, flgnm.FLG_FORMAT, "f", "json", "user data file format")
	batchMngMobDevCmd.Flags().StringVarP(&sheetRange, flgnm.FLG_SHEETRANGE, "s", "", "user data gsheet range")
}
