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
	flgnm "github.com/plusworx/gmin/utils/flagnames"
	gmess "github.com/plusworx/gmin/utils/gminmessages"
	lg "github.com/plusworx/gmin/utils/logging"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var setConfigCmd = &cobra.Command{
	Use:     "config",
	Aliases: []string{"cfg"},
	Example: `gmin set config --admin my.admin@mycompany.org
gmin set cfg -a my.admin@mycompany.org`,
	Short: "Sets gmin configuration information in config file",
	Long: `Sets gmin configuration information in config file.
	
N.B. This command does NOT set environment variables.`,
	RunE: doSetConfig,
}

func doSetConfig(cmd *cobra.Command, args []string) error {
	lg.Debugw("starting doSetConfig()",
		"args", args)
	defer lg.Debug("finished doSetConfig()")

	flgCustIDVal, err := cmd.Flags().GetString(flgnm.FLG_CUSTOMERID)
	if err != nil {
		lg.Error(err)
		return err
	}
	if flgCustIDVal != "" {
		viper.Set(cfg.CONFIGCUSTID, flgCustIDVal)
		err := viper.WriteConfig()
		if err != nil {
			lg.Error(err)
			return err
		}
		fmt.Println(cmn.GminMessage(fmt.Sprintf(gmess.INFO_CUSTOMERIDSET, flgCustIDVal)))
		lg.Infof(gmess.INFO_CUSTOMERIDSET, flgCustIDVal)
	}

	flgAdminVal, err := cmd.Flags().GetString(flgnm.FLG_ADMIN)
	if err != nil {
		lg.Error(err)
		return err
	}
	if flgAdminVal != "" {
		ok := valid.IsEmail(flgAdminVal)
		if !ok {
			err = fmt.Errorf(gmess.ERR_INVALIDEMAILADDRESS, flgAdminVal)
			lg.Error(err)
			return err
		}
		viper.Set(cfg.CONFIGADMIN, flgAdminVal)
		err := viper.WriteConfig()
		if err != nil {
			lg.Error(err)
			return err
		}
		fmt.Println(cmn.GminMessage(fmt.Sprintf(gmess.INFO_ADMINSET, flgAdminVal)))
		lg.Infof(gmess.INFO_ADMINSET, flgAdminVal)
	}

	flgCredPathVal, err := cmd.Flags().GetString(flgnm.FLG_CREDPATH)
	if err != nil {
		lg.Error(err)
		return err
	}
	if flgCredPathVal != "" {
		viper.Set(cfg.CONFIGCREDPATH, flgCredPathVal)
		err := viper.WriteConfig()
		if err != nil {
			lg.Error(err)
			return err
		}
		fmt.Println(cmn.GminMessage(fmt.Sprintf(gmess.INFO_CREDENTIALPATHSET, flgCredPathVal)))
		lg.Infof(gmess.INFO_CREDENTIALPATHSET, flgCredPathVal)
	}

	flgLogPathVal, err := cmd.Flags().GetString(flgnm.FLG_LOGPATH)
	if err != nil {
		lg.Error(err)
		return err
	}
	if flgLogPathVal != "" {
		viper.Set(cfg.CONFIGLOGPATH, flgLogPathVal)
		err := viper.WriteConfig()
		if err != nil {
			lg.Error(err)
			return err
		}
		fmt.Println(cmn.GminMessage(fmt.Sprintf(gmess.INFO_LOGPATHSET, flgLogPathVal)))
		lg.Infof(gmess.INFO_LOGPATHSET, flgLogPathVal)
	}

	if flgAdminVal == "" && flgCustIDVal == "" && flgCredPathVal == "" && flgLogPathVal == "" {
		cmd.Help()
	}

	return nil
}

func init() {
	setCmd.AddCommand(setConfigCmd)

	setConfigCmd.Flags().StringVarP(&adminEmail, flgnm.FLG_ADMIN, "a", "", "administrator email address")
	setConfigCmd.Flags().StringVarP(&customerID, flgnm.FLG_CUSTOMERID, "c", "", "customer id for domain")
	setConfigCmd.Flags().StringVarP(&logPath, flgnm.FLG_LOGPATH, "l", "", "log file path (multiple paths allowed separated by ~)")
	setConfigCmd.Flags().StringVarP(&credentialPath, flgnm.FLG_CREDPATH, "p", "", "service account credential file path")
	setConfigCmd.PersistentFlags().BoolVar(&silent, flgnm.FLG_SILENT, false, "suppress console output")
}
