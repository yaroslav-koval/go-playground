package main

import (
	"fmt"
	"sync"
)

func main() {
	ch()
	mut()
}

func ch() {
	// descriptor is at /usr/local/go/src/runtime/chan.go
	ch := make(chan int)
	c := ch

	ch2 := make(chan int, 1)
	c = ch2

	ch3 := make(chan int, 2)
	c = ch3
	_ = c
}

func mut() {
	m := sync.Mutex{}
	fmt.Printf("Mutex locked: %v\n", m.TryLock())
	// because of `_ noCopy` it's not copied
	cp := m
	fmt.Printf("Copy locked: %v\n", cp.TryLock())
}
