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
	gas "github.com/plusworx/gmin/utils/groupaliases"
	grps "github.com/plusworx/gmin/utils/groups"
	mems "github.com/plusworx/gmin/utils/members"
	ous "github.com/plusworx/gmin/utils/orgunits"
	uas "github.com/plusworx/gmin/utils/useraliases"
	usrs "github.com/plusworx/gmin/utils/users"
	"github.com/spf13/cobra"
)

var showAttrsCmd = &cobra.Command{
	Use:     "attributes <object> [composite attribute]",
	Aliases: []string{"attrs"},
	Args:    cobra.MinimumNArgs(1),
	Short:   "Shows object attribute information",
	Long: `Shows object attribute information.
	
	Examples:	gmin show attributes user -f pass
			gmin show attrs user name`,
	RunE: doShowAttrs,
}

func doShowAttrs(cmd *cobra.Command, args []string) error {
	obj := strings.ToLower(args[0])

	if composite {
		switch {
		case obj == "chromeosdevice" || obj == "crosdevice" || obj == "cdev":
			cdevs.ShowCompAttrs(filter)

		case obj == "user":
			usrs.ShowCompAttrs(filter)
		default:
			fmt.Printf("gmin: %v does not have any composite attributes\n", args[0])
		}
		return nil
	}

	if len(args) > 1 {
		err := doShowSubAttrs(obj, args[1])
		if err != nil {
			return err
		}
		return nil
	}

	switch {
	case obj == "chromeosdevice" || obj == "crosdevice" || obj == "cdev":
		cdevs.ShowAttrs(filter)
	case obj == "group" || obj == "grp":
		grps.ShowAttrs(filter)
	case obj == "group-alias" || obj == "galias" || obj == "ga":
		gas.ShowAttrs(filter)
	case obj == "group-member" || obj == "grp-member" || obj == "grp-mem" || obj == "gmember" || obj == "gmem":
		mems.ShowAttrs(filter)
	case obj == "orgunit" || obj == "ou":
		ous.ShowAttrs(filter)
	case obj == "user":
		usrs.ShowAttrs(filter)
	case obj == "user-alias" || obj == "ualias" || obj == "ua":
		uas.ShowAttrs(filter)
	default:
		return fmt.Errorf("gmin: error - %v not found", args[0])
	}

	return nil
}

func doShowSubAttrs(obj string, subAttr string) error {
	switch {
	case obj == "chromeosdevice" || obj == "crosdevice" || obj == "cdev":
		err := cdevs.ShowSubAttrs(subAttr, filter)
		if err != nil {
			return err
		}
	case obj == "group" || obj == "grp":
		return errors.New("gmin: error - groups do not have any composite attributes")
	case obj == "group-alias" || obj == "galias" || obj == "ga":
		return errors.New("gmin: error - group aliases do not have any composite attributes")
	case obj == "group-member" || obj == "grp-member" || obj == "grp-mem" || obj == "gmember" || obj == "gmem":
		return errors.New("gmin: error - group members do not have any composite attributes")
	case obj == "orgunit" || obj == "ou":
		return errors.New("gmin: error - orgunits do not have any composite attributes")
	case obj == "user":
		err := usrs.ShowSubAttrs(subAttr, filter)
		if err != nil {
			return err
		}
	case obj == "user-alias" || obj == "ualias" || obj == "ua":
		return errors.New("gmin: error - user aliases do not have any composite attributes")
	default:
		return fmt.Errorf("gmin: error - %v not found", obj)
	}
	return nil
}

func init() {
	showCmd.AddCommand(showAttrsCmd)

	showAttrsCmd.Flags().BoolVarP(&composite, "composite", "c", false, "show attributes that contain other attributes")
	showAttrsCmd.Flags().StringVarP(&filter, "filter", "f", "", "string used to filter results")
}
