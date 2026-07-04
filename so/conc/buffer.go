package conc

import (
	"solod.dev/so/c"
	"solod.dev/so/mem"
	"solod.dev/so/sync"
	"solod.dev/so/time"
)

// Status reports the outcome of a timed channel operation.
type Status int

const (
	Ok      Status = iota // the value was transferred
	Timeout               // the deadline elapsed before a transfer
	Closed                // the channel was closed
)

// Buffer is the non-generic engine behind a buffered [Chan]:
// a thread-safe FIFO of values stored by copy in a ring buffer. Each element
// is vsize bytes; Send copies from the source pointer into the ring and Recv
// copies out into the destination pointer. In most cases, using [Chan] is
// more convenient.
type Buffer struct {
	alloc mem.Allocator

	mu       sync.Mutex
	notEmpty sync.Cond // signaled when an item becomes available
	notFull  sync.Cond // signaled when a slot frees

	// buf is a ring buffer of values in the channel.
	// bhead and btail are indices into buf,
	// and bcount is the number of values buffered.
	buf    mem.Array
	bhead  int
	btail  int
	bcount int

	closed bool // true after Close
}

// NewBuffer creates a buffered channel holding
// up to size values of vsize bytes each.
func NewBuffer(alloc mem.Allocator, vsize int, size int) *Buffer {
	c.Assert(size > 0, "conc: buffered chan size must be > 0")

	ch := mem.Alloc[Buffer](alloc)
	ch.alloc = alloc
	ch.buf = mem.NewArray(alloc, vsize, size)
	ch.bhead, ch.btail, ch.bcount = 0, 0, 0
	ch.closed = false

	ch.mu.Init()
	ch.notEmpty.Init(&ch.mu)
	ch.notFull.Init(&ch.mu)
	return ch
}

// Send copies v into the channel, blocking while the channel
// is full (back-pressure). Panics if the channel is closed.
func (ch *Buffer) Send(v any) {
	ch.mu.Lock()
	for ch.bfull() && !ch.closed {
		ch.notFull.Wait()
	}
	if ch.closed {
		ch.mu.Unlock()
		panic("conc: send on closed channel")
	}
	ch.bpush(v)
	ch.notEmpty.Signal()
	ch.mu.Unlock()
}

// SendTimeout copies v into the channel, waiting up to d for room
// if the buffer is full. A zero or negative d makes it non-blocking.
//
// Returns Ok if the value was stored, Timeout if the deadline passed while
// the buffer stayed full, or Closed if the channel is closed.
func (ch *Buffer) SendTimeout(v any, d time.Duration) Status {
	deadline := time.Now().Add(d)
	ch.mu.Lock()
	timedOut := false
	for ch.bfull() && !ch.closed && !timedOut {
		dur := int64(time.Until(deadline))
		timedOut = ch.notFull.WaitFor(dur)
	}
	if ch.closed {
		ch.mu.Unlock()
		return Closed
	}
	if ch.bfull() {
		// Still full: the deadline passed before a slot freed.
		ch.mu.Unlock()
		return Timeout
	}
	// A slot may have freed right at the deadline; store anyway.
	ch.bpush(v)
	ch.notEmpty.Signal()
	ch.mu.Unlock()
	return Ok
}

// Recv copies the next value from the channel into dst.
// Reports whether a value was received: false means the channel
// is closed and drained, and dst is left untouched.
//
// dst must be a non-nil pointer to storage of at least vsize bytes.
func (ch *Buffer) Recv(dst any) bool {
	ch.mu.Lock()
	for ch.bempty() && !ch.closed {
		ch.notEmpty.Wait()
	}
	if ch.bempty() && ch.closed {
		ch.mu.Unlock()
		return false
	}
	ch.bpop(dst)
	ch.notFull.Signal()
	ch.mu.Unlock()
	return true
}

// RecvTimeout copies the next value from the channel into dst, waiting
// up to d for one if the buffer is empty. A zero or negative d makes it
// non-blocking.
//
// dst must be a non-nil pointer to storage of at least vsize bytes.
//
// Returns Ok with dst filled, Timeout if the deadline passed while the buffer
// stayed empty, or Closed if the channel is closed and drained. dst is left
// untouched unless Ok is returned.
func (ch *Buffer) RecvTimeout(dst any, d time.Duration) Status {
	deadline := time.Now().Add(d)
	ch.mu.Lock()
	timedOut := false
	for ch.bempty() && !ch.closed && !timedOut {
		dur := int64(time.Until(deadline))
		timedOut = ch.notEmpty.WaitFor(dur)
	}
	if !ch.bempty() {
		// A value is available (possibly delivered right at the deadline), so
		// it wins over both close and timeout.
		ch.bpop(dst)
		ch.notFull.Signal()
		ch.mu.Unlock()
		return Ok
	}
	if ch.closed {
		ch.mu.Unlock()
		return Closed
	}
	ch.mu.Unlock()
	return Timeout
}

// Close marks the channel closed. Subsequent sends panic; receivers drain any
// buffered items and then return false. Closing a closed channel panics.
func (ch *Buffer) Close() {
	ch.mu.Lock()
	if ch.closed {
		ch.mu.Unlock()
		panic("conc: close of closed channel")
	}
	ch.closed = true
	ch.notEmpty.Broadcast()
	ch.notFull.Broadcast()
	ch.mu.Unlock()
}

// bfull reports whether the ring buffer is at capacity.
func (ch *Buffer) bfull() bool { return ch.bcount == ch.buf.Len() }

// bempty reports whether the ring buffer is empty.
func (ch *Buffer) bempty() bool { return ch.bcount == 0 }

// bpush copies v into the tail of the ring buffer.
func (ch *Buffer) bpush(v any) {
	ch.buf.Store(ch.btail, v)
	ch.btail = (ch.btail + 1) % ch.buf.Len()
	ch.bcount++
}

// bpop copies the value at the head of the ring buffer into dst.
func (ch *Buffer) bpop(dst any) {
	ch.buf.Load(ch.bhead, dst)
	ch.bhead = (ch.bhead + 1) % ch.buf.Len()
	ch.bcount--
}

// Free releases the channel's resources. The channel is unusable afterward.
// Call it once fully done; a channel may be drained after Close.
func (ch *Buffer) Free() {
	ch.mu.Free()
	ch.notEmpty.Free()
	ch.notFull.Free()
	ch.buf.Free()
	mem.Free(ch.alloc, ch)
}
