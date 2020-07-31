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
	Use:     "chromeosdevices",
	Aliases: []string{"chromeosdevice", "crosdevices", "crosdevice", "cdevs", "cdev"},
	Short:   "Outputs a list of ChromeOS devices",
	Long:    `Outputs a list of ChromeOS devices.`,
	RunE:    doListCrOSDevs,
}

func doListCrOSDevs(cmd *cobra.Command, args []string) error {
	var (
		crosdevs *admin.ChromeOsDevices
	)

	ds, err := cmn.CreateDirectoryService(admin.AdminDirectoryDeviceChromeosReadonlyScope)
	if err != nil {
		return err
	}

	customerID, err := cfg.ReadConfigString("customerid")
	if err != nil {
		return err
	}

	cdlc := ds.Chromeosdevices.List(customerID)

	if attrs != "" {
		listAttrs, err := cmn.ParseOutputAttrs(attrs, cdevs.CrOSDevAttrMap)
		if err != nil {
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
			err = fmt.Errorf("gmin: error - %v is not a valid order by field", orderBy)
			return err
		}

		validOrderBy, err := cmn.IsValidAttr(ob, cdevs.CrOSDevAttrMap)
		if err != nil {
			return err
		}

		cdlc = cdevs.AddOrderBy(cdlc, validOrderBy)

		if sortOrder != "" {
			so := strings.ToLower(sortOrder)
			validSortOrder, err := cmn.IsValidAttr(so, cmn.ValidSortOrders)
			if err != nil {
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
			return fmt.Errorf("gmin: error - %v is not a valid projection type", projection)
		}

		listCall := cdevs.AddProjection(cdlc, proj)
		cdlc = listCall.(*admin.ChromeosdevicesListCall)
	}

	if query != "" {
		formattedQuery, err := cmn.ParseQuery(query, cdevs.QueryAttrMap)
		if err != nil {
			return err
		}

		cdlc = cdevs.AddQuery(cdlc, formattedQuery)
	}

	cdlc = cdevs.AddMaxResults(cdlc, maxResults)

	crosdevs, err = cdevs.DoList(cdlc)
	if err != nil {
		return err
	}

	if pages != "" {
		err = doCrOSDevPages(cdlc, crosdevs, pages)
		if err != nil {
			return err
		}
	}

	jsonData, err := json.MarshalIndent(crosdevs, "", "    ")
	if err != nil {
		return err
	}

	if count {
		fmt.Println(len(crosdevs.Chromeosdevices))
	} else {
		fmt.Println(string(jsonData))
	}

	return nil
}

func doCrOSDevAllPages(cdlc *admin.ChromeosdevicesListCall, crosdevs *admin.ChromeOsDevices) error {
	if crosdevs.NextPageToken != "" {
		cdlc = cdevs.AddPageToken(cdlc, crosdevs.NextPageToken)
		nxtCrOSDevs, err := cdevs.DoList(cdlc)
		if err != nil {
			return err
		}
		crosdevs.Chromeosdevices = append(crosdevs.Chromeosdevices, nxtCrOSDevs.Chromeosdevices...)
		crosdevs.Etag = nxtCrOSDevs.Etag
		crosdevs.NextPageToken = nxtCrOSDevs.NextPageToken

		if nxtCrOSDevs.NextPageToken != "" {
			doCrOSDevAllPages(cdlc, crosdevs)
		}
	}

	return nil
}

func doCrOSDevNumPages(cdlc *admin.ChromeosdevicesListCall, crosdevs *admin.ChromeOsDevices, numPages int) error {
	if crosdevs.NextPageToken != "" && numPages > 0 {
		cdlc = cdevs.AddPageToken(cdlc, crosdevs.NextPageToken)
		nxtCrOSDevs, err := cdevs.DoList(cdlc)
		if err != nil {
			return err
		}
		crosdevs.Chromeosdevices = append(crosdevs.Chromeosdevices, nxtCrOSDevs.Chromeosdevices...)
		crosdevs.Etag = nxtCrOSDevs.Etag
		crosdevs.NextPageToken = nxtCrOSDevs.NextPageToken

		if nxtCrOSDevs.NextPageToken != "" {
			doCrOSDevNumPages(cdlc, crosdevs, numPages-1)
		}
	}

	return nil
}

func doCrOSDevPages(cdlc *admin.ChromeosdevicesListCall, crosdevs *admin.ChromeOsDevices, pages string) error {
	if pages == "all" {
		err := doCrOSDevAllPages(cdlc, crosdevs)
		if err != nil {
			return err
		}
	} else {
		numPages, err := strconv.Atoi(pages)
		if err != nil {
			return errors.New("gmin: error - pages must be 'all' or a number")
		}

		if numPages > 1 {
			err = doCrOSDevNumPages(cdlc, crosdevs, numPages-1)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func init() {
	listCmd.AddCommand(listCrOSDevsCmd)

	listCrOSDevsCmd.Flags().StringVarP(&attrs, "attributes", "a", "", "required device attributes (separated by ~)")
	listCrOSDevsCmd.Flags().BoolVarP(&count, "count", "", false, "count number of entities returned")
	listCrOSDevsCmd.Flags().Int64VarP(&maxResults, "maxresults", "m", 200, "maximum number of results to return per page")
	listCrOSDevsCmd.Flags().StringVarP(&orderBy, "orderby", "o", "", "field by which results will be ordered")
	listCrOSDevsCmd.Flags().StringVarP(&pages, "pages", "p", "", "number of pages of results to be returned")
	listCrOSDevsCmd.Flags().StringVarP(&projection, "projection", "j", "", "type of projection")
	listCrOSDevsCmd.Flags().StringVarP(&query, "query", "q", "", "selection criteria to get devices (separated by ~)")
	listCrOSDevsCmd.Flags().StringVarP(&sortOrder, "sortorder", "s", "", "sort order of returned results")
	listCrOSDevsCmd.Flags().StringVarP(&orgUnit, "orgunitpath", "t", "", "sets orgunit path that returned devices belong to")
}
