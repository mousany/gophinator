package main

import "os"

func main() {
	path := os.Args[1]
	file, err := os.Create(path)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	file.Write([]byte("hello world\n"))
}
