package testing

//so:embed keep.h
var keep_h string

// Keep prevents the compiler from optimizing away code whose result would
// otherwise be unused, ensuring benchmarked work is actually executed. It emits
// no instructions and has no runtime cost beyond a compiler barrier.
//
// Pass the address of the value to retain. This is most useful for operations
// with no result to consume, such as a method with no return value:
//
//	for b.Loop() {
//		x.Store(1)
//		testing.Keep(&x)
//	}
//
//so:extern so_keep
func Keep[T any](p *T) {
	_ = p
}
