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

	cfg "github.com/plusworx/gmin/utils/config"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var showConfigCmd = &cobra.Command{
	Use:     "config",
	Aliases: []string{"cfg"},
	Args:    cobra.NoArgs,
	Example: `gmin show config`,
	Short:   "Shows configuration information",
	Long:    `Shows configuration information.`,
	RunE:    doShowConfig,
}

func doShowConfig(cmd *cobra.Command, args []string) error {
	fmt.Println("Configuration Information")
	fmt.Println("=========================")
	fmt.Println("Environment Variables")
	fmt.Println("---------------------")

	admin := os.Getenv(cfg.EnvPrefix + cfg.EnvVarAdmin)
	if admin != "" {
		fmt.Println(cfg.EnvPrefix+cfg.EnvVarAdmin, ": ", admin)
	}

	credPath := os.Getenv(cfg.EnvPrefix + cfg.EnvVarCredPath)
	if credPath != "" {
		fmt.Println(cfg.EnvPrefix+cfg.EnvVarCredPath, ": ", credPath)
	}

	custID := os.Getenv(cfg.EnvPrefix + cfg.EnvVarCustID)
	if custID != "" {
		fmt.Println(cfg.EnvPrefix+cfg.EnvVarCustID, ": ", custID)
	}

	logPath := os.Getenv(cfg.EnvPrefix + cfg.EnvVarLogPath)
	if logPath != "" {
		fmt.Println(cfg.EnvPrefix+cfg.EnvVarLogPath, ": ", logPath)
	}

	fmt.Println("")
	fmt.Println("Config File")
	fmt.Println("-----------")
	cfgAdmin := viper.GetString(cfg.ConfigAdmin)
	if cfgAdmin != "" {
		fmt.Println(cfg.ConfigAdmin, ":\t\t", cfgAdmin)
	}
	cfgCredPath := viper.GetString(cfg.ConfigCredPath)
	if cfgCredPath != "" {
		fmt.Println(cfg.ConfigCredPath, ":\t", cfgCredPath)
	}
	cfgCustID := viper.GetString(cfg.ConfigCustID)
	if cfgCustID != "" {
		fmt.Println(cfg.ConfigCustID, ":\t\t", cfgCustID)
	}
	cfgLogPath := viper.GetString(cfg.ConfigLogPath)
	if cfgLogPath != "" {
		fmt.Println(cfg.ConfigLogPath, ":\t\t", cfgLogPath)
	}

	return nil
}

func init() {
	showCmd.AddCommand(showConfigCmd)
}
