package glog

import (
	"fmt"
	"os"
	"strings"

	"github.com/designinlife/slib/types"

	"github.com/spf13/cast"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

type sugarLoggerConfig struct {
	disableTime   bool
	disableLevel  bool
	disableCaller bool
}

type SugarLoggerOption func(*sugarLoggerConfig)

func WithSugarDisableTime() SugarLoggerOption {
	return func(c *sugarLoggerConfig) {
		c.disableTime = true
	}
}

func WithSugarDisableLevel() SugarLoggerOption {
	return func(c *sugarLoggerConfig) {
		c.disableLevel = true
	}
}

func WithSugarDisableCaller() SugarLoggerOption {
	return func(c *sugarLoggerConfig) {
		c.disableCaller = true
	}
}

type defStdLevelEnabler struct {
	lv zapcore.Level
}

func (g *defStdLevelEnabler) Enabled(level zapcore.Level) bool {
	if level >= g.lv && level <= zapcore.InfoLevel {
		return true
	}

	return false
}

func initSugaredLogger(opts ...SugarLoggerOption) types.Logger {
	config := &sugarLoggerConfig{}

	for _, opt := range opts {
		opt(config)
	}

	pe1 := zap.NewProductionEncoderConfig()

	pe1.EncodeTime = zapcore.TimeEncoderOfLayout("2006-01-02 15:04:05.000")
	pe1.EncodeLevel = zapcore.CapitalLevelEncoder
	pe1.EncodeCaller = customEnccodeCaller
	pe1.ConsoleSeparator = " | "

	if os.Getppid() == 1 {
		config.disableTime = true
	}
	if config.disableTime {
		pe1.TimeKey = zapcore.OmitKey
	}
	if config.disableLevel {
		pe1.LevelKey = zapcore.OmitKey
	}
	if config.disableCaller {
		pe1.CallerKey = zapcore.OmitKey
	}

	pe2 := zap.NewProductionEncoderConfig()
	// pe.EncodeTime = zapcore.TimeEncoderOfLayout("2006-01-02 15:04:05.000")
	pe2.EncodeLevel = zapcore.CapitalColorLevelEncoder
	pe2.ConsoleSeparator = " | "
	pe2.EncodeCaller = customEnccodeCaller

	if os.Getppid() != 1 {
		pe2.EncodeTime = zapcore.TimeEncoderOfLayout("2006-01-02 15:04:05.000")
	} else {
		pe2.TimeKey = zapcore.OmitKey
		// pe1.LevelKey = zapcore.OmitKey
	}

	// fileEncoder := zapcore.NewJSONEncoder(pe)
	consoleEncoder1 := zapcore.NewConsoleEncoder(pe1)
	consoleEncoder2 := zapcore.NewConsoleEncoder(pe2)

	isDebug := cast.ToBool(os.Getenv("DEBUG"))
	logLevel := os.Getenv("LOG_LEVEL")
	logFile := os.Getenv("LOG_FILE")

	level := zap.InfoLevel
	if isDebug {
		level = zap.DebugLevel
	} else if strings.TrimSpace(logLevel) != "" {
		if lv, err := zapcore.ParseLevel(logLevel); err == nil {
			level = lv
		} else {
			_, _ = fmt.Fprintf(os.Stderr, "Invalid zap log level: %s\n", logLevel)
		}
	}

	enabler := &defStdLevelEnabler{lv: level}

	var cores []zapcore.Core

	if logFile != "" {
		pe3 := zap.NewProductionEncoderConfig()
		pe3.EncodeTime = zapcore.TimeEncoderOfLayout("2006-01-02 15:04:05.000")
		pe3.EncodeLevel = zapcore.CapitalLevelEncoder
		pe3.EncodeCaller = customEnccodeCaller
		// pe3.ConsoleSeparator = " "

		fileEncoder := zapcore.NewJSONEncoder(pe3)
		cores = append(cores, zapcore.NewCore(fileEncoder, zapcore.AddSync(getLogWriter(logFile)), level))
	}

	cores = append(cores, zapcore.NewCore(consoleEncoder1, zapcore.AddSync(os.Stdout), enabler))
	cores = append(cores, zapcore.NewCore(consoleEncoder2, zapcore.AddSync(os.Stdout), zap.WarnLevel))

	core := zapcore.NewTee(cores...)

	logger = zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1), zap.AddStacktrace(zap.FatalLevel)).Sugar()

	// logger.Debugf("Zap Logger installed.")

	return logger
}

func getLogWriter(filename string) zapcore.WriteSyncer {
	lumberJackLogger := &lumberjack.Logger{
		Filename:   filename,
		MaxSize:    10,
		MaxBackups: 5,
		MaxAge:     30,
		Compress:   false,
	}
	return zapcore.AddSync(lumberJackLogger)
}

func customEnccodeCaller(caller zapcore.EntryCaller, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(caller.TrimmedPath())
}
