package benchmarks

import (
	"sync"
	"testing"
)

func BenchmarkMutexes(b *testing.B) {
	b.Run("Mutex", func(b *testing.B) {
		m := &sync.Mutex{}

		b.ResetTimer()

		for range b.N {
			m.Lock()
			m.Unlock()
		}
	})

	b.Run("RWMutex reader RLock RUnlock", func(b *testing.B) {
		m := &sync.RWMutex{}

		b.ResetTimer()

		for range b.N {
			m.RLock()
			m.RUnlock()
		}
	})

	b.Run("RWMutex writer Lock Unlock", func(b *testing.B) {
		m := &sync.RWMutex{}

		b.ResetTimer()

		for range b.N {
			m.Lock()
			m.Unlock()
		}
	})
}
