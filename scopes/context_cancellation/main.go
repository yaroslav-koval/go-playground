package main

import (
	"context"
	"fmt"
)

func main() {
	ctx := context.Background()
	f1(ctx)
}

func f1(ctx context.Context) {
	ctx, cancel := context.WithCancel(ctx)
	_ = cancel

	f2(ctx)

	// only cancel() here can help avoid deadlock
	<-ctx.Done() // deadlock here

	fmt.Println("context is cancelled from f2")
}

func f2(ctx context.Context) {
	ctx, cancel := context.WithCancel(ctx)
	cancel()
	<-ctx.Done()
	fmt.Println("f2. context cancel signal received")
}
