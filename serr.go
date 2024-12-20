package logutils

import (
	"context"
	"errors"
	"log/slog"
)

var errorer *Logger

const (
	ERROR = "error"
)

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
	var serr StructuredError

	if errors.As(err, &serr) {
		additional = append(additional, serr.Attrs...)
	}
	attrs := append(e.attrs, additional...)
	return NewStructuredError(err, attrs...)
}

func SetDefaultErrorer(format Format) {
	opts := slog.HandlerOptions{
		AddSource:   true,
		ReplaceAttr: ReplaceSource,
	}
	errorer = NewLoggerFromDefault(&opts, format)
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
