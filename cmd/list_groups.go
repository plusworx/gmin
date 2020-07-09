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
	"fmt"

	cmn "github.com/plusworx/gmin/utils/common"
	grps "github.com/plusworx/gmin/utils/groups"
	"github.com/spf13/cobra"
	admin "google.golang.org/api/admin/directory/v1"
)

var listGroupsCmd = &cobra.Command{
	Use:     "groups",
	Aliases: []string{"group", "grp", "grps"},
	Short:   "Outputs a list of groups",
	Long:    `Outputs a list of groups.`,
	RunE:    doListGroups,
}

func doListGroups(cmd *cobra.Command, args []string) error {
	var (
		formattedAttrs string
		groups         *admin.Groups
		validAttrs     []string
	)

	ds, err := cmn.CreateDirectoryService(admin.AdminDirectoryGroupReadonlyScope)
	if err != nil {
		return err
	}

	glc := ds.Groups.List()

	if attrs != "" {
		validAttrs, err = cmn.ValidateAttrs(attrs, grps.GroupAttrMap)
		if err != nil {
			return err
		}

		formattedAttrs = grps.FormatAttrs(validAttrs, false)
	}

	if domain != "" {
		groups, err = groupDomainCall(domain, glc, formattedAttrs)
		if err != nil {
			return err
		}
	} else {
		groups, err = groupAllDomainCall(glc, formattedAttrs)
		if err != nil {
			return err
		}
	}

	jsonData, err := json.MarshalIndent(groups, "", "    ")
	if err != nil {
		return err
	}

	fmt.Println(string(jsonData))

	return nil
}

func groupAllDomainCall(glc *admin.GroupsListCall, fmtAttrs string) (*admin.Groups, error) {
	var (
		err            error
		formattedQuery string
		groups         *admin.Groups
	)

	if query != "" {
		formattedQuery, err = processQuery(query)
		if err != nil {
			return nil, err
		}
	}

	switch true {
	case formattedQuery == "" && attrs == "":
		groups, err = grps.AllDomain(glc)
	case formattedQuery != "" && attrs == "":
		groups, err = grps.AllDomainQuery(glc, formattedQuery)
	case formattedQuery == "" && attrs != "":
		groups, err = grps.AllDomainAttrs(glc, fmtAttrs)
	case formattedQuery != "" && attrs != "":
		groups, err = grps.AllDomainQueryAttrs(glc, formattedQuery, fmtAttrs)
	}

	return groups, err
}

func groupDomainCall(domain string, glc *admin.GroupsListCall, fmtAttrs string) (*admin.Groups, error) {
	var (
		err            error
		formattedQuery string
		groups         *admin.Groups
	)

	if query != "" {
		formattedQuery, err = processQuery(query)
		if err != nil {
			return nil, err
		}
	}

	switch true {
	case formattedQuery == "" && attrs == "":
		groups, err = grps.Domain(domain, glc)
	case formattedQuery != "" && attrs == "":
		groups, err = grps.DomainQuery(domain, glc, formattedQuery)
	case formattedQuery == "" && attrs != "":
		groups, err = grps.DomainAttrs(domain, glc, fmtAttrs)
	case formattedQuery != "" && attrs != "":
		groups, err = grps.DomainQueryAttrs(domain, glc, formattedQuery, fmtAttrs)
	}

	return groups, err
}

func init() {
	listCmd.AddCommand(listGroupsCmd)

	listGroupsCmd.Flags().StringVarP(&attrs, "attributes", "a", "", "required group attributes (separated by ~)")
	listGroupsCmd.Flags().StringVarP(&domain, "domain", "d", "", "domain from which to get groups")
	listGroupsCmd.Flags().StringVarP(&query, "query", "q", "", "selection criteria to get groups (separated by ~)")
}

func processQuery(query string) (string, error) {
	var formattedQuery string

	validQuery, err := cmn.ValidateQuery(query, grps.QueryAttrMap)
	if err != nil {
		return "", err
	}

	formattedQuery, err = grps.FormatQuery(validQuery)
	if err != nil {
		return "", err
	}

	return formattedQuery, nil
}
