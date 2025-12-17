package main

import "fmt"

func main() {
	arr()
	sl()
}

func arr() {
	a := [10]int{}
	a[0] = 0
	a[1] = 1
	fmt.Println("Array len: ", len(a))
	fmt.Println("Array cap: ", cap(a))
}

func sl() {
	a := make([]int, 0, 10)
	a = append(a, 0)
	a = append(a, 1)
	fmt.Println("Slice len: ", len(a))
	fmt.Println("Slice cap: ", cap(a))
}
