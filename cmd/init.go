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
	"log"
	"path/filepath"
	"strconv"

	valid "github.com/asaskevich/govalidator"
	"github.com/mitchellh/go-homedir"
	cfg "github.com/plusworx/gmin/utils/config"
	gmess "github.com/plusworx/gmin/utils/gminmessages"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Creates gmin config file",
	Long: `Asks for admin email address, credentials path, customer id, config file path, log file path,
log rotation count and log rotation time before creating gmin configuration file (.gmin.yaml) in the default
location (the home directory of the user) or at the specified path.

Log rotation count specifies the maximum number of log files that will be kept before the oldest is deleted.
Log rotation time specifies the amount of time before a new log file is created.

Defaults
--------
Credentials Path: <home directory>
Customer ID: my_customer
Config File Path: <home directory>
Log File Path: <home directory>
Log Rotation Count: 7
Log Rotation Time: 86400`,
	RunE: doInit,
}

func askForConfigPath() string {
	var response string

	fmt.Print("Please enter a full config file path (q to quit)\n(Press <Enter> for default value): ")

	_, err := fmt.Scanln(&response)

	if response == "" {
		return response
	}

	if err != nil {
		fmt.Println(gmess.ERR_INVALIDCONFIGPATH)
		return askForConfigPath()
	}

	return response
}

func askForCredentialPath() string {
	var response string

	fmt.Print("Please enter a full credentials file path (q to quit)\n(Press <Enter> for default value): ")

	_, err := fmt.Scanln(&response)

	if response == "" {
		return response
	}

	if err != nil {
		fmt.Println(gmess.ERR_INVALIDCREDPATH)
		return askForCredentialPath()
	}

	return response
}

func askForCustomerID() string {
	var response string

	fmt.Print("Please enter customer ID (q to quit)\n(Press <Enter> for default value): ")

	_, err := fmt.Scanln(&response)

	if response == "" {
		return response
	}

	if err != nil {
		fmt.Println(gmess.ERR_INVALIDCUSTID)
		return askForCustomerID()
	}

	return response
}

func askForEmail() string {
	var response string

	fmt.Print("Please enter an administrator email address (q to quit): ")

	_, err := fmt.Scanln(&response)
	if err != nil {
		fmt.Println(gmess.ERR_ADMINEMAILREQUIRED)
		return askForEmail()
	}

	if response == "q" {
		return response
	}

	ok := valid.IsEmail(response)
	if !ok {
		fmt.Println(gmess.ERR_INVALIDADMINEMAIL)
		return askForEmail()
	}

	return response
}

func askForLogPath() string {
	var response string

	fmt.Print("Please enter a full log file path (q to quit)\n(Press <Enter> for default value): ")

	_, err := fmt.Scanln(&response)

	if response == "" {
		return response
	}

	if err != nil {
		fmt.Println(gmess.ERR_INVALIDLOGPATH)
		return askForLogPath()
	}

	return response
}

func askForLogRotationCount() string {
	var response string

	fmt.Print("Please enter a log rotation count (q to quit)\n(Press <Enter> for default value): ")

	_, err := fmt.Scanln(&response)

	if response == "" || response == "q" {
		return response
	}

	if err != nil {
		fmt.Println(gmess.ERR_INVALIDLOGROTATIONCOUNT)
		return askForLogRotationCount()
	}

	ok := valid.IsNumeric(response)
	if !ok {
		fmt.Println(gmess.ERR_MUSTBENUMBER)
		return askForLogRotationCount()
	}

	return response
}

func askForLogRotationTime() string {
	var response string

	fmt.Print("Please enter a log rotation time (q to quit)\n(Press <Enter> for default value): ")

	_, err := fmt.Scanln(&response)

	if response == "" || response == "q" {
		return response
	}

	if err != nil {
		fmt.Println(gmess.ERR_INVALIDLOGROTATIONTIME)
		return askForLogRotationTime()
	}

	ok := valid.IsNumeric(response)
	if !ok {
		fmt.Println(gmess.ERR_MUSTBENUMBER)
		return askForLogRotationTime()
	}

	return response
}

func doAdminEmail(ans *cfg.Answers) bool {
	ans.AdminEmail = askForEmail()
	if ans.AdminEmail == "q" {
		fmt.Println(gmess.INFO_INITCANCELLED)
		return true
	}
	return false
}

func doConfigPath(ans *cfg.Answers) bool {
	ans.ConfigPath = askForConfigPath()
	if ans.ConfigPath == "q" {
		fmt.Println(gmess.INFO_INITCANCELLED)
		return true
	}
	if ans.ConfigPath == "" {
		hmDir, err := homedir.Dir()
		if err != nil {
			log.Fatal(err)
		}
		ans.ConfigPath = hmDir
	}
	return false
}

func doCredentialPath(ans *cfg.Answers) bool {
	ans.CredentialPath = askForCredentialPath()
	if ans.CredentialPath == "q" {
		fmt.Println(gmess.INFO_INITCANCELLED)
		return true
	}
	if ans.CredentialPath == "" {
		hmDir, err := homedir.Dir()
		if err != nil {
			log.Fatal(err)
		}
		ans.CredentialPath = hmDir
	}
	return false
}

func doCustomerID(ans *cfg.Answers) bool {
	ans.CustomerID = askForCustomerID()
	if ans.CustomerID == "q" {
		fmt.Println(gmess.INFO_INITCANCELLED)
		return true
	}
	if ans.CustomerID == "" {
		ans.CustomerID = cfg.DEFAULTCUSTID
	}
	return false
}

func doInit(cmd *cobra.Command, args []string) error {
	answers := new(cfg.Answers)

	quit := doAdminEmail(answers)
	if quit {
		return nil
	}
	quit = doConfigPath(answers)
	if quit {
		return nil
	}
	quit = doCredentialPath(answers)
	if quit {
		return nil
	}
	quit = doCustomerID(answers)
	if quit {
		return nil
	}
	quit = doLogPath(answers)
	if quit {
		return nil
	}
	quit = doLogRotationCount(answers)
	if quit {
		return nil
	}
	quit = doLogRotationTime(answers)
	if quit {
		return nil
	}

	cfgFile := cfg.File{Administrator: answers.AdminEmail, CredentialPath: answers.CredentialPath,
		CustomerID: answers.CustomerID, LogPath: answers.LogPath, LogRotationCount: answers.LogRotationCount,
		LogRotationTime: answers.LogRotationTime}

	path := filepath.Join(filepath.ToSlash(answers.ConfigPath), cfg.CONFIGFILENAME)

	cfgYaml, err := yaml.Marshal(&cfgFile)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(path, cfgYaml, 0644)
	if err != nil {
		return err
	}

	fmt.Println(gmess.INFO_INITCOMPLETED)

	return nil
}

func doLogPath(ans *cfg.Answers) bool {
	ans.LogPath = askForLogPath()
	if ans.LogPath == "q" {
		fmt.Println(gmess.INFO_INITCANCELLED)
		return true
	}
	if ans.LogPath == "" {
		hmDir, err := homedir.Dir()
		if err != nil {
			log.Fatal(err)
		}
		ans.LogPath = hmDir
	}
	return false
}

func doLogRotationCount(ans *cfg.Answers) bool {
	retVal := askForLogRotationCount()
	if retVal == "q" {
		fmt.Println(gmess.INFO_INITCANCELLED)
		return true
	}
	if retVal == "" {
		retVal = "0"
	}
	retInt, err := strconv.Atoi(retVal)
	if err != nil {
		log.Fatal(err)
	}
	ans.LogRotationCount = uint(retInt)

	if ans.LogRotationCount == 0 {
		ans.LogRotationCount = cfg.DEFAULTLOGROTATIONCOUNT
	}

	return false
}

func doLogRotationTime(ans *cfg.Answers) bool {
	retVal := askForLogRotationTime()
	if retVal == "q" {
		fmt.Println(gmess.INFO_INITCANCELLED)
		return true
	}
	if retVal == "" {
		retVal = "0"
	}
	retInt, err := strconv.Atoi(retVal)
	if err != nil {
		log.Fatal(err)
	}
	ans.LogRotationTime = retInt

	if ans.LogRotationTime == 0 {
		ans.LogRotationTime = cfg.DEFAULTLOGROTATIONTIME
	}

	return false
}

func init() {
	rootCmd.AddCommand(initCmd)
}
