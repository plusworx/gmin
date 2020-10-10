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

	valid "github.com/asaskevich/govalidator"
	cmn "github.com/plusworx/gmin/utils/common"
	cfg "github.com/plusworx/gmin/utils/config"
	gmess "github.com/plusworx/gmin/utils/gminmessages"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var setConfigCmd = &cobra.Command{
	Use:     "config",
	Aliases: []string{"cfg"},
	Example: `gmin set config --admin my.admin@mycompany.org
gmin set cfg -a my.admin@mycompany.org`,
	Short: "Sets gmin configuration information",
	Long:  `Sets gmin configuration information.`,
	RunE:  doSetConfig,
}

func doSetConfig(cmd *cobra.Command, args []string) error {
	if customerID != "" {
		viper.Set(cfg.ConfigCustID, customerID)
		err := viper.WriteConfig()
		if err != nil {
			return err
		}
		fmt.Println(cmn.GminMessage(fmt.Sprintf(gmess.InfoCustomerIDSet, customerID)))
	}

	if adminEmail != "" {
		ok := valid.IsEmail(adminEmail)
		if !ok {
			return fmt.Errorf(gmess.ErrInvalidEmailAddress, adminEmail)
		}
		viper.Set(cfg.ConfigAdmin, adminEmail)
		err := viper.WriteConfig()
		if err != nil {
			return err
		}
		fmt.Println(cmn.GminMessage(fmt.Sprintf(gmess.InfoCustomerIDSet, adminEmail)))
	}

	if credentialPath != "" {
		viper.Set(cfg.ConfigCredPath, credentialPath)
		err := viper.WriteConfig()
		if err != nil {
			return err
		}
		fmt.Println(cmn.GminMessage(fmt.Sprintf(gmess.InfoCredentialPathSet, credentialPath)))
	}

	if logPath != "" {
		viper.Set(cfg.ConfigLogPath, logPath)
		err := viper.WriteConfig()
		if err != nil {
			return err
		}
		fmt.Println(cmn.GminMessage(fmt.Sprintf(gmess.InfoLogPathSet, logPath)))
	}

	if adminEmail == "" && customerID == "" && credentialPath == "" && logPath == "" {
		cmd.Help()
	}

	return nil
}

func init() {
	setCmd.AddCommand(setConfigCmd)

	setConfigCmd.Flags().StringVarP(&adminEmail, "admin", "a", "", "administrator email address")
	setConfigCmd.Flags().StringVarP(&customerID, "customer-id", "c", "", "customer id for domain")
	setConfigCmd.Flags().StringVarP(&logPath, "log-path", "l", "", "log file path (multiple paths allowed separated by ~)")
	setConfigCmd.Flags().StringVarP(&credentialPath, "credential-path", "p", "", "service account credential file path")
}
