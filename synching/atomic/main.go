package main

import (
	"log/slog"
	"sync/atomic"
	"unsafe"
)

func main() {
	nativeAtomics()
	unsafeAtomics()
}

func nativeAtomics() {
	i := atomic.Int32{}
	// returns current value
	_ = i.Load()
	// adds delta, returns new value
	_ = i.Add(10)
	// sets value
	i.Store(15)
	// sets new value and returns old value
	_ = i.Swap(20)
	// swaps if old value matched
	ok := i.CompareAndSwap(20, 30)
	// bitwise AND operation
	i.And(63)
	// bitwise OR operation
	i.Or(63)

	b := atomic.Bool{}
	_ = b.Load()
	b.Store(true)
	_ = b.Swap(false)
	ok = b.CompareAndSwap(false, true)

	value := "my_value"
	value2 := "other_value"
	p := atomic.Pointer[string]{}
	_ = p.Load()
	p.Store(&value)
	_ = p.Swap(&value2)
	ok = p.CompareAndSwap(&value2, &value)

	// all
	_ = atomic.Value{} // any value
	_ = atomic.Int32{}
	_ = atomic.Uint32{}
	_ = atomic.Int64{}
	_ = atomic.Uint64{}
	_ = atomic.Uintptr{}
	_ = atomic.Pointer[string]{}

	// used in:
	slog.Default()
	// as "var defaultLogger atomic.Pointer[Logger]"

	_ = ok
}

func unsafeAtomics() {
	vInt32 := int32(10)
	vUint32 := uint32(10)
	vInt64 := int64(10)
	vUint64 := uint64(10)
	vPointer := unsafe.Pointer(&vInt32)
	vUintptr := uintptr(vPointer)

	_ = atomic.LoadInt32(&vInt32)
	_ = atomic.LoadUint32(&vUint32)
	_ = atomic.LoadInt64(&vInt64)
	_ = atomic.LoadUint64(&vUint64)
	_ = atomic.LoadPointer(&vPointer)
	_ = atomic.LoadUintptr(&vUintptr)

	// and all other methods like atomic.Int32 has in the function above
}
