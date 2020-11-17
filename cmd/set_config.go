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
	"github.com/spf13/pflag"
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

	var flagsPassed []string

	flagValueMap := map[string]interface{}{}
	setCfgFuncMap := map[string]func(interface{}) error{
		flgnm.FLG_ADMIN:            setCfgAdmin,
		flgnm.FLG_CREDPATH:         setCfgCredPath,
		flgnm.FLG_CUSTOMERID:       setCfgCustomerID,
		flgnm.FLG_LOGPATH:          setCfgLogPath,
		flgnm.FLG_LOGROTATIONCOUNT: setCfgLogRotCount,
		flgnm.FLG_LOGROTATIONTIME:  setCfgLogRotTime,
	}

	cmd.Flags().Visit(func(f *pflag.Flag) {
		flagsPassed = append(flagsPassed, f.Name)
	})

	for _, flg := range flagsPassed {
		val, err := cfg.GetFlagVal(cmd, flg)
		if err != nil {
			return err
		}
		flagValueMap[flg] = val
	}

	// Cycle through flags
	for key, val := range flagValueMap {
		scf, ok := setCfgFuncMap[key]
		if !ok {
			continue
		}
		err := scf(val)
		if err != nil {
			return err
		}
	}

	return nil
}

func init() {
	setCmd.AddCommand(setConfigCmd)

	setConfigCmd.Flags().StringP(flgnm.FLG_ADMIN, "a", "", "administrator email address")
	setConfigCmd.Flags().StringP(flgnm.FLG_CUSTOMERID, "c", "", "customer id for domain")
	setConfigCmd.Flags().StringP(flgnm.FLG_LOGPATH, "l", "", "log file path")
	setConfigCmd.Flags().UintP(flgnm.FLG_LOGROTATIONCOUNT, "r", 0, "max number of retained log files")
	setConfigCmd.Flags().IntP(flgnm.FLG_LOGROTATIONTIME, "t", 0, "time after which new log file created")
	setConfigCmd.Flags().StringP(flgnm.FLG_CREDPATH, "p", "", "service account credential file path")
}

func setCfgAdmin(flgVal interface{}) error {
	lg.Debug("starting setCfgAdmin()")
	defer lg.Debug("finished setCfgAdmin()")

	flgAdminVal := fmt.Sprintf("%v", flgVal)
	if flgAdminVal == "" {
		err := fmt.Errorf(gmess.ERR_EMPTYSTRING, "--"+flgnm.FLG_ADMIN)
		lg.Error(err)
		return err
	}

	ok := valid.IsEmail(flgAdminVal)
	if !ok {
		err := fmt.Errorf(gmess.ERR_INVALIDEMAILADDRESS, flgAdminVal)
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

	return nil
}

func setCfgCredPath(flgVal interface{}) error {
	lg.Debug("starting setCfgCredPath()")
	defer lg.Debug("finished setCfgCredPath()")

	flgCredPathVal := fmt.Sprintf("%v", flgVal)
	if flgCredPathVal == "" {
		err := fmt.Errorf(gmess.ERR_EMPTYSTRING, "--"+flgnm.FLG_CREDPATH)
		lg.Error(err)
		return err
	}

	viper.Set(cfg.CONFIGCREDPATH, flgCredPathVal)
	err := viper.WriteConfig()
	if err != nil {
		lg.Error(err)
		return err
	}
	fmt.Println(cmn.GminMessage(fmt.Sprintf(gmess.INFO_CREDENTIALPATHSET, flgCredPathVal)))
	lg.Infof(gmess.INFO_CREDENTIALPATHSET, flgCredPathVal)

	return nil
}

func setCfgCustomerID(flgVal interface{}) error {
	lg.Debug("starting setCfgCustomerID()")
	defer lg.Debug("finished setCfgCustomerID()")

	flgCustIDVal := fmt.Sprintf("%v", flgVal)
	if flgCustIDVal == "" {
		err := fmt.Errorf(gmess.ERR_EMPTYSTRING, "--"+flgnm.FLG_CUSTOMERID)
		lg.Error(err)
		return err
	}

	viper.Set(cfg.CONFIGCUSTID, flgCustIDVal)
	err := viper.WriteConfig()
	if err != nil {
		lg.Error(err)
		return err
	}
	fmt.Println(cmn.GminMessage(fmt.Sprintf(gmess.INFO_CUSTOMERIDSET, flgCustIDVal)))
	lg.Infof(gmess.INFO_CUSTOMERIDSET, flgCustIDVal)

	return nil
}

func setCfgLogPath(flgVal interface{}) error {
	lg.Debug("starting setCfgLogPath()")
	defer lg.Debug("finished setCfgLogPath()")

	flgLogPathVal := fmt.Sprintf("%v", flgVal)
	if flgLogPathVal == "" {
		err := fmt.Errorf(gmess.ERR_EMPTYSTRING, "--"+flgnm.FLG_LOGPATH)
		lg.Error(err)
		return err
	}

	viper.Set(cfg.CONFIGLOGPATH, flgLogPathVal)
	err := viper.WriteConfig()
	if err != nil {
		lg.Error(err)
		return err
	}
	fmt.Println(cmn.GminMessage(fmt.Sprintf(gmess.INFO_LOGPATHSET, flgLogPathVal)))
	lg.Infof(gmess.INFO_LOGPATHSET, flgLogPathVal)

	return nil
}

func setCfgLogRotCount(flgVal interface{}) error {
	lg.Debug("starting setCfgLogRotCount()")
	defer lg.Debug("finished setCfgLogRotCount()")

	flgLogRotCountVal := flgVal.(uint)
	if flgLogRotCountVal == 0 {
		err := fmt.Errorf(gmess.ERR_CANNOTBEZERO, "--"+flgnm.FLG_LOGROTATIONCOUNT)
		lg.Error(err)
		return err
	}

	viper.Set(cfg.CONFIGLOGROTATIONCOUNT, flgLogRotCountVal)
	err := viper.WriteConfig()
	if err != nil {
		lg.Error(err)
		return err
	}
	fmt.Println(cmn.GminMessage(fmt.Sprintf(gmess.INFO_LOGROTATIONCOUNTSET, flgLogRotCountVal)))
	lg.Infof(gmess.INFO_LOGROTATIONCOUNTSET, flgLogRotCountVal)

	return nil
}

func setCfgLogRotTime(flgVal interface{}) error {
	lg.Debug("starting setCfgLogRotTime()")
	defer lg.Debug("finished setCfgLogRotTime()")

	flgLogRotTimeVal := flgVal.(int)
	if flgLogRotTimeVal == 0 {
		err := fmt.Errorf(gmess.ERR_CANNOTBEZERO, "--"+flgnm.FLG_LOGROTATIONTIME)
		lg.Error(err)
		return err
	}

	viper.Set(cfg.CONFIGLOGROTATIONTIME, flgLogRotTimeVal)
	err := viper.WriteConfig()
	if err != nil {
		lg.Error(err)
		return err
	}
	fmt.Println(cmn.GminMessage(fmt.Sprintf(gmess.INFO_LOGROTATIONTIMESET, flgLogRotTimeVal)))
	lg.Infof(gmess.INFO_LOGROTATIONTIMESET, flgLogRotTimeVal)

	return nil
}
