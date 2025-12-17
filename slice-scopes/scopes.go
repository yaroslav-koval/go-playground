package slice_scopes

import "fmt"

type Animal struct {
	name       string
	population int
}

func nullifyPopulations(animals []Animal) []Animal {
	fmt.Println("Animals before modification:", animals)

	for i := range animals {
		animals[i].population = 0
	}

	fmt.Println("Animals after modification:", animals)

	return animals
}
