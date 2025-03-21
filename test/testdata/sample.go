package main

import (
	"fmt"
	"os"
)

func main() {
	fmt.Println("TEMP")
	fmt.Println(os.Args[1:])
	fmt.Println("env: ", len(os.Environ()))
}