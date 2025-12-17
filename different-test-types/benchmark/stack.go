package benchmark

type Stack[T any] interface {
	Push(value T)
	Fetch() (T, bool)
	GetLast() (T, bool)
}
