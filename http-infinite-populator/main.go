package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
)

type response struct {
	Message string `json:"message"`
}

func main() {
	http.HandleFunc("/spawn", func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Add("Content-Type", "application/json")
		resp, _ := json.Marshal(&response{Message: "Started goroutine population"})

		go populateInfiniteGoroutines()

		_, err := writer.Write(resp)
		if err != nil {
			fmt.Printf("error on writing responce: %s\n", err)
		}
	})
	http.HandleFunc("/spawn-recursively", func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Add("Content-Type", "application/json")
		resp, _ := json.Marshal(&response{Message: "Started recursive"})

		go populateStackRecursively()

		_, err := writer.Write(resp)
		if err != nil {
			fmt.Printf("error on writing responce: %s\n", err)
		}
	})

	fmt.Printf("Server started on localhost:8080\n")
	if err := http.ListenAndServe("localhost:8080", nil); err != nil {
		panic(err)
	}
}

func populateInfiniteGoroutines() {
	exitCh := make(chan int)

	for {
		a := rand.Int()
		b := rand.Int()
		go createVariableInStack(exitCh, a, b)
	}
}

func createVariableInStack(ch chan int, a, b int) {
	sum := a + b
	// block
	ch <- sum
}

func populateStackRecursively() {
	infiniteTraverse(1)
}

func infiniteTraverse(value int) {
	infiniteTraverse(value + 1)
}
