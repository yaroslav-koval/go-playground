package benchmark

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"runtime"
	"testing"
	"unsafe"
)

// 'testing.B' directives

// b.Loop()
// b.ReportAllocs()
// b.ReportMetric()
// b.RunParallel()
// b.SetBytes()
// b.SetParallelism()

// b.StartTimer()
// b.ResetTimer()
// b.StopTimer()
// b.Elapsed()

type stackTestStruct struct {
	value int
}

func BenchmarkUpdateLastInSlice(b *testing.B) {
	b.Log("Outer benchmark. N is", b.N)
	s := stackSlice[stackTestStruct]{}
	s.Push(stackTestStruct{100})

	b.StartTimer()

	b.Run("FetchAndPush", func(b *testing.B) {
		b.Log("FetchAndPush benchmark. N is", b.N)
		for range b.N {
			last, ok := s.Fetch()
			if !ok {
				b.Errorf("stackSlice must be initialized\n")
				b.FailNow()
			}
			last.value = 500
			s.Push(last)
		}
	})

	b.StopTimer()
	b.Log("Elapsed:", b.Elapsed())
	b.ResetTimer()
	b.StartTimer()

	v, ok := s.GetLast()
	require.True(b, ok)
	assert.Equal(b, 500, v.value)

	b.Run("UpdateLast", func(b *testing.B) {
		b.Log("UpdateLast benchmark. N is", b.N)
		for range b.N {
			last, ok := s.GetLast()
			if !ok {
				b.Errorf("stackSlice must be initialized\n")
				b.FailNow()
			}
			last.value = 500
		}
	})

	b.StopTimer()
	b.Log("Elapsed:", b.Elapsed())

	v, ok = s.GetLast()
	require.True(b, ok)
	assert.Equal(b, 500, v.value)
}

func BenchmarkUpdateLastInLinkedList(b *testing.B) {
	b.Log("Outer benchmark. N is", b.N)
	s := stackLinkedList[stackTestStruct]{}
	s.Push(stackTestStruct{100})

	b.StartTimer()

	b.Run("FetchAndPush", func(b *testing.B) {
		b.Log("FetchAndPush benchmark. N is", b.N)
		for range b.N {
			last, ok := s.Fetch()
			if !ok {
				b.Errorf("stackSlice must be initialized\n")
				b.FailNow()
			}
			last.value = 500
			s.Push(last)
		}
	})

	b.StopTimer()
	b.Log("Elapsed:", b.Elapsed())
	b.ResetTimer()
	b.StartTimer()

	v, ok := s.GetLast()
	require.True(b, ok)
	assert.Equal(b, 500, v.value)

	b.Run("UpdateLast", func(b *testing.B) {
		b.Log("UpdateLast benchmark. N is", b.N)
		for range b.N {
			last, ok := s.GetLast()
			if !ok {
				b.Errorf("stackSlice must be initialized\n")
				b.FailNow()
			}
			last.value = 500
		}
	})

	b.StopTimer()
	b.Log("Elapsed:", b.Elapsed())

	v, ok = s.GetLast()
	require.True(b, ok)
	assert.Equal(b, 500, v.value)
}

func BenchmarkLoop(b *testing.B) {
	s := []int{}
	i := 0

	b.Log("Started benchmark. b.N:", b.N)

	for b.Loop() {
		i++
		s = append(s, 1)
	}

	b.Logf("Stopped benchmark. Last iteration %d. b.N: %d\n", i, b.N)
}

type allocsTest struct {
	value   int
	pointer *int
}

func BenchmarkReportAllocs(b *testing.B) {
	// ReportAllocs starts report allocations in the heap (not stack)
	b.ReportAllocs()

	var outscopedValue interface{}
	for b.Loop() {
		outscopedValue = &allocsTest{
			value:   b.N,
			pointer: &b.N,
		}
		outscopedValue = &allocsTest{
			value:   b.N,
			pointer: &b.N,
		}

	}

	// needed to store value in heap
	b.Logf("Outscoped value: %v\n", outscopedValue)

	// Explanation. ReportAllocs adds next reports
	// 32 B/op: in 64-bit OS allocsTest{int,pointer} consumes 16 bytes (8+8). It was allocated 2 times
	// 2 allocs/op: 2 memory allocations
	// so, system allocated 16-bytes struct 2 times, so we have 32B/op and 2 allocs/op
}

func BenchmarkReportAllocsOnSlice(b *testing.B) {
	b.ReportAllocs()

	s := []allocsTest{}

	for b.Loop() {
		// Doesn't matter len or cap provided - it benchmarks size in the heap
		s = make([]allocsTest, 0, 10)
		s = make([]allocsTest, 10)
	}

	b.Logf("Slice len: %d. b.N: %d\n", len(s), b.N)

	// Explanation. Look at BenchmarkReportAllocs for basic info.
	// 320 B/op	  2 allocs/op
	// each make allocates 10*16 bytes = 160.
	// Make is called 2 times during loop = 2 allocs/op.
	// so, 160*2 = 320 B/op
}

// TODO research why 1792 and not 1600?
// runtime/sizeclasses.go
func BenchmarkMakeRoundup(b *testing.B) {
	b.ReportAllocs()

	s := []allocsTest{}

	for b.Loop() {
		s = make([]allocsTest, 100)
	}

	b.Logf("Slice len: %d, cap: %d. b.N: %d\n", len(s), cap(s), b.N)

	s = append(s, allocsTest{})
	b.Logf("Slice len: %d, cap: %d", len(s), cap(s))

	// Explanation. Look at BenchmarkReportAllocsOnSlice for basic info.
	// 1792 B/op	       1 allocs/op
	//

	runtime.Breakpoint()
}

func BenchmarkReportMetric(b *testing.B) {
	i := 0.
	for b.Loop() {
		i++
	}
	b.ReportMetric(i, "httpCalls/op")
}

func BenchmarkRunParallel(b *testing.B) {
	b.Log("Outer benchmark. N is", b.N)
	s := stackLinkedListAsync[stackTestStruct]{}
	for range 500 {
		s.Push(stackTestStruct{100})
	}

	// works as maxGoroutines=2*GOMAXPROCS
	b.SetParallelism(2)

	b.RunParallel(func(pb *testing.PB) {
		b.Log("Number of goroutines:", runtime.NumGoroutine())

		for pb.Next() {
			last, ok := s.Fetch()
			if !ok {
				b.Errorf("stackSlice must be initialized\n")
				b.FailNow()
			}
			last.value = 500
			s.Push(last)
		}
	})

	v, ok := s.GetLast()
	require.True(b, ok)
	assert.Equal(b, 500, v.value)
}

type setBytesBenchmarkStruct struct {
	value int
}

func BenchmarkSetBytes(b *testing.B) {
	b.Log("Outer benchmark b.N is", b.N)

	s := []setBytesBenchmarkStruct{{value: 10}}

	// let's say that I process n bytes during each single operation (= b.Loop iteration)
	// Input in megabytes = {n from SetBytes} / {1 000 000: bytes to Megabytes}
	// Megabytes processed = {Input in megabytes} * {Number of iterations}
	// result = {Megabytes processed} / {Elapsed time}
	b.SetBytes(int64(unsafe.Sizeof(setBytesBenchmarkStruct{})))
	// close example here:
	// setBytesBenchmarkStruct size is 8B or 0.000008MB
	// processed: 0.000008 * 1000000000 (num of iter) = 8,000MB
	// result: 8000MB / 1s = 8000 MB/s

	b.StartTimer()
	for b.Loop() {
		item := s[0]
		item.value += 1
		s = s[0:0]
		s = append(s, item)
	}
	b.StopTimer()

	b.Logf("Finished b.N: %d. Elapsed: %v\n", b.N, b.Elapsed())
}
