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
	flgnm "github.com/plusworx/gmin/utils/flagnames"
	"github.com/spf13/cobra"
)

var batchManageCmd = &cobra.Command{
	Use:     "batch-manage",
	Aliases: []string{"bmanage", "bmng"},
	Args:    cobra.NoArgs,
	Short:   "Manages a batch of Google Workspace devices",
	Long:    "Manages a batch of Google Workspace devices.",
	Run:     doBatchManage,
}

func doBatchManage(cmd *cobra.Command, args []string) {
	cmd.Help()
}

func init() {
	rootCmd.AddCommand(batchManageCmd)
	batchManageCmd.PersistentFlags().BoolVar(&silent, flgnm.FLG_SILENT, false, "suppress console output")
	batchManageCmd.PersistentFlags().StringVar(&logLevel, flgnm.FLG_LOGLEVEL, "info", "log level (debug, info, error, warn)")

	batchManageCmd.PersistentPreRunE = preRun
}
