package main

import (
	"arena"
)

func AllocateGarbage(isArena bool) [][]int {
	var alloc allocator
	defer func() {
		alloc.Cleanup()
	}()

	if !isArena {
		alloc = &defaultAllocator{}
	} else {
		alloc = &arenaAllocator{
			ar: arena.NewArena(),
		}
	}

	iterations := 10000

	slices := make([][]int, 0, iterations)

	for range iterations {
		s := alloc.CreateSlice(1000)

		for i := range s {
			s[i] = i
		}

		// removing 10'th item
		// capacity is unchanged
		s = append(s[:10], s[11:]...)

		slices = append(slices, s)
	}

	return slices
}

type allocator interface {
	CreateSlice(size int) []int
	Cleanup()
}

type defaultAllocator struct {
}

func (da *defaultAllocator) CreateSlice(size int) []int {
	return make([]int, size)
}

func (da *defaultAllocator) Cleanup() {}

type arenaAllocator struct {
	ar *arena.Arena
}

func (aa arenaAllocator) CreateSlice(size int) []int {
	return arena.MakeSlice[int](aa.ar, size, size)
}

func (aa arenaAllocator) Cleanup() {
	aa.ar.Free()
}
