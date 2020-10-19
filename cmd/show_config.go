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

	lg "github.com/plusworx/gmin/utils/logging"

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
	lg.Debug("starting doShowConfig()")
	defer lg.Debug("finished doShowConfig()")

	fmt.Println("gmin Configuration Information")
	fmt.Println("==============================")
	fmt.Println("Environment Variables")
	fmt.Println("---------------------")

	admin := os.Getenv(cfg.ENVPREFIX + cfg.ENVVARADMIN)
	credPath := os.Getenv(cfg.ENVPREFIX + cfg.ENVVARCREDPATH)
	custID := os.Getenv(cfg.ENVPREFIX + cfg.ENVVARCUSTID)
	logPath := os.Getenv(cfg.ENVPREFIX + cfg.ENVVARLOGPATH)

	if admin == "" && credPath == "" && custID == "" && logPath == "" {
		fmt.Println(gmess.INFO_ENVVARSNOTFOUND)
	}
	if admin != "" {
		fmt.Println(cfg.ENVPREFIX+cfg.ENVVARADMIN+":", admin)
	}
	if credPath != "" {
		fmt.Println(cfg.ENVPREFIX+cfg.ENVVARCREDPATH+":", credPath)
	}
	if custID != "" {
		fmt.Println(cfg.ENVPREFIX+cfg.ENVVARCUSTID+":", custID)
	}
	if logPath != "" {
		fmt.Println(cfg.ENVPREFIX+cfg.ENVVARLOGPATH+":", logPath)
	}

	fmt.Println("")
	fmt.Println("Config File")
	fmt.Println("-----------")

	cfgFilePath := viper.GetViper().ConfigFileUsed()
	yamlFile, err := ioutil.ReadFile(cfgFilePath)
	if err != nil {
		fmt.Println(gmess.INFO_CONFIGFILENOTFOUND)
	}

	fmt.Println(string(yamlFile))

	return nil
}

func init() {
	showCmd.AddCommand(showConfigCmd)
}
