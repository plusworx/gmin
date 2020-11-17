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
	grps "github.com/plusworx/gmin/utils/groups"
	grpset "github.com/plusworx/gmin/utils/groupsettings"
	lg "github.com/plusworx/gmin/utils/logging"
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
user, usr`,
	RunE: doShowFlagVals,
}

func doShowFlagVals(cmd *cobra.Command, args []string) error {
	lg.Debugw("starting doShowFlagVals()",
		"args", args)
	defer lg.Debug("finished doShowFlagVals()")

	aliasSlice := [][]string{
		ca.CDevAliases,
		ca.GroupAliases,
		ca.GMAliases,
		ca.GSAliases,
		ca.MDevAliases,
		ca.OUAliases,
		ca.UserAliases,
	}

	// Functions are mapped to index values governed by aliasSlice
	flagFuncMap := map[int]func(int, []string, string) error{
		0: cdevs.ShowFlagValues,
		1: grps.ShowFlagValues,
		2: gmems.ShowFlagValues,
		3: grpset.ShowFlagValues,
		4: mdevs.ShowFlagValues,
		5: ous.ShowFlagValues,
		6: usrs.ShowFlagValues,
	}

	if len(args) > 2 {
		return errors.New(gmess.ERR_MAX2ARGSEXCEEDED)
	}

	object := strings.ToLower(args[0])

	flgFilterVal, err := cmd.Flags().GetString(flgnm.FLG_FILTER)
	if err != nil {
		return err
	}

	if object == "global" {
		err := cmn.ShowGlobalFlagValues(len(args), args, flgFilterVal)
		if err != nil {
			return err
		}
		return nil
	}

	for idx, alSlice := range aliasSlice {
		exists := cmn.SliceContainsStr(alSlice, object)
		if exists {
			showFunc := flagFuncMap[idx]
			err := showFunc(len(args), args, flgFilterVal)
			if err != nil {
				return err
			}
			return nil
		}
	}

	return fmt.Errorf(gmess.ERR_OBJECTNOTRECOGNIZED, args[0])
}

func init() {
	showCmd.AddCommand(showFlagValsCmd)
	showFlagValsCmd.Flags().StringP(flgnm.FLG_FILTER, "f", "", "string used to filter results")
}
