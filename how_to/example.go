package main

import (
	"fmt"
	"strconv"
)

func main() {
	var number int = 12
	_ = number
	arrNumber := func(n int) []int {
				s := strconv.Itoa(n)
				out := make([]int, len(s))
				for i := 0; i < len(s); i++ {
					out[i] = int(s[i] - '0')
				}
				return out
			}(number)
	fmt.Println(arrNumber)
}
