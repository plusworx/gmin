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
	flgnm "github.com/plusworx/gmin/utils/flagnames"
	gd "github.com/plusworx/gmin/utils/gendatastructs"
	gmess "github.com/plusworx/gmin/utils/gminmessages"
	gpars "github.com/plusworx/gmin/utils/gminparsers"
	lg "github.com/plusworx/gmin/utils/logging"
	mdevs "github.com/plusworx/gmin/utils/mobiledevices"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	admin "google.golang.org/api/admin/directory/v1"
)

var listMobDevsCmd = &cobra.Command{
	Use:     "mobile-devices",
	Aliases: []string{"mobile-device", "mob-devices", "mob-device", "mob-devs", "mob-dev", "mdevs", "mdev"},
	Args:    cobra.NoArgs,
	Example: `gmin list mobile-devices --pages all --count
gmin ls mdevs --pages all`,
	Short: "Outputs a list of mobile devices",
	Long:  `Outputs a list of mobile devices.`,
	RunE:  doListMobDevs,
}

func doListMobDevs(cmd *cobra.Command, args []string) error {
	lg.Debugw("starting doListMobDevs()",
		"args", args)
	defer lg.Debug("finished doListMobDevs()")

	var (
		flagsPassed []string
		mobdevs     *admin.MobileDevices
	)

	flagValueMap := map[string]interface{}{}

	// Collect names of command flags passed in
	cmd.Flags().Visit(func(f *pflag.Flag) {
		flagsPassed = append(flagsPassed, f.Name)
	})

	// Populate flag value map
	for _, flg := range flagsPassed {
		val, err := mdevs.GetFlagVal(cmd, flg)
		if err != nil {
			return err
		}
		flagValueMap[flg] = val
	}

	srv, err := cmn.CreateService(cmn.SRVTYPEADMIN, admin.AdminDirectoryDeviceMobileReadonlyScope)
	if err != nil {
		return err
	}
	ds := srv.(*admin.Service)

	customerID, err := cfg.ReadConfigString(cfg.CONFIGCUSTID)
	if err != nil {
		return err
	}

	mdlc := ds.Mobiledevices.List(customerID)

	err = lstMobDevProcessFlags(mdlc, flagValueMap)
	if err != nil {
		return err
	}

	// If maxresults flag wasn't passed in then use the default value
	_, maxResultsPresent := flagValueMap[flgnm.FLG_MAXRESULTS]
	if !maxResultsPresent {
		err := lstMobDevMaxResults(mdlc, int64(100))
		if err != nil {
			return err
		}
	}

	mobdevs, err = mdevs.DoList(mdlc)
	if err != nil {
		return err
	}

	err = lstMobDevPages(mdlc, mobdevs, flagValueMap)
	if err != nil {
		lg.Error(err)
		return err
	}

	flgCountVal, countPresent := flagValueMap[flgnm.FLG_COUNT]
	if countPresent {
		countVal := flgCountVal.(bool)
		if countVal {
			fmt.Println(len(mobdevs.Mobiledevices))
			return nil
		}
	}

	jsonData, err := json.MarshalIndent(mobdevs, "", "    ")
	if err != nil {
		lg.Error(err)
		return err
	}
	fmt.Println(string(jsonData))

	return nil
}

func doMobDevAllPages(mdlc *admin.MobiledevicesListCall, mobdevs *admin.MobileDevices) error {
	lg.Debug("starting doMobDevAllPages()")
	defer lg.Debug("finished doMobDevAllPages()")

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
	lg.Debugw("starting doMobDevNumPages()",
		"numPages", numPages)
	defer lg.Debug("finished doMobDevNumPages()")

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
	lg.Debugw("starting doMobDevPages()",
		"pages", pages)
	defer lg.Debug("finished doMobDevPages()")

	if pages == "all" {
		err := doMobDevAllPages(mdlc, mobdevs)
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

	listMobDevsCmd.Flags().StringP(flgnm.FLG_ATTRIBUTES, "a", "", "required device attributes (separated by ~)")
	listMobDevsCmd.Flags().Bool(flgnm.FLG_COUNT, false, "count number of entities returned")
	listMobDevsCmd.Flags().Int64P(flgnm.FLG_MAXRESULTS, "m", 100, "maximum number of results to return per page")
	listMobDevsCmd.Flags().StringP(flgnm.FLG_ORDERBY, "o", "", "field by which results will be ordered")
	listMobDevsCmd.Flags().StringP(flgnm.FLG_PAGES, "p", "", "number of pages of results to be returned ('all' or a number)")
	listMobDevsCmd.Flags().StringP(flgnm.FLG_PROJECTION, "j", "", "type of projection")
	listMobDevsCmd.Flags().StringP(flgnm.FLG_QUERY, "q", "", "selection criteria to get devices (separated by ~)")
	listMobDevsCmd.Flags().StringP(flgnm.FLG_SORTORDER, "s", "", "sort order of returned results")
}

func lstMobDevAttributes(mdlc *admin.MobiledevicesListCall, flagVal interface{}) error {
	lg.Debug("starting lstMobDevAttributes()")
	defer lg.Debug("finished lstMobDevAttributes()")

	attrsVal := fmt.Sprintf("%v", flagVal)
	if attrsVal != "" {
		listAttrs, err := gpars.ParseOutputAttrs(attrsVal, mdevs.MobDevAttrMap)
		if err != nil {
			return err
		}
		formattedAttrs := mdevs.STARTMOBDEVICESFIELD + listAttrs + mdevs.ENDFIELD

		listCall := mdevs.AddFields(mdlc, formattedAttrs)
		mdlc = listCall.(*admin.MobiledevicesListCall)
	}

	return nil
}

func lstMobDevMaxResults(mdlc *admin.MobiledevicesListCall, flagVal interface{}) error {
	lg.Debug("starting lstMobDevMaxResults()")
	defer lg.Debug("finished lstMobDevMaxResults()")

	flgMaxResultsVal := flagVal.(int64)

	mdlc = mdevs.AddMaxResults(mdlc, flgMaxResultsVal)
	return nil
}

func lstMobDevOrderBy(mdlc *admin.MobiledevicesListCall, inData interface{}) error {
	lg.Debug("starting lstMobDevOrderBy()")
	defer lg.Debug("finished lstMobDevOrderBy()")

	var (
		err          error
		validOrderBy string
	)

	inStruct := inData.(gd.TwoStrStruct)

	orderVal := inStruct.Element1
	if orderVal != "" {
		ob := strings.ToLower(orderVal)
		ok := cmn.SliceContainsStr(mdevs.ValidOrderByStrs, ob)
		if !ok {
			err := fmt.Errorf(gmess.ERR_INVALIDORDERBY, orderVal)
			lg.Error(err)
			return err
		}

		validOrderBy, err = cmn.IsValidAttr(ob, mdevs.MobDevAttrMap)
		if err != nil {
			lg.Error(err)
			return err
		}
		mdlc = mdevs.AddOrderBy(mdlc, validOrderBy)

		flgSrtOrdByVal := inStruct.Element2
		if flgSrtOrdByVal != "" {
			so := strings.ToLower(flgSrtOrdByVal)
			validSortOrder, err := cmn.IsValidAttr(so, cmn.ValidSortOrders)
			if err != nil {
				lg.Error(err)
				return err
			}

			mdlc = mdevs.AddSortOrder(mdlc, validSortOrder)
		}
	}
	return nil
}

func lstMobDevPages(mdlc *admin.MobiledevicesListCall, mobdevs *admin.MobileDevices, flgValMap map[string]interface{}) error {
	lg.Debug("starting lstMobDevPages()")
	defer lg.Debug("finished lstMobDevPages()")

	flgPagesVal, pagesPresent := flgValMap[flgnm.FLG_PAGES]
	if !pagesPresent {
		return nil
	}
	if flgPagesVal != "" {
		pagesVal := fmt.Sprintf("%v", flgPagesVal)
		err := doMobDevPages(mdlc, mobdevs, pagesVal)
		if err != nil {
			lg.Error(err)
			return err
		}
	}
	return nil
}

func lstMobDevProcessFlags(mdlc *admin.MobiledevicesListCall, flgValMap map[string]interface{}) error {
	lg.Debug("starting lstMobDevProcessFlags()")
	defer lg.Debug("finished lstMobDevProcessFlags()")

	lstMobDevFuncMap := map[string]func(*admin.MobiledevicesListCall, interface{}) error{
		flgnm.FLG_ATTRIBUTES: lstMobDevAttributes,
		flgnm.FLG_MAXRESULTS: lstMobDevMaxResults,
		flgnm.FLG_ORDERBY:    lstMobDevOrderBy,
		flgnm.FLG_PROJECTION: lstMobDevProjection,
		flgnm.FLG_QUERY:      lstMobDevQuery,
	}

	// Cycle through flags that build the mdlc excluding pages and count
	for key, val := range flgValMap {
		// Order by has dependent sort order so deal with that
		if key == flgnm.FLG_ORDERBY {
			retStruct, err := lstMobDevSortOrderBy(flgValMap)
			if err != nil {
				return err
			}
			val = retStruct
		}

		lmdf, ok := lstMobDevFuncMap[key]
		if !ok {
			continue
		}
		err := lmdf(mdlc, val)
		if err != nil {
			return err
		}
	}
	return nil
}

func lstMobDevProjection(mdlc *admin.MobiledevicesListCall, flagVal interface{}) error {
	flgProjectionVal := fmt.Sprintf("%v", flagVal)
	if flgProjectionVal != "" {
		proj := strings.ToLower(flgProjectionVal)
		ok := cmn.SliceContainsStr(mdevs.ValidProjections, proj)
		if !ok {
			err := fmt.Errorf(gmess.ERR_INVALIDPROJECTIONTYPE, flgProjectionVal)
			lg.Error(err)
			return err
		}

		listCall := mdevs.AddProjection(mdlc, proj)
		mdlc = listCall.(*admin.MobiledevicesListCall)
	}
	return nil
}

func lstMobDevQuery(mdlc *admin.MobiledevicesListCall, flagVal interface{}) error {
	lg.Debug("starting lstMobDevQuery()")
	defer lg.Debug("finished lstMobDevQuery()")

	qryVal := fmt.Sprintf("%v", flagVal)
	if qryVal != "" {
		formattedQuery, err := gpars.ParseQuery(qryVal, mdevs.QueryAttrMap)
		if err != nil {
			lg.Error(err)
			return err
		}

		mdlc = mdevs.AddQuery(mdlc, formattedQuery)
	}
	return nil
}

func lstMobDevSortOrderBy(flgValMap map[string]interface{}) (gd.TwoStrStruct, error) {
	lg.Debug("starting lstMobDevSortOrderBy()")
	defer lg.Debug("finished lstMobDevSortOrderBy()")

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
