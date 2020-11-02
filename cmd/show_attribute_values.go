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
	flgnm "github.com/plusworx/gmin/utils/flagnames"
	gmess "github.com/plusworx/gmin/utils/gminmessages"
	grpset "github.com/plusworx/gmin/utils/groupsettings"
	lg "github.com/plusworx/gmin/utils/logging"
	gmems "github.com/plusworx/gmin/utils/members"
	mdevs "github.com/plusworx/gmin/utils/mobiledevices"
	usrs "github.com/plusworx/gmin/utils/users"
	"github.com/spf13/cobra"
)

var showAttrValsCmd = &cobra.Command{
	Use:     "attribute-values <object> [field with predefined values]",
	Aliases: []string{"attr-vals", "avals"},
	Args:    cobra.RangeArgs(1, 3),
	Example: `gmin show attribute-values user email type
gmin show avals user email type`,
	Short: "Shows object field predefined value information",
	Long: `Shows object field predefined value information.

Valid objects are:
chromeos-device, cros-device, cros-dev, cdev
group-member, grp-member, grp-mem, gmember, gmem
group-settings, grp-settings, grp-set, gsettings, gset
mobile-device, mob-device, mob-dev, mdev
user, usr`,
	RunE: doShowAttrVals,
}

func doShowAttrVals(cmd *cobra.Command, args []string) error {
	lg.Debugw("starting doShowAttrVals()",
		"args", args)
	defer lg.Debug("finished doShowAttrVals")

	lArgs := len(args)

	flgFilterVal, err := cmd.Flags().GetString(flgnm.FLG_FILTER)
	if err != nil {
		lg.Error(err)
		return err
	}
	lowerFilter := strings.ToLower(flgFilterVal)

	if lArgs > 3 {
		err = errors.New(gmess.ERR_MAX3ARGSEXCEEDED)
		lg.Error(err)
		return err
	}

	object := strings.ToLower(args[0])

	switch {
	case cmn.SliceContainsStr(ca.CDevAliases, object):
		err := cdevs.ShowAttrValues(lArgs, args, lowerFilter)
		if err != nil {
			return err
		}
	case cmn.SliceContainsStr(ca.GMAliases, object):
		err := gmems.ShowAttrValues(lArgs, args, lowerFilter)
		if err != nil {
			return err
		}
	case cmn.SliceContainsStr(ca.GSAliases, object):
		err := grpset.ShowAttrValues(lArgs, args, lowerFilter)
		if err != nil {
			return err
		}
	case cmn.SliceContainsStr(ca.MDevAliases, object):
		err := mdevs.ShowAttrValues(lArgs, args, lowerFilter)
		if err != nil {
			return err
		}
	case cmn.SliceContainsStr(ca.UserAliases, object):
		err := usrs.ShowAttrValues(lArgs, args, lowerFilter)
		if err != nil {
			return err
		}
	default:
		err = fmt.Errorf(gmess.ERR_OBJECTNOTRECOGNIZED, args[0])
		lg.Error(err)
		return err
	}
	return nil
}

func init() {
	showCmd.AddCommand(showAttrValsCmd)
	showAttrValsCmd.Flags().StringP(flgnm.FLG_FILTER, "f", "", "string used to filter results")
}
