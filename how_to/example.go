package main

import (
	"fmt"
	"strconv"
)

func main() {
	
	isPalindrome := func(n int) bool {
		nArr := func(n int) []int {
				s := strconv.Itoa(n)
				out := make([]int, len(s))
				for i := 0; i < len(s); i++ {
					out[i] = int(s[i] - '0')
				}
				return out
			}(n)
		new := []int{}
		var final bool = false
		_ = final
		for index, _ := range nArr {
			_ = index
			func() { new = append(new, nArr[((len(nArr) - 1) - index)]) }()
		}
		for index, value := range nArr {
			_ = index
			if (value != new[index]) {
				final = false
			} else {
				final = true
			}
		}
		return final
	}
	
	fmt.Println(isPalindrome(121))
}
