package main

import (
	"net/http"
	_ "net/http/pprof"
)

const numOfWorkers = 10

func main() {
	generateGoroutines()

	http.ListenAndServe("localhost:6061", nil)
}

func generateGoroutines() {
	ch := make(chan int, 1024)
	ch2 := make(chan int, 1024)

	for i := 0; i < numOfWorkers; i++ {
		go produce(ch)
		go consume(ch, ch2)
		go reader(ch2)
	}
}

func produce(ch chan<- int) {
	i := 0
	for {
		i++
		ch <- i * 5 / 4
	}
}

func consume(ch <-chan int, ch2 chan<- int) {
	for v := range ch {
		ch2 <- v * 10 / 6 * 7 / 4 * 5 / 3 * 50 / 20 * 1234 / 532234 * 234234
	}
}

func reader(ch <-chan int) {
	for v := range ch {
		_ = v
	}
}
