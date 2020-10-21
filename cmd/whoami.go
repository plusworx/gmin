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
	"os"

	cmn "github.com/plusworx/gmin/utils/common"
	cfg "github.com/plusworx/gmin/utils/config"
	flgnm "github.com/plusworx/gmin/utils/flagnames"
	gmess "github.com/plusworx/gmin/utils/gminmessages"
	lg "github.com/plusworx/gmin/utils/logging"
	"github.com/spf13/cobra"
)

// whoamiCmd represents the whoami command
var whoamiCmd = &cobra.Command{
	Use:   "whoami",
	Short: "Displays the email address of the user impersonated by the service account",
	Long:  `Displays the email address of the user impersonated by the service account.`,
	RunE:  doWhoami,
}

func doWhoami(cmd *cobra.Command, args []string) error {
	lg.Debugw("starting doWhoami()",
		"args", args)
	defer lg.Debug("finished doWhomai()")

	var err error

	email := os.Getenv(cfg.ENVPREFIX + cfg.ENVVARADMIN)

	if email == "" {
		email, err = cfg.ReadConfigString(cfg.CONFIGADMIN)
		if err != nil {
			return err
		}
	}

	fmt.Println(cmn.GminMessage(fmt.Sprintf(gmess.INFO_ADMINIS, email)))

	return nil
}

func init() {
	rootCmd.AddCommand(whoamiCmd)
	whoamiCmd.Flags().StringVar(&logLevel, flgnm.FLG_LOGLEVEL, "info", "log level (debug, info, error, warn)")

	whoamiCmd.PreRunE = preRunForDisplayCmds
}
