package main

import (
	"fmt"
	"sync"
	"time"
)

func main() {
	receive() // panic here
	send()    // panic here
	fmt.Println()
	sel()
}

func send() {
	var ch chan struct{}

	wg := sync.WaitGroup{}

	wg.Go(func() {
		time.Sleep(100 * time.Millisecond)

		// panic
		v, ok := <-ch
		fmt.Println(v, ok)
	})

	// panic
	ch <- struct{}{}

	wg.Wait()
}

func receive() {
	var ch chan struct{}

	wg := sync.WaitGroup{}

	wg.Go(func() {
		time.Sleep(100 * time.Millisecond)

		// panic
		ch <- struct{}{}
	})

	// panic
	v, ok := <-ch
	fmt.Println(v, ok)

	wg.Wait()
}

func sel() {
	var ch chan struct{}

	// second case just ignored, no panic
	select {
	case <-time.After(100 * time.Millisecond):
		fmt.Println("timeouted")
	case ch <- struct{}{}:
		fmt.Println("sent value")
	}
}
