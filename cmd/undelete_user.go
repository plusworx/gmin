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
	"fmt"

	cmn "github.com/plusworx/gmin/utils/common"
	flgnm "github.com/plusworx/gmin/utils/flagnames"
	gmess "github.com/plusworx/gmin/utils/gminmessages"
	lg "github.com/plusworx/gmin/utils/logging"
	"github.com/spf13/cobra"
	admin "google.golang.org/api/admin/directory/v1"
)

var undeleteUserCmd = &cobra.Command{
	Use:     "user <id>",
	Aliases: []string{"usr"},
	Args:    cobra.ExactArgs(1),
	Example: `gmin undelete user 417578192529765228417
gmin und user 308127142904731923463 -o /Marketing`,
	Short: "Undeletes user",
	Long: `Undeletes user and reinstates to specified orgunit.
			  
N.B. Must use id and not email address.`,
	RunE: doUndeleteUser,
}

func doUndeleteUser(cmd *cobra.Command, args []string) error {
	lg.Debugw("starting doUndeleteUser()",
		"args", args)
	defer lg.Debug("finished doUndeleteUser()")

	var userUndelete *admin.UserUndelete
	userUndelete = new(admin.UserUndelete)

	flgOUVal, err := cmd.Flags().GetString(flgnm.FLG_ORGUNIT)
	if err != nil {
		lg.Error(err)
		return err
	}
	if flgOUVal == "" {
		userUndelete.OrgUnitPath = "/"
	} else {
		userUndelete.OrgUnitPath = flgOUVal
	}

	srv, err := cmn.CreateService(cmn.SRVTYPEADMIN, admin.AdminDirectoryUserScope)
	if err != nil {
		return err
	}
	ds := srv.(*admin.Service)

	uuc := ds.Users.Undelete(args[0], userUndelete)

	err = uuc.Do()
	if err != nil {
		lg.Error(err)
		return err
	}

	fmt.Println(cmn.GminMessage(fmt.Sprintf(gmess.INFO_USERUNDELETED, args[0])))
	lg.Infof(gmess.INFO_USERUNDELETED, args[0])

	return nil
}

func init() {
	undeleteCmd.AddCommand(undeleteUserCmd)
	undeleteUserCmd.Flags().StringVarP(&orgUnit, flgnm.FLG_ORGUNIT, "o", "", "path of orgunit to restore user to")
}
