package main

import "fmt"

func main() {
	var name string = "hermes"
	_ = name
	var age int = 45
	_ = age
	var is_male bool = true
	_ = is_male
	if is_male {
		fmt.Println("is male")
	} else if (is_male == false) {
		fmt.Println("is female")
	} else {
		fmt.Println("default value")
	}
	
	helloName := func(name string) string {
		return fmt.Sprintf("hello, %v !", name)
	}
	
	var hello string = helloName("hermes")
	_ = hello
	fmt.Println(hello)
	
	writeScreen := func() {
		fmt.Println("hello, world")
	}
	
	writeScreen()
	names := []string{"hermes", "lucas"}
	for index, value := range names {
		_ = index
		fmt.Println(value)
	}
	func() { names = append(names, "gusta") }()
	fmt.Println(names)
}
