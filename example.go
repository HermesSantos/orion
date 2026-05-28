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
	func() interface{} { val := names[len(names)-1]; names = names[:len(names)-1]; return val }()
	func() {
		for _i, _v := range names {
			if fmt.Sprintf("%v", _v) == fmt.Sprintf("%v", "hermes") {
				names = append(names[:_i], names[_i+1:]...)
				break
			}
		}
	}()
	fmt.Println(names[0])
	fmt.Println(names[len(names)-1])
	fmt.Println(names)
}
