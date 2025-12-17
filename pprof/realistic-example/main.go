package main

import (
	"net/http"
	_ "net/http/pprof"
	"runtime"
)

// Configuration for profiling
func enableProfiling() {
	// Enable mutex and block profiling
	runtime.SetMutexProfileFraction(1)
	runtime.SetBlockProfileRate(1)

	// Start pprof server on a different port
	go func() {
		http.ListenAndServe("localhost:6061", nil)
	}()
}

// Simulate a memory-intensive operation
func processRequest(w http.ResponseWriter, r *http.Request) {
	// Allocate memory unnecessarily
	data := make([]byte, 10<<20) // 10MB
	_ = data

	w.Write([]byte("Processed"))
}

func main() {
	enableProfiling()

	// Main application server
	http.HandleFunc("/process", processRequest)
	http.ListenAndServe(":8080", nil)
}
