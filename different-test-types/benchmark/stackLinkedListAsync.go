package benchmark

import "sync"

var _ Stack[int] = &stackLinkedListAsync[int]{}

type stackLinkedListAsync[T any] struct {
	m  sync.Mutex
	ll stackLinkedList[T]
}

func (s *stackLinkedListAsync[T]) Push(value T) {
	s.m.Lock()
	defer s.m.Unlock()
	s.ll.Push(value)
}

func (s *stackLinkedListAsync[T]) Fetch() (T, bool) {
	s.m.Lock()
	defer s.m.Unlock()
	return s.ll.Fetch()
}

func (s *stackLinkedListAsync[T]) GetLast() (T, bool) {
	s.m.Lock()
	defer s.m.Unlock()
	return s.ll.GetLast()
}
