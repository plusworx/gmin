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
	ous "github.com/plusworx/gmin/utils/orgunits"
	"github.com/spf13/cobra"
	admin "google.golang.org/api/admin/directory/v1"
)

var batchUpdOUCmd = &cobra.Command{
	Use:     "orgunits -i <input file path>",
	Aliases: []string{"orgunit", "ous", "ou"},
	Example: `gmin batch-update orgunits -i inputfile.json
gmin bupd ous -i inputfile.csv -f csv
gmin bupd ou -i 1odyAIp3jGspd3M4xeepxWD6aeQIUuHBgrZB2OHSu8MI -s 'Sheet1!A1:K25' -f gsheet`,
	Short: "Updates a batch of orgunits",
	Long: `Updates a batch of orgunits where orgunit details are provided in a Google Sheet, CSV/JSON input file or piped JSON.
			  
The contents of the JSON file or piped input should look something like this:

{"ouKey":"Credit","parentOrgUnitPath":"/Finance","name":"Credit_Control"}
{"ouKey":"Audit","parentOrgUnitPath":"/Finance","name":"Audit_Governance"}
{"ouKey":"Planning","parentOrgUnitPath":"/Finance","name":"Planning_Reporting"}

N.B. ouKey (full orgunit path or id) must be provided.

CSV and Google sheets must have a header row with the following column names being the only ones that are valid:

blockInheritance [value true or false]
description
name
ouKey [required]
parentOrgUnitPath

The column names are case insensitive and can be in any order.`,
	RunE: doBatchUpdOU,
}

func doBatchUpdOU(cmd *cobra.Command, args []string) error {
	lg.Debugw("starting doBatchUpdOU()",
		"args", args)
	defer lg.Debug("finished doBatchUpdOU()")

	var ouParams []ous.OrgUnitParams

	srv, err := cmn.CreateService(cmn.SRVTYPEADMIN, admin.AdminDirectoryOrgunitScope)
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

	callParams := btch.CallParams{CallType: cmn.CALLTYPEUPDATE, ObjectType: cmn.OBJTYPEORGUNIT}
	inputParams := btch.ProcessInputParams{
		Format:      lwrFmt,
		InputFlgVal: inputFlgVal,
		Scanner:     scanner,
		SheetRange:  rangeFlgVal,
	}

	objs, err := buoProcessInput(callParams, inputParams)
	if err != nil {
		return err
	}

	for _, opObj := range objs {
		ouParams = append(ouParams, opObj.(ous.OrgUnitParams))
	}

	err = buoProcessObjects(ds, ouParams)
	if err != nil {
		return err
	}

	return nil
}

func buoProcessInput(callParams btch.CallParams, inputParams btch.ProcessInputParams) ([]interface{}, error) {
	lg.Debug("starting buoProcessInput()")
	defer lg.Debug("finished buoProcessInput()")

	var (
		err  error
		objs []interface{}
	)

	switch inputParams.Format {
	case "csv":
		objs, err = btch.ProcessCSVFile(callParams, inputParams.InputFlgVal, ous.OrgUnitAttrMap)
		if err != nil {
			return nil, err
		}
	case "json":
		objs, err = btch.ProcessJSON(callParams, inputParams.InputFlgVal, inputParams.Scanner, ous.OrgUnitAttrMap)
		if err != nil {
			return nil, err
		}
	case "gsheet":
		objs, err = btch.ProcessGSheet(callParams, inputParams.InputFlgVal, inputParams.SheetRange, ous.OrgUnitAttrMap)
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

func buoProcessObjects(ds *admin.Service, ouParams []ous.OrgUnitParams) error {
	lg.Debugw("starting buoProcessObjects()",
		"ouParams", ouParams)
	defer lg.Debug("finished buoProcessObjects()")

	wg := new(sync.WaitGroup)

	customerID, err := cfg.ReadConfigString(cfg.CONFIGCUSTID)
	if err != nil {
		lg.Error(err)
		return err
	}

	for _, op := range ouParams {
		ouuc := ds.Orgunits.Update(customerID, op.OUKey, op.OrgUnit)

		wg.Add(1)

		// Sleep for 2 seconds because only 1 orgunit can be updated per second but 1 second interval
		// still can result in rate limit errors
		time.Sleep(2 * time.Second)

		go buoUpdate(op.OrgUnit, wg, ouuc, op.OUKey)
	}

	wg.Wait()

	return nil
}

func buoUpdate(orgunit *admin.OrgUnit, wg *sync.WaitGroup, ouuc *admin.OrgunitsUpdateCall, ouKey string) {
	lg.Debugw("starting buoUpdate()",
		"ouKey", ouKey)
	defer lg.Debug("finished buoUpdate()")

	defer wg.Done()

	b := backoff.NewExponentialBackOff()
	b.MaxElapsedTime = 32 * time.Second

	err := backoff.Retry(func() error {
		var err error
		_, err = ouuc.Do()
		if err == nil {
			fmt.Println(cmn.GminMessage(fmt.Sprintf(gmess.INFO_OUUPDATED, ouKey)))
			lg.Infof(gmess.INFO_OUUPDATED, ouKey)
			return err
		}
		if !cmn.IsErrRetryable(err) {
			return backoff.Permanent(fmt.Errorf(gmess.ERR_BATCHOU, err.Error(), ouKey))
		}
		// Log the retries
		lg.Warnw(err.Error(),
			"retrying", b.GetElapsedTime().String(),
			"orgunit", ouKey)
		return fmt.Errorf(gmess.ERR_BATCHOU, err.Error(), ouKey)
	}, b)
	if err != nil {
		// Log final error
		lg.Error(err)
		fmt.Println(cmn.GminMessage(err.Error()))
	}
}

func init() {
	batchUpdateCmd.AddCommand(batchUpdOUCmd)

	batchUpdOUCmd.Flags().StringP(flgnm.FLG_INPUTFILE, "i", "", "filepath to orgunit data file or sheet id")
	batchUpdOUCmd.Flags().StringP(flgnm.FLG_FORMAT, "f", "json", "user data file format")
	batchUpdOUCmd.Flags().StringP(flgnm.FLG_SHEETRANGE, "s", "", "user data gsheet range")
}
