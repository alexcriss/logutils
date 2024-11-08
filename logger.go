package logutils

import (
	"context"
	"log/slog"
	"os"
	"strings"
)

const (
	LevelTrace     = slog.Level(-8)
	LevelTraceName = "TRACE"
)

var (
	defaultLogger  *Logger = &Logger{Logger: slog.Default()}
	defaultOptions slog.HandlerOptions
)

func ReplaceLogLevel(groups []string, a slog.Attr) slog.Attr {
	if a.Key == slog.LevelKey {
		level := a.Value.Any().(slog.Level)
		var replaced string
		switch level {
		case LevelTrace:
			replaced = LevelTraceName
		case slog.LevelDebug:
			replaced = slog.LevelDebug.String()
		case slog.LevelInfo:
			replaced = slog.LevelInfo.String()
		case slog.LevelWarn:
			replaced = slog.LevelWarn.String()
		case slog.LevelError:
			replaced = slog.LevelError.String()
		default:
			replaced = "UNKNOWN"
		}
		a.Value = slog.StringValue(replaced)
	}
	return a
}

func NewDefaultLogger(loglevel string) {
	var level slog.Level
	switch strings.ToUpper(loglevel) {
	case LevelTraceName:
		level = LevelTrace
	case slog.LevelDebug.String():
		level = slog.LevelDebug
	case slog.LevelInfo.String():
		level = slog.LevelInfo
	case slog.LevelWarn.String():
		level = slog.LevelWarn
	case slog.LevelError.String():
		level = slog.LevelError
	default:
		level = slog.LevelError
	}

	opts := slog.HandlerOptions{
		AddSource:   true,
		Level:       level,
		ReplaceAttr: ReplaceLogLevel,
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &opts))
	defaultLogger = &Logger{
		Logger: logger,
	}
	defaultOptions = opts
}

func DefaultLogger() *Logger {
	return defaultLogger
}

func NewLoggerFromDefault(options *slog.HandlerOptions) *Logger {
	options.Level = defaultOptions.Level
	return &Logger{
		Logger: slog.New(slog.NewJSONHandler(os.Stdout, options)),
	}
}

type Logger struct {
	*slog.Logger
}

func (l *Logger) With(args ...slog.Attr) *Logger {
	iargs := []interface{}{}
	for _, arg := range args {
		iargs = append(iargs, arg)
	}
	return &Logger{
		Logger: l.Logger.With(iargs...),
	}

}

func (l *Logger) Trace(msg string, additional ...slog.Attr) {
	l.LogAttrs(context.Background(), LevelTrace, msg, additional...)
}
