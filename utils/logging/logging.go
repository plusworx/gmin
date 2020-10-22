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
	"os"
	"path/filepath"
	"strconv"
	"time"

	rlogs "github.com/lestrrat-go/file-rotatelogs"
	cmn "github.com/plusworx/gmin/utils/common"
	cfg "github.com/plusworx/gmin/utils/config"
	gmess "github.com/plusworx/gmin/utils/gminmessages"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var logger *zap.SugaredLogger

func createLogger(loglevel string) (*zap.Logger, error) {
	var (
		encCfg zapcore.EncoderConfig
		err    error
		zlog   *zap.Logger
	)

	// Get logpath
	logpath := viper.GetString(cfg.CONFIGLOGPATH)
	if logpath == "" {
		// Look for environment variable
		logpath = os.Getenv(cfg.ENVPREFIX + cfg.ENVVARLOGPATH)
		if logpath == "" {
			err = fmt.Errorf(gmess.ERR_NOTFOUNDINCONFIG, cfg.CONFIGLOGPATH)
			return nil, err
		}
	}

	// Get logging properties
	logRotationCount := viper.GetUint(cfg.CONFIGLOGROTATIONCOUNT)
	if logRotationCount == 0 {
		envVal := os.Getenv(cfg.ENVPREFIX + cfg.ENVVARLOGROTATIONCOUNT)
		if envVal == "" {
			err = fmt.Errorf(gmess.ERR_NOTFOUNDINCONFIG, cfg.CONFIGLOGROTATIONCOUNT)
			return nil, err
		}
		retNum, err := strconv.Atoi(envVal)
		if err != nil {
			return nil, err
		}
		retUint := uint(retNum)
		if retUint == 0 {
			logRotationCount = cfg.DEFAULTLOGROTATIONCOUNT
		} else {
			logRotationCount = retUint
		}
	}

	logRotationTime := viper.GetInt(cfg.CONFIGLOGROTATIONTIME)
	if logRotationTime == 0 {
		envVal := os.Getenv(cfg.ENVPREFIX + cfg.ENVVARLOGROTATIONTIME)
		if envVal == "" {
			err = fmt.Errorf(gmess.ERR_NOTFOUNDINCONFIG, cfg.CONFIGLOGROTATIONTIME)
			return nil, err
		}
		retNum, err := strconv.Atoi(envVal)
		if err != nil {
			return nil, err
		}
		if retNum == 0 {
			logRotationTime = cfg.DEFAULTLOGROTATIONTIME
		} else {
			logRotationTime = retNum
		}
	}

	// Configure log rotation
	t := time.Now()

	fmtLogName := fmt.Sprintf(cfg.LOGFILE,
		t.Year(), t.Month(), t.Day())

	rotator, err := rlogs.New(
		// filepath.Join(filepath.ToSlash(logpath), cfg.LOGFILE),
		filepath.Join(filepath.ToSlash(logpath), fmtLogName),
		rlogs.WithLinkName(filepath.Join(filepath.ToSlash(logpath), "current")),
		rlogs.WithMaxAge(-1),
		rlogs.WithRotationCount(logRotationCount),
		rlogs.WithRotationTime(time.Duration(logRotationTime)*time.Second))
	if err != nil {
		return nil, err
	}

	// Add log rotation to zap logging
	w := zapcore.AddSync(rotator)

	// initialize the JSON encoding config
	encCfg.CallerKey = "caller"
	encCfg.EncodeCaller = zapcore.ShortCallerEncoder
	encCfg.EncodeDuration = zapcore.SecondsDurationEncoder
	encCfg.EncodeLevel = zapcore.LowercaseLevelEncoder
	encCfg.EncodeTime = zapcore.ISO8601TimeEncoder
	encCfg.FunctionKey = zapcore.OmitKey
	encCfg.LevelKey = "level"
	encCfg.LineEnding = zapcore.DefaultLineEnding
	encCfg.MessageKey = "msg"
	encCfg.NameKey = "logger"
	encCfg.StacktraceKey = "stacktrace"
	encCfg.TimeKey = "ts"

	zlevel, err := getLogLevel(loglevel)
	if err != nil {
		return nil, err
	}

	// Create zap logger object
	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(encCfg),
		w,
		zlevel)

	zlog = zap.New(core)

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

// getLogLevel gets requested logging zapcore.Level
func getLogLevel(loglevel string) (zapcore.Level, error) {
	switch loglevel {
	case "debug":
		return zapcore.DebugLevel, nil
	case "info":
		return zapcore.InfoLevel, nil
	case "error":
		return zapcore.ErrorLevel, nil
	case "warn":
		return zapcore.WarnLevel, nil
	default:
		return -99, fmt.Errorf(gmess.ERR_INVALIDLOGLEVEL, loglevel)
	}
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
	cfg.Logger = logger
	cmn.Logger = logger
	return nil
}

// Warnw wrapper function for logger.Warnw
func Warnw(msg string, keysandvalues ...interface{}) {
	logger.Warnw(msg, keysandvalues...)
}
