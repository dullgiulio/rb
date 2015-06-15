package main

import (
	"flag"
	"fmt"

	"github.com/dullgiulio/rb"
)

func main() {
	flag.Parse()

	structs := rb.NewStructs()

	for _, arg := range flag.Args() {
		structs.ParseFile(arg)
	}

	for _, s := range structs.All() {
		fmt.Printf("%s\n", s)
	}
}
