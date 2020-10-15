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
	gas "github.com/plusworx/gmin/utils/groupaliases"
	grps "github.com/plusworx/gmin/utils/groups"
	grpset "github.com/plusworx/gmin/utils/groupsettings"
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
	Aliases: []string{"attrs", "attr"},
	Args:    cobra.RangeArgs(1, 3),
	Example: `gmin show attributes user -f pass
gmin show attrs user name`,
	Short: "Shows object attribute information",
	Long: `Shows object attribute information.
	
Valid objects are:
chromeos-device, cros-device, cros-dev, cdev
group, grp
group-alias, grp-alias, galias, ga
group-member, grp-member, grp-mem, gmember, gmem
group-settings,	grp-settings, grp-set, gsettings, gset
mobile-device, mob-device, mob-dev, mdev
orgunit, ou
schema, sc
user, usr
user-alias, ualias, ua`,
	RunE: doShowAttrs,
}

func doShowAttrs(cmd *cobra.Command, args []string) error {
	logger.Debugw("starting doShowAttrs()",
		"args", args)

	lArgs := len(args)
	object := strings.ToLower(args[0])

	flgFilterVal, err := cmd.Flags().GetString(flgnm.FLG_FILTER)
	if err != nil {
		return err
	}
	lowerFilter := strings.ToLower(flgFilterVal)

	flgQueryableVal, err := cmd.Flags().GetBool(flgnm.FLG_QUERYABLE)
	if err != nil {
		return err
	}
	if flgQueryableVal && lArgs > 1 {
		return errors.New(gmess.ERR_QUERYABLEFLAG1ARG)
	}

	flgCompositeVal, err := cmd.Flags().GetBool(flgnm.FLG_COMPOSITE)
	if err != nil {
		return err
	}
	if flgCompositeVal && flgQueryableVal {
		return errors.New(gmess.ERR_QUERYANDCOMPOSITEFLAGS)
	}

	ok := cmn.SliceContainsStr(cmn.ValidPrimaryShowArgs, object)
	if !ok {
		return fmt.Errorf(gmess.ERR_OBJECTNOTFOUND, args[0])
	}

	if cmn.SliceContainsStr(ca.CDevAliases, object) {
		err := saChromeOSDev(args, lArgs, args[0], lowerFilter, flgQueryableVal, flgCompositeVal)
		if err != nil {
			return err
		}
	}

	if cmn.SliceContainsStr(ca.GroupAliases, object) {
		err := saGroup(args, lArgs, args[0], lowerFilter, flgQueryableVal, flgCompositeVal)
		if err != nil {
			return err
		}
	}

	if cmn.SliceContainsStr(ca.GAAliases, object) {
		err := saGroupAlias(args, lArgs, args[0], lowerFilter, flgQueryableVal, flgCompositeVal)
		if err != nil {
			return err
		}
	}

	if cmn.SliceContainsStr(ca.GMAliases, object) {
		err := saGroupMember(args, lArgs, args[0], lowerFilter, flgQueryableVal, flgCompositeVal)
		if err != nil {
			return err
		}
	}

	if cmn.SliceContainsStr(ca.GSAliases, object) {
		err := saGroupSettings(args, lArgs, args[0], lowerFilter, flgQueryableVal, flgCompositeVal)
		if err != nil {
			return err
		}
	}

	if cmn.SliceContainsStr(ca.MDevAliases, object) {
		err := saMobileDev(args, lArgs, args[0], lowerFilter, flgQueryableVal, flgCompositeVal)
		if err != nil {
			return err
		}
	}

	if cmn.SliceContainsStr(ca.OUAliases, object) {
		err := saOrgUnit(args, lArgs, args[0], lowerFilter, flgQueryableVal, flgCompositeVal)
		if err != nil {
			return err
		}
	}

	if cmn.SliceContainsStr(ca.SCAliases, object) {
		err := saSchema(args, lArgs, args[0], lowerFilter, flgQueryableVal, flgCompositeVal)
		if err != nil {
			return err
		}
	}

	if cmn.SliceContainsStr(ca.UserAliases, object) {
		err := saUser(args, lArgs, args[0], lowerFilter, flgQueryableVal, flgCompositeVal)
		if err != nil {
			return err
		}
	}

	if cmn.SliceContainsStr(ca.UAAliases, object) {
		err := saUserAlias(args, lArgs, args[0], lowerFilter, flgQueryableVal, flgCompositeVal)
		if err != nil {
			return err
		}
	}

	logger.Debug("finished doShowAttrs()")
	return nil
}

func init() {
	showCmd.AddCommand(showAttrsCmd)

	showAttrsCmd.Flags().BoolVarP(&composite, flgnm.FLG_COMPOSITE, "c", false, "show attributes that contain other attributes")
	showAttrsCmd.Flags().StringVarP(&filter, flgnm.FLG_FILTER, "f", "", "string used to filter results")
	showAttrsCmd.Flags().BoolVarP(&queryable, flgnm.FLG_QUERYABLE, "q", false, "show attributes that can be used in a query")
}

func saChromeOSDev(args []string, lArgs int, objectName string, filter string, queryable bool, composite bool) error {
	logger.Debugw("starting saChromeOSDev()",
		"args", args,
		"lArgs", lArgs,
		"objectName", objectName)

	if queryable {
		cmn.ShowQueryableAttrs(filter, cdevs.QueryAttrMap)
		return nil
	}

	if lArgs == 1 {
		if composite {
			cdevs.ShowCompAttrs(filter)
			return nil
		}
		cdevs.ShowAttrs(filter)
		return nil
	}

	if lArgs == 2 {
		if composite {
			return fmt.Errorf(gmess.ERR_NOCOMPOSITEATTRS, args[lArgs-1])
		}
		err := cdevs.ShowSubAttrs(args[lArgs-1], filter)
		if err != nil {
			return err
		}
	}

	if lArgs > 2 {
		return fmt.Errorf(gmess.ERR_NOCOMPOSITEATTRS, args[2])
	}

	logger.Debug("finished saChromeOSDev()")
	return nil
}

func saGroup(args []string, lArgs int, objectName string, filter string, queryable bool, composite bool) error {
	logger.Debugw("starting saGroup()",
		"args", args,
		"lArgs", lArgs,
		"objectName", objectName)

	if queryable {
		cmn.ShowQueryableAttrs(filter, grps.QueryAttrMap)
		return nil
	}

	if lArgs == 1 {
		if composite {
			return fmt.Errorf(gmess.ERR_NOCOMPOSITEATTRS, objectName)
		}
		grps.ShowAttrs(filter)
	}

	if lArgs > 1 {
		return fmt.Errorf(gmess.ERR_NOCOMPOSITEATTRS, objectName)
	}

	logger.Debug("finished saGroup()")
	return nil
}

func saGroupAlias(args []string, lArgs int, objectName string, filter string, queryable bool, composite bool) error {
	logger.Debugw("starting saGroupAlias()",
		"args", args,
		"lArgs", lArgs,
		"objectName", objectName)

	if queryable {
		return fmt.Errorf(gmess.ERR_NOQUERYABLEATTRS, objectName)
	}

	if lArgs == 1 {
		if composite {
			return fmt.Errorf(gmess.ERR_NOCOMPOSITEATTRS, objectName)
		}
		gas.ShowAttrs(filter)
	}

	if lArgs > 1 {
		return fmt.Errorf(gmess.ERR_NOCOMPOSITEATTRS, objectName)
	}

	logger.Debug("finished saGroupAlias()")
	return nil
}

func saGroupMember(args []string, lArgs int, objectName string, filter string, queryable bool, composite bool) error {
	logger.Debugw("starting saGroupMember()",
		"args", args,
		"lArgs", lArgs,
		"objectName", objectName)

	if queryable {
		return fmt.Errorf(gmess.ERR_NOQUERYABLEATTRS, objectName)
	}

	if lArgs == 1 {
		if composite {
			return fmt.Errorf(gmess.ERR_NOCOMPOSITEATTRS, objectName)
		}
		mems.ShowAttrs(filter)
	}

	if lArgs > 1 {
		return fmt.Errorf(gmess.ERR_NOCOMPOSITEATTRS, objectName)
	}

	logger.Debug("finished saGroupMember()")
	return nil
}

func saGroupSettings(args []string, lArgs int, objectName string, filter string, queryable bool, composite bool) error {
	logger.Debugw("starting saGroupSettings()",
		"args", args,
		"lArgs", lArgs,
		"objectName", objectName)

	if queryable {
		return fmt.Errorf(gmess.ERR_NOQUERYABLEATTRS, objectName)
	}

	if lArgs == 1 {
		if composite {
			return fmt.Errorf(gmess.ERR_NOCOMPOSITEATTRS, objectName)
		}
		grpset.ShowAttrs(filter)
	}

	if lArgs > 1 {
		return fmt.Errorf(gmess.ERR_NOCOMPOSITEATTRS, objectName)
	}

	logger.Debug("finished saGroupSettings()")
	return nil
}

func saMobileDev(args []string, lArgs int, objectName string, filter string, queryable bool, composite bool) error {
	logger.Debugw("starting saMobileDev()",
		"args", args,
		"lArgs", lArgs,
		"objectName", objectName)

	if queryable {
		cmn.ShowQueryableAttrs(filter, mdevs.QueryAttrMap)
		return nil
	}

	if lArgs == 1 {
		if composite {
			mdevs.ShowCompAttrs(filter)
			return nil
		}
		mdevs.ShowAttrs(filter)
		return nil
	}

	if lArgs == 2 {
		if composite {
			return fmt.Errorf(gmess.ERR_NOCOMPOSITEATTRS, args[lArgs-1])
		}
		err := mdevs.ShowSubAttrs(args[lArgs-1], filter)
		if err != nil {
			return err
		}
	}

	if lArgs > 2 {
		return fmt.Errorf(gmess.ERR_NOCOMPOSITEATTRS, args[2])
	}

	logger.Debug("finished saMobileDev()")
	return nil
}

func saOrgUnit(args []string, lArgs int, objectName string, filter string, queryable bool, composite bool) error {
	logger.Debugw("starting saOrgUnit()",
		"args", args,
		"lArgs", lArgs,
		"objectName", objectName)

	if queryable {
		return fmt.Errorf(gmess.ERR_NOQUERYABLEATTRS, objectName)
	}

	if lArgs == 1 {
		if composite {
			return fmt.Errorf(gmess.ERR_NOCOMPOSITEATTRS, objectName)
		}
		ous.ShowAttrs(filter)
	}

	if lArgs > 1 {
		return fmt.Errorf(gmess.ERR_NOCOMPOSITEATTRS, objectName)
	}

	logger.Debug("finished saOrgUnit()")
	return nil
}

func saSchema(args []string, lArgs int, objectName string, filter string, queryable bool, composite bool) error {
	logger.Debugw("starting saSchema()",
		"args", args,
		"lArgs", lArgs,
		"objectName", objectName)

	if queryable {
		return fmt.Errorf(gmess.ERR_NOQUERYABLEATTRS, objectName)
	}

	if lArgs == 1 {
		if composite {
			scs.ShowCompAttrs(filter)
			return nil
		}
		scs.ShowAttrs(filter)
	}

	if lArgs == 2 {
		if composite {
			err := scs.ShowSubCompAttrs(args[lArgs-1], filter)
			if err != nil {
				return err
			}
			return nil
		}
		err := scs.ShowSubAttrs(args[lArgs-1], filter)
		if err != nil {
			return err
		}
		return nil
	}

	if lArgs == 3 {
		if composite {
			return fmt.Errorf(gmess.ERR_NOCOMPOSITEATTRS, args[lArgs-1])
		}
		err := scs.ShowSubSubAttrs(args[lArgs-1])
		if err != nil {
			return err
		}
	}

	logger.Debug("finished saSchema()")
	return nil
}

func saUser(args []string, lArgs int, objectName string, filter string, queryable bool, composite bool) error {
	logger.Debugw("starting saUser()",
		"args", args,
		"lArgs", lArgs,
		"objectName", objectName)

	if queryable {
		cmn.ShowQueryableAttrs(filter, usrs.QueryAttrMap)
		return nil
	}

	if lArgs == 1 {
		if composite {
			usrs.ShowCompAttrs(filter)
			return nil
		}
		usrs.ShowAttrs(filter)
	}

	if lArgs == 2 {
		if composite {
			return fmt.Errorf(gmess.ERR_NOCOMPOSITEATTRS, args[lArgs-1])
		}
		err := usrs.ShowSubAttrs(args[lArgs-1], filter)
		if err != nil {
			return err
		}
	}

	if lArgs > 2 {
		return fmt.Errorf(gmess.ERR_NOCOMPOSITEATTRS, args[2])
	}

	logger.Debug("finished saUser()")
	return nil
}

func saUserAlias(args []string, lArgs int, objectName string, filter string, queryable bool, composite bool) error {
	logger.Debugw("starting saUserAlias()",
		"args", args,
		"lArgs", lArgs,
		"objectName", objectName)

	if queryable {
		return fmt.Errorf(gmess.ERR_NOQUERYABLEATTRS, objectName)
	}

	if lArgs == 1 {
		if composite {
			return fmt.Errorf(gmess.ERR_NOCOMPOSITEATTRS, objectName)
		}
		uas.ShowAttrs(filter)
	}

	if lArgs > 1 {
		return fmt.Errorf(gmess.ERR_NOCOMPOSITEATTRS, objectName)
	}

	logger.Debug("finished saUserAlias()")
	return nil
}
