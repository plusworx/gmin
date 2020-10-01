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
	"os"

	cmn "github.com/plusworx/gmin/utils/common"
	"github.com/spf13/cobra"
)

var batchCreateCmd = &cobra.Command{
	Use:     "batch-create",
	Aliases: []string{"bcreate", "bcrt"},
	Args:    cobra.NoArgs,
	Short:   "Creates a batch of G Suite entities",
	Long:    "Creates a batch of G Suite entities.",
	Run:     doBatchCreate,
}

func doBatchCreate(cmd *cobra.Command, args []string) {
	cmd.Help()
}

func init() {
	rootCmd.AddCommand(batchCreateCmd)
	batchCreateCmd.PersistentFlags().BoolVar(&silent, "silent", false, "suppress console output")
	batchCreateCmd.PersistentFlags().StringVar(&logLevel, "loglevel", "info", "log level (debug, info, error, warn)")

	batchCreateCmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		// Set log level
		zlog, err := setupLogging(logLevel)
		if err != nil {
			return err
		}
		logger = zlog.Sugar()
		// Suppress console output according to silent flag
		if silent {
			os.Stdout = nil
			os.Stderr = nil
		}
		// Log current user
		logger.Infow("User Information",
			"Username", cmn.Username(),
			"Hostname", cmn.Hostname(),
			"IP Address", cmn.IPAddress())
		return nil
	}
}
