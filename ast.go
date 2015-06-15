package rb

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
)

type consumer interface {
	consume(ast.Node) consumer
}

type structField struct {
	field *ast.Field
	name  *ast.Ident
	def   []ast.Node
}

func newStructField(field *ast.Field) *structField {
	return &structField{
		field: field,
		def:   make([]ast.Node, 0),
	}
}

func (f *structField) String() string {
	var b bytes.Buffer
	fmt.Fprintf(&b, "%s { ", f.name.Name)
	for _, n := range f.def {
		fmt.Fprintf(&b, "%T(%v), ", n, n)
	}
	return b.String()
}

func (f *structField) consume(n ast.Node) consumer {
	if _, ok := n.(*ast.Field); ok {
		return nil
	}
	if n.Pos() > f.field.End() {
		return nil
	}
	if f.name == nil {
		if i, ok := n.(*ast.Ident); ok {
			f.name = i
			return f
		}
	} else {
		f.def = append(f.def, n)
	}
	return f
}

type structDef struct {
	node   ast.Node
	name   *ast.Ident
	fields []*structField
}

func newStructDef(n ast.Node) *structDef {
	return &structDef{
		node:   n,
		fields: make([]*structField, 0),
	}
}

func (s *structDef) add(f *structField) {
	s.fields = append(s.fields, f)
}

func (s *structDef) String() string {
	var b bytes.Buffer
	fmt.Fprintf(&b, "%s {\n", s.name.Name)
	for _, field := range s.fields {
		fmt.Fprintf(&b, "\t%T{ %v },\n", field, field)
	}
	fmt.Fprintf(&b, "}")
	return b.String()
}

type structsFilter struct {
	curr    *structDef
	structs chan *structDef
}

func newStructsFilter() *structsFilter {
	return &structsFilter{
		structs: make(chan *structDef),
	}
}

func (f *structsFilter) consume(n ast.Node) consumer {
	if _, ok := n.(*ast.TypeSpec); ok {
		f.curr = newStructDef(n)
		return f
	}
	if f.curr == nil {
		return f
	}
	if f.curr.name == nil {
		if i, ok := n.(*ast.Ident); ok {
			f.curr.name = i
		}
		return f
	}
	if n.Pos() < f.curr.node.End() {
		if nf, ok := n.(*ast.Field); ok {
			field := newStructField(nf)
			f.curr.add(field)
			return field
		}
		return f
	}
	f.emit()
	return f
}

func (f *structsFilter) close() {
	if f.curr != nil {
		f.emit()
	}
	close(f.structs)
}

func (f *structsFilter) emit() {
	f.structs <- f.curr
	f.curr = nil
}

type Structs struct {
	fset *token.FileSet
	sf   *structsFilter
	strs []*structDef
}

func NewStructs() *Structs {
	return &Structs{
		fset: token.NewFileSet(),
		sf:   newStructsFilter(),
		strs: make([]*structDef, 0),
	}
}

func (s *Structs) All() []*structDef {
	return s.strs
}

func (s *Structs) ParseFile(filename string) {
	fset := token.NewFileSet()

	f, err := parser.ParseFile(fset, filename, nil, 0)
	if err != nil {
		fmt.Println(err)
		return
	}

	var c consumer
	c = s.sf

	go func() {
		ast.Inspect(f, func(n ast.Node) bool {
			if n == nil {
				return false
			}
			c = c.consume(n)
			if c == nil {
				c = s.sf.consume(n)
			}
			return true
		})
		s.sf.close()
	}()

	for se := range s.sf.structs {
		s.strs = append(s.strs, se)
	}
}
