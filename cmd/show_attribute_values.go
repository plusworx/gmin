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
	gmems "github.com/plusworx/gmin/utils/members"
	mdevs "github.com/plusworx/gmin/utils/mobiledevices"
	usrs "github.com/plusworx/gmin/utils/users"
	"github.com/spf13/cobra"
)

var showAttrValsCmd = &cobra.Command{
	Use:     "attribute-values <object> [field with predefined values]",
	Aliases: []string{"attr-vals", "avals"},
	Args:    cobra.MinimumNArgs(1),
	Example: `gmin show attribute-values user email type
gmin show avals user email type`,
	Short: "Shows object field predefined value information",
	Long: `Shows object field predefined value information.

Valid objects are:
chromeosdevice, crosdevice, crosdev, cdev
group-member, grp-member, grp-mem, gmember, gmem
mobiledevice, mobdevice, mobdev, mdev
user`,
	RunE: doShowAttrVals,
}

func doShowAttrVals(cmd *cobra.Command, args []string) error {
	if len(args) > 3 {
		return errors.New(cmn.ErrMax3ArgsExceeded)
	}

	obj := strings.ToLower(args[0])

	switch {
	case obj == "chromeosdevice" || obj == "crosdevice" || obj == "crosdev" || obj == "cdev":
		err := cdevs.ShowAttrValues(len(args), args)
		if err != nil {
			return err
		}
	case obj == "group-member" || obj == "grp-member" || obj == "grp-mem" || obj == "gmember" || obj == "gmem":
		err := gmems.ShowAttrValues(len(args), args)
		if err != nil {
			return err
		}
	case obj == "mobiledevice" || obj == "mobdevice" || obj == "mobdev" || obj == "mdev":
		err := mdevs.ShowAttrValues(len(args), args)
		if err != nil {
			return err
		}
	case obj == "user":
		err := usrs.ShowAttrValues(len(args), args)
		if err != nil {
			return err
		}
	default:
		return fmt.Errorf(cmn.ErrObjectNotRecognized, args[0])
	}
	return nil
}

func init() {
	showCmd.AddCommand(showAttrValsCmd)
}
