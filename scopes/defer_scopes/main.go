package main

import "fmt"

func main() {
	v := 1
	// #4
	defer fmt.Println(v)

	v = 2
	defer func() {
		// #3
		fmt.Println(v)
	}()

	v = 3
	defer func(val int) {
		// #2
		fmt.Println(val)
	}(v)

	v = 4
	// #1
	fmt.Println(v)
}
