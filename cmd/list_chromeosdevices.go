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
	flgnm "github.com/plusworx/gmin/utils/flagnames"
	gmess "github.com/plusworx/gmin/utils/gminmessages"
	gpars "github.com/plusworx/gmin/utils/gminparsers"
	lg "github.com/plusworx/gmin/utils/logging"
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
	lg.Debugw("starting doListCrOSDevs()",
		"args", args)
	defer lg.Debug("finished doListCrOSDevs()")

	var (
		crosdevs *admin.ChromeOsDevices
	)

	srv, err := cmn.CreateService(cmn.SRVTYPEADMIN, admin.AdminDirectoryDeviceChromeosReadonlyScope)
	if err != nil {
		return err
	}
	ds := srv.(*admin.Service)

	customerID, err := cfg.ReadConfigString(cfg.CONFIGCUSTID)
	if err != nil {
		return err
	}

	cdlc := ds.Chromeosdevices.List(customerID)

	flgAttrsVal, err := cmd.Flags().GetString(flgnm.FLG_ATTRIBUTES)
	if err != nil {
		lg.Error(err)
		return err
	}
	if flgAttrsVal != "" {
		listAttrs, err := gpars.ParseOutputAttrs(flgAttrsVal, cdevs.CrOSDevAttrMap)
		if err != nil {
			return err
		}
		formattedAttrs := cdevs.STARTCHROMEDEVICESFIELD + listAttrs + cdevs.ENDFIELD
		listCall := cdevs.AddFields(cdlc, formattedAttrs)
		cdlc = listCall.(*admin.ChromeosdevicesListCall)
	}

	flgOrderByVal, err := cmd.Flags().GetString(flgnm.FLG_ORDERBY)
	if err != nil {
		lg.Error(err)
		return err
	}
	if flgOrderByVal != "" {
		ob := strings.ToLower(flgOrderByVal)
		ok := cmn.SliceContainsStr(cdevs.ValidOrderByStrs, ob)
		if !ok {
			err = fmt.Errorf(gmess.ERR_INVALIDORDERBY, flgOrderByVal)
			lg.Error(err)
			return err
		}

		validOrderBy, err := cmn.IsValidAttr(ob, cdevs.CrOSDevAttrMap)
		if err != nil {
			return err
		}

		cdlc = cdevs.AddOrderBy(cdlc, validOrderBy)

		flgSrtOrdVal, err := cmd.Flags().GetString(flgnm.FLG_SORTORDER)
		if err != nil {
			lg.Error(err)
			return err
		}
		if flgSrtOrdVal != "" {
			so := strings.ToLower(flgSrtOrdVal)
			validSortOrder, err := cmn.IsValidAttr(so, cmn.ValidSortOrders)
			if err != nil {
				return err
			}

			cdlc = cdevs.AddSortOrder(cdlc, validSortOrder)
		}
	}

	flgOUVal, err := cmd.Flags().GetString(flgnm.FLG_ORGUNITPATH)
	if err != nil {
		lg.Error(err)
		return err
	}
	if flgOUVal != "" {
		cdlc = cdevs.AddQuery(cdlc, flgOUVal)
	}

	flgProjectionVal, err := cmd.Flags().GetString(flgnm.FLG_PROJECTION)
	if err != nil {
		lg.Error(err)
		return err
	}
	if flgProjectionVal != "" {
		proj := strings.ToLower(flgProjectionVal)
		ok := cmn.SliceContainsStr(cdevs.ValidProjections, proj)
		if !ok {
			err = fmt.Errorf(gmess.ERR_INVALIDPROJECTIONTYPE, flgProjectionVal)
			lg.Error(err)
			return err
		}

		listCall := cdevs.AddProjection(cdlc, proj)
		cdlc = listCall.(*admin.ChromeosdevicesListCall)
	}

	flgQueryVal, err := cmd.Flags().GetString(flgnm.FLG_QUERY)
	if err != nil {
		lg.Error(err)
		return err
	}
	if flgQueryVal != "" {
		formattedQuery, err := gpars.ParseQuery(flgQueryVal, cdevs.QueryAttrMap)
		if err != nil {
			return err
		}

		cdlc = cdevs.AddQuery(cdlc, formattedQuery)
	}

	flgMaxResultsVal, err := cmd.Flags().GetInt64(flgnm.FLG_MAXRESULTS)
	if err != nil {
		lg.Error(err)
		return err
	}
	cdlc = cdevs.AddMaxResults(cdlc, flgMaxResultsVal)

	crosdevs, err = cdevs.DoList(cdlc)
	if err != nil {
		return err
	}

	flgPagesVal, err := cmd.Flags().GetString(flgnm.FLG_PAGES)
	if err != nil {
		lg.Error(err)
		return err
	}
	if flgPagesVal != "" {
		err = doCrOSDevPages(cdlc, crosdevs, flgPagesVal)
		if err != nil {
			return err
		}
	}

	jsonData, err := json.MarshalIndent(crosdevs, "", "    ")
	if err != nil {
		lg.Error(err)
		return err
	}

	flgCountVal, err := cmd.Flags().GetBool(flgnm.FLG_COUNT)
	if err != nil {
		lg.Error(err)
		return err
	}
	if flgCountVal {
		fmt.Println(len(crosdevs.Chromeosdevices))
	} else {
		fmt.Println(string(jsonData))
	}

	return nil
}

func doCrOSDevAllPages(cdlc *admin.ChromeosdevicesListCall, crosdevs *admin.ChromeOsDevices) error {
	lg.Debug("starting doCrOSDevAllPages()")
	defer lg.Debug("finished doCrOSDevAllPages()")

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
	lg.Debugw("starting doCrOSDevNumPages()",
		"numPages", numPages)
	defer lg.Debug("finished doCrOSDevNumPages()")

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
	lg.Debugw("starting doCrOSDevPages()",
		"pages", pages)
	defer lg.Debug("finished doCrOSDevPages()")

	if pages == "all" {
		err := doCrOSDevAllPages(cdlc, crosdevs)
		if err != nil {
			return err
		}
	} else {
		numPages, err := strconv.Atoi(pages)
		if err != nil {
			err = errors.New(gmess.ERR_INVALIDPAGESARGUMENT)
			lg.Error(err)
			return err
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

	listCrOSDevsCmd.Flags().StringP(flgnm.FLG_ATTRIBUTES, "a", "", "required device attributes (separated by ~)")
	listCrOSDevsCmd.Flags().Bool(flgnm.FLG_COUNT, false, "count number of entities returned")
	listCrOSDevsCmd.Flags().Int64P(flgnm.FLG_MAXRESULTS, "m", 200, "maximum number of results to return per page")
	listCrOSDevsCmd.Flags().StringP(flgnm.FLG_ORDERBY, "o", "", "field by which results will be ordered")
	listCrOSDevsCmd.Flags().StringP(flgnm.FLG_PAGES, "p", "", "number of pages of results to be returned ('all' or a number)")
	listCrOSDevsCmd.Flags().StringP(flgnm.FLG_PROJECTION, "j", "", "type of projection")
	listCrOSDevsCmd.Flags().StringP(flgnm.FLG_QUERY, "q", "", "selection criteria to get devices (separated by ~)")
	listCrOSDevsCmd.Flags().StringP(flgnm.FLG_SORTORDER, "s", "", "sort order of returned results")
	listCrOSDevsCmd.Flags().StringP(flgnm.FLG_ORGUNITPATH, "t", "", "sets orgunit path that returned devices belong to")
}
