package main

import (
	"testing"
)

func BenchmarkNonArena(b *testing.B) {
	b.ReportAllocs()

	for b.Loop() {
		AllocateGarbage(false)
	}
}

func BenchmarkArena(b *testing.B) {
	b.ReportAllocs()

	for b.Loop() {
		AllocateGarbage(true)
	}
}
