package logutils

import (
	"context"
	"fmt"
	"log/slog"
	"runtime"
	"strings"
)

var errorer *Logger

type StructuredError struct {
	Attrs []slog.Attr
	Err   error
}

func (serr StructuredError) Error() string {
	return serr.Err.Error()
}

func NewStructuredError(err error, attrs ...slog.Attr) StructuredError {
	return StructuredError{Err: err, Attrs: attrs}
}

func (serr StructuredError) Unwrap() error {
	return serr.Err
}

type Errorer struct {
	attrs []slog.Attr
}

func NewErrorer(attrs ...slog.Attr) *Errorer {
	return &Errorer{
		attrs: attrs,
	}
}

func (e *Errorer) Error(err error, additional ...slog.Attr) error {
	if err == nil {
		return nil
	}
	if serr, ok := err.(StructuredError); ok {
		additional = append(additional, serr.Attrs...)
	}
	attrs := append(e.attrs, additional...)
	return NewStructuredError(err, attrs...)
}

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
			if strings.Contains(frame.File, "runtime/") || strings.Contains(frame.File, "log/slog") || strings.Contains(frame.File, "utils/serr") {
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

func SetDefaultErrorer() {
	opts := slog.HandlerOptions{
		AddSource:   true,
		ReplaceAttr: ReplaceSource,
	}
	errorer = NewLoggerFromDefault(&opts)
}

func ErrLog(level slog.Level, msg string, err error, additional ...slog.Attr) {
	attrs := []slog.Attr{slog.Any(ERROR, err.Error())}
	attrs = append(attrs, additional...)
	if serr, ok := err.(StructuredError); ok {
		attrs = append(attrs, serr.Attrs...)
		errorer.LogAttrs(context.Background(), level, msg, attrs...)
	} else {
		errorer.LogAttrs(context.Background(), level, msg, attrs...)
	}
}

func ErrTrace(msg string, err error, additional ...slog.Attr) {
	ErrLog(LevelTrace, msg, err, additional...)
}

func ErrDebug(msg string, err error, additional ...slog.Attr) {
	ErrLog(slog.LevelDebug, msg, err, additional...)
}

func ErrInfo(msg string, err error, additional ...slog.Attr) {
	ErrLog(slog.LevelInfo, msg, err, additional...)
}

func ErrWarn(msg string, err error, additional ...slog.Attr) {
	ErrLog(slog.LevelWarn, msg, err, additional...)
}

func ErrError(msg string, err error, additional ...slog.Attr) {
	ErrLog(slog.LevelError, msg, err, additional...)
}
