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

var batchMoveCrOSDevCmd = &cobra.Command{
	Use:     "chromeos-devices -i <input file>",
	Aliases: []string{"chromeos-device", "cros-devices", "cros-device", "cros-devs", "cros-dev", "cdevs", "cdev"},
	Example: `gmin batch-move chromeos-devices -i inputfile.txt
gmin bmv cdevs -i inputfile.csv -f csv
gmin bmv cdev -i 1odyAIp3jGspd3M4xeepxWD6aeQIUuHBgrZB2OHSu8MI -s 'Sheet1!A1:C25' -f gsheet`,
	Short: "Moves a batch of ChromeOS devices to another orgunit",
	Long: `Moves a batch of ChromeOS devices to another orgunit where device details are provided in a Google Sheet, CSV/JSON input file or piped JSON.
			
The JSON file or piped input should contain ChromeOS device move details like this:

{"deviceId":"5ac7be73-5996-394e-9c30-62d41a8f10e8","orgUnitPath":"/IT"}
{"deviceId":"6ac9bd33-7095-453e-6c39-22d48a8f13e8","orgUnitPath":"/Engineering"}
{"deviceId":"6bc4be13-9916-494e-9c39-62d45c8f40e9","orgUnitPath":"/Sales"}

CSV and Google sheets must have a header row with the following column names being the only ones that are valid:

deviceId [required]
orgUnitPath [required]

The column names are case insensitive and can be in any order.`,
	RunE: doBatchMoveCrOSDev,
}

func doBatchMoveCrOSDev(cmd *cobra.Command, args []string) error {
	lg.Debugw("starting doBatchMoveCrOSDev()",
		"args", args)
	defer lg.Debug("finished doBatchMoveCrOSDev()")

	var movedDevs []cdevs.MovedDevice

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

	rangeFlgVal, err := cmd.Flags().GetString(flgnm.FLG_SHEETRANGE)
	if err != nil {
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

	callParams := btch.CallParams{CallType: cmn.CALLTYPEMOVE, ObjectType: cmn.OBJTYPECROSDEV}
	inputParams := btch.ProcessInputParams{
		Format:      lwrFmt,
		InputFlgVal: inputFlgVal,
		Scanner:     scanner,
		SheetRange:  rangeFlgVal,
	}

	objs, err := bmvcProcessInput(callParams, inputParams)
	if err != nil {
		return err
	}

	for _, cdevObj := range objs {
		movedDevs = append(movedDevs, cdevObj.(cdevs.MovedDevice))
	}

	err = bmvcProcessObjects(ds, movedDevs)
	if err != nil {
		return err
	}

	return nil
}

func bmvcPerformMove(deviceID string, ouPath string, wg *sync.WaitGroup, cdmc *admin.ChromeosdevicesMoveDevicesToOuCall) {
	lg.Debugw("starting bmvcPerformMove()",
		"deviceID", deviceID,
		"ouPath", ouPath)
	defer lg.Debug("finished bmvcPerformMove()")

	defer wg.Done()

	b := backoff.NewExponentialBackOff()
	b.MaxElapsedTime = 32 * time.Second

	err := backoff.Retry(func() error {
		var err error
		err = cdmc.Do()
		if err == nil {
			fmt.Println(cmn.GminMessage(fmt.Sprintf(gmess.INFO_CDEVMOVEPERFORMED, deviceID, ouPath)))
			lg.Infof(gmess.INFO_CDEVMOVEPERFORMED, deviceID, ouPath)
			return err
		}
		if !cmn.IsErrRetryable(err) {
			return backoff.Permanent(fmt.Errorf(gmess.ERR_BATCHCHROMEOSDEVICE, err.Error(), deviceID))
		}
		// Log the retries
		lg.Warnw(err.Error(),
			"retrying", b.GetElapsedTime().String(),
			"ChromeOS device", deviceID,
			"orgunit", ouPath)
		return fmt.Errorf(gmess.ERR_BATCHCHROMEOSDEVICE, err.Error(), deviceID)
	}, b)
	if err != nil {
		// Log final error
		lg.Error(err)
		fmt.Println(cmn.GminMessage(err.Error()))
	}
}

func bmvcProcessInput(callParams btch.CallParams, inputParams btch.ProcessInputParams) ([]interface{}, error) {
	lg.Debug("starting bmvcProcessInput()")
	defer lg.Debug("finished bmvcProcessInput()")

	var (
		err  error
		objs []interface{}
	)

	switch inputParams.Format {
	case "csv":
		objs, err = btch.ProcessCSVFile(callParams, inputParams.InputFlgVal, cdevs.CrOSDevAttrMap)
		if err != nil {
			return nil, err
		}
	case "json":
		objs, err = btch.ProcessJSON(callParams, inputParams.InputFlgVal, inputParams.Scanner, cdevs.CrOSDevAttrMap)
		if err != nil {
			return nil, err
		}
	case "gsheet":
		objs, err = btch.ProcessGSheet(callParams, inputParams.InputFlgVal, inputParams.SheetRange, cdevs.CrOSDevAttrMap)
		if err != nil {
			return nil, err
		}
	default:
		err = fmt.Errorf(gmess.ERR_INVALIDFILEFORMAT, inputParams.Format)
		lg.Error(err)
		return nil, err
	}

	return objs, nil
}

func bmvcProcessObjects(ds *admin.Service, movedDevs []cdevs.MovedDevice) error {
	lg.Debug("starting bmvcProcessObjects()")
	defer lg.Debug("finished bmvcProcessObjects()")

	customerID, err := cfg.ReadConfigString(cfg.CONFIGCUSTID)
	if err != nil {
		lg.Error(err)
		return err
	}

	wg := new(sync.WaitGroup)

	for _, md := range movedDevs {
		move := admin.ChromeOsMoveDevicesToOu{}
		deviceIDs := []string{}

		deviceIDs = append(deviceIDs, md.DeviceId)
		move.DeviceIds = deviceIDs

		cdmc := ds.Chromeosdevices.MoveDevicesToOu(customerID, md.OrgUnitPath, &move)

		wg.Add(1)

		go bmvcPerformMove(md.DeviceId, md.OrgUnitPath, wg, cdmc)
	}

	wg.Wait()

	return nil
}

func init() {
	batchMoveCmd.AddCommand(batchMoveCrOSDevCmd)

	batchMoveCrOSDevCmd.Flags().StringP(flgnm.FLG_INPUTFILE, "i", "", "filepath to device data file")
	batchMoveCrOSDevCmd.Flags().StringP(flgnm.FLG_FORMAT, "f", "json", "user data file format")
	batchMoveCrOSDevCmd.Flags().StringP(flgnm.FLG_SHEETRANGE, "s", "", "user data gsheet range")
}
