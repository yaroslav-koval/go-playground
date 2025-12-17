package main

import "fmt"

// Alias
type secretTypeAlias = string

const (
	tlsAlias secretTypeAlias = "tls"
)

// New basic type
type secretType string

const (
	tlsSecretType secretType = "tls"
)

func main() {
	var s string
	_ = s

	var al secretTypeAlias
	_ = al
	al = "resdf"
	s = al

	var t secretType
	_ = t
	t = "dsfa"
	s = t

	printAlias(tlsAlias)
	printAlias("random string")

	printType(tlsSecretType)
	printType("random string")

	const str = "sdf"
	printType(str)

	var str2 = "sdf"
	printType(str2)
}

func printAlias(value secretTypeAlias) {
	fmt.Println(value)
}

func printType(value secretType) {
	fmt.Println(value)
}
