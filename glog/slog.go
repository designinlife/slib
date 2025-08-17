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
	"time"

	"github.com/designinlife/slib/errors"

	slibos "github.com/designinlife/slib/os"

	"gopkg.in/natefinch/lumberjack.v2"
)

type textOnlyHandler struct {
	w     io.Writer
	level slog.Leveler
	cfg   *slogLoggerConfig
}

func (h *textOnlyHandler) Enabled(_ context.Context, level slog.Level) bool {
	return level >= h.level.Level()
}

func (h *textOnlyHandler) Handle(_ context.Context, record slog.Record) error {
	var err error
	var builder strings.Builder

	if !h.cfg.onlyMessage {
		builder.WriteString(record.Time.Format("2006-01-02 15:04:05.000"))
		builder.WriteString(" | ")
	}

	builder.WriteString(colorizeSlog(h.cfg.useAnsiColor, record.Level, strings.ToUpper(rightPad(strings.TrimSpace(record.Level.String()), 5, ' '))))
	builder.WriteString(" | ")

	if record.Level >= h.cfg.callerLevel {
		_, file, line, ok := runtime.Caller(6)

		if ok {
			fn := h.trimmedPath(file)
			builder.WriteString(fmt.Sprintf("%s:%d | ", fn, line))
		}
	}

	builder.WriteString(colorizeSlog(h.cfg.useAnsiColor, record.Level, strings.TrimSpace(record.Message)))

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

// mixedHandler 同时写入多个 slog.Handler。
type mixedHandler struct {
	handlers []slog.Handler
}

// Enabled 检查所有 handler 是否启用某个 level。
func (m *mixedHandler) Enabled(ctx context.Context, level slog.Level) bool {
	for _, h := range m.handlers {
		if h.Enabled(ctx, level) {
			return true
		}
	}
	return false
}

// Handle 同时写入所有 handler。
func (m *mixedHandler) Handle(ctx context.Context, r slog.Record) error {
	var err error
	for _, h := range m.handlers {
		if e := h.Handle(ctx, r); e != nil {
			err = e
		}
	}
	return err
}

// WithAttrs 为每个 handler 添加属性。
func (m *mixedHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	nh := &mixedHandler{}
	for _, h := range m.handlers {
		nh.handlers = append(nh.handlers, h.WithAttrs(attrs))
	}
	return nh
}

// WithGroup 为每个 handler 添加分组。
func (m *mixedHandler) WithGroup(name string) slog.Handler {
	nh := &mixedHandler{}
	for _, h := range m.handlers {
		nh.handlers = append(nh.handlers, h.WithGroup(name))
	}
	return nh
}

type customJSONHandler struct {
	h    slog.Handler
	skip int
}

func (m *customJSONHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return m.h.Enabled(ctx, level)
}

func (m *customJSONHandler) Handle(ctx context.Context, r slog.Record) error {
	return m.h.Handle(ctx, r)
}

func (m *customJSONHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &customJSONHandler{h: m.h.WithAttrs(attrs), skip: m.skip}
}

func (m *customJSONHandler) WithGroup(name string) slog.Handler {
	return &customJSONHandler{h: m.h.WithGroup(name), skip: m.skip}
}

func (m *customJSONHandler) Log(ctx context.Context, level slog.Level, msg string, attrs ...slog.Attr) {
	if !m.h.Enabled(ctx, level) {
		return
	}
	var pcs [1]uintptr
	runtime.Callers(m.skip, pcs[:])
	rec := slog.NewRecord(time.Now(), level, msg, pcs[0])
	rec.AddAttrs(attrs...)
	_ = m.h.Handle(ctx, rec)
}

func newJSONHandlerWithSkip(w io.Writer, level slog.Level, skip int) *customJSONHandler {
	h := slog.NewJSONHandler(w, &slog.HandlerOptions{
		Level:     level,
		AddSource: false,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			if a.Key == slog.TimeKey {
				if t, ok := a.Value.Any().(time.Time); ok {
					a.Key = "ts"
					a.Value = slog.StringValue(t.Format("2006-01-02 15:04:05.000"))
				}
			}
			return a
		},
	})
	return &customJSONHandler{h: h, skip: skip}
}

type customSLogger struct {
	logger *slog.Logger
}

type slogLoggerConfig struct {
	Handler         slog.Handler
	useTextHandler  bool
	useMixedHandler bool
	useAnsiColor    bool
	onlyMessage     bool
	level           slog.Level
	callerLevel     slog.Level
	compress        bool
}

type SlogLoggerOption func(*slogLoggerConfig)

func WithSlogUseTextHandler() SlogLoggerOption {
	return func(c *slogLoggerConfig) {
		c.useTextHandler = true
	}
}

func WithSlogUseMixedHandler() SlogLoggerOption {
	return func(c *slogLoggerConfig) {
		c.useMixedHandler = true
	}
}

func WithSlogUseAnsiColor() SlogLoggerOption {
	return func(c *slogLoggerConfig) {
		c.useAnsiColor = true
	}
}

func WithSlogOnlyMessage() SlogLoggerOption {
	return func(c *slogLoggerConfig) {
		c.onlyMessage = true
	}
}

func WithSlogHandler(h slog.Handler) SlogLoggerOption {
	return func(c *slogLoggerConfig) {
		c.Handler = h
	}
}

func WithSlogLevel(level slog.Level) SlogLoggerOption {
	return func(c *slogLoggerConfig) {
		c.level = level
	}
}

func WithSlogCallerLevel(level int) SlogLoggerOption {
	return func(c *slogLoggerConfig) {
		c.callerLevel = slog.Level(level)
	}
}

func WithSlogCompress() SlogLoggerOption {
	return func(c *slogLoggerConfig) {
		c.compress = true
	}
}

// NewSlogLogger 创建 slog 的日志实例。
func NewSlogLogger(opts ...SlogLoggerOption) Logger {
	// 获取环境变量
	debugEnabled := isTrue(os.Getenv("DEBUG"))
	isAnsiColor := isTrue(os.Getenv("LOG_COLOR"))
	logLevel := os.Getenv("LOG_LEVEL")
	logFile := os.Getenv("LOG_FILE")
	logMaxSize := slibos.GetEnvDefault("LOG_MAX_SIZE", DefaultLogMaxSize)
	logMaxBackups := slibos.GetEnvDefault("LOG_MAX_BACKUPS", DefaultLogMaxBackups)
	logMaxAge := slibos.GetEnvDefault("LOG_MAX_AGE", DefaultLogMaxAge)

	config := &slogLoggerConfig{
		callerLevel: slog.LevelWarn,
	}

	for _, opt := range opts {
		opt(config)
	}

	if os.Getppid() == 1 {
		config.onlyMessage = true
	}

	if isAnsiColor {
		config.useAnsiColor = true
	}

	if config.useMixedHandler && logFile == "" {
		config.useMixedHandler = false
	}

	if logLevel == "" {
		logLevel = config.level.String()
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

	var handler slog.Handler

	if config.Handler != nil {
		handler = config.Handler
	} else if config.useMixedHandler {
		consoleHandler := &textOnlyHandler{w: os.Stdout, level: level, cfg: config}
		fileWriter := &lumberjack.Logger{
			Filename:   logFile,
			MaxSize:    logMaxSize, // MB
			MaxBackups: logMaxBackups,
			MaxAge:     logMaxAge, // days
			Compress:   config.compress,
		}
		jsonHandler := newJSONHandlerWithSkip(fileWriter, level, 3)

		handler = &mixedHandler{
			handlers: []slog.Handler{consoleHandler, jsonHandler},
		}
	} else if config.useTextHandler {
		if logFile != "" {
			fileWriter := &lumberjack.Logger{
				Filename:   logFile,
				MaxSize:    logMaxSize, // MB
				MaxBackups: logMaxBackups,
				MaxAge:     logMaxAge, // days
				Compress:   config.compress,
			}

			writer = io.MultiWriter(os.Stdout, fileWriter)
		}
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
func (l *customSLogger) Debug(args ...any) {
	l.logger.Debug(fmt.Sprint(args...))
}

func (l *customSLogger) Debugf(format string, args ...any) {
	l.logger.Debug(fmt.Sprintf(format, args...))
}

func (l *customSLogger) Debugln(args ...any) {
	l.logger.Debug(fmt.Sprintln(args...))
}

// Info 级别日志。
func (l *customSLogger) Info(args ...any) {
	l.logger.Info(fmt.Sprint(args...))
}

func (l *customSLogger) Infof(format string, args ...any) {
	l.logger.Info(fmt.Sprintf(format, args...))
}

func (l *customSLogger) Infoln(args ...any) {
	l.logger.Info(fmt.Sprintln(args...))
}

// Warn Warn/Warning 级别日志 (Warning 作为 Warn 的别名)。
func (l *customSLogger) Warn(args ...any) {
	l.logger.Warn(fmt.Sprint(args...))
}

func (l *customSLogger) Warnf(format string, args ...any) {
	l.logger.Warn(fmt.Sprintf(format, args...))
}

func (l *customSLogger) Warnln(args ...any) {
	l.logger.Warn(fmt.Sprintln(args...))
}

func (l *customSLogger) Warning(args ...any) {
	l.Warn(args...)
}

func (l *customSLogger) Warningf(format string, args ...any) {
	l.Warnf(format, args...)
}

func (l *customSLogger) Warningln(args ...any) {
	l.Warnln(args...)
}

// Error 级别日志 (带文件和行号)。
func (l *customSLogger) Error(args ...any) {
	l.withSource().Error(fmt.Sprint(args...))
}

func (l *customSLogger) Errorf(format string, args ...any) {
	l.withSource().Error(fmt.Sprintf(format, args...))
}

func (l *customSLogger) Errorln(args ...any) {
	l.withSource().Error(fmt.Sprintln(args...))
}

// Fatal 级别日志 (带文件和行号后退出)。
func (l *customSLogger) Fatal(args ...any) {
	l.withSource().Error(fmt.Sprint(args...))
	os.Exit(1)
}

func (l *customSLogger) Fatalf(format string, args ...any) {
	l.withSource().Error(fmt.Sprintf(format, args...))
	os.Exit(1)
}

func (l *customSLogger) Fatalln(args ...any) {
	l.withSource().Error(fmt.Sprintln(args...))
	os.Exit(1)
}

// Panic 级别日志 (带文件和行号后抛出 panic)。
func (l *customSLogger) Panic(args ...any) {
	msg := fmt.Sprint(args...)
	l.withSource().Error(msg)
	panic(msg)
}

func (l *customSLogger) Panicf(format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	l.withSource().Error(msg)
	panic(msg)
}

func (l *customSLogger) Panicln(args ...any) {
	msg := fmt.Sprintln(args...)
	l.withSource().Error(msg)
	panic(msg)
}
