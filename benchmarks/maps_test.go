package benchmarks

import "testing"

func BenchmarkMaps(b *testing.B) {
	// ~40 ns/op
	b.Run("map insert", func(b *testing.B) {
		m := make(map[int]int, b.N)

		b.ResetTimer()

		for i := range b.N {
			m[i] = i
		}
	})

	// ~40 ns/op
	b.Run("map update", func(b *testing.B) {
		m := make(map[int]int, b.N)

		for i := range b.N {
			m[i] = i + 1
		}

		b.ResetTimer()

		for i := range b.N {
			m[i] = i
		}
	})

	// ~60 ns/op
	b.Run("map delete", func(b *testing.B) {
		m := make(map[int]int, b.N)

		for i := range b.N {
			m[i] = i
		}

		b.ResetTimer()

		for i := range b.N {
			delete(m, i)
		}
	})
}
