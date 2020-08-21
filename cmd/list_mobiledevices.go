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

	cmn "github.com/plusworx/gmin/utils/common"
	cfg "github.com/plusworx/gmin/utils/config"
	mdevs "github.com/plusworx/gmin/utils/mobiledevices"
	"github.com/spf13/cobra"
	admin "google.golang.org/api/admin/directory/v1"
)

var listMobDevsCmd = &cobra.Command{
	Use:     "mobiledevices",
	Aliases: []string{"mobiledevice", "mobdevices", "mobdevice", "mobdevs", "mobdev", "mdevs", "mdev"},
	Args:    cobra.NoArgs,
	Short:   "Outputs a list of mobile devices",
	Long: `Outputs a list of mobile devices.
	
	Examples:	gmin list mobiledevices --pages all --count
			gmin ls mdevs --pages all`,
	RunE: doListMobDevs,
}

func doListMobDevs(cmd *cobra.Command, args []string) error {
	var (
		mobdevs *admin.MobileDevices
	)

	ds, err := cmn.CreateDirectoryService(admin.AdminDirectoryDeviceMobileReadonlyScope)
	if err != nil {
		return err
	}

	customerID, err := cfg.ReadConfigString("customerid")
	if err != nil {
		return err
	}

	mdlc := ds.Mobiledevices.List(customerID)

	if attrs != "" {
		listAttrs, err := cmn.ParseOutputAttrs(attrs, mdevs.MobDevAttrMap)
		if err != nil {
			return err
		}
		formattedAttrs := mdevs.StartMobDevicesField + listAttrs + mdevs.EndField
		listCall := mdevs.AddFields(mdlc, formattedAttrs)
		mdlc = listCall.(*admin.MobiledevicesListCall)
	}

	if orderBy != "" {
		ob := strings.ToLower(orderBy)
		ok := cmn.SliceContainsStr(mdevs.ValidOrderByStrs, ob)
		if !ok {
			err = fmt.Errorf("gmin: error - %v is not a valid order by field", orderBy)
			return err
		}

		validOrderBy, err := cmn.IsValidAttr(ob, mdevs.MobDevAttrMap)
		if err != nil {
			return err
		}

		mdlc = mdevs.AddOrderBy(mdlc, validOrderBy)

		if sortOrder != "" {
			so := strings.ToLower(sortOrder)
			validSortOrder, err := cmn.IsValidAttr(so, cmn.ValidSortOrders)
			if err != nil {
				return err
			}

			mdlc = mdevs.AddSortOrder(mdlc, validSortOrder)
		}
	}

	if orgUnit != "" {
		mdlc = mdevs.AddQuery(mdlc, orgUnit)
	}

	if projection != "" {
		proj := strings.ToLower(projection)
		ok := cmn.SliceContainsStr(mdevs.ValidProjections, proj)
		if !ok {
			return fmt.Errorf("gmin: error - %v is not a valid projection type", projection)
		}

		listCall := mdevs.AddProjection(mdlc, proj)
		mdlc = listCall.(*admin.MobiledevicesListCall)
	}

	if query != "" {
		formattedQuery, err := cmn.ParseQuery(query, mdevs.QueryAttrMap)
		if err != nil {
			return err
		}

		mdlc = mdevs.AddQuery(mdlc, formattedQuery)
	}

	mdlc = mdevs.AddMaxResults(mdlc, maxResults)

	mobdevs, err = mdevs.DoList(mdlc)
	if err != nil {
		return err
	}

	if pages != "" {
		err = doMobDevPages(mdlc, mobdevs, pages)
		if err != nil {
			return err
		}
	}

	jsonData, err := json.MarshalIndent(mobdevs, "", "    ")
	if err != nil {
		return err
	}

	if count {
		fmt.Println(len(mobdevs.Mobiledevices))
	} else {
		fmt.Println(string(jsonData))
	}

	return nil
}

func doMobDevAllPages(mdlc *admin.MobiledevicesListCall, mobdevs *admin.MobileDevices) error {
	if mobdevs.NextPageToken != "" {
		mdlc = mdevs.AddPageToken(mdlc, mobdevs.NextPageToken)
		nxtMobDevs, err := mdevs.DoList(mdlc)
		if err != nil {
			return err
		}
		mobdevs.Mobiledevices = append(mobdevs.Mobiledevices, nxtMobDevs.Mobiledevices...)
		mobdevs.Etag = nxtMobDevs.Etag
		mobdevs.NextPageToken = nxtMobDevs.NextPageToken

		if nxtMobDevs.NextPageToken != "" {
			doMobDevAllPages(mdlc, mobdevs)
		}
	}

	return nil
}

func doMobDevNumPages(mdlc *admin.MobiledevicesListCall, mobdevs *admin.MobileDevices, numPages int) error {
	if mobdevs.NextPageToken != "" && numPages > 0 {
		mdlc = mdevs.AddPageToken(mdlc, mobdevs.NextPageToken)
		nxtMobDevs, err := mdevs.DoList(mdlc)
		if err != nil {
			return err
		}
		mobdevs.Mobiledevices = append(mobdevs.Mobiledevices, nxtMobDevs.Mobiledevices...)
		mobdevs.Etag = nxtMobDevs.Etag
		mobdevs.NextPageToken = nxtMobDevs.NextPageToken

		if nxtMobDevs.NextPageToken != "" {
			doMobDevNumPages(mdlc, mobdevs, numPages-1)
		}
	}

	return nil
}

func doMobDevPages(mdlc *admin.MobiledevicesListCall, mobdevs *admin.MobileDevices, pages string) error {
	if pages == "all" {
		err := doMobDevAllPages(mdlc, mobdevs)
		if err != nil {
			return err
		}
	} else {
		numPages, err := strconv.Atoi(pages)
		if err != nil {
			return errors.New("gmin: error - pages must be 'all' or a number")
		}

		if numPages > 1 {
			err = doMobDevNumPages(mdlc, mobdevs, numPages-1)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func init() {
	listCmd.AddCommand(listMobDevsCmd)

	listMobDevsCmd.Flags().StringVarP(&attrs, "attributes", "a", "", "required device attributes (separated by ~)")
	listMobDevsCmd.Flags().BoolVarP(&count, "count", "", false, "count number of entities returned")
	listMobDevsCmd.Flags().Int64VarP(&maxResults, "maxresults", "m", 100, "maximum number of results to return per page")
	listMobDevsCmd.Flags().StringVarP(&orderBy, "orderby", "o", "", "field by which results will be ordered")
	listMobDevsCmd.Flags().StringVarP(&pages, "pages", "p", "", "number of pages of results to be returned ('all' or a number)")
	listMobDevsCmd.Flags().StringVarP(&projection, "projection", "j", "", "type of projection")
	listMobDevsCmd.Flags().StringVarP(&query, "query", "q", "", "selection criteria to get devices (separated by ~)")
	listMobDevsCmd.Flags().StringVarP(&sortOrder, "sortorder", "s", "", "sort order of returned results")
}
