package slice_scopes

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// Description:
// Type of the slice is an Animal structure, it is changed inside nullifyPopulations because slice itself is a pointer
// to memory.

// Behavior can be changed in new Go versions due to escape analysis optimization.
// If the change done, a slice can ba passed entirely in stack and bypass behavior described above.

func TestNullifyPopulations(t *testing.T) {
	animals := []Animal{
		{
			name:       "Zebra",
			population: 50,
		},
		{
			name:       "Kangoo",
			population: 70,
		},
	}

	nullifyPopulations(animals)

	for _, animal := range animals {
		assert.Equal(t, 0, animal.population)
	}
}
