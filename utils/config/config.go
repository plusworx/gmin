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

package config

import (
	"fmt"

	"github.com/spf13/viper"
)

const (
	// CredentialFile holds service account credentials
	CredentialFile string = "gmin_credentials"
	// ConfigFile is configuration file name
	ConfigFile string = ".gmin.yaml"
)

// File holds configuration data
type File struct {
	Administrator  string `yaml:"administrator"`
	CredentialPath string `yaml:"credentialpath"`
	CustomerID     string `yaml:"customerid"`
}

// ReadConfigString gets a string item from config file
func ReadConfigString(s string) (string, error) {
	var err error

	str := viper.GetString(s)
	if str == "" {
		err = fmt.Errorf("gmin: error - %v not found in config file", s)
	}

	return str, err
}
