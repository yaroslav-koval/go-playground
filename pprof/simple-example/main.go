package main

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"
	"runtime/pprof"
	"strings"
	"time"
)

// A function with memory leak
func leakyFunction() {
	leak := make([]string, 0)

	for range 100_000 {
		b1 := strings.Builder{}
		for range 100 {
			b1.WriteString("little memory leak")
		}

		b2 := strings.Builder{}
		for range 100 {
			b1.WriteString("little memory leak")
		}

		leak = append(leak, b1.String())
		leak = append(leak, b2.String())
	}
}

// CPU-intensive function
func cpuIntensive() *int {
	v := 0

	for range 1 << 32 {
		v = v + 1
	}

	return &v
}

func main() {
	//runtime.SetBlockProfileRate(1)

	defer func(t time.Time) {
		slog.Info(fmt.Sprintf("Finished in %v", time.Since(t)))
	}(time.Now())

	home, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}

	proj := filepath.Join(home, "repos/go-playground/pprof/simple-example")
	cpuPath := filepath.Join(proj, "cpu.prof")
	memPath := filepath.Join(proj, "mem.prof")

	_ = os.Remove(cpuPath)

	_ = os.Remove(memPath)

	f, err := os.OpenFile(cpuPath, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		panic(err)
	}

	defer func() {
		err = f.Close()
		if err != nil {
			panic(err)
		}
	}()

	err = pprof.StartCPUProfile(f)
	if err != nil {
		panic(err)
	}

	defer pprof.StopCPUProfile()

	cpuIntensive()

	// write heap profile

	memF, err := os.OpenFile(memPath, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		panic(err)
	}
	defer memF.Close()

	runtime.GC()

	leakyFunction()

	err = pprof.WriteHeapProfile(memF)
	if err != nil {
		panic(err)
	}
}
