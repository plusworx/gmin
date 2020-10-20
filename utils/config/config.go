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

	gmess "github.com/plusworx/gmin/utils/gminmessages"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

const (
	// CONFIGADMIN is config file administrator variable name
	CONFIGADMIN string = "administrator"
	// CONFIGCUSTID is config file customer id variable name
	CONFIGCUSTID string = "customerid"
	// CONFIGCREDPATH is config file credential path variable name
	CONFIGCREDPATH string = "credentialpath"
	// CONFIGFILENAME is configuration file name
	CONFIGFILENAME string = ".gmin.yaml"
	// CONFIGFILEPREFIX is name of gmin config file without the .yaml suffix
	CONFIGFILEPREFIX string = ".gmin"
	// CONFIGLOGPATH is config file log path variable name
	CONFIGLOGPATH string = "logpath"
	// CREDENTIALFILE service account credentials file name
	CREDENTIALFILE string = "gmin_credentials"
	// DEFAULTCUSTID is default customer id value
	DEFAULTCUSTID string = "my_customer"
	// ENVPREFIX is prefix for gmin environment variables
	ENVPREFIX string = "GMIN"
	// ENVVARADMIN is gmin administrator environment variable suffix
	ENVVARADMIN string = "_ADMINISTRATOR"
	// ENVVARCREDPATH is gmin credential path environment variable suffix
	ENVVARCREDPATH string = "_CREDENTIALPATH"
	// ENVVARCUSTID is gmin custormer id environment variable suffix
	ENVVARCUSTID string = "_CUSTOMERID"
	// ENVVARLOGPATH is gmin log path environment variable suffix
	ENVVARLOGPATH string = "_LOGPATH"
	// LOGFILE is default gmin log file name
	LOGFILE string = "gmin-%d-%02d-%02d.log"
)

// File holds configuration data
type File struct {
	Administrator  string `yaml:"administrator"`
	CredentialPath string `yaml:"credentialpath"`
	CustomerID     string `yaml:"customerid"`
	LogPath        string `yaml:"logpath"`
}

// Logger passed from logging package
var Logger *zap.SugaredLogger

// ReadConfigString gets a string item from config file
func ReadConfigString(key string) (string, error) {
	Logger.Debugw("starting ReadConfigString()",
		"key", key)
	defer Logger.Debug("finished ReadConfigString()")

	var err error

	// Use viper directly here because logging may not be set up yet
	str := viper.GetString(key)
	if str == "" {
		err = fmt.Errorf(gmess.ERR_NOTFOUNDINCONFIG, key)
		Logger.Error(err)
	}
	return str, err
}
