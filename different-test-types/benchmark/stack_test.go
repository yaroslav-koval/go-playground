package benchmark

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestStackSlice(t *testing.T) {
	t.Run("based on slice", func(t *testing.T) {
		s := &stackSlice[int]{}
		testStack(t, s)
	})

	t.Run("based on linked list", func(t *testing.T) {
		s := &stackLinkedList[int]{}
		testStack(t, s)
	})
}

func testStack(t *testing.T, s Stack[int]) {
	t.Helper()

	value, ok := s.Fetch()
	assert.False(t, ok)
	assert.Equal(t, 0, value)

	s.Push(1)
	s.Push(2)
	s.Push(3)

	value, ok = s.GetLast()
	require.True(t, ok)
	assert.Equal(t, 3, value)

	value, ok = s.Fetch()
	require.True(t, ok)
	assert.Equal(t, 3, value)

	value, ok = s.Fetch()
	require.True(t, ok)
	assert.Equal(t, 2, value)

	value, ok = s.Fetch()
	require.True(t, ok)
	assert.Equal(t, 1, value)

	_, ok = s.Fetch()
	require.False(t, ok)
}
