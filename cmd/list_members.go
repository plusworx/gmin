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

	cmn "github.com/plusworx/gmin/utils/common"
	flgnm "github.com/plusworx/gmin/utils/flagnames"
	gmess "github.com/plusworx/gmin/utils/gminmessages"
	gpars "github.com/plusworx/gmin/utils/gminparsers"
	lg "github.com/plusworx/gmin/utils/logging"
	mems "github.com/plusworx/gmin/utils/members"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	admin "google.golang.org/api/admin/directory/v1"
)

var listMembersCmd = &cobra.Command{
	Use:     "group-members <group email address or id>",
	Aliases: []string{"group-member", "grp-members", "grp-member", "grp-mems", "grp-mem", "gmembers", "gmember", "gmems", "gmem"},
	Args:    cobra.ExactArgs(1),
	Example: `gmin list group-members mygroup@mycompany.com -r OWNER~MANAGER
gmin ls gmems mygroup@mycompany.com -a email`,
	Short: "Outputs a list of group members",
	Long:  `Outputs a list of group members. Must specify a group email address or id.`,
	RunE:  doListMembers,
}

func doListMembers(cmd *cobra.Command, args []string) error {
	lg.Debugw("starting doListMembers()",
		"args", args)
	defer lg.Debug("finished doListMembers()")

	var (
		flagsPassed []string
		jsonData    []byte
		members     *admin.Members
	)

	flagValueMap := map[string]interface{}{}

	// Collect names of command flags passed in
	cmd.Flags().Visit(func(f *pflag.Flag) {
		flagsPassed = append(flagsPassed, f.Name)
	})

	// Populate flag value map
	for _, flg := range flagsPassed {
		val, err := mems.GetFlagVal(cmd, flg)
		if err != nil {
			return err
		}
		flagValueMap[flg] = val
	}

	srv, err := cmn.CreateService(cmn.SRVTYPEADMIN, admin.AdminDirectoryGroupMemberReadonlyScope)
	if err != nil {
		return err
	}
	ds := srv.(*admin.Service)

	mlc := ds.Members.List(args[0])

	err = lstMemProcessFlags(mlc, flagValueMap)
	if err != nil {
		return err
	}

	// If maxresults flag wasn't passed in then use the default value
	_, maxResultsPresent := flagValueMap[flgnm.FLG_MAXRESULTS]
	if !maxResultsPresent {
		err := lstMemMaxResults(mlc, int64(200))
		if err != nil {
			return err
		}
	}

	members, err = mems.DoList(mlc)
	if err != nil {
		return err
	}

	err = lstMemPages(mlc, members, flagValueMap)
	if err != nil {
		lg.Error(err)
		return err
	}

	flgCountVal, countPresent := flagValueMap[flgnm.FLG_COUNT]
	if countPresent {
		countVal := flgCountVal.(bool)
		if countVal {
			fmt.Println(len(members.Members))
			return nil
		}
	}

	jsonData, err = json.MarshalIndent(members, "", "    ")
	if err != nil {
		lg.Error(err)
		return err
	}
	fmt.Println(string(jsonData))

	return nil
}

func doMemAllPages(mlc *admin.MembersListCall, members *admin.Members) error {
	lg.Debug("starting doMemAllPages()")
	defer lg.Debug("finished doMemAllPages()")

	if members.NextPageToken != "" {
		mlc = mems.AddPageToken(mlc, members.NextPageToken)
		nxtMems, err := mems.DoList(mlc)
		if err != nil {
			return err
		}
		members.Members = append(members.Members, nxtMems.Members...)
		members.Etag = nxtMems.Etag
		members.NextPageToken = nxtMems.NextPageToken

		if nxtMems.NextPageToken != "" {
			doMemAllPages(mlc, members)
		}
	}

	return nil
}

func doMemNumPages(mlc *admin.MembersListCall, members *admin.Members, numPages int) error {
	lg.Debugw("starting doMemNumPages()",
		"numPages", numPages)
	defer lg.Debug("finished doMemNumPages()")

	if members.NextPageToken != "" && numPages > 0 {
		mlc = mems.AddPageToken(mlc, members.NextPageToken)
		nxtMems, err := mems.DoList(mlc)
		if err != nil {
			return err
		}
		members.Members = append(members.Members, nxtMems.Members...)
		members.Etag = nxtMems.Etag
		members.NextPageToken = nxtMems.NextPageToken

		if nxtMems.NextPageToken != "" {
			doMemNumPages(mlc, members, numPages-1)
		}
	}

	return nil
}

func doMemPages(mlc *admin.MembersListCall, members *admin.Members, pages string) error {
	lg.Debugw("starting doMemPages()",
		"pages", pages)
	defer lg.Debug("finished doMemPages()")

	if pages == "all" {
		err := doMemAllPages(mlc, members)
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
			err = doMemNumPages(mlc, members, numPages-1)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func init() {
	listCmd.AddCommand(listMembersCmd)

	listMembersCmd.Flags().StringP(flgnm.FLG_ATTRIBUTES, "a", "", "required member attributes (separated by ~)")
	listMembersCmd.Flags().Bool(flgnm.FLG_COUNT, false, "count number of entities returned")
	listMembersCmd.Flags().Int64P(flgnm.FLG_MAXRESULTS, "m", 200, "maximum number or results to return")
	listMembersCmd.Flags().StringP(flgnm.FLG_PAGES, "p", "", "number of pages of results to be returned ('all' or a number)")
	listMembersCmd.Flags().StringP(flgnm.FLG_ROLES, "r", "", "roles to filter results by (separated by ~)")
}

func lstMemAttributes(mlc *admin.MembersListCall, flagVal interface{}) error {
	lg.Debug("starting lstMemAttributes()")
	defer lg.Debug("finished lstMemAttributes()")

	attrsVal := fmt.Sprintf("%v", flagVal)
	if attrsVal != "" {
		listAttrs, err := gpars.ParseOutputAttrs(attrsVal, mems.MemberAttrMap)
		if err != nil {
			return err
		}
		formattedAttrs := mems.STARTMEMBERSFIELD + listAttrs + mems.ENDFIELD

		listCall := mems.AddFields(mlc, formattedAttrs)
		mlc = listCall.(*admin.MembersListCall)
	}

	return nil
}

func lstMemMaxResults(mlc *admin.MembersListCall, flagVal interface{}) error {
	lg.Debug("starting lstMemMaxResults()")
	defer lg.Debug("finished lstMemMaxResults()")

	flgMaxResultsVal := flagVal.(int64)

	mlc = mems.AddMaxResults(mlc, flgMaxResultsVal)
	return nil
}

func lstMemPages(mlc *admin.MembersListCall, members *admin.Members, flgValMap map[string]interface{}) error {
	lg.Debug("starting lstMemPages()")
	defer lg.Debug("finished lstMemPages()")

	flgPagesVal, pagesPresent := flgValMap[flgnm.FLG_PAGES]
	if !pagesPresent {
		return nil
	}
	if flgPagesVal != "" {
		pagesVal := fmt.Sprintf("%v", flgPagesVal)
		err := doMemPages(mlc, members, pagesVal)
		if err != nil {
			lg.Error(err)
			return err
		}
	}
	return nil
}

func lstMemProcessFlags(mlc *admin.MembersListCall, flgValMap map[string]interface{}) error {
	lg.Debug("starting lstMemProcessFlags()")
	defer lg.Debug("finished lstMemProcessFlags()")

	lstMemFuncMap := map[string]func(*admin.MembersListCall, interface{}) error{
		flgnm.FLG_ATTRIBUTES: lstMemAttributes,
		flgnm.FLG_MAXRESULTS: lstMemMaxResults,
		flgnm.FLG_ROLES:      lstMemRoles,
	}

	// Cycle through flags that build the mlc excluding pages and count
	for key, val := range flgValMap {
		lmf, ok := lstMemFuncMap[key]
		if !ok {
			continue
		}
		err := lmf(mlc, val)
		if err != nil {
			return err
		}
	}
	return nil
}

func lstMemRoles(mlc *admin.MembersListCall, flagVal interface{}) error {
	lg.Debug("starting lstMemRoles()")
	defer lg.Debug("finished lstMemRoles()")

	flgRolesVal := fmt.Sprintf("%v", flagVal)
	if flgRolesVal != "" {
		formattedRoles, err := gpars.ParseOutputAttrs(flgRolesVal, mems.RoleMap)
		if err != nil {
			return err
		}
		mlc = mems.AddRoles(mlc, formattedRoles)
	}
	return nil
}
