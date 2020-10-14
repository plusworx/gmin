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
	"sync"
	"time"

	"github.com/cenkalti/backoff/v4"
	cmn "github.com/plusworx/gmin/utils/common"
	cfg "github.com/plusworx/gmin/utils/config"
	gmess "github.com/plusworx/gmin/utils/gminmessages"
	ous "github.com/plusworx/gmin/utils/orgunits"
	"github.com/spf13/cobra"
	admin "google.golang.org/api/admin/directory/v1"
	sheet "google.golang.org/api/sheets/v4"
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

n input Google sheet must have a header row with the following column names being the only ones that are valid:

ouKey [required]

The column name is case insensitive.`,
	RunE: doBatchDelOrgUnit,
}

func doBatchDelOrgUnit(cmd *cobra.Command, args []string) error {
	logger.Debugw("starting doBatchDelOrgUnit()",
		"args", args)

	var orgunits []string

	ds, err := cmn.CreateDirectoryService(admin.AdminDirectoryOrgunitScope)
	if err != nil {
		logger.Error(err)
		return err
	}

	inputFlgVal, err := cmd.Flags().GetString("input-file")
	if err != nil {
		logger.Error(err)
		return err
	}

	scanner, err := cmn.InputFromStdIn(inputFlgVal)
	if err != nil {
		logger.Error(err)
		return err
	}

	if inputFlgVal == "" && scanner == nil {
		err := errors.New(gmess.ERR_NOINPUTFILE)
		logger.Error(err)
		return err
	}

	formatFlgVal, err := cmd.Flags().GetString("format")
	if err != nil {
		logger.Error(err)
		return err
	}
	lwrFmt := strings.ToLower(formatFlgVal)

	ok := cmn.SliceContainsStr(cmn.ValidFileFormats, lwrFmt)
	if !ok {
		err = fmt.Errorf(gmess.ERR_INVALIDFILEFORMAT, formatFlgVal)
		logger.Error(err)
		return err
	}

	switch {
	case lwrFmt == "text":
		orgunits, err = bdoProcessTextFile(ds, inputFlgVal, scanner)
		if err != nil {
			logger.Error(err)
			return err
		}
	case lwrFmt == "gsheet":
		rangeFlgVal, err := cmd.Flags().GetString("sheet-range")
		if err != nil {
			logger.Error(err)
			return err
		}

		orgunits, err = bdoProcessGSheet(ds, inputFlgVal, rangeFlgVal)
		if err != nil {
			logger.Error(err)
			return err
		}
	default:
		return fmt.Errorf(gmess.ERR_INVALIDFILEFORMAT, formatFlgVal)
	}

	err = bdoProcessDeletion(ds, orgunits)
	if err != nil {
		logger.Error(err)
		return err
	}

	logger.Debug("finished doBatchDelOrgUnit()")
	return nil
}

func bdoDelete(wg *sync.WaitGroup, oudc *admin.OrgunitsDeleteCall, ouPath string) {
	logger.Debugw("starting bdoDelete()",
		"ouPath", ouPath)

	defer wg.Done()

	b := backoff.NewExponentialBackOff()
	b.MaxElapsedTime = 32 * time.Second

	err := backoff.Retry(func() error {
		var err error

		err = oudc.Do()
		if err == nil {
			logger.Infof(gmess.INFO_OUDELETED, ouPath)
			fmt.Println(cmn.GminMessage(fmt.Sprintf(gmess.INFO_OUDELETED, ouPath)))
			return err
		}
		if !cmn.IsErrRetryable(err) {
			return backoff.Permanent(fmt.Errorf(gmess.ERR_BATCHOU, err.Error(), ouPath))
		}
		// Log the retries
		logger.Warnw(err.Error(),
			"retrying", b.GetElapsedTime().String(),
			"orgunit", ouPath)
		return fmt.Errorf(gmess.ERR_BATCHOU, err.Error(), ouPath)
	}, b)
	if err != nil {
		// Log final error
		logger.Error(err)
		fmt.Println(cmn.GminMessage(err.Error()))
	}
	logger.Debug("finished bdoDelete()")
}

func bdoFromFileFactory(hdrMap map[int]string, ouData []interface{}) (string, error) {
	logger.Debugw("starting bdoFromFileFactory()",
		"hdrMap", hdrMap)

	var orgunit string

	for idx, val := range ouData {
		attrName := hdrMap[idx]
		attrVal := fmt.Sprintf("%v", val)

		if attrName == "ouKey" {
			orgunit = attrVal
		}
	}
	logger.Debug("finished bdoFromFileFactory()")
	return orgunit, nil
}

func bdoProcessDeletion(ds *admin.Service, orgunits []string) error {
	logger.Debug("starting bdoProcessDeletion()")

	customerID, err := cfg.ReadConfigString("customerid")
	if err != nil {
		logger.Error(err)
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

	logger.Debug("finished bdoProcessDeletion()")
	return nil
}

func bdoProcessGSheet(ds *admin.Service, sheetID string, sheetrange string) ([]string, error) {
	logger.Debugw("starting bdoProcessGSheet()",
		"sheetID", sheetID,
		"sheetrange", sheetrange)

	var orgunits []string

	if sheetrange == "" {
		err := errors.New(gmess.ERR_NOSHEETRANGE)
		logger.Error(err)
		return nil, err
	}

	ss, err := cmn.CreateSheetService(sheet.DriveReadonlyScope)
	if err != nil {
		logger.Error(err)
		return nil, err
	}

	ssvgc := ss.Spreadsheets.Values.Get(sheetID, sheetrange)
	sValRange, err := ssvgc.Do()
	if err != nil {
		logger.Error(err)
		return nil, err
	}

	if len(sValRange.Values) == 0 {
		err = fmt.Errorf(gmess.ERR_NOSHEETDATAFOUND, sheetID, sheetrange)
		logger.Error(err)
		return nil, err
	}

	hdrMap := cmn.ProcessHeader(sValRange.Values[0])
	err = cmn.ValidateHeader(hdrMap, ous.OrgUnitAttrMap)
	if err != nil {
		logger.Error(err)
		return nil, err
	}

	for idx, row := range sValRange.Values {
		if idx == 0 {
			continue
		}

		ouVar, err := bdoFromFileFactory(hdrMap, row)
		if err != nil {
			logger.Error(err)
			return nil, err
		}

		orgunits = append(orgunits, ouVar)
	}

	logger.Debug("finished bdoProcessGSheet()")
	return orgunits, nil
}

func bdoProcessTextFile(ds *admin.Service, filePath string, scanner *bufio.Scanner) ([]string, error) {
	logger.Debugw("starting bdoProcessTextFile()",
		"filePath", filePath)

	var orgunits []string

	if filePath != "" {
		file, err := os.Open(filePath)
		if err != nil {
			logger.Error(err)
			return nil, err
		}
		defer file.Close()
		scanner = bufio.NewScanner(file)
	}

	for scanner.Scan() {
		orgunit := scanner.Text()
		orgunits = append(orgunits, orgunit)
	}

	logger.Debug("finished bdoProcessTextFile()")
	return orgunits, nil
}

func init() {
	batchDelCmd.AddCommand(batchDelOrgUnitCmd)

	batchDelOrgUnitCmd.Flags().StringVarP(&inputFile, "input-file", "i", "", "filepath to orgunit data text file")
	batchDelOrgUnitCmd.Flags().StringVarP(&delFormat, "format", "f", "text", "orgunit data file format (text or gsheet)")
	batchDelOrgUnitCmd.Flags().StringVarP(&sheetRange, "sheet-range", "s", "", "orgunit data gsheet range")
}
