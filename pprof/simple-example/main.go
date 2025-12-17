package main

import (
	"net/http"
	_ "net/http/pprof" // Import pprof
	"time"
)

// A function with memory leak
func leakyFunction() {
	leak := make([]string, 0)

	for {
		leak = append(leak, "memory leak")
		time.Sleep(time.Millisecond * 100)
	}
}

// CPU-intensive function
func cpuIntensive() *int {
	v := 0

	go func() {
		for {
			for i := 0; i < 1000000; i++ {
				//v = v + 1
			}
			time.Sleep(time.Millisecond * 100)
		}
	}()

	return &v
}

func main() {
	//runtime.SetBlockProfileRate(1)

	// Enable pprof
	go func() {
		http.ListenAndServe("localhost:6060", nil)
	}()

	// Start problematic routines
	go leakyFunction()

	go cpuIntensive()

	// Keep the program running
	select {}
}
