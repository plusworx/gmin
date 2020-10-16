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

package logging

import (
	"fmt"
	"time"

	cmn "github.com/plusworx/gmin/utils/common"
	cfg "github.com/plusworx/gmin/utils/config"
	gmess "github.com/plusworx/gmin/utils/gminmessages"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var logger *zap.SugaredLogger

// createLogger sets log level and creates logger
func createLogger(loglevel string) (*zap.Logger, error) {
	var (
		err  error
		zlog *zap.Logger
	)

	zconf := zap.NewProductionConfig()
	err = setLogLevel(zconf, loglevel)

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

// Debug wrapper function for logger.Debug
func Debug(args ...interface{}) {
	logger.Debug(args...)
}

// Debugw wrapper function for logger.Debugw
func Debugw(msg string, keysandvalues ...interface{}) {
	logger.Debugw(msg, keysandvalues...)
}

// Error wrapper function for logger.Error
func Error(args ...interface{}) {
	logger.Error(args...)
}

// GetLogger gets logger
func GetLogger() *zap.SugaredLogger {
	return logger
}

// Info wrapper function for logger.Info
func Info(args ...interface{}) {
	logger.Info(args...)
}

// Infof wrapper function for logger.Infof
func Infof(template string, args ...interface{}) {
	logger.Infof(template, args...)
}

// Infow wrapper function for logger.Infow
func Infow(msg string, keysandvalues ...interface{}) {
	logger.Infow(msg, keysandvalues...)
}

// InitLogging initialises logger
func InitLogging(loglevel string) error {
	zlog, err := createLogger(loglevel)
	if err != nil {
		return err
	}
	logger = zlog.Sugar()
	return nil
}

// setLogLevel sets log level in zconf
func setLogLevel(zconf zap.Config, loglevel string) error {
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
		return fmt.Errorf(gmess.ERR_INVALIDLOGLEVEL, loglevel)
	}

	return nil
}

// Warnw wrapper function for logger.Warnw
func Warnw(msg string, keysandvalues ...interface{}) {
	logger.Warnw(msg, keysandvalues...)
}
