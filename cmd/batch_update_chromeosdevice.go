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

var batchUpdCrOSDevCmd = &cobra.Command{
	Use:     "chromeos-devices -i <input file>",
	Aliases: []string{"chromeos-device", "cros-devices", "cros-device", "cros-devs", "cros-dev", "cdevs", "cdev"},
	Example: `gmin batch-update chromeos-devices -i inputfile.json
gmin bupd cdevs -i inputfile.csv -f csv
gmin bupd cdev -i 1odyAIp3jGspd3M4xeepxWD6aeQIUuHBgrZB2OHSu8MI -s 'Sheet1!A1:K25' -f gsheet`,
	Short: "Updates a batch of ChromeOS devices",
	Long: `Updates a batch of ChromeOS devices with device details provided in a Google Sheet, CSV/JSON input file or piped JSON.
			
The JSON file or piped input should contain device update details like this:

{"deviceId":"5ac7be43-5906-394e-7c39-62d45a8f10e8","annotatedAssetId":"CB1","annotatedLocation":"Batcave","annotatedUser":"Bruce Wayne","notes":"Test machine","orgUnitPath":"/Anticrime"}
{"deviceId":"4ac7be43-5906-394e-7c39-62d45a8f10e8","annotatedAssetId":"CB2","annotatedLocation":"Wayne Manor","annotatedUser":"Alfred Pennyworth","notes":"Another test machine","orgUnitPath":"/Anticorruption"}
{"deviceId":"3ac7be43-5906-394e-7c39-62d45a8f10e8","annotatedAssetId":"CB3","annotatedLocation":"Wayne Towers","annotatedUser":"The Big Enchilada","notes":"Yet another test machine","orgUnitPath":"/Legal"}

N.B. deviceId must be provided.

CSV and Google sheets must have a header row with the following column names being the only ones that are valid:

annotatedAssetId
annotatedLocation
annotatedUser
deviceId [required]
notes
orgUnitPath

The column names are case insensitive and can be in any order.`,
	RunE: doBatchUpdCrOSDev,
}

func doBatchUpdCrOSDev(cmd *cobra.Command, args []string) error {
	lg.Debugw("starting doBatchUpdCrOSDev()",
		"args", args)
	defer lg.Debug("finished doBatchUpdCrOSDev()")

	var (
		crosdevs []*admin.ChromeOsDevice
		objs     []interface{}
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

	callParams := btch.CallParams{CallType: cmn.CALLTYPEUPDATE, ObjectType: cmn.OBJTYPECROSDEV}

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
			lg.Error(err)
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
		crosdevs = append(crosdevs, cdevObj.(*admin.ChromeOsDevice))
	}

	err = bucProcessObjects(ds, crosdevs)
	if err != nil {
		return err
	}

	return nil
}

func bucProcessObjects(ds *admin.Service, crosdevs []*admin.ChromeOsDevice) error {
	lg.Debug("starting bucProcessObjects()")
	wg := new(sync.WaitGroup)
	defer lg.Debug("finished bucProcessObjects()")

	customerID, err := cfg.ReadConfigString(cfg.CONFIGCUSTID)
	if err != nil {
		lg.Error(err)
		return err
	}

	for _, c := range crosdevs {
		cduc := ds.Chromeosdevices.Update(customerID, c.DeviceId, c)

		wg.Add(1)

		go bucUpdate(c, wg, cduc)
	}

	wg.Wait()

	return nil
}

func bucUpdate(crosdev *admin.ChromeOsDevice, wg *sync.WaitGroup, cduc *admin.ChromeosdevicesUpdateCall) {
	lg.Debug("starting bucUpdate()")
	defer lg.Debug("finished bucUpdate()")

	defer wg.Done()

	b := backoff.NewExponentialBackOff()
	b.MaxElapsedTime = 32 * time.Second

	err := backoff.Retry(func() error {
		var err error
		_, err = cduc.Do()
		if err == nil {
			fmt.Println(cmn.GminMessage(fmt.Sprintf(gmess.INFO_CDEVUPDATED, crosdev.DeviceId)))
			lg.Infof(gmess.INFO_CDEVUPDATED, crosdev.DeviceId)
			return err
		}
		if !cmn.IsErrRetryable(err) {
			return backoff.Permanent(fmt.Errorf(gmess.ERR_BATCHCHROMEOSDEVICE, err.Error(), crosdev.DeviceId))
		}
		// Log the retries
		lg.Warnw(err.Error(),
			"retrying", b.GetElapsedTime().String(),
			"ChromeOS device", crosdev.DeviceId)
		return fmt.Errorf(gmess.ERR_BATCHCHROMEOSDEVICE, err.Error(), crosdev.DeviceId)
	}, b)
	if err != nil {
		// Log final error
		lg.Error(err)
		fmt.Println(cmn.GminMessage(err.Error()))
	}
}

func init() {
	batchUpdateCmd.AddCommand(batchUpdCrOSDevCmd)

	batchUpdCrOSDevCmd.Flags().StringP(flgnm.FLG_INPUTFILE, "i", "", "filepath to device data file or sheet id")
	batchUpdCrOSDevCmd.Flags().StringP(flgnm.FLG_FORMAT, "f", "json", "user data file format")
	batchUpdCrOSDevCmd.Flags().StringP(flgnm.FLG_SHEETRANGE, "s", "", "user data gsheet range")
}
