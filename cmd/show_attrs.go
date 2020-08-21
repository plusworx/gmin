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

	cdevs "github.com/plusworx/gmin/utils/chromeosdevices"
	cmn "github.com/plusworx/gmin/utils/common"
	gas "github.com/plusworx/gmin/utils/groupaliases"
	grps "github.com/plusworx/gmin/utils/groups"
	mems "github.com/plusworx/gmin/utils/members"
	mdevs "github.com/plusworx/gmin/utils/mobiledevices"
	ous "github.com/plusworx/gmin/utils/orgunits"
	scs "github.com/plusworx/gmin/utils/schemas"
	uas "github.com/plusworx/gmin/utils/useraliases"
	usrs "github.com/plusworx/gmin/utils/users"
	"github.com/spf13/cobra"
)

var showAttrsCmd = &cobra.Command{
	Use:     "attributes <object> [composite attributes]",
	Aliases: []string{"attrs"},
	Args:    cobra.MinimumNArgs(1),
	Short:   "Shows object attribute information",
	Long: `Shows object attribute information.
	
	Valid objects are:
	chromeosdevice, crosdevice, crosdev, cdev
	group, grp
	group-alias, grp-alias, galias, ga
	group-member, grp-member, grp-mem, gmember, gmem
	mobiledevice, mobdevice, mobdev, mdev
	orgunit, ou
	schema, sc
	user
	user-alias, ualias, ua

	Examples:	gmin show attributes user -f pass
			gmin show attrs user name`,
	RunE: doShowAttrs,
}

func doShowAttrs(cmd *cobra.Command, args []string) error {
	lArgs := len(args)

	if lArgs > 3 {
		return fmt.Errorf("gmin: error - maximum of 3 arguments exceeded")
	}

	if composite && queryable {
		return errors.New("gmin: error - cannot provide both --composite and --queryable flags")
	}

	if queryable && lArgs > 1 {
		return errors.New("gmin: error - only one argument is allowed with --queryable flag")
	}

	ok := validateShowSlice(strings.ToLower(args[0]), cmn.ValidPrimaryShowArgs)
	if !ok {
		return fmt.Errorf("gmin: error - %v not found", args[0])
	}

	err := processShowArgs(args, lArgs)
	if err != nil {
		return err
	}

	return nil
}

func init() {
	showCmd.AddCommand(showAttrsCmd)

	showAttrsCmd.Flags().BoolVarP(&composite, "composite", "c", false, "show attributes that contain other attributes")
	showAttrsCmd.Flags().StringVarP(&filter, "filter", "f", "", "string used to filter results")
	showAttrsCmd.Flags().BoolVarP(&queryable, "queryable", "q", false, "show attributes that can be used in a query")
}

func processShowArgs(args []string, lArgs int) error {
	switch {
	case lArgs == 3:
		err := threeArgs(args[0], args[1], args[2])
		if err != nil {
			return err
		}
	case lArgs == 2:
		err := twoArgs(args[0], args[1])
		if err != nil {
			return err
		}
	default:
		err := oneArg(args[0])
		if err != nil {
			return err
		}
	}
	return nil
}

func threeArgs(arg1 string, arg2 string, arg3 string) error {
	obj := strings.ToLower(arg1)
	subAttr := strings.ToLower(arg3)

	switch {
	case obj == "schema" || obj == "sc":
		if composite {
			return fmt.Errorf("gmin: error - %v does not have any composite attributes", arg3)
		}
		err := scs.ShowSubSubAttrs(subAttr)
		if err != nil {
			return err
		}
	default:
		return fmt.Errorf("gmin: error - %v does not have any composite attributes", arg2)
	}

	return nil
}

func twoArgs(arg1 string, arg2 string) error {
	obj := strings.ToLower(arg1)
	switch {
	case obj == "chromeosdevice" || obj == "crosdevice" || obj == "crosdev" || obj == "cdev":
		if composite {
			return fmt.Errorf("gmin: error - %v does not have any composite attributes", arg2)
		}
		err := cdevs.ShowSubAttrs(arg2, filter)
		if err != nil {
			return err
		}
	case obj == "mobiledevice" || obj == "mobdevice" || obj == "mobdev" || obj == "mdev":
		if composite {
			return fmt.Errorf("gmin: error - %v does not have any composite attributes", arg2)
		}
		err := mdevs.ShowSubAttrs(arg2, filter)
		if err != nil {
			return err
		}
	case obj == "schema" || obj == "sc":
		subAttr := strings.ToLower(arg2)
		if composite {
			err := scs.ShowSubCompAttrs(subAttr, filter)
			if err != nil {
				return err
			}
			break
		}
		err := scs.ShowSubAttrs(arg2, filter)
		if err != nil {
			return err
		}
	case obj == "user":
		if composite {
			return fmt.Errorf("gmin: error - %v does not have any composite attributes", arg2)
		}
		err := usrs.ShowSubAttrs(arg2, filter)
		if err != nil {
			return err
		}
	default:
		return fmt.Errorf("gmin: error - %v does not have any composite attributes", arg1)
	}

	return nil
}

func oneArg(arg string) error {
	obj := strings.ToLower(arg)

	if queryable {
		switch {
		case obj == "chromeosdevice" || obj == "crosdevice" || obj == "crosdev" || obj == "cdev":
			cmn.ShowQueryableAttrs(filter, cdevs.QueryAttrMap)
		case obj == "group" || obj == "grp":
			cmn.ShowQueryableAttrs(filter, grps.QueryAttrMap)
		case obj == "mobiledevice" || obj == "mobdevice" || obj == "mobdev" || obj == "mdev":
			cmn.ShowQueryableAttrs(filter, mdevs.QueryAttrMap)
		case obj == "user":
			cmn.ShowQueryableAttrs(filter, usrs.QueryAttrMap)
		default:
			return fmt.Errorf("gmin: error - %v does not have any queryable attributes", arg)
		}
		return nil
	}

	switch {
	case obj == "chromeosdevice" || obj == "crosdevice" || obj == "crosdev" || obj == "cdev":
		if composite {
			cdevs.ShowCompAttrs(filter)
			break
		}
		cdevs.ShowAttrs(filter)
	case obj == "group" || obj == "grp":
		if composite {
			return errors.New("gmin: error - groups do not have any composite attributes")
		}
		grps.ShowAttrs(filter)
	case obj == "group-alias" || obj == "grp-alias" || obj == "galias" || obj == "ga":
		if composite {
			return errors.New("gmin: error - group aliases do not have any composite attributes")
		}
		gas.ShowAttrs(filter)
	case obj == "group-member" || obj == "grp-member" || obj == "grp-mem" || obj == "gmember" || obj == "gmem":
		if composite {
			return errors.New("gmin: error - group members do not have any composite attributes")
		}
		mems.ShowAttrs(filter)
	case obj == "mobiledevice" || obj == "mobdevice" || obj == "mobdev" || obj == "mdev":
		if composite {
			mdevs.ShowCompAttrs(filter)
			break
		}
		mdevs.ShowAttrs(filter)
	case obj == "orgunit" || obj == "ou":
		if composite {
			return errors.New("gmin: error - orgunits do not have any composite attributes")
		}
		ous.ShowAttrs(filter)
	case obj == "schema" || obj == "sc":
		if composite {
			scs.ShowCompAttrs(filter)
			break
		}
		scs.ShowAttrs(filter)
	case obj == "user":
		if composite {
			usrs.ShowCompAttrs(filter)
			break
		}
		usrs.ShowAttrs(filter)
	case obj == "user-alias" || obj == "ualias" || obj == "ua":
		if composite {
			return errors.New("gmin: error - user aliases do not have any composite attributes")
		}
		uas.ShowAttrs(filter)
	}

	return nil
}

func validateShowSlice(arg string, attrSlice []string) bool {
	ok := cmn.SliceContainsStr(attrSlice, arg)
	return ok
}
