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
	flgnm "github.com/plusworx/gmin/utils/flagnames"
	gmess "github.com/plusworx/gmin/utils/gminmessages"
	grps "github.com/plusworx/gmin/utils/groups"
	lg "github.com/plusworx/gmin/utils/logging"
	"github.com/spf13/cobra"
	admin "google.golang.org/api/admin/directory/v1"
)

var batchCrtGroupCmd = &cobra.Command{
	Use:     "groups -i <input file path or google sheet id>",
	Aliases: []string{"group", "grps", "grp"},
	Example: `gmin batch-create groups -i inputfile.json
gmin bcrt grps -i inputfile.csv -f csv
gmin bcrt grp -i 1odyAIp3jGspd3M4xeepxWD6aeQIUuHBgrZB2OHSu8MI -s 'Sheet1!A1:K25' -f gsheet`,
	Short: "Creates a batch of groups",
	Long: `Creates a batch of groups where group details are provided in a Google Sheet, CSV/JSON input file or piped JSON.
				  
The contents of the JSON file or piped input should look something like this:
	
{"description":"Finance group","email":"finance@mycompany.com","name":"Finance"}
{"description":"Marketing group","email":"marketing@mycompany.com","name":"Marketing"}
{"description":"Sales group","email":"sales@mycompany.com","name":"Sales"}
	
CSV and Google sheets must have a header row with the following column names being the only ones that are valid:
	
description
email [required]
name

The column names are case insensitive and can be in any order.`,
	RunE: doBatchCrtGroup,
}

func doBatchCrtGroup(cmd *cobra.Command, args []string) error {
	lg.Debugw("starting doBatchCrtGroup()",
		"args", args)
	defer lg.Debug("finished doBatchCrtGroup()")

	var (
		groups []*admin.Group
		objs   []interface{}
	)

	srv, err := cmn.CreateService(cmn.SRVTYPEADMIN, admin.AdminDirectoryGroupScope)
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

	callParams := btch.CallParams{CallType: cmn.CALLTYPECREATE, ObjectType: cmn.OBJTYPEGROUP}

	switch {
	case lwrFmt == "csv":
		objs, err = btch.ProcessCSVFile(callParams, inputFlgVal, grps.GroupAttrMap)
		if err != nil {
			return err
		}
	case lwrFmt == "json":
		objs, err = btch.ProcessJSON(callParams, inputFlgVal, scanner, grps.GroupAttrMap)
		if err != nil {
			return err
		}
	case lwrFmt == "gsheet":
		rangeFlgVal, err := cmd.Flags().GetString(flgnm.FLG_SHEETRANGE)
		if err != nil {
			return err
		}

		objs, err = btch.ProcessGSheet(callParams, inputFlgVal, rangeFlgVal, grps.GroupAttrMap)
		if err != nil {
			return err
		}
	default:
		err = fmt.Errorf(gmess.ERR_INVALIDFILEFORMAT, formatFlgVal)
		lg.Error(err)
		return err
	}

	for _, grpObj := range objs {
		groups = append(groups, grpObj.(*admin.Group))
	}

	err = bcgProcessObjects(ds, groups)
	if err != nil {
		return err
	}

	return nil
}

func bcgCreate(group *admin.Group, wg *sync.WaitGroup, gic *admin.GroupsInsertCall) {
	lg.Debug("starting bcgCreate()")
	defer lg.Debug("finished bcgCreate()")

	defer wg.Done()

	b := backoff.NewExponentialBackOff()
	b.MaxElapsedTime = 32 * time.Second

	err := backoff.Retry(func() error {
		var err error
		newGroup, err := gic.Do()
		if err == nil {
			fmt.Println(cmn.GminMessage(fmt.Sprintf(gmess.INFO_GROUPCREATED, newGroup.Email)))
			lg.Infof(gmess.INFO_GROUPCREATED, newGroup.Email)
			return err
		}
		if !cmn.IsErrRetryable(err) {
			return backoff.Permanent(fmt.Errorf(gmess.ERR_BATCHGROUP, err.Error(), group.Email))
		}
		// Log the retries
		lg.Warnw(err.Error(),
			"retrying", b.GetElapsedTime().String(),
			"group", group.Email)
		return fmt.Errorf(gmess.ERR_BATCHGROUP, err.Error(), group.Email)
	}, b)
	if err != nil {
		// Log final error
		lg.Error(err)
		fmt.Println(cmn.GminMessage(err.Error()))
	}
}

func bcgProcessObjects(ds *admin.Service, groups []*admin.Group) error {
	lg.Debug("starting bcgProcessObjects()")
	defer lg.Debug("finished bcgProcessObjects()")

	wg := new(sync.WaitGroup)

	for _, g := range groups {
		if g.Email == "" {
			err := errors.New(gmess.ERR_NOGROUPEMAILADDRESS)
			lg.Error(err)
			return err
		}

		gic := ds.Groups.Insert(g)

		wg.Add(1)

		go bcgCreate(g, wg, gic)
	}

	wg.Wait()

	return nil
}

func init() {
	batchCreateCmd.AddCommand(batchCrtGroupCmd)

	batchCrtGroupCmd.Flags().StringP(flgnm.FLG_INPUTFILE, "i", "", "filepath to group data file or sheet id")
	batchCrtGroupCmd.Flags().StringP(flgnm.FLG_FORMAT, "f", "json", "user data file format")
	batchCrtGroupCmd.Flags().StringP(flgnm.FLG_SHEETRANGE, "s", "", "user data gsheet range")
}
