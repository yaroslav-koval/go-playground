package main

import (
	"arena"
	"context"
	"fmt"
	"os/signal"
	"runtime"
	"runtime/trace"
	"syscall"

	"github.com/yaroslav-koval/go-playground/pkg/tracing"
)

const (
	iterations = 10000
	sliceSize  = 1000
)

func main() {
	err := run()
	if err != nil {
		fmt.Println("Error:", err)
	}
}

func run() (err error) {
	tr, err := tracing.StartTracing("arena")
	if err != nil {
		return err
	}

	defer func() {
		err = tr.Close()
	}()

	fmt.Println("Trace file:", tr.FileName)

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer cancel()

	trace.WithRegion(ctx, "make_allocation", func() {
		allocateSliceByMake()
	})

	trace.WithRegion(ctx, "multiple_arenas", func() {
		allocateSliceByArenaFrequentCleanup()
	})

	trace.WithRegion(ctx, "single_arena", func() {
		allocateSliceByArena()
	})

	return
}

func allocateSliceByMake() {
	storer := make([][]int, iterations)

	for j := range iterations {
		s := make([]int, sliceSize)

		storer[j] = s

		for i := range s {
			s[i] = i
		}
	}

	runtime.KeepAlive(storer)

	runtime.GC()
}

func allocateSliceByArena() {
	a := newArenaAllocator()
	defer a.Cleanup()

	for range iterations {
		s := a.CreateSlice(sliceSize)

		for i := range s {
			s[i] = i
		}
	}
}

func allocateSliceByArenaFrequentCleanup() {
	cyclesToClean := iterations / 20
	a := newArenaAllocator()

	for j := range iterations {

		s := a.CreateSlice(sliceSize)

		for i := range s {
			s[i] = i
		}

		if j%cyclesToClean == 0 {
			a.Cleanup()

			a = newArenaAllocator()
		}
	}

	a.Cleanup()
}

func newArenaAllocator() *arenaAllocator {
	return &arenaAllocator{
		ar: arena.NewArena(),
	}
}

type arenaAllocator struct {
	ar *arena.Arena
}

func (aa arenaAllocator) CreateSlice(size int) []int {
	return arena.MakeSlice[int](aa.ar, size, size)
}

func (aa arenaAllocator) Cleanup() {
	aa.ar.Free()
}
