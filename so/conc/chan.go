package conc

import (
	"solod.dev/so/c"
	"solod.dev/so/mem"
	"solod.dev/so/time"
)

// Chan is a thread-safe FIFO channel, similar to Go's built-in `chan T`.
// It carries values by copy: Send copies its argument into the channel and
// Recv copies a value out, just like `chan T` in Go. To pass large payloads
// without copying, use a channel of pointers (Chan[*T]).
//
// It supports buffered mode (created with n > 0) and unbuffered rendezvous mode
// (n == 0), where each send blocks until a receiver takes the value. Exactly one
// of the two backing engines is non-nil, chosen at creation time.
type Chan[T any] struct {
	buf *Buffer     // non-nil for buffered channels (n > 0)
	rdv *Rendezvous // non-nil for unbuffered channels (n == 0)
}

// NewChan creates a channel of T backed by alloc. n is the buffer size:
// n > 0 makes it buffered, n == 0 makes it an unbuffered rendezvous channel.
// Call [Chan.Free] exactly once when done.
//
//so:inline
func NewChan[T any](alloc mem.Allocator, n int) Chan[T] {
	_n := n
	_vsize := c.Sizeof[T]()
	var _ch Chan[T]
	c.Assert(_n >= 0, "conc: chan size must be >= 0")
	if _n > 0 {
		_ch.buf = NewBuffer(alloc, _vsize, _n)
	} else {
		_ch.rdv = NewRendezvous(alloc, _vsize)
	}
	return _ch
}

// Send copies v into the channel, blocking until there is room (buffered)
// or a receiver takes it (unbuffered). Sending on a closed channel panics.
//
// Send is thread-safe.
//
//so:inline
func (ch *Chan[T]) Send(v T) {
	_v := v
	if ch.buf != nil {
		ch.buf.Send(any(&_v))
	} else {
		ch.rdv.Send(any(&_v))
	}
}

// SendTimeout copies v into the channel, waiting up to d for room (buffered)
// or a receiver (unbuffered). A zero or negative d makes it non-blocking.
//
// Returns [Ok] if the value was sent, [Timeout] if the deadline passed first,
// or [Closed] if the channel is closed. Unlike [Chan.Send], it does not panic
// on a closed channel.
//
// A non-blocking send on an unbuffered channel reports [Timeout] unless a
// receiver is already parked and takes the value in the brief window it is offered.
//
// SendTimeout is thread-safe.
//
//so:inline
func (ch *Chan[T]) SendTimeout(v T, d time.Duration) Status {
	_v := v
	var _st Status
	if ch.buf != nil {
		_st = ch.buf.SendTimeout(any(&_v), d)
	} else {
		_st = ch.rdv.SendTimeout(any(&_v), d)
	}
	return _st
}

// Recv copies the next value into dst and reports whether one was received.
// It returns false when the channel is closed and no buffered values remain,
// in which case dst is left untouched.
//
// Recv is thread-safe.
//
//so:inline
func (ch *Chan[T]) Recv(dst *T) bool {
	var _ok bool
	if ch.buf != nil {
		_ok = ch.buf.Recv(any(dst))
	} else {
		_ok = ch.rdv.Recv(any(dst))
	}
	return _ok
}

// RecvTimeout copies the next value into dst, waiting up to d for one. A zero
// or negative d makes it non-blocking. Returns [Ok] with dst filled, [Timeout]
// if the deadline passed first, or [Closed] if the channel is closed and
// drained. dst is left untouched unless [Ok] is returned.
//
// RecvTimeout is thread-safe.
//
//so:inline
func (ch *Chan[T]) RecvTimeout(dst *T, d time.Duration) Status {
	var _st Status
	if ch.buf != nil {
		_st = ch.buf.RecvTimeout(any(dst), d)
	} else {
		_st = ch.rdv.RecvTimeout(any(dst), d)
	}
	return _st
}

// Close closes the channel. Subsequent sends panic; receivers drain remaining
// buffered values and then report false. Closing a closed channel panics.
//
// Close is thread-safe and may run concurrently with Send and Recv but
// it must be called exactly once; a repeated or concurrent Close panics.
//
//so:inline
func (ch *Chan[T]) Close() {
	if ch.buf != nil {
		ch.buf.Close()
	} else {
		ch.rdv.Close()
	}
}

// Free releases the channel's resources. The channel is unusable afterward.
// Free should only be called once; it's not thread-safe.
//
//so:inline
func (ch *Chan[T]) Free() {
	if ch.buf != nil {
		ch.buf.Free()
	} else {
		ch.rdv.Free()
	}
}
