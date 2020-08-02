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
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var setConfigCmd = &cobra.Command{
	Use:     "config",
	Aliases: []string{"cfg"},
	Short:   "Sets gmin configuration information",
	RunE:    doSetConfig,
}

func doSetConfig(cmd *cobra.Command, args []string) error {
	if customerID != "" {
		viper.Set("customerid", customerID)
		err := viper.WriteConfig()
		if err != nil {
			return err
		}
		fmt.Printf("**** gmin: customer id set to %v ****\n", customerID)
	}

	if adminEmail != "" {
		ok := valid.IsEmail(adminEmail)
		if !ok {
			return fmt.Errorf("gmin: error - %v is not a valid email address", adminEmail)
		}
		viper.Set("administrator", adminEmail)
		err := viper.WriteConfig()
		if err != nil {
			return err
		}
		fmt.Printf("**** gmin: administrator set to %v ****\n", adminEmail)
	}

	if credentialPath != "" {
		viper.Set("credentialpath", credentialPath)
		err := viper.WriteConfig()
		if err != nil {
			return err
		}
		fmt.Printf("**** gmin: service account credential path set to %v ****\n", credentialPath)
	}

	if adminEmail == "" && customerID == "" && credentialPath == "" {
		cmd.Help()
	}

	return nil
}

func init() {
	setCmd.AddCommand(setConfigCmd)

	setConfigCmd.Flags().StringVarP(&adminEmail, "admin", "a", "", "administrator email address")
	setConfigCmd.Flags().StringVarP(&customerID, "customerid", "c", "", "customer id for domain")
	setConfigCmd.Flags().StringVarP(&credentialPath, "credentialpath", "p", "", "service account credential file path")
}
