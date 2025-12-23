package main

import (
	"arena"
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"runtime/trace"
	"strings"
	"syscall"
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
	dir, ok := os.LookupEnv("PWD")
	if !ok {
		return errors.New("env PWD not found")
	}

	// PWD contains go.mod directory, if run in IDE
	if !strings.HasSuffix(dir, "arena") {
		dir = filepath.Join(dir, "arena")
	}

	traceFileName := filepath.Join(dir, "trace")

	fmt.Println("Trace file:", traceFileName)

	var (
		f *os.File
	)

	f, err = os.OpenFile(traceFileName, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}

	defer func() {
		err = f.Close()
	}()

	err = trace.Start(f)
	if err != nil {
		return err
	}

	defer trace.Stop()

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
