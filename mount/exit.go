package main

import (
	"fmt"
	"os"
	"strconv"
)

func main() {
	code, err := strconv.Atoi(os.Args[1])
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	fmt.Println("exiting with code", code)
	os.Exit(code)
}
