package main

import "fmt"

func main() {
	slice := []string{"Golang", "Java", "daiki"}
	options := slice[2:]

	if len(options) > 0 {
		fmt.Println(slice)
		fmt.Println(options)
	}
}
