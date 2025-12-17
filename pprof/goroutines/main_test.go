package main

import (
	"fmt"
	"runtime"
	"testing"
	"time"
)

func TestGenerateGoroutines(t *testing.T) {
	generateGoroutines()

	time.Sleep(3 * time.Second)

	printGoroutines()
}

func printGoroutines() {
	buf := make([]byte, 1<<16)
	runtime.Stack(buf, true)
	fmt.Printf("%s\n", buf)
}
