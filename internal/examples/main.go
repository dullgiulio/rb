package main

import (
	"go/ast"
)

type TestStruct1 struct {
	Field        int
	AnotherField *ast.Ident
	unexported   string
}

type TestStruct2 struct {
	Field   *ast.Ident
	useless string
}

func main() {
	// Do nothing really.
}
