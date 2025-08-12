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

func Debug(args ...any) {
	logger.Debug(args...)
}

func Info(args ...any) {
	logger.Info(args...)
}

func Warn(args ...any) {
	logger.Warn(args...)
}

func Error(args ...any) {
	logger.Error(args...)
}

func Panic(args ...any) {
	logger.Panic(args...)
}

func Fatal(args ...any) {
	logger.Fatal(args...)
}

func Debugf(format string, args ...any) {
	logger.Debugf(format, args...)
}

func Infof(format string, args ...any) {
	logger.Infof(format, args...)
}

func Warnf(format string, args ...any) {
	logger.Warnf(format, args...)
}

func Errorf(format string, args ...any) {
	logger.Errorf(format, args...)
}

func Panicf(format string, args ...any) {
	logger.Panicf(format, args...)
}

func Fatalf(format string, args ...any) {
	logger.Fatalf(format, args...)
}

func Debugln(args ...any) {
	logger.Debugln(args...)
}

func Infoln(args ...any) {
	logger.Infoln(args...)
}

func Warnln(args ...any) {
	logger.Warnln(args...)
}

func Errorln(args ...any) {
	logger.Errorln(args...)
}

func Panicln(args ...any) {
	logger.Panicln(args...)
}

func Fatalln(args ...any) {
	logger.Fatalln(args...)
}
