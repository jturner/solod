package main

import (
	"solod.dev/so/conc"
	"solod.dev/so/mem"
	"solod.dev/so/time"
)

func testChan() {
	testChan_Buffered()
	testChan_ProducerConsumer()
	testChan_Unbuffered()
	testChan_UnbufferedMultiProducer()
	testChan_CloseDrain()
	testChan_TimeoutBuffered()
	testChan_TimeoutExpires()
	testChan_TimeoutHandoff()
	testChan_TimeoutSend()
}

// Fills a buffered channel without blocking
// and checks that values come back in FIFO order.
func testChan_Buffered() {
	print("- chan buffered...")
	ch := conc.NewChan[int](mem.System, 4)
	for i := range 4 {
		ch.Send(i * 10)
	}
	var v int
	for i := range 4 {
		if !ch.Recv(&v) || v != i*10 {
			panic("wrong buffered value")
		}
	}
	ch.Free()
	println("ok")
}

// sumTask carries a channel and the resulting sum between threads.
type sumTask struct {
	ch  conc.Chan[int]
	sum int
}

// consume receives values until the channel is closed and accumulates them.
func consume(arg any) any {
	task := arg.(*sumTask)
	var v int
	for task.ch.Recv(&v) {
		task.sum += v
	}
	return nil
}

// Sends 0..n-1 from the main thread through a small buffered channel
// while a worker thread sums them, exercising back-pressure.
func testChan_ProducerConsumer() {
	print("- chan producer/consumer...")
	const n = 1000
	task := sumTask{ch: conc.NewChan[int](mem.System, 8), sum: 0}

	thr := conc.Go(consume, &task, nil)
	for i := range n {
		task.ch.Send(i)
	}
	task.ch.Close()
	thr.Wait()

	// Sum of 0..999.
	if task.sum != 499500 {
		panic("wrong producer/consumer sum")
	}
	task.ch.Free()
	println("ok")
}

// seqTask for sending a sequence of values to a channel.
type seqTask struct {
	ch conc.Chan[int]
	n  int
}

// produceSeq sends 0..n-1 to the channel and then closes it.
func produceSeq(arg any) any {
	task := arg.(*seqTask)
	for i := 0; i < task.n; i++ {
		task.ch.Send(i)
	}
	task.ch.Close()
	return nil
}

// Receives from an unbuffered channel fed by a worker thread
// and checks the handoff order.
func testChan_Unbuffered() {
	print("- chan unbuffered...")
	task := seqTask{ch: conc.NewChan[int](mem.System, 0), n: 10}

	want := 0
	thr := conc.Go(produceSeq, &task, nil)
	var v int
	for task.ch.Recv(&v) {
		if v != want {
			panic("wrong unbuffered handoff order")
		}
		want++
	}
	thr.Wait()

	if want != 10 {
		panic("missing unbuffered values")
	}
	task.ch.Free()
	println("ok")
}

// rangeTask for sending a range of values to a channel.
type rangeTask struct {
	ch   conc.Chan[int]
	base int
	n    int
}

// produceRange sends base..base+n-1 to the channel.
func produceRange(arg any) {
	task := arg.(*rangeTask)
	for i := 0; i < task.n; i++ {
		task.ch.Send(task.base + i)
	}
}

// Runs several producer threads sending on a single unbuffered channel while
// the main thread receives. Each value 0..N-1 is sent exactly once across
// producers; the receiver checks none is lost or duplicated. This exercises
// the rendezvous handshake with concurrent senders.
func testChan_UnbufferedMultiProducer() {
	print("- chan unbuffered multi-producer...")
	const producers = 4
	const perProducer = 250
	const total = producers * perProducer

	ch := conc.NewChan[int](mem.System, 0)
	opts := conc.PoolOpts{NumThreads: producers}
	p := conc.NewPool(mem.System, opts)

	tasks := make([]rangeTask, producers)
	for i := range tasks {
		tasks[i] = rangeTask{ch: ch, base: i * perProducer, n: perProducer}
		p.Go(produceRange, &tasks[i])
	}

	seen := make([]bool, total)
	var v int
	for range total {
		if !ch.Recv(&v) {
			panic("unexpected close")
		}
		if v < 0 || v >= total || seen[v] {
			panic("lost or duplicated unbuffered value")
		}
		seen[v] = true
	}
	p.Free()
	ch.Free()
	println("ok")
}

// Checks that buffered values survive Close and are drained in order
// before Recv reports the channel closed.
func testChan_CloseDrain() {
	print("- chan close drain...")
	ch := conc.NewChan[int](mem.System, 4)
	for i := 1; i <= 3; i++ {
		ch.Send(i)
	}
	ch.Close()

	seen := 0
	want := 1
	var v int
	for ch.Recv(&v) {
		if v != want {
			panic("wrong drained value")
		}
		want++
		seen++
	}
	if seen != 3 {
		panic("did not drain all buffered values")
	}
	ch.Free()
	println("ok")
}

// Exercises non-blocking SendTimeout/RecvTimeout (d == 0) on a buffered channel
// from a single thread, where the outcomes are fully deterministic: sends fail
// once full, receives fail once empty, and a drained closed channel reports
// Closed.
func testChan_TimeoutBuffered() {
	print("- chan timeout buffered...")
	ch := conc.NewChan[int](mem.System, 2)

	// The buffer holds 2; the third non-blocking send must time out.
	if ch.SendTimeout(10, 0) != conc.Ok || ch.SendTimeout(20, 0) != conc.Ok {
		panic("SendTimeout should succeed with room")
	}
	if ch.SendTimeout(30, 0) != conc.Timeout {
		panic("SendTimeout should time out when full")
	}

	// Drain in FIFO order, then a non-blocking receive must time out.
	var v int
	if ch.RecvTimeout(&v, 0) != conc.Ok || v != 10 {
		panic("wrong first RecvTimeout value")
	}
	if ch.RecvTimeout(&v, 0) != conc.Ok || v != 20 {
		panic("wrong second RecvTimeout value")
	}
	if ch.RecvTimeout(&v, 0) != conc.Timeout {
		panic("RecvTimeout should time out when empty")
	}

	// After close with no buffered values, a receive reports Closed.
	ch.Close()
	if ch.RecvTimeout(&v, 0) != conc.Closed {
		panic("RecvTimeout should report Closed")
	}
	ch.Free()
	println("ok")
}

// Checks that timed operations actually give up at the deadline when no peer
// ever appears: both a send and a receive on an idle unbuffered channel must
// return Timeout rather than block forever.
func testChan_TimeoutExpires() {
	print("- chan timeout expires...")
	ch := conc.NewChan[int](mem.System, 0)

	if ch.SendTimeout(1, 10*time.Millisecond) != conc.Timeout {
		panic("SendTimeout should time out with no receiver")
	}
	var v int
	if ch.RecvTimeout(&v, 10*time.Millisecond) != conc.Timeout {
		panic("RecvTimeout should time out with no sender")
	}
	ch.Free()
	println("ok")
}

// Receives from an unbuffered channel with a deadline while a worker thread
// feeds it with blocking sends. The loop tolerates timeouts and stops on
// Closed, checking the handoff order.
func testChan_TimeoutHandoff() {
	print("- chan timeout handoff...")
	task := seqTask{ch: conc.NewChan[int](mem.System, 0), n: 10}

	thr := conc.Go(produceSeq, &task, nil)
	want := 0
	var v int
	for {
		st := task.ch.RecvTimeout(&v, 50*time.Millisecond)
		if st == conc.Closed {
			break
		}
		if st == conc.Timeout {
			continue // no sender ready yet; keep polling
		}
		if v != want {
			panic("wrong timeout handoff order")
		}
		want++
	}
	thr.Wait()

	if want != 10 {
		panic("missing timeout handoff values")
	}
	task.ch.Free()
	println("ok")
}

// Sends on an unbuffered channel with a deadline while a worker thread drains
// it with blocking receives. Each send retries until a receiver takes it.
func testChan_TimeoutSend() {
	print("- chan timeout send...")
	const n = 100
	task := sumTask{ch: conc.NewChan[int](mem.System, 0), sum: 0}

	thr := conc.Go(consume, &task, nil)
	for i := range n {
		for task.ch.SendTimeout(i, 50*time.Millisecond) != conc.Ok {
			// No receiver ready yet; keep retrying.
		}
	}
	task.ch.Close()
	thr.Wait()

	// Sum of 0..99.
	if task.sum != 4950 {
		panic("wrong timeout send sum")
	}
	task.ch.Free()
	println("ok")
}
