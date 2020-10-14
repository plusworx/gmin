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
	Long: `Asks for admin email address, credentials path, customer id, config file path and log file path
before creating gmin configuration file (.gmin.yaml) in the default location (the home directory of the user)
or at the specified path.

N.B. Log file path must include the name of the log file and can contain multiple paths separated by tilde '~'.
stdout is a valid path and will direct output to the console.

Example log paths
-----------------
/tmp/logs/gmin.log
stdout~/home/myuser/logs/gmin.log

Defaults
--------
Credentials Path: <home directory>
Customer ID: my_customer
Config File Path: <home directory>
Log File Path: <home directory>/gmin.log`,
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
		fmt.Println(gmess.ERRINVALIDCONFIGPATH)
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
		fmt.Println(gmess.ERRINVALIDCREDPATH)
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
		fmt.Println(gmess.ERRINVALIDCUSTID)
		return askForCustomerID()
	}

	return response
}

func askForEmail() string {
	var response string

	fmt.Print("Please enter an administrator email address (q to quit): ")

	_, err := fmt.Scanln(&response)
	if err != nil {
		fmt.Println(gmess.ERRADMINEMAILREQUIRED)
		return askForEmail()
	}

	if response == "q" {
		return response
	}

	ok := valid.IsEmail(response)
	if !ok {
		fmt.Println(gmess.ERRINVALIDADMINEMAIL)
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
		fmt.Println(gmess.ERRINVALIDLOGPATH)
		return askForLogPath()
	}

	return response
}

func doInit(cmd *cobra.Command, args []string) error {
	answers := struct {
		AdminEmail     string
		ConfigPath     string
		CredentialPath string
		CustomerID     string
		LogPath        string
	}{}

	answers.AdminEmail = askForEmail()
	if answers.AdminEmail == "q" {
		fmt.Println(gmess.INFOINITCANCELLED)
		return nil
	}
	answers.ConfigPath = askForConfigPath()
	if answers.ConfigPath == "q" {
		fmt.Println(gmess.INFOINITCANCELLED)
		return nil
	}
	answers.CredentialPath = askForCredentialPath()
	if answers.CredentialPath == "q" {
		fmt.Println(gmess.INFOINITCANCELLED)
		return nil
	}
	answers.CustomerID = askForCustomerID()
	if answers.CustomerID == "q" {
		fmt.Println(gmess.INFOINITCANCELLED)
		return nil
	}
	answers.LogPath = askForLogPath()
	if answers.LogPath == "q" {
		fmt.Println(gmess.INFOINITCANCELLED)
		return nil
	}

	if answers.ConfigPath == "" {
		hmDir, err := homedir.Dir()
		if err != nil {
			log.Fatal(err)
		}
		answers.ConfigPath = hmDir
	}

	if answers.CredentialPath == "" {
		hmDir, err := homedir.Dir()
		if err != nil {
			log.Fatal(err)
		}
		answers.CredentialPath = hmDir
	}

	if answers.CustomerID == "" {
		answers.CustomerID = cfg.DEFAULTCUSTID
	}

	if answers.LogPath == "" {
		hmDir, err := homedir.Dir()
		if err != nil {
			log.Fatal(err)
		}
		answers.LogPath = filepath.Join(filepath.ToSlash(hmDir), cfg.LOGFILE)
	}

	cfgFile := cfg.File{Administrator: answers.AdminEmail, CredentialPath: answers.CredentialPath,
		CustomerID: answers.CustomerID, LogPath: answers.LogPath}

	path := filepath.Join(filepath.ToSlash(answers.ConfigPath), cfg.CONFIGFILENAME)

	cfgYaml, err := yaml.Marshal(&cfgFile)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(path, cfgYaml, 0644)
	if err != nil {
		return err
	}

	fmt.Println(gmess.INFOINITCOMPLETED)

	return nil
}

func init() {
	rootCmd.AddCommand(initCmd)
}
