package main

import (
	"fmt"
	"strconv"
)

func main() {
	
	isPalindrome := func(n int) {
		nArr := func(n int) []int {
				s := strconv.Itoa(n)
				out := make([]int, len(s))
				for i := 0; i < len(s); i++ {
					out[i] = int(s[i] - '0')
				}
				return out
			}(n)
		new := []int{}
		for index, value := range nArr {
			_ = index
			if (len(new) == 0) {
				func() { new = append(new, value) }()
			} else {
				fmt.Println((len(new) - 1))
				new[(len(new) - 1)] = value
			}
		}
		fmt.Println(new)
	}
	
	isPalindrome(12)
}
