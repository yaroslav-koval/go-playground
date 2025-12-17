package benchmark

var _ Stack[int] = &stackSlice[int]{}

type stackSlice[T any] struct {
	stack []T
}

func (s *stackSlice[T]) Push(value T) {
	s.stack = append(s.stack, value)
}

func (s *stackSlice[T]) Fetch() (T, bool) {
	if len(s.stack) == 0 {
		var zero T
		return zero, false
	}

	last := s.stack[len(s.stack)-1]
	s.stack = s.stack[:len(s.stack)-1]
	return last, true
}

func (s *stackSlice[T]) GetLast() (T, bool) {
	if len(s.stack) == 0 {
		var zero T
		return zero, false
	}

	return s.stack[len(s.stack)-1], true
}
