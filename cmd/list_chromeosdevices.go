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
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"

	cdevs "github.com/plusworx/gmin/utils/chromeosdevices"
	cmn "github.com/plusworx/gmin/utils/common"
	cfg "github.com/plusworx/gmin/utils/config"
	"github.com/spf13/cobra"
	admin "google.golang.org/api/admin/directory/v1"
)

var listCrOSDevsCmd = &cobra.Command{
	Use:     "chromeos-devices",
	Aliases: []string{"chromeos-device", "cros-devices", "cros-device", "cros-devs", "cros-dev", "cdevs", "cdev"},
	Args:    cobra.NoArgs,
	Example: `gmin list chromeos-devices --pages all --count
gmin ls cdevs --pages all`,
	Short: "Outputs a list of ChromeOS devices",
	Long:  `Outputs a list of ChromeOS devices.`,
	RunE:  doListCrOSDevs,
}

func doListCrOSDevs(cmd *cobra.Command, args []string) error {
	logger.Debugw("starting doListCrOSDevs()",
		"args", args)

	var (
		crosdevs *admin.ChromeOsDevices
	)

	ds, err := cmn.CreateDirectoryService(admin.AdminDirectoryDeviceChromeosReadonlyScope)
	if err != nil {
		logger.Error(err)
		return err
	}

	customerID, err := cfg.ReadConfigString("customerid")
	if err != nil {
		logger.Error(err)
		return err
	}

	cdlc := ds.Chromeosdevices.List(customerID)

	if attrs != "" {
		listAttrs, err := cmn.ParseOutputAttrs(attrs, cdevs.CrOSDevAttrMap)
		if err != nil {
			logger.Error(err)
			return err
		}
		formattedAttrs := cdevs.StartChromeDevicesField + listAttrs + cdevs.EndField
		listCall := cdevs.AddFields(cdlc, formattedAttrs)
		cdlc = listCall.(*admin.ChromeosdevicesListCall)
	}

	if orderBy != "" {
		ob := strings.ToLower(orderBy)
		ok := cmn.SliceContainsStr(cdevs.ValidOrderByStrs, ob)
		if !ok {
			err = fmt.Errorf(cmn.ErrInvalidOrderBy, orderBy)
			logger.Error(err)
			return err
		}

		validOrderBy, err := cmn.IsValidAttr(ob, cdevs.CrOSDevAttrMap)
		if err != nil {
			logger.Error(err)
			return err
		}

		cdlc = cdevs.AddOrderBy(cdlc, validOrderBy)

		if sortOrder != "" {
			so := strings.ToLower(sortOrder)
			validSortOrder, err := cmn.IsValidAttr(so, cmn.ValidSortOrders)
			if err != nil {
				logger.Error(err)
				return err
			}

			cdlc = cdevs.AddSortOrder(cdlc, validSortOrder)
		}
	}

	if orgUnit != "" {
		cdlc = cdevs.AddQuery(cdlc, orgUnit)
	}

	if projection != "" {
		proj := strings.ToLower(projection)
		ok := cmn.SliceContainsStr(cdevs.ValidProjections, proj)
		if !ok {
			err = fmt.Errorf(cmn.ErrInvalidProjectionType, projection)
			logger.Error(err)
			return err
		}

		listCall := cdevs.AddProjection(cdlc, proj)
		cdlc = listCall.(*admin.ChromeosdevicesListCall)
	}

	if query != "" {
		formattedQuery, err := cmn.ParseQuery(query, cdevs.QueryAttrMap)
		if err != nil {
			logger.Error(err)
			return err
		}

		cdlc = cdevs.AddQuery(cdlc, formattedQuery)
	}

	cdlc = cdevs.AddMaxResults(cdlc, maxResults)

	crosdevs, err = cdevs.DoList(cdlc)
	if err != nil {
		logger.Error(err)
		return err
	}

	if pages != "" {
		err = doCrOSDevPages(cdlc, crosdevs, pages)
		if err != nil {
			logger.Error(err)
			return err
		}
	}

	jsonData, err := json.MarshalIndent(crosdevs, "", "    ")
	if err != nil {
		logger.Error(err)
		return err
	}

	if count {
		fmt.Println(len(crosdevs.Chromeosdevices))
	} else {
		fmt.Println(string(jsonData))
	}

	logger.Debug("finished doListCrOSDevs()")
	return nil
}

func doCrOSDevAllPages(cdlc *admin.ChromeosdevicesListCall, crosdevs *admin.ChromeOsDevices) error {
	logger.Debug("starting doCrOSDevAllPages()")

	if crosdevs.NextPageToken != "" {
		cdlc = cdevs.AddPageToken(cdlc, crosdevs.NextPageToken)
		nxtCrOSDevs, err := cdevs.DoList(cdlc)
		if err != nil {
			logger.Error(err)
			return err
		}
		crosdevs.Chromeosdevices = append(crosdevs.Chromeosdevices, nxtCrOSDevs.Chromeosdevices...)
		crosdevs.Etag = nxtCrOSDevs.Etag
		crosdevs.NextPageToken = nxtCrOSDevs.NextPageToken

		if nxtCrOSDevs.NextPageToken != "" {
			doCrOSDevAllPages(cdlc, crosdevs)
		}
	}

	logger.Debug("finished doCrOSDevAllPages()")
	return nil
}

func doCrOSDevNumPages(cdlc *admin.ChromeosdevicesListCall, crosdevs *admin.ChromeOsDevices, numPages int) error {
	logger.Debugw("starting doCrOSDevNumPages()",
		"numPages", numPages)

	if crosdevs.NextPageToken != "" && numPages > 0 {
		cdlc = cdevs.AddPageToken(cdlc, crosdevs.NextPageToken)
		nxtCrOSDevs, err := cdevs.DoList(cdlc)
		if err != nil {
			logger.Error(err)
			return err
		}
		crosdevs.Chromeosdevices = append(crosdevs.Chromeosdevices, nxtCrOSDevs.Chromeosdevices...)
		crosdevs.Etag = nxtCrOSDevs.Etag
		crosdevs.NextPageToken = nxtCrOSDevs.NextPageToken

		if nxtCrOSDevs.NextPageToken != "" {
			doCrOSDevNumPages(cdlc, crosdevs, numPages-1)
		}
	}

	logger.Debug("finished doCrOSDevNumPages()")
	return nil
}

func doCrOSDevPages(cdlc *admin.ChromeosdevicesListCall, crosdevs *admin.ChromeOsDevices, pages string) error {
	logger.Debugw("starting doCrOSDevPages()",
		"pages", pages)

	if pages == "all" {
		err := doCrOSDevAllPages(cdlc, crosdevs)
		if err != nil {
			logger.Error(err)
			return err
		}
	} else {
		numPages, err := strconv.Atoi(pages)
		if err != nil {
			err = errors.New(cmn.ErrInvalidPagesArgument)
			logger.Error(err)
			return err
		}

		if numPages > 1 {
			err = doCrOSDevNumPages(cdlc, crosdevs, numPages-1)
			if err != nil {
				logger.Error(err)
				return err
			}
		}
	}

	logger.Debug("finished doCrOSDevPages()")
	return nil
}

func init() {
	listCmd.AddCommand(listCrOSDevsCmd)

	listCrOSDevsCmd.Flags().StringVarP(&attrs, "attributes", "a", "", "required device attributes (separated by ~)")
	listCrOSDevsCmd.Flags().BoolVarP(&count, "count", "", false, "count number of entities returned")
	listCrOSDevsCmd.Flags().Int64VarP(&maxResults, "max-results", "m", 200, "maximum number of results to return per page")
	listCrOSDevsCmd.Flags().StringVarP(&orderBy, "order-by", "o", "", "field by which results will be ordered")
	listCrOSDevsCmd.Flags().StringVarP(&pages, "pages", "p", "", "number of pages of results to be returned ('all' or a number)")
	listCrOSDevsCmd.Flags().StringVarP(&projection, "projection", "j", "", "type of projection")
	listCrOSDevsCmd.Flags().StringVarP(&query, "query", "q", "", "selection criteria to get devices (separated by ~)")
	listCrOSDevsCmd.Flags().StringVarP(&sortOrder, "sort-order", "s", "", "sort order of returned results")
	listCrOSDevsCmd.Flags().StringVarP(&orgUnit, "orgunit-path", "t", "", "sets orgunit path that returned devices belong to")
}
