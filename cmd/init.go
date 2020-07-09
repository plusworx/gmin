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
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Creates gmin config file",
	Long: `Asks for admin email address and config file path and creates gmin
	configuration file (.gmin.yaml) in the default location,
	which is the home directory of the user, or at the specified path.`,
	RunE: doInit,
}

func askForConfigPath() string {
	var response string

	fmt.Print("Please enter a full config file path\n(Press <Enter> for default value): ")

	_, err := fmt.Scanln(&response)
	if response == "" {
		return response
	}

	if err != nil {
		fmt.Println(err)
		return askForConfigPath()
	}

	return response
}

func askForCustomerID() string {
	var response string

	fmt.Print("Please enter customer ID\n(Press <Enter> for default value): ")

	_, err := fmt.Scanln(&response)
	if response == "" {
		return response
	}

	if err != nil {
		fmt.Println(err)
		return askForCustomerID()
	}

	return response
}

func askForEmail() string {
	var response string

	fmt.Print("Please enter an administrator email address: ")

	_, err := fmt.Scanln(&response)
	if err != nil {
		fmt.Println("an email address is required - try again")
		return askForEmail()
	}

	ok := valid.IsEmail(response)
	if !ok {
		fmt.Println("invalid email address - try again")
		return askForEmail()
	}

	return response
}

func doInit(cmd *cobra.Command, args []string) error {
	answers := struct {
		AdminEmail string
		ConfigPath string
		CustomerID string
	}{}

	answers.AdminEmail = askForEmail()
	answers.ConfigPath = askForConfigPath()
	answers.CustomerID = askForCustomerID()

	if answers.ConfigPath == "" {
		hmDir, err := homedir.Dir()
		if err != nil {
			log.Fatal(err)
		}

		answers.ConfigPath = hmDir
	}

	if answers.CustomerID == "" {
		answers.CustomerID = "my_customer"
	}

	cfgFile := cfg.File{Administrator: answers.AdminEmail, CustomerID: answers.CustomerID}

	path := filepath.Join(filepath.ToSlash(answers.ConfigPath), cfg.FileName)

	cfgYaml, err := yaml.Marshal(&cfgFile)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(path, cfgYaml, 0644)
	if err != nil {
		return err
	}

	fmt.Println("**** gmin: init completed successfully ****")

	return nil
}

func init() {
	rootCmd.AddCommand(initCmd)
}
