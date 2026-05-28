package main

import "fmt"

func main() {
	names := []string{"joao", "maria"}
	for i, value := range names {
		_ = i
		fmt.Println(value)
	}
}
