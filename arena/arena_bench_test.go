package main

import (
	"fmt"
	"testing"
)

func BenchmarkNonArena(b *testing.B) {
	b.ReportAllocs()

	slices := AllocateGarbage(false)

	sum := 0
	for _, s := range slices {
		for i := range s {
			sum += i
		}
	}

	fmt.Println(sum)
}

func BenchmarkArena(b *testing.B) {
	b.ReportAllocs()

	slices := AllocateGarbage(true)

	sum := 0
	for _, s := range slices {
		for i := range s {
			sum += i
		}
	}

	fmt.Println(sum)
}
