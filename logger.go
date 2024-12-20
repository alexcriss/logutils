package logutils

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"runtime"
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

type Format string

const (
	JSON = "json"
	TEXT = "text"
)

func ReplaceSource(groups []string, a slog.Attr) slog.Attr {
	// Run through the default function for replacing log levels
	a = ReplaceLogLevel(groups, a)

	if a.Key == slog.SourceKey {
		pc := make([]uintptr, 15)
		n := runtime.Callers(0, pc)
		if n == 0 {
			return slog.Attr{}
		}

		pc = pc[:n] // pass only valid pcs to runtime.CallersFrames
		frames := runtime.CallersFrames(pc)
		for {
			frame, more := frames.Next()
			// We skip everything in the call stack that is before this function and the function itself
			if strings.Contains(frame.File, "runtime/") ||
				strings.Contains(frame.File, "github.com/alexcriss/logutils") ||
				strings.Contains(frame.File, "log/slog") ||
				strings.Contains(frame.File, "serr.go") {
				if !more {
					break
				} else {
					continue
				}
			}

			return slog.Attr{Key: slog.SourceKey, Value: slog.StringValue(fmt.Sprintf("%s:%d", frame.File, frame.Line))}
		}
	}
	return a
}

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

func NewDefaultLogger(loglevel string, format Format) {
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
		ReplaceAttr: ReplaceSource,
	}

	var logger *slog.Logger
	switch format {
	case JSON:
		logger = slog.New(slog.NewJSONHandler(os.Stdout, &opts))
	case TEXT:
		logger = slog.New(slog.NewTextHandler(os.Stdout, &opts))
	}
	defaultLogger = &Logger{
		Logger: logger,
	}
	defaultOptions = opts
}

func DefaultLogger() *Logger {
	return defaultLogger
}

func NewLoggerFromDefault(options *slog.HandlerOptions, format Format) *Logger {
	options.Level = defaultOptions.Level
	var logger *slog.Logger
	switch format {
	case JSON:
		logger = slog.New(slog.NewJSONHandler(os.Stdout, options))
	case TEXT:
		logger = slog.New(slog.NewTextHandler(os.Stdout, options))
	}
	return &Logger{
		Logger: logger,
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
