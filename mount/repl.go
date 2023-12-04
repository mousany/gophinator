package main

import (
	"fmt"
)

func main() {
	fmt.Println("Tell me something good...")
	for {
		var s string
		fmt.Scanln(&s)
		if s == "quit" {
			break
		}
		fmt.Println("You said:", s)
	}
}
