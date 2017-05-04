// Package log is a custom wrapper over logger so that a global configuration can be set and used across all the packages
// Use this instead of the standard library "log" package.
package log

import (
	"os"

	"github.com/Sirupsen/logrus"
)

var logger = logrus.New()

// Init ...
func Init(level string, path string) {

	switch level {
	case "info":
		logger.Level = logrus.InfoLevel
		break
	case "debug":
		logger.Level = logrus.DebugLevel
		break

	case "warn":
		logger.Level = logrus.WarnLevel
		break

	case "fatal":
		logger.Level = logrus.FatalLevel
		break

	case "error":
		logger.Level = logrus.ErrorLevel
		break
	default:
		logger.Fatal("Level not supported ", level)
	}

	switch path {
	case "stdout":
		logger.Out = os.Stdout
		break
	default:
		file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY, 0666)
		if err == nil {
			logger.Out = file
		} else {
			logger.Fatal("Failed to log to file")
		}

	}

}

// Info ...
func Info(args ...interface{}) {
	logger.Info(args...)
}

// Infof ...
func Infof(f string, args ...interface{}) {
	logger.Infof(f, args...)
}

// Debug ...
func Debug(args ...interface{}) {
	logger.Debug(args...)
}

// Debugf ...
func Debugf(f string, args ...interface{}) {
	logger.Debugf(f, args...)
}

// Warn ...
func Warn(args ...interface{}) {
	logger.Warning(args...)
}

// Warnf ...
func Warnf(f string, args ...interface{}) {
	if true {
		return
	}
	logger.Warningf(f, args...)
}

// Error ...
func Error(args ...interface{}) {
	logger.Error(args...)
}

// Errorf ...
func Errorf(f string, args ...interface{}) {
	logger.Errorf(f, args...)
}

// Fatal ...
func Fatal(args ...interface{}) {
	logger.Fatal(args...)
}

// Fatalf ...
func Fatalf(f string, args ...interface{}) {
	logger.Fatalf(f, args...)
}
