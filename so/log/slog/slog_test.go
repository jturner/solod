package slog

import (
	"testing"

	"github.com/nalgeon/be"
	"solod.dev/so/strings"
	"solod.dev/so/time"
)

func TestLevel(t *testing.T) {
	t.Run("String", func(t *testing.T) {
		be.Equal(t, LevelDebug.String(), "DEBUG")
		be.Equal(t, LevelInfo.String(), "INFO")
		be.Equal(t, LevelWarn.String(), "WARN")
		be.Equal(t, LevelError.String(), "ERROR")
	})

	t.Run("unknown", func(t *testing.T) {
		be.Equal(t, Level(42).String(), "UNKNOWN")
	})
}

func TestValue(t *testing.T) {
	t.Run("String", func(t *testing.T) {
		v := StringValue("hello")
		be.Equal(t, v.Kind(), KindString)
		be.Equal(t, v.String(), "hello")
	})

	t.Run("Int", func(t *testing.T) {
		v := IntValue(42)
		be.Equal(t, v.Kind(), KindInt64)
		be.Equal(t, v.Int(), 42)
	})

	t.Run("Int64", func(t *testing.T) {
		v := Int64Value(-100)
		be.Equal(t, v.Kind(), KindInt64)
		be.Equal(t, v.Int64(), int64(-100))
	})

	t.Run("Uint64", func(t *testing.T) {
		v := Uint64Value(999)
		be.Equal(t, v.Kind(), KindUint64)
		be.Equal(t, v.Uint64(), uint64(999))
	})

	t.Run("Float64", func(t *testing.T) {
		v := Float64Value(3.14)
		be.Equal(t, v.Kind(), KindFloat64)
		be.Equal(t, v.Float64(), 3.14)
	})

	t.Run("Bool/true", func(t *testing.T) {
		v := BoolValue(true)
		be.Equal(t, v.Kind(), KindBool)
		be.True(t, v.Bool())
	})

	t.Run("Bool/false", func(t *testing.T) {
		v := BoolValue(false)
		be.True(t, !v.Bool())
	})

	t.Run("Time", func(t *testing.T) {
		ts := time.Unix(1234567890, 0)
		v := TimeValue(ts)
		be.Equal(t, v.Kind(), KindTime)
		be.True(t, v.Time().Equal(ts))
	})

	t.Run("Duration", func(t *testing.T) {
		d := 5 * time.Second
		v := DurationValue(d)
		be.Equal(t, v.Kind(), KindDuration)
		be.Equal(t, v.Duration(), d)
	})
}

func TestAttr(t *testing.T) {
	t.Run("String", func(t *testing.T) {
		a := String("name", "alice")
		be.Equal(t, a.Key, "name")
		be.Equal(t, a.Value.String(), "alice")
	})

	t.Run("Int", func(t *testing.T) {
		a := Int("count", 5)
		be.Equal(t, a.Key, "count")
		be.Equal(t, a.Value.Int(), 5)
	})

	t.Run("Int64", func(t *testing.T) {
		a := Int64("big", int64(-9999))
		be.Equal(t, a.Key, "big")
		be.Equal(t, a.Value.Int64(), int64(-9999))
	})

	t.Run("Uint64", func(t *testing.T) {
		a := Uint64("id", uint64(12345))
		be.Equal(t, a.Key, "id")
		be.Equal(t, a.Value.Uint64(), uint64(12345))
	})

	t.Run("Float64", func(t *testing.T) {
		a := Float64("score", 9.5)
		be.Equal(t, a.Key, "score")
		be.Equal(t, a.Value.Float64(), 9.5)
	})

	t.Run("Bool", func(t *testing.T) {
		a := Bool("ok", false)
		be.Equal(t, a.Key, "ok")
		be.True(t, !a.Value.Bool())
	})

	t.Run("Time", func(t *testing.T) {
		ts := time.Unix(1234567890, 0)
		a := Time("timestamp", ts)
		be.Equal(t, a.Key, "timestamp")
		be.True(t, a.Value.Time().Equal(ts))
	})

	t.Run("Duration", func(t *testing.T) {
		d := 5 * time.Second
		a := Duration("timeout", d)
		be.Equal(t, a.Key, "timeout")
		be.Equal(t, a.Value.Duration(), d)
	})
}

func TestTextHandler(t *testing.T) {
	t.Run("Enabled", func(t *testing.T) {
		var sb strings.Builder
		h := NewTextHandler(&sb, LevelWarn)
		be.True(t, !h.Enabled(LevelInfo))
		be.True(t, h.Enabled(LevelWarn))
		be.True(t, h.Enabled(LevelError))
		sb.Free()
	})

	t.Run("basic", func(t *testing.T) {
		var sb strings.Builder
		h := NewTextHandler(&sb, LevelInfo)
		r := Record{
			Time:    time.Unix(1700000000, 0),
			Message: "hello",
			Level:   LevelInfo,
		}
		err := h.Handle(r)
		be.Err(t, err, nil)
		be.Equal(t, sb.String(), "2023-11-14T22:13:20Z INFO hello\n")
		sb.Free()
	})

	t.Run("attrs", func(t *testing.T) {
		var sb strings.Builder
		h := NewTextHandler(&sb, LevelInfo)
		r := Record{
			Time:    time.Unix(1700000000, 0),
			Message: "request",
			Level:   LevelWarn,
			Attrs:   []Attr{String("method", "GET"), Int("status", 200)},
		}
		err := h.Handle(r)
		be.Err(t, err, nil)
		be.Equal(t, sb.String(), "2023-11-14T22:13:20Z WARN request method=GET status=200\n")
		sb.Free()
	})

	t.Run("quoted string", func(t *testing.T) {
		var sb strings.Builder
		h := NewTextHandler(&sb, LevelInfo)
		r := Record{
			Time:    time.Unix(1700000000, 0),
			Message: "test",
			Level:   LevelInfo,
			Attrs:   []Attr{String("msg", "hello world")},
		}
		_ = h.Handle(r)
		be.Equal(t, sb.String(), "2023-11-14T22:13:20Z INFO test msg=\"hello world\"\n")
		sb.Free()
	})

	t.Run("bool attr", func(t *testing.T) {
		var sb strings.Builder
		h := NewTextHandler(&sb, LevelInfo)
		r := Record{
			Time:    time.Unix(1700000000, 0),
			Message: "flags",
			Level:   LevelInfo,
			Attrs:   []Attr{Bool("yes", true), Bool("no", false)},
		}
		_ = h.Handle(r)
		be.Equal(t, sb.String(), "2023-11-14T22:13:20Z INFO flags yes=true no=false\n")
		sb.Free()
	})

	t.Run("float64 attr", func(t *testing.T) {
		var sb strings.Builder
		h := NewTextHandler(&sb, LevelInfo)
		r := Record{
			Time:    time.Unix(1700000000, 0),
			Message: "metric",
			Level:   LevelInfo,
			Attrs:   []Attr{Float64("elapsed", 1.5)},
		}
		_ = h.Handle(r)
		be.Equal(t, sb.String(), "2023-11-14T22:13:20Z INFO metric elapsed=1.5\n")
		sb.Free()
	})

	t.Run("uint64 attr", func(t *testing.T) {
		var sb strings.Builder
		h := NewTextHandler(&sb, LevelInfo)
		r := Record{
			Time:    time.Unix(1700000000, 0),
			Message: "id",
			Level:   LevelInfo,
			Attrs:   []Attr{Uint64("val", uint64(42))},
		}
		_ = h.Handle(r)
		be.Equal(t, sb.String(), "2023-11-14T22:13:20Z INFO id val=42\n")
		sb.Free()
	})

	t.Run("time attr", func(t *testing.T) {
		var sb strings.Builder
		h := NewTextHandler(&sb, LevelInfo)
		r := Record{
			Time:    time.Unix(1700000000, 0),
			Message: "event",
			Level:   LevelInfo,
			Attrs:   []Attr{Time("at", time.Unix(1700000000, 0))},
		}
		_ = h.Handle(r)
		be.Equal(t, sb.String(), "2023-11-14T22:13:20Z INFO event at=2023-11-14T22:13:20Z\n")
		sb.Free()
	})

	t.Run("duration attr", func(t *testing.T) {
		var sb strings.Builder
		h := NewTextHandler(&sb, LevelInfo)
		r := Record{
			Time:    time.Unix(1700000000, 0),
			Message: "timing",
			Level:   LevelInfo,
			Attrs:   []Attr{Duration("took", 3*time.Second+500*time.Millisecond)},
		}
		_ = h.Handle(r)
		be.Equal(t, sb.String(), "2023-11-14T22:13:20Z INFO timing took=3.5s\n")
		sb.Free()
	})

	t.Run("empty attr", func(t *testing.T) {
		var sb strings.Builder
		h := NewTextHandler(&sb, LevelInfo)
		r := Record{
			Time:    time.Unix(1700000000, 0),
			Message: "test",
			Level:   LevelInfo,
			Attrs:   []Attr{String("val", "")},
		}
		_ = h.Handle(r)
		be.Equal(t, sb.String(), "2023-11-14T22:13:20Z INFO test val=\"\"\n")
		sb.Free()
	})

	t.Run("quoted attr", func(t *testing.T) {
		var sb strings.Builder
		h := NewTextHandler(&sb, LevelInfo)
		r := Record{
			Time:    time.Unix(1700000000, 0),
			Message: "test",
			Level:   LevelInfo,
			Attrs:   []Attr{String("expr", "a=b")},
		}
		_ = h.Handle(r)
		be.Equal(t, sb.String(), "2023-11-14T22:13:20Z INFO test expr=\"a=b\"\n")
		sb.Free()
	})
}

func TestLogger(t *testing.T) {
	t.Run("New", func(t *testing.T) {
		var sb strings.Builder
		h := NewTextHandler(&sb, LevelInfo)
		l := New(&h)
		be.Equal(t, l.Handler(), Handler(&h))
		sb.Free()
	})

	t.Run("Enabled", func(t *testing.T) {
		var sb strings.Builder
		h := NewTextHandler(&sb, LevelWarn)
		l := New(&h)
		be.True(t, !l.Enabled(LevelInfo))
		be.True(t, l.Enabled(LevelWarn))
		sb.Free()
	})

	t.Run("Info", func(t *testing.T) {
		var sb strings.Builder
		h := NewTextHandler(&sb, LevelInfo)
		l := New(&h)
		l.Info("hello", String("key", "val"))
		got := sb.String()
		// Check suffix (time prefix varies).
		be.True(t, len(got) > 20)
		be.True(t, strings.Contains(got, "INFO hello key=val\n"))
		sb.Free()
	})

	t.Run("Debug/filtered", func(t *testing.T) {
		var sb strings.Builder
		h := NewTextHandler(&sb, LevelInfo)
		l := New(&h)
		l.Debug("hidden")
		be.Equal(t, sb.String(), "")
		sb.Free()
	})

	t.Run("Warn", func(t *testing.T) {
		var sb strings.Builder
		h := NewTextHandler(&sb, LevelInfo)
		l := New(&h)
		l.Warn("caution")
		be.True(t, strings.Contains(sb.String(), "WARN caution\n"))
		sb.Free()
	})

	t.Run("Error", func(t *testing.T) {
		var sb strings.Builder
		h := NewTextHandler(&sb, LevelInfo)
		l := New(&h)
		l.Error("fail", Int("code", 500))
		be.True(t, strings.Contains(sb.String(), "ERROR fail code=500\n"))
		sb.Free()
	})

	t.Run("Log", func(t *testing.T) {
		var sb strings.Builder
		h := NewTextHandler(&sb, LevelInfo)
		l := New(&h)
		l.Log(LevelInfo, "via log")
		be.True(t, strings.Contains(sb.String(), "INFO via log\n"))
		sb.Free()
	})
}

func TestDefault(t *testing.T) {
	// Save and restore default logger.
	saved := defaultLogger
	defer func() { defaultLogger = saved }()

	t.Run("ensureDefault", func(t *testing.T) {
		// Reset the lazy-init guard so the default is rebuilt from scratch.
		defaultOnce.Init()
		defaultLogger = nil
		d := Default()
		be.True(t, d != nil)
		be.True(t, d.Enabled(LevelInfo))
		be.True(t, !d.Enabled(LevelDebug))
	})

	t.Run("SetDefault", func(t *testing.T) {
		var sb strings.Builder
		h := NewTextHandler(&sb, LevelDebug)
		l := New(&h)
		SetDefault(&l)
		be.True(t, Default() == &l)
		be.True(t, Default().Enabled(LevelDebug))
		sb.Free()
	})

	t.Run("Info", func(t *testing.T) {
		var sb strings.Builder
		h := NewTextHandler(&sb, LevelInfo)
		l := New(&h)
		SetDefault(&l)
		Info("pkg info", Int("port", 8080))
		be.True(t, strings.Contains(sb.String(), "INFO pkg info port=8080\n"))
		sb.Free()
	})

	t.Run("Debug", func(t *testing.T) {
		var sb strings.Builder
		h := NewTextHandler(&sb, LevelDebug)
		l := New(&h)
		SetDefault(&l)
		Debug("pkg debug")
		be.True(t, strings.Contains(sb.String(), "DEBUG pkg debug\n"))
		sb.Free()
	})

	t.Run("Warn", func(t *testing.T) {
		var sb strings.Builder
		h := NewTextHandler(&sb, LevelInfo)
		l := New(&h)
		SetDefault(&l)
		Warn("pkg warn")
		be.True(t, strings.Contains(sb.String(), "WARN pkg warn\n"))
		sb.Free()
	})

	t.Run("Error", func(t *testing.T) {
		var sb strings.Builder
		h := NewTextHandler(&sb, LevelInfo)
		l := New(&h)
		SetDefault(&l)
		Error("pkg error")
		be.True(t, strings.Contains(sb.String(), "ERROR pkg error\n"))
		sb.Free()
	})

	t.Run("Log", func(t *testing.T) {
		var sb strings.Builder
		h := NewTextHandler(&sb, LevelInfo)
		l := New(&h)
		SetDefault(&l)
		Log(LevelInfo, "pkg log")
		be.True(t, strings.Contains(sb.String(), "INFO pkg log\n"))
		sb.Free()
	})
}
