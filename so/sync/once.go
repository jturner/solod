package sync

import "solod.dev/so/sync/atomic"

// Once runs a function exactly once, even when Do
// is called concurrently from multiple threads.
//
// The zero value is not usable; call [Once.Init] before use.
// A Once must not be copied after Init.
type Once struct {
	mu   Mutex
	done atomic.Bool
}

// Init prepares o for use. It must be called exactly once before
// any other method. A Once must not be copied after Init.
func (o *Once) Init() {
	o.mu.Init()
	o.done.Store(false)
}

// Do calls f if and only if Do is being called for the first time for this o.
// If Do is called concurrently, the callers block until the one call to f
// returns; every Do returns only after f has completed.
//
// Because no call to Do returns until the one call to f returns,
// f must not call Do on the same o, or it will deadlock.
func (o *Once) Do(f func()) {
	if o.done.Load() {
		return
	}
	o.mu.Lock()
	if !o.done.Load() {
		f()
		o.done.Store(true)
	}
	o.mu.Unlock()
}

// Free releases the resources held by o. The Once is unusable afterward.
func (o *Once) Free() {
	o.mu.Free()
}
