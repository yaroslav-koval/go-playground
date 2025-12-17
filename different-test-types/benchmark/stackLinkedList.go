package benchmark

var _ Stack[int] = &stackLinkedList[int]{}

type stackLinkedListNode[T any] struct {
	value T
	prev  *stackLinkedListNode[T]
}

type stackLinkedList[T any] struct {
	tail *stackLinkedListNode[T]
}

func (s *stackLinkedList[T]) Push(value T) {
	if s.tail == nil {
		s.tail = &stackLinkedListNode[T]{
			value: value,
			prev:  nil,
		}
		return
	}

	s.tail = &stackLinkedListNode[T]{
		value: value,
		prev:  s.tail,
	}
}

func (s *stackLinkedList[T]) Fetch() (T, bool) {
	if s.tail == nil {
		var zero T
		return zero, false
	}

	value := s.tail.value
	s.tail = s.tail.prev
	return value, true
}

func (s *stackLinkedList[T]) GetLast() (T, bool) {
	if s.tail == nil {
		var zero T
		return zero, false
	}

	return s.tail.value, true
}
