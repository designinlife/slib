package glog

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cast"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"

	slibos "github.com/designinlife/slib/os"
)

type sugarLoggerConfig struct {
	disableTime   bool
	disableLevel  bool
	disableCaller bool
	level         zapcore.Level
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

func WithSugarLevel(level zapcore.Level) SugarLoggerOption {
	return func(c *sugarLoggerConfig) {
		c.level = level
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

type sugarLogger struct {
	logger *zap.SugaredLogger
}

func (s *sugarLogger) Debug(args ...any) {
	s.logger.Debug(args...)
}

func (s *sugarLogger) Debugf(format string, args ...any) {
	s.logger.Debugf(format, args...)
}

func (s *sugarLogger) Debugln(args ...any) {
	s.logger.Debugln(args...)
}

func (s *sugarLogger) Info(args ...any) {
	s.logger.Info(args...)
}

func (s *sugarLogger) Infof(format string, args ...any) {
	s.logger.Infof(format, args...)
}

func (s *sugarLogger) Infoln(args ...any) {
	s.logger.Infoln(args...)
}

func (s *sugarLogger) Warn(args ...any) {
	s.logger.Warn(args...)
}

func (s *sugarLogger) Warnf(format string, args ...any) {
	s.logger.Warnf(format, args...)
}

func (s *sugarLogger) Warnln(args ...any) {
	s.logger.Warnln(args...)
}

func (s *sugarLogger) Error(args ...any) {
	s.logger.Error(args...)
}

func (s *sugarLogger) Errorf(format string, args ...any) {
	s.logger.Errorf(format, args...)
}

func (s *sugarLogger) Errorln(args ...any) {
	s.logger.Errorln(args...)
}

func (s *sugarLogger) Fatal(args ...any) {
	s.logger.Fatal(args...)
}

func (s *sugarLogger) Fatalf(format string, args ...any) {
	s.logger.Fatalf(format, args...)
}

func (s *sugarLogger) Fatalln(args ...any) {
	s.logger.Fatalln(args...)
}

func (s *sugarLogger) Panic(args ...any) {
	s.logger.Panic(args...)
}

func (s *sugarLogger) Panicf(format string, args ...any) {
	s.logger.Panicf(format, args...)
}

func (s *sugarLogger) Panicln(args ...any) {
	s.logger.Panicln(args...)
}

func newSugarLogger(logger *zap.SugaredLogger) Logger {
	return &sugarLogger{logger: logger}
}

func initSugaredLogger(opts ...SugarLoggerOption) Logger {
	config := &sugarLoggerConfig{}

	for _, opt := range opts {
		opt(config)
	}

	pe1 := zap.NewProductionEncoderConfig()

	pe1.EncodeTime = zapcore.TimeEncoderOfLayout("2006-01-02 15:04:05.000")
	// pe1.EncodeLevel = zapcore.CapitalLevelEncoder
	pe1.EncodeLevel = func(level zapcore.Level, encoder zapcore.PrimitiveArrayEncoder) {
		encoder.AppendString(rightPad(level.CapitalString(), 5, ' '))
	}
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
	// pe2.EncodeLevel = zapcore.CapitalColorLevelEncoder
	pe2.EncodeLevel = func(level zapcore.Level, encoder zapcore.PrimitiveArrayEncoder) {
		encoder.AppendString(rightPad(level.CapitalString(), 5, ' '))
	}
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
	logMaxSize := slibos.GetEnvDefault("LOG_MAX_SIZE", 10)
	logMaxBackups := slibos.GetEnvDefault("LOG_MAX_BACKUPS", 5)
	logMaxAge := slibos.GetEnvDefault("LOG_MAX_AGE", 30)

	level := zap.InfoLevel

	if logLevel == "" {
		logLevel = config.level.String()
	}

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
		// pe3.EncodeLevel = zapcore.CapitalLevelEncoder
		pe3.EncodeLevel = func(level zapcore.Level, encoder zapcore.PrimitiveArrayEncoder) {
			encoder.AppendString(level.CapitalString())
		}
		pe3.EncodeCaller = customEnccodeCaller
		// pe3.ConsoleSeparator = " "

		fileEncoder := zapcore.NewJSONEncoder(pe3)
		cores = append(cores, zapcore.NewCore(fileEncoder, zapcore.AddSync(getLogWriter(logFile, logMaxSize, logMaxBackups, logMaxAge)), level))
	}

	cores = append(cores, zapcore.NewCore(consoleEncoder1, zapcore.AddSync(os.Stdout), enabler))
	cores = append(cores, zapcore.NewCore(consoleEncoder2, zapcore.AddSync(os.Stdout), zap.WarnLevel))

	core := zapcore.NewTee(cores...)

	logger = newSugarLogger(zap.New(core, zap.AddCaller(), zap.AddCallerSkip(2)).Sugar())

	// logger.Debugf("Zap Logger installed.")

	return logger
}

func getLogWriter(filename string, maxSize, maxBackups, maxAge int) zapcore.WriteSyncer {
	lumberJackLogger := &lumberjack.Logger{
		Filename:   filename,
		MaxSize:    maxSize,
		MaxBackups: maxBackups,
		MaxAge:     maxAge,
		Compress:   false,
	}
	return zapcore.AddSync(lumberJackLogger)
}

func customEnccodeCaller(caller zapcore.EntryCaller, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(caller.TrimmedPath())
}
