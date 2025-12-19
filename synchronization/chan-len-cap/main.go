package main

import (
	"fmt"
	"sync"
)

func main() {
	ch := make(chan int, 5)

	wg := sync.WaitGroup{}
	wg.Add(2)
	go func() {
		defer wg.Done()
		ch <- 1
	}()

	go func() {
		defer wg.Done()
		ch <- 2
	}()

	wg.Wait()

	fmt.Printf("Chan len (%v) & cap (%v)\n", len(ch), cap(ch))
}
