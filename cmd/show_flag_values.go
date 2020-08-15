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
	grps "github.com/plusworx/gmin/utils/groups"
	ous "github.com/plusworx/gmin/utils/orgunits"
	usrs "github.com/plusworx/gmin/utils/users"
	"github.com/spf13/cobra"
)

var showFlagValsCmd = &cobra.Command{
	Use:     "flag-values <object> [flag with predefined values]",
	Aliases: []string{"flag-vals", "fvals"},
	Args:    cobra.MinimumNArgs(1),
	Short:   "Shows object flag predefined value information",
	Long: `Shows object flag predefined value information.
	
	Examples:	gmin show flag-values user projection
			gmin show fvals user orderby`,
	RunE: doShowFlagVals,
}

func doShowFlagVals(cmd *cobra.Command, args []string) error {
	if len(args) > 2 {
		return errors.New("gmin: error - exceeded maximum 2 arguments")
	}

	obj := strings.ToLower(args[0])

	switch {
	case obj == "chromeosdevice" || obj == "crosdevice" || obj == "cdev":
		err := cdevs.ShowFlagValues(len(args), args)
		if err != nil {
			return err
		}
	case obj == "orgunit" || obj == "ou":
		err := ous.ShowFlagValues(len(args), args)
		if err != nil {
			return err
		}
	case obj == "user":
		err := usrs.ShowFlagValues(len(args), args)
		if err != nil {
			return err
		}
	case obj == "group" || obj == "grp":
		err := grps.ShowFlagValues(len(args), args)
		if err != nil {
			return err
		}
	default:
		return fmt.Errorf("gmin: error - %v is not recognized", args[0])
	}
	return nil
}

func init() {
	showCmd.AddCommand(showFlagValsCmd)
}