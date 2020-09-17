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
	"errors"
	"fmt"
	"log"
	"os"
	"strings"

	"time"

	"github.com/mitchellh/go-homedir"
	cmn "github.com/plusworx/gmin/utils/common"
	cfg "github.com/plusworx/gmin/utils/config"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	adminEmail       string
	assetID          string
	attrs            string
	archived         bool
	blockInherit     bool
	cfgFile          string
	credentialPath   string
	changePassword   bool
	composite        bool
	count            bool
	customerID       string
	customField      string
	deleted          bool
	deliverySetting  string
	domain           string
	filter           string
	firstName        string
	forceSend        string
	format           string
	gal              bool
	group            string
	groupDesc        string
	groupEmail       string
	groupName        string
	inputFile        string
	lastName         string
	location         string
	logger           *zap.SugaredLogger
	logLevel         string
	logPath          string
	maxResults       int64
	noChangePassword bool
	noGAL            bool
	notes            string
	orderBy          string
	orgUnit          string
	orgUnitDesc      string
	orgUnitName      string
	pages            string
	parentOUPath     string
	password         string
	projection       string
	query            string
	queryable        bool
	sheetRange       string
	reason           string
	recoveryEmail    string
	recoveryPhone    string
	role             string
	searchType       string
	sortOrder        string
	suspended        bool
	unblockInherit   bool
	unSuspended      bool
	userEmail        string
	userKey          string
	viewType         string
)

var rootCmd = &cobra.Command{
	Use:   "gmin",
	Short: "gmin is a CLI for administering G Suite domains",
	Long: `gmin is a commandline interface (CLI) that enables the 
	administration of G Suite domains. It aims to provide tools that
	make G Suite administration from the commandline more manageable.`,
	SilenceErrors: true,
	SilenceUsage:  true,
	Version:       "v0.8.0",
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(formatError(err.Error()))
		os.Exit(1)
	}
}

func formatError(errStr string) string {
	var retStr string

	if strings.HasPrefix(errStr, "gmin") {
		retStr = cmn.Timestamp() + " " + errStr
		return retStr
	}

	retStr = cmn.Timestamp() + " " + "gmin: error - " + errStr
	return retStr
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.gmin.yaml)")
}

func initConfig() {
	viper.SetEnvPrefix(cfg.EnvPrefix)

	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := homedir.Dir()
		if err != nil {
			log.Fatal(err)
		}

		viper.AddConfigPath(home)
		viper.SetConfigName(cfg.ConfigFilePrefix)
	}

	viper.AutomaticEnv()
	viper.ReadInConfig()
}

func setupLogging(loglevel string) (*zap.Logger, error) {
	var (
		err  error
		zlog *zap.Logger
	)

	zconf := zap.NewProductionConfig()

	switch loglevel {
	case "info":
		zconf.Level.SetLevel(zapcore.InfoLevel)
	case "warn":
		zconf.Level.SetLevel(zapcore.WarnLevel)
	case "error":
		zconf.Level.SetLevel(zapcore.ErrorLevel)
	case "fatal":
		zconf.Level.SetLevel(zapcore.FatalLevel)
	default:
		return nil, errors.New("gmin: error - loglevel " + loglevel + " is invalid")
	}

	zconf.EncoderConfig.EncodeTime = zapcore.TimeEncoder(func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
		enc.AppendString(t.Local().Format("2006-01-02T15:04:05Z0700"))
	})
	logpath, err := cfg.ReadConfigString("logpath")
	if err != nil {
		log.Fatal(err)
	}

	lpaths := cmn.ParseTildeField(logpath)
	zconf.OutputPaths = lpaths

	if loglevel == "debug" {
		zlog, err = zap.NewDevelopment()
	} else {
		zlog, err = zconf.Build()
	}
	if err != nil {
		return nil, err
	}
	return zlog, nil
}
