package benchmarks

import (
	"runtime"
	"testing"
)

// BenchmarkGoroutinesSwitch measures the cost of one goroutine switch, and
// compares it with a switch that also has to change the OS thread.
//
// Run: go test -run '^$' -bench GoroutinesSwitch ./benchmarks/
//
// Both subtests use the same ping-pong over two unbuffered channels. Each side
// sends, then blocks on the reply, so only one goroutine can make progress at
// a time. One round trip forces exactly 2 switches, so ns/op / 2 is one
// switch. The number also holds the channel's own work on every send and
// receive, so it is an upper bound, not the bare switch alone.
//
// Both set GOMAXPROCS to 1 inside. With more Ps the runtime moves the goroutines
// between threads and wakes idle threads that spin, which adds OS calls to
// the number (checked with ps -M: 5 busy threads with 12 Ps, 1 with 1 P).
//
// goroutines: both goroutines share one P and one thread, so a switch never
// leaves the Go scheduler (checked with top: about 180 OS context switches
// per second at 6M round trips per second).
//
// main locked to OS thread: the loop goroutine calls runtime.LockOSThread.
// Its thread then runs no other goroutine, and it runs on no other thread.
// So every handoff must move the single P to another thread and back, and
// each goroutine switch becomes one OS thread switch (checked with top: 2 OS
// context switches per round trip). The other goroutine is not locked: the
// runtime gives it an idle thread, and the idle list is LIFO, so in practice
// it is the same thread every time (checked with ps -M: 2 busy threads).
// Do not lock both goroutines: the P then passes through a third idle
// thread, so each goroutine switch costs 2 thread switches.
//
// macOS cannot pin a thread to a core, so the woken thread may start on
// another core. The locked number is a thread handoff that may cross cores,
// not a context switch on one core. It also holds the runtime's own work to
// pass the P between threads.
//
// # Measured
//
// Apple M3 Pro, macOS, Go 1.27.1, 2026-09-27, ns/op per round trip:
//   - goroutines: 163-170, one switch about 80-85ns.
//   - main locked to OS thread: 5,200-5,500, one switch about 2.6-2.8us,
//     about 30x a goroutine switch.
func BenchmarkGoroutinesSwitch(b *testing.B) {
	b.Run("goroutines", func(b *testing.B) {
		defer runtime.GOMAXPROCS(runtime.GOMAXPROCS(1))

		channelPingPong(b)
	})

	b.Run("main locked to OS thread", func(b *testing.B) {
		defer runtime.GOMAXPROCS(runtime.GOMAXPROCS(1))

		runtime.LockOSThread()
		defer runtime.UnlockOSThread()

		channelPingPong(b)
	})
}

func channelPingPong(b *testing.B) {
	b.Helper()

	ping := make(chan struct{})
	pong := make(chan struct{})
	done := make(chan struct{})

	go func() {
		defer close(done)

		for range ping {
			pong <- struct{}{}
		}
	}()

	b.ResetTimer()

	for range b.N {
		ping <- struct{}{}
		<-pong
	}

	b.StopTimer()
	close(ping)
	<-done
}
