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
	"log"
	"os"
	"strings"

	"time"

	"github.com/mitchellh/go-homedir"
	cmn "github.com/plusworx/gmin/utils/common"
	cfg "github.com/plusworx/gmin/utils/config"
	flgnm "github.com/plusworx/gmin/utils/flagnames"
	gmess "github.com/plusworx/gmin/utils/gminmessages"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	adminEmail       string
	approveMems      string
	archiveOnly      bool
	assetID          string
	assistContent    string
	attrs            string
	archived         bool
	banUser          string
	blockInherit     bool
	cfgFile          string
	collabInbox      bool
	contactOwner     string
	credentialPath   string
	changePassword   bool
	composite        bool
	count            bool
	customerID       string
	customField      string
	deleted          bool
	delFormat        string
	deliverySetting  string
	denyNotification bool
	denyText         string
	discoverGroup    string
	domain           string
	extMems          bool
	filter           string
	firstName        string
	footerText       string
	forceSend        string
	format           string
	gal              bool
	group            string
	groupDesc        string
	groupEmail       string
	groupName        string
	incFooter        bool
	inputFile        string
	isArchived       bool
	join             string
	language         string
	lastName         string
	leave            string
	location         string
	logger           *zap.SugaredLogger
	logLevel         string
	logPath          string
	maxResults       int64
	messageMod       string
	modContent       string
	modMems          string
	notes            string
	orderBy          string
	orgUnit          string
	orgUnitDesc      string
	orgUnitName      string
	pages            string
	parentOUPath     string
	password         string
	postAsGroup      bool
	postMessage      string
	projection       string
	query            string
	queryable        bool
	sheetRange       string
	reason           string
	recoveryEmail    string
	recoveryPhone    string
	repliesOnTop     bool
	replyEmail       string
	replyTo          string
	role             string
	searchType       string
	silent           bool
	sortOrder        string
	spamMod          string
	suspended        bool
	userEmail        string
	userKey          string
	viewGroup        string
	viewMems         string
	viewType         string
	webPosting       bool
)

var rootCmd = &cobra.Command{
	Use:   "gmin",
	Short: "gmin is a CLI for administering Google Workspace domains",
	Long: `gmin is a friendly Google Workspace administration CLI (command line interface)
written in Go. It provides tools that make Google Workspace administration from the
command line more manageable.`,
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

	rootCmd.PersistentFlags().StringVar(&cfgFile, flgnm.FLG_CONFIG, "", "config file (default is $HOME/.gmin.yaml)")
}

func initConfig() {
	viper.SetEnvPrefix(cfg.ENVPREFIX)

	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := homedir.Dir()
		if err != nil {
			log.Fatal(err)
		}

		viper.AddConfigPath(home)
		viper.SetConfigName(cfg.CONFIGFILEPREFIX)
	}

	viper.AutomaticEnv()
	viper.ReadInConfig()
}

func preRun(cmd *cobra.Command, args []string) error {
	// Set loglevel
	logFlgVal, err := cmd.Flags().GetString(flgnm.FLG_LOGLEVEL)
	if err != nil {
		return err
	}
	zlog, err := setupLogging(logFlgVal)
	if err != nil {
		return err
	}
	logger = zlog.Sugar()
	// Suppress console output according to silent flag
	silentFlgVal, err := cmd.Flags().GetBool(flgnm.FLG_SILENT)
	if err != nil {
		return err
	}
	if silentFlgVal {
		os.Stdout = nil
		os.Stderr = nil
	}
	// Get gmin admin email address
	admAddr, err := cfg.ReadConfigString(cfg.CONFIGADMIN)
	if err != nil {
		return err
	}
	// Log current user
	logger.Infow("User Information",
		"gmin admin", admAddr,
		"Username", cmn.Username(),
		"Hostname", cmn.Hostname(),
		"IP Address", cmn.IPAddress())
	return nil
}

func setupLogging(loglevel string) (*zap.Logger, error) {
	var (
		err  error
		zlog *zap.Logger
	)

	zconf := zap.NewProductionConfig()

	switch loglevel {
	case "debug":
		zconf.Level.SetLevel(zapcore.DebugLevel)
	case "info":
		zconf.Level.SetLevel(zapcore.InfoLevel)
	case "error":
		zconf.Level.SetLevel(zapcore.ErrorLevel)
	case "warn":
		zconf.Level.SetLevel(zapcore.WarnLevel)
	default:
		return nil, fmt.Errorf(gmess.ERR_INVALIDLOGLEVEL, loglevel)
	}

	zconf.EncoderConfig.EncodeTime = zapcore.TimeEncoder(func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
		enc.AppendString(t.Local().Format("2006-01-02T15:04:05Z0700"))
	})
	logpath, err := cfg.ReadConfigString(cfg.CONFIGLOGPATH)
	if err != nil {
		return nil, err
	}

	lpaths := cmn.ParseTildeField(logpath)
	zconf.OutputPaths = lpaths

	zlog, err = zconf.Build()
	if err != nil {
		return nil, err
	}
	return zlog, nil
}
