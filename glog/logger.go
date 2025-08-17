package glog

import (
	"fmt"
	"log/slog"
	"os"
	"strings"

	"go.uber.org/zap/zapcore"
	"golang.org/x/term"
)

// Logger 标准日志接口。
type Logger interface {
	Debug(args ...any)
	Debugf(format string, args ...any)
	Debugln(args ...any)
	Info(args ...any)
	Infof(format string, args ...any)
	Infoln(args ...any)
	Warn(args ...any)
	Warnf(format string, args ...any)
	Warnln(args ...any)
	Error(args ...any)
	Errorf(format string, args ...any)
	Errorln(args ...any)
	Fatal(args ...any)
	Fatalf(format string, args ...any)
	Fatalln(args ...any)
	Panic(args ...any)
	Panicf(format string, args ...any)
	Panicln(args ...any)
}

var logger Logger

func init() {
	InitDefaultLogger()
}

func isTrue(s string) bool {
	switch strings.ToLower(s) {
	case "true", "t", "yes", "y", "on":
		return true
	default:
		return false
	}
}

func rightPad(str string, size int, padChar rune) string {
	if len(str) >= size {
		return str
	}
	padding := strings.Repeat(string(padChar), size-len(str))
	return str + padding
}

func colorizeSlog(level slog.Level, str string) string {
	if !term.IsTerminal(int(os.Stdout.Fd())) {
		return str
	}

	switch level {
	case slog.LevelDebug:
		return fmt.Sprintf("\x1b[1;34m%s\x1b[0m", str)
	case slog.LevelWarn:
		return fmt.Sprintf("\x1b[1;33m%s\x1b[0m", str)
	case slog.LevelError:
		return fmt.Sprintf("\x1b[1;31m%s\x1b[0m", str)
	default:
		return str
	}
}

func colorizeZaplog(enable bool, level zapcore.Level, str string) string {
	if !enable {
		return str
	}
	if !term.IsTerminal(int(os.Stdout.Fd())) {
		return str
	}

	switch level {
	case zapcore.DebugLevel:
		return fmt.Sprintf("\x1b[1;34m%s\x1b[0m", str)
	case zapcore.WarnLevel:
		return fmt.Sprintf("\x1b[1;33m%s\x1b[0m", str)
	case zapcore.ErrorLevel:
		return fmt.Sprintf("\x1b[1;31m%s\x1b[0m", str)
	default:
		return str
	}
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

func InitSlogLogger(opts ...SlogLoggerOption) {
	logger = NewSlogLogger(opts...)
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
