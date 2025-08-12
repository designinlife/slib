package glog

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/designinlife/slib/errors"
	"github.com/designinlife/slib/str"
	"github.com/designinlife/slib/types"

	"golang.org/x/term"
)

type textOnlyHandler struct {
	w     io.Writer
	level slog.Leveler
	cfg   *stdLoggerConfig
}

func (h *textOnlyHandler) Enabled(_ context.Context, level slog.Level) bool {
	return level >= h.level.Level()
}

func (h *textOnlyHandler) Handle(_ context.Context, record slog.Record) error {
	var err error
	var builder strings.Builder

	isTerm := term.IsTerminal(int(os.Stdout.Fd()))

	if h.cfg.OnlyMessage {
		if record.Level >= h.cfg.CallerLevel {
			_, file, line, ok := runtime.Caller(5)

			if ok {
				fn := h.trimmedPath(file)
				builder.WriteString(fmt.Sprintf("%s:%d | ", fn, line))
			}
		}
		if h.cfg.UseColor && isTerm {
			if record.Level >= slog.LevelError {
				builder.WriteString("\033[1;31m")
			}
			builder.WriteString(strings.TrimSpace(record.Message))
			if record.Level >= slog.LevelError {
				builder.WriteString("\033[0m")
			}
		} else {
			builder.WriteString(strings.TrimSpace(record.Message))
		}
	} else {
		builder.WriteString(record.Time.Format("2006-01-02 15:04:05"))
		builder.WriteString(" | ")
		builder.WriteString(strings.ToUpper(record.Level.String()))
		builder.WriteString(" | ")

		if record.Level >= h.cfg.CallerLevel {
			_, file, line, ok := runtime.Caller(5)

			if ok {
				fn := h.trimmedPath(file)
				builder.WriteString(fmt.Sprintf("%s:%d | ", fn, line))
			}

			if h.cfg.UseColor && isTerm {
				builder.WriteString("\033[1;31m")
			}
		}
		builder.WriteString(strings.TrimSpace(record.Message))
		if h.cfg.UseColor && isTerm && record.Level >= slog.LevelError {
			builder.WriteString("\033[0m")
		}
	}

	_, err = fmt.Fprintln(h.w, builder.String())
	if err != nil {
		return errors.Wrap(err, "textOnlyHandler Handle Fprintln failed")
	}

	return nil
}

func (h *textOnlyHandler) WithAttrs(_ []slog.Attr) slog.Handler {
	// Since we're ignoring attributes, we don't need to do anything here.
	return h
}

func (h *textOnlyHandler) WithGroup(_ string) slog.Handler {
	// Groups are also ignored in this simple example
	return h
}

func (h *textOnlyHandler) trimmedPath(file string) string {
	idx := strings.LastIndexByte(file, '/')
	if idx == -1 {
		return file
	}
	dirName := filepath.Base(filepath.Dir(file))

	return strings.ReplaceAll(filepath.Join(dirName, file[idx+1:]), "\\", "/")
}

type customSLogger struct {
	logger *slog.Logger
}

type stdLoggerConfig struct {
	Handler        slog.Handler
	UseTextHandler bool
	UseColor       bool
	OnlyMessage    bool
	CallerLevel    slog.Level
}

type StdLoggerOption func(*stdLoggerConfig)

func WithUseTextHandler() StdLoggerOption {
	return func(c *stdLoggerConfig) {
		c.UseTextHandler = true
	}
}

func WithUseColor() StdLoggerOption {
	return func(c *stdLoggerConfig) {
		c.UseColor = true
	}
}

func WithOnlyMessage() StdLoggerOption {
	return func(c *stdLoggerConfig) {
		c.OnlyMessage = true
	}
}

func WithHandler(h slog.Handler) StdLoggerOption {
	return func(c *stdLoggerConfig) {
		c.Handler = h
	}
}

func WithCallerLevel(level int) StdLoggerOption {
	return func(c *stdLoggerConfig) {
		c.CallerLevel = slog.Level(level)
	}
}

// NewStdLogger 创建 slog 的日志实例。
func NewStdLogger(opts ...StdLoggerOption) types.ExtraLogger {
	// 获取环境变量
	debugEnabled := str.IsTrue(os.Getenv("DEBUG"))
	logLevel := os.Getenv("LOG_LEVEL")
	logFile := os.Getenv("LOG_FILE")

	config := &stdLoggerConfig{
		CallerLevel: slog.LevelError,
	}

	for _, opt := range opts {
		opt(config)
	}

	// 设置日志级别
	var level slog.Level
	switch strings.ToLower(logLevel) {
	case "debug":
		level = slog.LevelDebug
	case "warn", "warning":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	default:
		level = slog.LevelInfo
	}
	// DEBUG=true 时强制使用 Debug 级别
	if debugEnabled {
		level = slog.LevelDebug
	}

	// 配置输出目标
	var writer io.Writer = os.Stdout
	if logFile != "" {
		file, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o600)
		if err == nil {
			writer = io.MultiWriter(os.Stdout, file)
		}
	}

	var handler slog.Handler

	if config.Handler != nil {
		handler = config.Handler
	} else if config.UseTextHandler {
		handler = &textOnlyHandler{w: writer, level: level, cfg: config}
	} else {
		handler = slog.NewJSONHandler(writer, &slog.HandlerOptions{})
	}

	return &customSLogger{
		logger: slog.New(handler),
	}
}

// 添加文件和行号的辅助函数。
func (l *customSLogger) withSource() *slog.Logger {
	pc, file, line, _ := runtime.Caller(3)
	return l.logger.With(
		"file", filepath.Base(file),
		"line", line,
		"func", runtime.FuncForPC(pc).Name(),
	)
}

// Debug 级别日志。
func (l *customSLogger) Debug(args ...interface{}) {
	l.logger.Debug(fmt.Sprint(args...))
}

func (l *customSLogger) Debugf(format string, args ...interface{}) {
	l.logger.Debug(fmt.Sprintf(format, args...))
}

func (l *customSLogger) Debugln(args ...interface{}) {
	l.logger.Debug(fmt.Sprintln(args...))
}

// Info 级别日志。
func (l *customSLogger) Info(args ...interface{}) {
	l.logger.Info(fmt.Sprint(args...))
}

func (l *customSLogger) Infof(format string, args ...interface{}) {
	l.logger.Info(fmt.Sprintf(format, args...))
}

func (l *customSLogger) Infoln(args ...interface{}) {
	l.logger.Info(fmt.Sprintln(args...))
}

// Warn Warn/Warning 级别日志 (Warning 作为 Warn 的别名)。
func (l *customSLogger) Warn(args ...interface{}) {
	l.logger.Warn(fmt.Sprint(args...))
}

func (l *customSLogger) Warnf(format string, args ...interface{}) {
	l.logger.Warn(fmt.Sprintf(format, args...))
}

func (l *customSLogger) Warnln(args ...interface{}) {
	l.logger.Warn(fmt.Sprintln(args...))
}

func (l *customSLogger) Warning(args ...interface{}) {
	l.Warn(args...)
}

func (l *customSLogger) Warningf(format string, args ...interface{}) {
	l.Warnf(format, args...)
}

func (l *customSLogger) Warningln(args ...interface{}) {
	l.Warnln(args...)
}

// Error 级别日志 (带文件和行号)。
func (l *customSLogger) Error(args ...interface{}) {
	l.withSource().Error(fmt.Sprint(args...))
}

func (l *customSLogger) Errorf(format string, args ...interface{}) {
	l.withSource().Error(fmt.Sprintf(format, args...))
}

func (l *customSLogger) Errorln(args ...interface{}) {
	l.withSource().Error(fmt.Sprintln(args...))
}

// Fatal 级别日志 (带文件和行号后退出)。
func (l *customSLogger) Fatal(args ...interface{}) {
	l.withSource().Error(fmt.Sprint(args...))
	os.Exit(1)
}

func (l *customSLogger) Fatalf(format string, args ...interface{}) {
	l.withSource().Error(fmt.Sprintf(format, args...))
	os.Exit(1)
}

func (l *customSLogger) Fatalln(args ...interface{}) {
	l.withSource().Error(fmt.Sprintln(args...))
	os.Exit(1)
}

// Panic 级别日志 (带文件和行号后抛出 panic)。
func (l *customSLogger) Panic(args ...interface{}) {
	msg := fmt.Sprint(args...)
	l.withSource().Error(msg)
	panic(msg)
}

func (l *customSLogger) Panicf(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	l.withSource().Error(msg)
	panic(msg)
}

func (l *customSLogger) Panicln(args ...interface{}) {
	msg := fmt.Sprintln(args...)
	l.withSource().Error(msg)
	panic(msg)
}
