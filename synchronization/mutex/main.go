package main

import (
	"fmt"
	"sync"
)

var m = sync.Mutex{}

func lockMutex() {
	fmt.Println("Locking...")
	m.Lock()
	fmt.Println("Locked")
}

func unlockMutex() {
	fmt.Println("Unlocking...")
	m.Unlock()
	fmt.Println("unlocked")

}

func main() {
	// any workers can lock and unlock mutex
	someLocker()
	someUnlocker()

	// but we get panic("unlock of unlocked mutex") when we unlock unlocked mutex
	unlockMutex() // <- Panic here
}

func someLocker() {
	lockMutex()

	// emulate load
	prevValue := 1
	curValue := 2
	for i := 0; i < 10; i++ {
		prevValue, curValue = curValue, curValue+prevValue
	}
}

func someUnlocker() {
	// emulate load
	prevValue := 1
	curValue := 2
	for i := 0; i < 10; i++ {
		prevValue, curValue = curValue, curValue+prevValue
	}

	unlockMutex()
}
