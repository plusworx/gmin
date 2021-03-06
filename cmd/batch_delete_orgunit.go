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

var batchDelOrgUnitCmd = &cobra.Command{
	Use:     "orgunits [-i input file path]",
	Aliases: []string{"orgunit", "ous", "ou"},
	Example: `gmin batch-delete orgunits -i inputfile.txt
gmin bdel ous -i inputfile.txt
gmin ls ous -o TestOU -a orgunitpath | jq '.organizationUnits[] | .orgUnitPath' -r | gmin bdel ou`,
	Short: "Deletes a batch of orgunits",
	Long: `Deletes a batch of orgunits where orgunit details are provided in a text input file or through a pipe.
			
The input file or piped in data should provide the orgunit paths or ids to be deleted on separate lines like this:

Engineering/Skunkworx
Engineering/SecretOps
Engineering/Surplus

An input Google sheet must have a header row with the following column names being the only ones that are valid:

ouKey [required]

The column name is case insensitive.`,
	RunE: doBatchDelOrgUnit,
}

func doBatchDelOrgUnit(cmd *cobra.Command, args []string) error {
	lg.Debugw("starting doBatchDelOrgUnit()",
		"args", args)
	defer lg.Debug("finished doBatchDelOrgUnit()")

	var orgunits []string

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
		orgunits, err = btch.DeleteProcessTextFile(inputFlgVal, scanner)
		if err != nil {
			return err
		}
	case lwrFmt == "gsheet":
		rangeFlgVal, err := cmd.Flags().GetString(flgnm.FLG_SHEETRANGE)
		if err != nil {
			return err
		}

		orgunits, err = btch.DeleteProcessGSheet(inputFlgVal, rangeFlgVal, ous.OrgUnitAttrMap, ous.KEYNAME)
		if err != nil {
			return err
		}
	default:
		err = fmt.Errorf(gmess.ERR_INVALIDFILEFORMAT, formatFlgVal)
		lg.Error(err)
		return err
	}

	err = bdoProcessDeletion(ds, orgunits)
	if err != nil {
		return err
	}

	return nil
}

func bdoDelete(wg *sync.WaitGroup, oudc *admin.OrgunitsDeleteCall, ouPath string) {
	lg.Debugw("starting bdoDelete()",
		"ouPath", ouPath)
	defer lg.Debug("finished bdoDelete()")

	defer wg.Done()

	b := backoff.NewExponentialBackOff()
	b.MaxElapsedTime = 32 * time.Second

	err := backoff.Retry(func() error {
		var err error

		err = oudc.Do()
		if err == nil {
			fmt.Println(cmn.GminMessage(fmt.Sprintf(gmess.INFO_OUDELETED, ouPath)))
			lg.Infof(gmess.INFO_OUDELETED, ouPath)
			return err
		}
		if !cmn.IsErrRetryable(err) {
			return backoff.Permanent(fmt.Errorf(gmess.ERR_BATCHOU, err.Error(), ouPath))
		}
		// Log the retries
		lg.Warnw(err.Error(),
			"retrying", b.GetElapsedTime().String(),
			"orgunit", ouPath)
		return fmt.Errorf(gmess.ERR_BATCHOU, err.Error(), ouPath)
	}, b)
	if err != nil {
		// Log final error
		lg.Error(err)
		fmt.Println(cmn.GminMessage(err.Error()))
	}
}

func bdoProcessDeletion(ds *admin.Service, orgunits []string) error {
	lg.Debug("starting bdoProcessDeletion()")
	defer lg.Debug("finished bdoProcessDeletion()")

	customerID, err := cfg.ReadConfigString(cfg.CONFIGCUSTID)
	if err != nil {
		return err
	}

	wg := new(sync.WaitGroup)

	for _, orgunit := range orgunits {
		if orgunit[0] == '/' {
			orgunit = orgunit[1:]
		}

		oudc := ds.Orgunits.Delete(customerID, orgunit)

		// Sleep for 2 seconds because only 1 orgunit can be deleted per second but 1 second interval
		// still results in rate limit errors
		time.Sleep(2 * time.Second)

		wg.Add(1)

		go bdoDelete(wg, oudc, orgunit)
	}

	wg.Wait()

	return nil
}

func init() {
	batchDelCmd.AddCommand(batchDelOrgUnitCmd)

	batchDelOrgUnitCmd.Flags().StringVarP(&inputFile, flgnm.FLG_INPUTFILE, "i", "", "filepath to orgunit data text file")
	batchDelOrgUnitCmd.Flags().StringVarP(&delFormat, flgnm.FLG_FORMAT, "f", "text", "orgunit data file format (text or gsheet)")
	batchDelOrgUnitCmd.Flags().StringVarP(&sheetRange, flgnm.FLG_SHEETRANGE, "s", "", "orgunit data gsheet range")
}
