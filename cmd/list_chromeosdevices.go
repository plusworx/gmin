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
	gd "github.com/plusworx/gmin/utils/gendatastructs"
	gmess "github.com/plusworx/gmin/utils/gminmessages"
	gpars "github.com/plusworx/gmin/utils/gminparsers"
	lg "github.com/plusworx/gmin/utils/logging"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
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
		crosdevs    *admin.ChromeOsDevices
		flagsPassed []string
	)

	flagValueMap := map[string]interface{}{}

	// Collect names of command flags passed in
	cmd.Flags().Visit(func(f *pflag.Flag) {
		flagsPassed = append(flagsPassed, f.Name)
	})

	// Populate flag value map
	for _, flg := range flagsPassed {
		val, err := cdevs.GetFlagVal(cmd, flg)
		if err != nil {
			return err
		}
		flagValueMap[flg] = val
	}

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

	err = lstCrOSDevProcessFlags(cdlc, flagValueMap)
	if err != nil {
		return err
	}

	// If maxresults flag wasn't passed in then use the default value
	_, maxResultsPresent := flagValueMap[flgnm.FLG_MAXRESULTS]
	if !maxResultsPresent {
		err := lstCrOSDevMaxResults(cdlc, int64(200))
		if err != nil {
			return err
		}
	}

	crosdevs, err = cdevs.DoList(cdlc)
	if err != nil {
		return err
	}

	err = lstCrOSDevPages(cdlc, crosdevs, flagValueMap)
	if err != nil {
		lg.Error(err)
		return err
	}

	flgCountVal, countPresent := flagValueMap[flgnm.FLG_COUNT]
	if countPresent {
		countVal := flgCountVal.(bool)
		if countVal {
			fmt.Println(len(crosdevs.Chromeosdevices))
			return nil
		}
	}

	jsonData, err := json.MarshalIndent(crosdevs, "", "    ")
	if err != nil {
		lg.Error(err)
		return err
	}
	fmt.Println(string(jsonData))

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

func lstCrOSDevAttributes(cdlc *admin.ChromeosdevicesListCall, flagVal interface{}) error {
	lg.Debug("starting lstCrOSDevAttributes()")
	defer lg.Debug("finished lstCrOSDevAttributes()")

	attrsVal := fmt.Sprintf("%v", flagVal)
	if attrsVal != "" {
		listAttrs, err := gpars.ParseOutputAttrs(attrsVal, cdevs.CrOSDevAttrMap)
		if err != nil {
			return err
		}
		formattedAttrs := cdevs.STARTCHROMEDEVICESFIELD + listAttrs + cdevs.ENDFIELD

		listCall := cdevs.AddFields(cdlc, formattedAttrs)
		cdlc = listCall.(*admin.ChromeosdevicesListCall)
	}

	return nil
}

func lstCrOSDevMaxResults(cdlc *admin.ChromeosdevicesListCall, flagVal interface{}) error {
	lg.Debug("starting lstCrOSDevMaxResults()")
	defer lg.Debug("finished lstCrOSDevMaxResults()")

	flgMaxResultsVal := flagVal.(int64)

	cdlc = cdevs.AddMaxResults(cdlc, flgMaxResultsVal)
	return nil
}

func lstCrOSDevOrderBy(cdlc *admin.ChromeosdevicesListCall, inData interface{}) error {
	lg.Debug("starting lstCrOSDevOrderBy()")
	defer lg.Debug("finished lstCrOSDevOrderBy()")

	var (
		err          error
		validOrderBy string
	)

	inStruct := inData.(gd.TwoStrStruct)

	orderVal := inStruct.Element1
	if orderVal != "" {
		ob := strings.ToLower(orderVal)
		ok := cmn.SliceContainsStr(cdevs.ValidOrderByStrs, ob)
		if !ok {
			err := fmt.Errorf(gmess.ERR_INVALIDORDERBY, orderVal)
			lg.Error(err)
			return err
		}

		validOrderBy, err = cmn.IsValidAttr(ob, cdevs.CrOSDevAttrMap)
		if err != nil {
			lg.Error(err)
			return err
		}
		cdlc = cdevs.AddOrderBy(cdlc, validOrderBy)

		flgSrtOrdByVal := inStruct.Element2
		if flgSrtOrdByVal != "" {
			so := strings.ToLower(flgSrtOrdByVal)
			validSortOrder, err := cmn.IsValidAttr(so, cmn.ValidSortOrders)
			if err != nil {
				lg.Error(err)
				return err
			}

			cdlc = cdevs.AddSortOrder(cdlc, validSortOrder)
		}
	}
	return nil
}

func lstCrOSDevPages(cdlc *admin.ChromeosdevicesListCall, crosdevs *admin.ChromeOsDevices, flgValMap map[string]interface{}) error {
	lg.Debug("starting lstCrOSDevPages()")
	defer lg.Debug("finished lstCrOSDevPages()")

	flgPagesVal, pagesPresent := flgValMap[flgnm.FLG_PAGES]
	if !pagesPresent {
		return nil
	}
	if flgPagesVal != "" {
		pagesVal := fmt.Sprintf("%v", flgPagesVal)
		err := doCrOSDevPages(cdlc, crosdevs, pagesVal)
		if err != nil {
			lg.Error(err)
			return err
		}
	}
	return nil
}

func lstCrOSDevOUPath(cdlc *admin.ChromeosdevicesListCall, flagVal interface{}) error {
	flgOUVal := fmt.Sprintf("%v", flagVal)
	if flgOUVal != "" {
		cdlc = cdevs.AddQuery(cdlc, flgOUVal)
	}
	return nil
}

func lstCrOSDevProcessFlags(cdlc *admin.ChromeosdevicesListCall, flgValMap map[string]interface{}) error {
	lg.Debug("starting lstCrOSDevProcessFlags()")
	defer lg.Debug("finished lstCrOSDevProcessFlags()")

	lstCrOSDevFuncMap := map[string]func(*admin.ChromeosdevicesListCall, interface{}) error{
		flgnm.FLG_ATTRIBUTES:  lstCrOSDevAttributes,
		flgnm.FLG_MAXRESULTS:  lstCrOSDevMaxResults,
		flgnm.FLG_ORDERBY:     lstCrOSDevOrderBy,
		flgnm.FLG_ORGUNITPATH: lstCrOSDevOUPath,
		flgnm.FLG_PROJECTION:  lstCrOSDevProjection,
		flgnm.FLG_QUERY:       lstCrOSDevQuery,
	}

	// Cycle through flags that build the mdlc excluding pages and count
	for key, val := range flgValMap {
		// Order by has dependent sort order so deal with that
		if key == flgnm.FLG_ORDERBY {
			retStruct, err := lstCrOSDevSortOrderBy(flgValMap)
			if err != nil {
				return err
			}
			val = retStruct
		}

		lcdf, ok := lstCrOSDevFuncMap[key]
		if !ok {
			continue
		}
		err := lcdf(cdlc, val)
		if err != nil {
			return err
		}
	}
	return nil
}

func lstCrOSDevProjection(cdlc *admin.ChromeosdevicesListCall, flagVal interface{}) error {
	flgProjectionVal := fmt.Sprintf("%v", flagVal)
	if flgProjectionVal != "" {
		proj := strings.ToLower(flgProjectionVal)
		ok := cmn.SliceContainsStr(cdevs.ValidProjections, proj)
		if !ok {
			err := fmt.Errorf(gmess.ERR_INVALIDPROJECTIONTYPE, flgProjectionVal)
			lg.Error(err)
			return err
		}

		listCall := cdevs.AddProjection(cdlc, proj)
		cdlc = listCall.(*admin.ChromeosdevicesListCall)
	}
	return nil
}

func lstCrOSDevQuery(cdlc *admin.ChromeosdevicesListCall, flagVal interface{}) error {
	lg.Debug("starting lstCrOSDevQuery()")
	defer lg.Debug("finished lstCrOSDevQuery()")

	qryVal := fmt.Sprintf("%v", flagVal)
	if qryVal != "" {
		formattedQuery, err := gpars.ParseQuery(qryVal, cdevs.QueryAttrMap)
		if err != nil {
			lg.Error(err)
			return err
		}

		cdlc = cdevs.AddQuery(cdlc, formattedQuery)
	}
	return nil
}

func lstCrOSDevSortOrderBy(flgValMap map[string]interface{}) (gd.TwoStrStruct, error) {
	lg.Debug("starting lstCrOSDevSortOrderBy()")
	defer lg.Debug("finished lstCrOSDevSortOrderBy()")

	outData := gd.TwoStrStruct{}

	orderByVal := flgValMap[flgnm.FLG_ORDERBY]
	sortOrderVal, sortOrderPresent := flgValMap[flgnm.FLG_SORTORDER]

	outData.Element1 = fmt.Sprintf("%v", orderByVal)
	if sortOrderPresent {
		outData.Element2 = fmt.Sprintf("%v", sortOrderVal)
	} else {
		outData.Element2 = ""
	}

	return outData, nil
}
