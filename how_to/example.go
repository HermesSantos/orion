package main

import "fmt"

func main() {
	names := []string{"sasuke", "itachi"}
	fmt.Println(names)
	func() { names = append(names, "naruto") }()
	fmt.Println(names)
}
