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
	"github.com/spf13/cobra"
	admin "google.golang.org/api/admin/directory/v1"
)

var undeleteUserCmd = &cobra.Command{
	Use:  "user <id>",
	Args: cobra.ExactArgs(1),
	Example: `gmin undelete user 417578192529765228417
gmin und user 308127142904731923463 -o /Marketing`,
	Short: "Undeletes user",
	Long: `Undeletes user and reinstates to specified orgunit.
			  
N.B. Must use id and not email address.`,
	RunE: doUndeleteUser,
}

func doUndeleteUser(cmd *cobra.Command, args []string) error {
	logger.Debugw("starting doUndeleteUser()",
		"args", args)

	var userUndelete *admin.UserUndelete
	userUndelete = new(admin.UserUndelete)

	if orgUnit == "" {
		userUndelete.OrgUnitPath = "/"
	} else {
		userUndelete.OrgUnitPath = orgUnit
	}

	ds, err := cmn.CreateDirectoryService(admin.AdminDirectoryUserScope)
	if err != nil {
		logger.Error(err)
		return err
	}

	uuc := ds.Users.Undelete(args[0], userUndelete)

	err = uuc.Do()
	if err != nil {
		logger.Error(err)
		return err
	}

	logger.Infof(cmn.InfoUserUndeleted, args[0])
	fmt.Println(cmn.GminMessage(fmt.Sprintf(cmn.InfoUserUndeleted, args[0])))

	logger.Debug("finished doUndeleteUser()")
	return nil
}

func init() {
	undeleteCmd.AddCommand(undeleteUserCmd)
	undeleteUserCmd.Flags().StringVarP(&orgUnit, "orgunit", "o", "", "path of orgunit to restore user to")
}
