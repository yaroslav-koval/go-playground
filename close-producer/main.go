package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

// task is to close producer properly
// deprecated;
// better example in https://github.com/yaroslav-koval/hange, look for fileprovider

const (
	producersNumber    = 3 // better to keep less than 10 for readability
	recordsPerProducer = 100
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())

	pipe := make(chan int)

	doneCh := startProducing(ctx, pipe)
	doneConsumptionCh := startConsuming(pipe)

	select {
	case s := <-createSignalsChan():
		fmt.Println("Graceful shutdown. Signal: " + s.String())
		cancel()
	case <-doneCh:
	}

	select {
	case <-time.After(3 * time.Second):
		fmt.Println("Timeout!")
	case <-doneConsumptionCh:
		fmt.Println("Consumption is finished")
	}
}

func createSignalsChan() chan os.Signal {
	sigCh := make(chan os.Signal)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	return sigCh
}

func startProducing(ctx context.Context, ch chan<- int) <-chan struct{} {
	ctx, cancel := context.WithCancel(ctx)
	done := make(chan struct{})
	wg := sync.WaitGroup{}

	for i := 0; i < producersNumber; i++ {
		wg.Add(1)

		// pass 'i' if go <=1.21
		go func() {
			dp := newDataPopulator(i*recordsPerProducer, i*recordsPerProducer+recordsPerProducer)

			for {
				time.Sleep(time.Millisecond * 100)

				select {
				case <-ctx.Done():
					fmt.Println("Producer got cancellation signal")
					wg.Done()
					return
				default:
					message, ok := dp.readData()
					if !ok {
						cancel()
						continue
					}
					ch <- message
				}
			}
		}()
	}

	go func() {
		wg.Wait()
		// important: only when all the producers are finished the chan can be closed.
		// in other way there's a chance to get a panic ("send to closed channel")
		// if we have producers in sendq of chan
		close(ch)
		fmt.Println("All the producers finished")
		done <- struct{}{}
	}()

	return done
}

func startConsuming(ch <-chan int) <-chan struct{} {
	doneCh := make(chan struct{})

	go func() {
		for {
			v, ok := <-ch
			if !ok {
				doneCh <- struct{}{}
				return
			}
			fmt.Printf("Value: %d\n", v)
		}
	}()

	return doneCh
}
