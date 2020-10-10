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
	"io/ioutil"
	"os"

	cfg "github.com/plusworx/gmin/utils/config"
	gmess "github.com/plusworx/gmin/utils/gminmessages"
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
	logger.Debug("starting doShowConfig()")

	fmt.Println("gmin Configuration Information")
	fmt.Println("==============================")
	fmt.Println("Environment Variables")
	fmt.Println("---------------------")

	admin := os.Getenv(cfg.EnvPrefix + cfg.EnvVarAdmin)
	credPath := os.Getenv(cfg.EnvPrefix + cfg.EnvVarCredPath)
	custID := os.Getenv(cfg.EnvPrefix + cfg.EnvVarCustID)
	logPath := os.Getenv(cfg.EnvPrefix + cfg.EnvVarLogPath)

	if admin == "" && credPath == "" && custID == "" && logPath == "" {
		fmt.Println(gmess.InfoEnvVarsNotFound)
	}
	if admin != "" {
		fmt.Println(cfg.EnvPrefix+cfg.EnvVarAdmin+":", admin)
	}
	if credPath != "" {
		fmt.Println(cfg.EnvPrefix+cfg.EnvVarCredPath+":", credPath)
	}
	if custID != "" {
		fmt.Println(cfg.EnvPrefix+cfg.EnvVarCustID+":", custID)
	}
	if logPath != "" {
		fmt.Println(cfg.EnvPrefix+cfg.EnvVarLogPath+":", logPath)
	}

	fmt.Println("")
	fmt.Println("Config File")
	fmt.Println("-----------")

	cfgFilePath := viper.GetViper().ConfigFileUsed()
	yamlFile, err := ioutil.ReadFile(cfgFilePath)
	if err != nil {
		fmt.Println(gmess.InfoConfigFileNotFound)
	}

	fmt.Println(string(yamlFile))

	logger.Debug("finished doShowConfig()")
	return nil
}

func init() {
	showCmd.AddCommand(showConfigCmd)
}
