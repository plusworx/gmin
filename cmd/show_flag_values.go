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
	ca "github.com/plusworx/gmin/utils/commandaliases"
	cmn "github.com/plusworx/gmin/utils/common"
	gmess "github.com/plusworx/gmin/utils/gminmessages"
	grps "github.com/plusworx/gmin/utils/groups"
	grpset "github.com/plusworx/gmin/utils/groupsettings"
	gmems "github.com/plusworx/gmin/utils/members"
	mdevs "github.com/plusworx/gmin/utils/mobiledevices"
	ous "github.com/plusworx/gmin/utils/orgunits"
	usrs "github.com/plusworx/gmin/utils/users"
	"github.com/spf13/cobra"
)

var showFlagValsCmd = &cobra.Command{
	Use:     "flag-values <object> [flag with predefined values]",
	Aliases: []string{"flag-vals", "fvals"},
	Args:    cobra.MinimumNArgs(1),
	Example: `gmin show flag-values user projection
gmin show fvals user orderby`,
	Short: "Shows object flag predefined value information",
	Long: `Shows object flag predefined value information.

Valid objects are:
chromeos-device, cros-device, cros-dev, cdev
global
group, grp
group-member, grp-member, grp-mem, gmember, gmem
mobile-device, mob-device, mob-dev, mdev
orgunit, ou
user`,
	RunE: doShowFlagVals,
}

func doShowFlagVals(cmd *cobra.Command, args []string) error {
	if len(args) > 2 {
		return errors.New(gmess.ERRMAX2ARGSEXCEEDED)
	}

	object := strings.ToLower(args[0])

	switch {
	case cmn.SliceContainsStr(ca.CDevAliases, object):
		err := cdevs.ShowFlagValues(len(args), args, filter)
		if err != nil {
			return err
		}
	case object == "global":
		err := cmn.ShowGlobalFlagValues(len(args), args, filter)
		if err != nil {
			return err
		}
	case cmn.SliceContainsStr(ca.GroupAliases, object):
		err := grps.ShowFlagValues(len(args), args, filter)
		if err != nil {
			return err
		}
	case cmn.SliceContainsStr(ca.GMAliases, object):
		err := gmems.ShowFlagValues(len(args), args, filter)
		if err != nil {
			return err
		}
	case cmn.SliceContainsStr(ca.GSAliases, object):
		err := grpset.ShowFlagValues(len(args), args, filter)
		if err != nil {
			return err
		}
	case cmn.SliceContainsStr(ca.MDevAliases, object):
		err := mdevs.ShowFlagValues(len(args), args, filter)
		if err != nil {
			return err
		}
	case cmn.SliceContainsStr(ca.OUAliases, object):
		err := ous.ShowFlagValues(len(args), args, filter)
		if err != nil {
			return err
		}
	case cmn.SliceContainsStr(ca.UserAliases, object):
		err := usrs.ShowFlagValues(len(args), args, filter)
		if err != nil {
			return err
		}
	default:
		return fmt.Errorf(gmess.ERROBJECTNOTRECOGNIZED, args[0])
	}
	return nil
}

func init() {
	showCmd.AddCommand(showFlagValsCmd)
	showFlagValsCmd.Flags().StringVarP(&filter, "filter", "f", "", "string used to filter results")
}
