package main

import (
	"fmt"

	"os"
)

func tree(dir string, indent string) {
	items, _ := os.ReadDir(dir)
	for _, item := range items {
		if item.IsDir() {
			fmt.Println(indent + item.Name())
			tree(item.Name(), indent+"  ")
		} else {
			fmt.Println(indent + item.Name())
		}
	}
}

func main() {
	var dir string
	if len(os.Args) < 2 {
		dir = "."
	} else {
		dir = os.Args[1]
	}

	tree(dir, "")
}
