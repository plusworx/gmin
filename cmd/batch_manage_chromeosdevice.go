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
	cdevs "github.com/plusworx/gmin/utils/chromeosdevices"
	cmn "github.com/plusworx/gmin/utils/common"
	cfg "github.com/plusworx/gmin/utils/config"
	flgnm "github.com/plusworx/gmin/utils/flagnames"
	gmess "github.com/plusworx/gmin/utils/gminmessages"
	lg "github.com/plusworx/gmin/utils/logging"
	"github.com/spf13/cobra"
	admin "google.golang.org/api/admin/directory/v1"
)

var batchMngCrOSDevCmd = &cobra.Command{
	Use:     "chromeos-devices -i <input file>",
	Aliases: []string{"chromeos-device", "cros-devices", "cros-device", "cros-devs", "cros-dev", "cdevs", "cdev"},
	Example: `gmin batch-manage chromeos-devices -i inputfile.json
gmin bmng cdevs -i inputfile.csv -f csv
gmin bmng cdev -i 1odyAIp3jGspd3M4xeepxWD6aeQIUuHBgrZB2OHSu8MI -s 'Sheet1!A1:C25' -f gsheet`,
	Short: "Manages a batch of ChromeOS devices",
	Long: `Manages a batch of ChromeOS devices where device details are provided in a Google Sheet, CSV/JSON input file or piped JSON.
			  
The JSON file or piped input should contain ChromeOS device management details like this:

{"deviceId":"5ac7be73-5996-394e-9c30-62d41a8f10e8","action":"disable"}
{"deviceId":"6ac9bd33-7095-453e-6c39-22d48a8f13e8","action":"reenable"}
{"deviceId":"6bc4be13-9916-494e-9c39-62d45c8f40e9","action":"deprovision","deprovisionReason":"retiring_device"}

N.B. If you are deprovisioning devices then a deprovision reason must be provided.

CSV and Google sheets must have a header row with the following column names being the only ones that are valid:

action [required]
deprovisionReason
deviceId [required]

The column names are case insensitive and can be in any order.

Valid actions are:
deprovision
disable
reenable

Valid deprovision reasons are:
different_model_replacement
retiring_device
same_model_replacement
upgrade_transfer`,
	RunE: doBatchMngCrOSDev,
}

func doBatchMngCrOSDev(cmd *cobra.Command, args []string) error {
	lg.Debugw("starting doBatchMngCrOSDev()",
		"args", args)
	defer lg.Debug("finished doBatchMngCrOSDev()")

	var (
		managedDevs []cdevs.ManagedDevice
		objs        []interface{}
	)

	srv, err := cmn.CreateService(cmn.SRVTYPEADMIN, admin.AdminDirectoryDeviceChromeosScope)
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

	callParams := btch.CallParams{CallType: cmn.CALLTYPEMANAGE, ObjectType: cmn.OBJTYPECROSDEV}

	switch {
	case lwrFmt == "csv":
		objs, err = btch.ProcessCSVFile(callParams, inputFlgVal, cdevs.CrOSDevAttrMap)
		if err != nil {
			return err
		}
	case lwrFmt == "json":
		objs, err = btch.ProcessJSON(callParams, inputFlgVal, scanner, cdevs.CrOSDevAttrMap)
		if err != nil {
			return err
		}
	case lwrFmt == "gsheet":
		rangeFlgVal, err := cmd.Flags().GetString(flgnm.FLG_SHEETRANGE)
		if err != nil {
			return err
		}

		objs, err = btch.ProcessGSheet(callParams, inputFlgVal, rangeFlgVal, cdevs.CrOSDevAttrMap)
		if err != nil {
			return err
		}
	default:
		err = fmt.Errorf(gmess.ERR_INVALIDFILEFORMAT, formatFlgVal)
		lg.Error(err)
		return err
	}

	for _, cdevObj := range objs {
		managedDevs = append(managedDevs, cdevObj.(cdevs.ManagedDevice))
	}

	err = bmngcProcessObjects(ds, managedDevs)
	if err != nil {
		return err
	}

	return nil
}

func bmngcPerformAction(deviceID string, action string, wg *sync.WaitGroup, cdac *admin.ChromeosdevicesActionCall) {
	lg.Debugw("starting bmngcPerformAction()",
		"action", action,
		"deviceID", deviceID)
	defer lg.Debug("finished bmngcPerformAction()")

	defer wg.Done()

	b := backoff.NewExponentialBackOff()
	b.MaxElapsedTime = 32 * time.Second

	err := backoff.Retry(func() error {
		var err error
		err = cdac.Do()
		if err == nil {
			fmt.Println(cmn.GminMessage(fmt.Sprintf(gmess.INFO_CDEVACTIONPERFORMED, action, deviceID)))
			lg.Infof(gmess.INFO_CDEVACTIONPERFORMED, action, deviceID)
			return err
		}
		if !cmn.IsErrRetryable(err) {
			return backoff.Permanent(fmt.Errorf(gmess.ERR_BATCHCHROMEOSDEVICE, err.Error(), deviceID))
		}
		// Log the retries
		lg.Warnw(err.Error(),
			"retrying", b.GetElapsedTime().String(),
			"ChromeOS device", deviceID)
		return fmt.Errorf(gmess.ERR_BATCHCHROMEOSDEVICE, err.Error(), deviceID)
	}, b)
	if err != nil {
		// Log final error
		lg.Error(err)
		fmt.Println(cmn.GminMessage(err.Error()))
	}
}

func bmngcProcessObjects(ds *admin.Service, managedDevs []cdevs.ManagedDevice) error {
	lg.Debug("starting bmngcProcessObjects()")
	defer lg.Debug("finished bmngcProcessObjects()")

	customerID, err := cfg.ReadConfigString(cfg.CONFIGCUSTID)
	if err != nil {
		return err
	}

	wg := new(sync.WaitGroup)

	for _, md := range managedDevs {
		devAction := admin.ChromeOsDeviceAction{}

		devAction.Action = md.Action
		devAction.DeprovisionReason = md.DeprovisionReason

		cdac := ds.Chromeosdevices.Action(customerID, md.DeviceId, &devAction)

		wg.Add(1)

		go bmngcPerformAction(md.DeviceId, md.Action, wg, cdac)
	}

	wg.Wait()

	return nil
}

func init() {
	batchManageCmd.AddCommand(batchMngCrOSDevCmd)

	batchMngCrOSDevCmd.Flags().StringVarP(&inputFile, flgnm.FLG_INPUTFILE, "i", "", "filepath to device data file")
	batchMngCrOSDevCmd.Flags().StringVarP(&format, flgnm.FLG_FORMAT, "f", "json", "user data file format")
	batchMngCrOSDevCmd.Flags().StringVarP(&sheetRange, flgnm.FLG_SHEETRANGE, "s", "", "user data gsheet range")
}
