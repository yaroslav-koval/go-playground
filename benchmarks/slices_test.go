package benchmarks

import "testing"

// all of these tests take in average 1 ns/op or less.
// Doesn't matter, benchmarked with allocation or without.
func BenchmarkSlices(b *testing.B) {
	b.Run("slice insert", func(b *testing.B) {
		s := make([]int, b.N)

		b.ResetTimer()

		for i := range b.N {
			s[i] = i
		}
	})

	b.Run("slice append", func(b *testing.B) {
		a := make([]int, 0, b.N)

		b.ResetTimer()

		for i := range b.N {
			a = append(a, i)
		}
	})

	b.Run("slice allocation + insert", func(b *testing.B) {
		s := make([]int, b.N)

		for i := range b.N {
			s[i] = i
		}
	})

	b.Run("slice allocation + append", func(b *testing.B) {
		a := make([]int, 0, b.N)

		for i := range b.N {
			a = append(a, i)
		}
	})
}
