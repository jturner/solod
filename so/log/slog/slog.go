// Copyright 2022 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package slog provides structured logging, in which log records include a message,
// a severity level, and various other attributes expressed as key-value pairs.
//
// It defines a type, [Logger], which provides several methods such as [Logger.Info]
// and [Logger.Error]) for reporting events of interest.
//
// Each Logger is associated with a [Handler] which processes the log records
// produced by the Logger (typically by writing them to standard error or a file).
//
// A Logger output method creates a [Record] from the method arguments
// and passes it to the Handler, which decides how to handle it.
// There is a default Logger accessible through top-level functions
// (such as [Info] and [Error]) that call the corresponding Logger methods.
//
// A log record consists of a time, a level, a message, and a set of key-value
// pairs, where the keys are strings and the values may be of any type. For example:
//
//	2024-01-15T10:30:00Z INFO hello world user=john count=42
//
// The package provides top-level functions like [Info] and [Error] that call the
// corresponding methods of the [Default] Logger. The default logger uses a [TextHandler]
// that writes to [os.Stderr] and handles only records at or above [LevelInfo].
//
// Loosely based on the [log/slog] package.
//
// [log/slog]: https://github.com/golang/go/blob/go1.26.2/src/log/slog
package slog

import (
	"solod.dev/so/os"
	"solod.dev/so/sync"
	"solod.dev/so/time"
)

// A Logger records structured information about each call to its
// Log, Debug, Info, Warn, and Error methods.
// For each call, it creates a [Record] and passes it to a [Handler].
type Logger struct {
	handler Handler
}

// New creates a new Logger with the given non-nil Handler.
func New(h Handler) Logger {
	if h == nil {
		panic("slog: nil Handler")
	}
	return Logger{handler: h}
}

// Debug logs at [LevelDebug].
func (l *Logger) Debug(msg string, attrs ...Attr) {
	l.Log(LevelDebug, msg, attrs...)
}

// Info logs at [LevelInfo].
func (l *Logger) Info(msg string, attrs ...Attr) {
	l.Log(LevelInfo, msg, attrs...)
}

// Warn logs at [LevelWarn].
func (l *Logger) Warn(msg string, attrs ...Attr) {
	l.Log(LevelWarn, msg, attrs...)
}

// Error logs at [LevelError].
func (l *Logger) Error(msg string, attrs ...Attr) {
	l.Log(LevelError, msg, attrs...)
}

// Log logs a message at the given level with the given attrs.
func (l *Logger) Log(level Level, msg string, attrs ...Attr) {
	if !l.handler.Enabled(level) {
		return
	}
	r := Record{
		Time:    time.Now(),
		Message: msg,
		Level:   level,
		Attrs:   attrs,
	}
	_ = l.handler.Handle(r)
}

// Enabled reports whether the logger handles records at the given level.
func (l *Logger) Enabled(level Level) bool {
	return l.handler.Enabled(level)
}

// Handler returns the logger's Handler.
func (l *Logger) Handler() Handler {
	return l.handler
}

// Default logger. Uses a [TextHandler] that writes to [os.Stderr]
// and handles only records at or above [LevelInfo].
// Unlike the regular [Logger], the default logger is thread-safe.
var defaultOnce sync.Once
var defaultSync syncHandler
var defaultHandler TextHandler
var defaultLogger_ Logger
var defaultLogger *Logger

// SetDefault sets the default [Logger] which is used by
// the top-level functions [Info], [Debug] and so on.
//
// The installed logger's thread safety is the caller's responsibility;
// SetDefault by itself does not make the logger safe for concurrent use.
//
// SetDefault is not thread-safe. Call it during startup,
// before logging from multiple threads.
func SetDefault(l *Logger) {
	defaultLogger = l
	// Consume the once so a later lazy [ensureDefault] becomes a no-op.
	defaultOnce.Do(noop)
}

// noop does nothing. Used to consume defaultOnce in SetDefault.
func noop() {}

// Default returns the default [Logger].
func Default() *Logger {
	ensureDefault()
	return defaultLogger
}

// Debug calls [Logger.Debug] on the default logger.
func Debug(msg string, attrs ...Attr) {
	ensureDefault()
	defaultLogger.Log(LevelDebug, msg, attrs...)
}

// Info calls [Logger.Info] on the default logger.
func Info(msg string, attrs ...Attr) {
	ensureDefault()
	defaultLogger.Log(LevelInfo, msg, attrs...)
}

// Warn calls [Logger.Warn] on the default logger.
func Warn(msg string, attrs ...Attr) {
	ensureDefault()
	defaultLogger.Log(LevelWarn, msg, attrs...)
}

// Error calls [Logger.Error] on the default logger.
func Error(msg string, attrs ...Attr) {
	ensureDefault()
	defaultLogger.Log(LevelError, msg, attrs...)
}

// Log calls [Logger.Log] on the default logger.
func Log(level Level, msg string, attrs ...Attr) {
	ensureDefault()
	defaultLogger.Log(level, msg, attrs...)
}

// ensureDefault lazy-initializes the default logger exactly once.
// The build is deferred to first use (rather than init) because of the
// dependency on os.Stderr, which a sibling package's constructor may not
// have set at init time.
func ensureDefault() {
	defaultOnce.Do(buildDefault)
}

// buildDefault constructs the default logger.
func buildDefault() {
	defaultSync.mu.Init()
	defaultHandler = NewTextHandler(os.Stderr, LevelInfo)
	defaultSync.inner = &defaultHandler
	defaultLogger_ = New(&defaultSync)
	defaultLogger = &defaultLogger_
}

func init() {
	defaultOnce.Init()
}
