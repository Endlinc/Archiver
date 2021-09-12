package main

import (
	"Archiver/function"
	"fmt"
)

func main() {
	fmt.Println("Let's see how it goes.")

	err := function.Compress("/Users/ming/test.tar", []string{"/Users/ming", "/Users/ming/symlink"})
	if err != nil {
		fmt.Println(err.Error())
		return
	}
}
