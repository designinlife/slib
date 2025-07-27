package glog

import "github.com/designinlife/slib/types"

var logger types.Logger

func init() {
	InitDefaultLogger()
}

// InitDefaultLogger Initialize default Logger.
// Note: This method should be called in the init function of the main entry file when used in a project.
func InitDefaultLogger() {
	logger = initSugaredLogger()
	// logger = NewStdLogger(WithUseTextHandler())
}

func InitSugarLogger(opts ...SugarLoggerOption) {
	logger = initSugaredLogger(opts...)
}

func InitStdLogger(opts ...StdLoggerOption) {
	logger = NewStdLogger(opts...)
}

func Debug(args ...interface{}) {
	logger.Debug(args...)
}

func Info(args ...interface{}) {
	logger.Info(args...)
}

func Warn(args ...interface{}) {
	logger.Warn(args...)
}

func Error(args ...interface{}) {
	logger.Error(args...)
}

func Panic(args ...interface{}) {
	logger.Panic(args...)
}

func Fatal(args ...interface{}) {
	logger.Fatal(args...)
}

func Debugf(format string, args ...interface{}) {
	logger.Debugf(format, args...)
}

func Infof(format string, args ...interface{}) {
	logger.Infof(format, args...)
}

func Warnf(format string, args ...interface{}) {
	logger.Warnf(format, args...)
}

func Errorf(format string, args ...interface{}) {
	logger.Errorf(format, args...)
}

func Panicf(format string, args ...interface{}) {
	logger.Panicf(format, args...)
}

func Fatalf(format string, args ...interface{}) {
	logger.Fatalf(format, args...)
}

func Debugln(args ...interface{}) {
	logger.Debugln(args...)
}

func Infoln(args ...interface{}) {
	logger.Infoln(args...)
}

func Warnln(args ...interface{}) {
	logger.Warnln(args...)
}

func Errorln(args ...interface{}) {
	logger.Errorln(args...)
}

func Panicln(args ...interface{}) {
	logger.Panicln(args...)
}

func Fatalln(args ...interface{}) {
	logger.Fatalln(args...)
}
