// Copyright 2022 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package slog

import (
	"solod.dev/so/io"
	"solod.dev/so/strconv"
	"solod.dev/so/sync"
	"solod.dev/so/time"
)

// A Level is the importance or severity of a log event.
// The higher the level, the more important or severe the event.
type Level int

const (
	LevelDebug Level = -4
	LevelInfo  Level = 0
	LevelWarn  Level = 4
	LevelError Level = 8
)

// String returns a name for the level.
func (l Level) String() string {
	if l == LevelDebug {
		return "DEBUG"
	}
	if l == LevelInfo {
		return "INFO"
	}
	if l == LevelWarn {
		return "WARN"
	}
	if l == LevelError {
		return "ERROR"
	}
	return "UNKNOWN"
}

// Record holds information about a log event.
type Record struct {
	Time    time.Time
	Message string
	Level   Level
	Attrs   []Attr
}

// Handler handles log records produced by a [Logger].
//
// A typical handler may print log records to standard error,
// or write them to a file or database, or perhaps augment them
// with additional attributes and pass them on to another handler.
//
// Users of the slog package should not invoke Handler methods directly.
// They should use the methods of Logger instead.
type Handler interface {
	// Enabled reports whether the handler handles records at the given level.
	// The handler ignores records whose level is lower.
	Enabled(level Level) bool

	// Handle handles the Record.
	// It will only be called when Enabled returns true.
	// Logger discards any errors from Handle.
	Handle(r Record) error
}

// TextHandler is a [Handler] that writes Records to an [io.Writer] as a
// sequence of key=value pairs separated by spaces and followed by a newline.
//
// Each record is formatted as:
//
//	2024-01-15T10:30:00Z INFO hello world user=john count=42
//
// Timestamp (RFC3339), level, message, then key=value pairs.
// String values with spaces are quoted.
type TextHandler struct {
	w     io.Writer
	level Level
}

// NewTextHandler creates a TextHandler that writes to w
// and handles only records at or above the given level.
func NewTextHandler(w io.Writer, level Level) TextHandler {
	return TextHandler{w: w, level: level}
}

// Enabled reports whether the handler handles records at the given level.
// The handler ignores records whose level is lower.
func (h *TextHandler) Enabled(level Level) bool {
	return level >= h.level
}

// Handle formats the record and writes it to the writer.
func (h *TextHandler) Handle(r Record) error {
	// Time
	var tbuf [30]byte
	tstr := r.Time.Format(tbuf[:], time.RFC3339, time.UTC)
	io.WriteString(h.w, tstr)

	// Level
	io.WriteString(h.w, " ")
	io.WriteString(h.w, r.Level.String())

	// Message
	io.WriteString(h.w, " ")
	io.WriteString(h.w, r.Message)

	// Attrs
	for i := 0; i < len(r.Attrs); i++ {
		io.WriteString(h.w, " ")
		io.WriteString(h.w, r.Attrs[i].Key)
		io.WriteString(h.w, "=")
		writeValue(h.w, r.Attrs[i].Value)
	}

	_, err := io.WriteString(h.w, "\n")
	return err
}

// syncHandler is a thread-safe Handle wrapper.
type syncHandler struct {
	inner Handler
	mu    sync.Mutex
}

func (h *syncHandler) Enabled(level Level) bool {
	return h.inner.Enabled(level)
}

func (h *syncHandler) Handle(r Record) error {
	h.mu.Lock()
	err := h.inner.Handle(r)
	h.mu.Unlock()
	return err
}

// writeValue formats a Value and writes it to w.
func writeValue(w io.Writer, v Value) {
	if v.kind == KindString {
		writeQuoted(w, v.str)
		return
	}
	var buf [64]byte
	if v.kind == KindInt64 {
		w.Write(strconv.AppendInt(buf[:0], int64(v.num), 10))
		return
	}
	if v.kind == KindUint64 {
		w.Write(strconv.AppendUint(buf[:0], v.num, 10))
		return
	}
	if v.kind == KindFloat64 {
		w.Write(strconv.AppendFloat(buf[:0], v.float(), 'g', -1, 64))
		return
	}
	if v.kind == KindBool {
		if v.bool() {
			io.WriteString(w, "true")
		} else {
			io.WriteString(w, "false")
		}
		return
	}
	if v.kind == KindTime {
		t := v.time()
		io.WriteString(w, t.Format(buf[:], time.RFC3339, time.UTC))
		return
	}
	if v.kind == KindDuration {
		d := v.duration()
		io.WriteString(w, d.String(buf[:]))
		return
	}
}

// writeQuoted writes s to w, quoting it if it contains spaces or is empty.
func writeQuoted(w io.Writer, s string) {
	// TODO: Currently, it doesn't handle embedded quotes.
	// Should probably use strconv.Quote once it's available.
	if len(s) == 0 || needsQuote(s) {
		io.WriteString(w, "\"")
		io.WriteString(w, s)
		io.WriteString(w, "\"")
		return
	}
	io.WriteString(w, s)
}

// needsQuote reports whether s needs to be quoted.
func needsQuote(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] == ' ' || s[i] == '"' || s[i] == '=' {
			return true
		}
	}
	return false
}
