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
	mdevs "github.com/plusworx/gmin/utils/mobiledevices"
	"github.com/spf13/cobra"
	admin "google.golang.org/api/admin/directory/v1"
	sheet "google.golang.org/api/sheets/v4"
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
	logger.Debugw("starting doBatchDelMobDev()",
		"args", args)

	ds, err := cmn.CreateDirectoryService(admin.AdminDirectoryDeviceMobileScope)
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

	lwrFmt := strings.ToLower(delFormat)

	ok := cmn.SliceContainsStr(cmn.ValidFileFormats, lwrFmt)
	if !ok {
		err = fmt.Errorf(cmn.ErrInvalidFileFormat, delFormat)
		logger.Error(err)
		return err
	}

	switch {
	case lwrFmt == "text":
		err := bdmdProcessTextFile(ds, inputFile, scanner)
		if err != nil {
			logger.Error(err)
			return err
		}
	case lwrFmt == "gsheet":
		err := bdmdProcessGSheet(ds, inputFile)
		if err != nil {
			logger.Error(err)
			return err
		}
	default:
		return fmt.Errorf(cmn.ErrInvalidFileFormat, format)
	}

	logger.Debug("finished doBatchDelMobDev()")
	return nil
}

func bdmdDeleteObject(wg *sync.WaitGroup, mdc *admin.MobiledevicesDeleteCall, resourceID string) {
	logger.Debugw("starting bdmdDeleteObject()",
		"resourceID", resourceID)

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
		logger.Warnw(err.Error(),
			"retrying", b.GetElapsedTime().String(),
			"resourceID", resourceID)
		return fmt.Errorf(cmn.ErrBatchMobileDevice, err.Error(), resourceID)
	}, b)
	if err != nil {
		// Log final error
		logger.Error(err)
		fmt.Println(cmn.GminMessage(err.Error()))
	}
	logger.Debug("finished bdmdDeleteObject()")
}

func bdmdFromFileFactory(hdrMap map[int]string, mobDevData []interface{}) (string, error) {
	logger.Debugw("starting bdmdFromFileFactory()",
		"hdrMap", hdrMap)

	var mobResID string

	for idx, val := range mobDevData {
		attrName := hdrMap[idx]
		attrVal := fmt.Sprintf("%v", val)

		if attrName == "resourceId" {
			mobResID = attrVal
		}
	}
	logger.Debug("finished bdmdFromFileFactory()")
	return mobResID, nil
}

func bdmdProcessDeletion(ds *admin.Service, mobdevs []string) error {
	logger.Debug("starting bdmdProcessDeletion()")

	customerID, err := cfg.ReadConfigString("customerid")
	if err != nil {
		logger.Error(err)
		return err
	}

	wg := new(sync.WaitGroup)

	for _, mobResID := range mobdevs {
		mdc := ds.Mobiledevices.Delete(customerID, mobResID)

		wg.Add(1)

		go bdmdDeleteObject(wg, mdc, mobResID)
	}

	wg.Wait()

	logger.Debug("finished bdmdProcessDeletion()")
	return nil
}

func bdmdProcessGSheet(ds *admin.Service, sheetID string) error {
	logger.Debugw("starting bdmdProcessGSheet()",
		"sheetID", sheetID)

	var mobdevs []string

	if sheetRange == "" {
		err := errors.New(cmn.ErrNoSheetRange)
		logger.Error(err)
		return err
	}

	ss, err := cmn.CreateSheetService(sheet.DriveReadonlyScope)
	if err != nil {
		logger.Error(err)
		return err
	}

	ssvgc := ss.Spreadsheets.Values.Get(sheetID, sheetRange)
	sValRange, err := ssvgc.Do()
	if err != nil {
		logger.Error(err)
		return err
	}

	if len(sValRange.Values) == 0 {
		err = fmt.Errorf(cmn.ErrNoSheetDataFound, sheetID, sheetRange)
		logger.Error(err)
		return err
	}

	hdrMap := cmn.ProcessHeader(sValRange.Values[0])
	err = cmn.ValidateHeader(hdrMap, mdevs.MobDevAttrMap)
	if err != nil {
		logger.Error(err)
		return err
	}

	for idx, row := range sValRange.Values {
		if idx == 0 {
			continue
		}

		mobDevVar, err := bdmdFromFileFactory(hdrMap, row)
		if err != nil {
			logger.Error(err)
			return err
		}

		mobdevs = append(mobdevs, mobDevVar)
	}

	err = bdmdProcessDeletion(ds, mobdevs)
	if err != nil {
		logger.Error(err)
		return err
	}
	logger.Debug("finished bdmdProcessGSheet()")
	return nil
}

func bdmdProcessTextFile(ds *admin.Service, filePath string, scanner *bufio.Scanner) error {
	logger.Debugw("starting bdmdProcessTextFile()",
		"filePath", filePath)

	var mobdevs []string

	if filePath != "" {
		file, err := os.Open(filePath)
		if err != nil {
			logger.Error(err)
			return err
		}
		defer file.Close()
		scanner = bufio.NewScanner(file)
	}

	for scanner.Scan() {
		mobdev := scanner.Text()
		mobdevs = append(mobdevs, mobdev)
	}

	err := bdmdProcessDeletion(ds, mobdevs)
	if err != nil {
		logger.Error(err)
		return err
	}

	logger.Debug("finished bdmdProcessTextFile()")
	return nil
}

func init() {
	batchDelCmd.AddCommand(batchDelMobDevCmd)

	batchDelMobDevCmd.Flags().StringVarP(&inputFile, "input-file", "i", "", "filepath to mobile device data text file")
	batchDelMobDevCmd.Flags().StringVarP(&delFormat, "format", "f", "text", "mobile device data file format (text or gsheet)")
	batchDelMobDevCmd.Flags().StringVarP(&sheetRange, "sheet-range", "s", "", "mobile device data gsheet range")
}
