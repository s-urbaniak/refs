package refs

import (
	"fmt"
	"go/ast"
	"go/importer"
	"go/parser"
	"go/token"
	"go/types"
	"testing"
)

var src string = `
package b

type foo struct {
    s string
    i int
}
var f = &foo{}

func bar(f *foo) {
    f.s = "baz"
    f.i = 1
}
`

type ParsedNode struct {
	Ident   *ast.Ident
	Name    string
	fileset *token.FileSet
}

func (n *ParsedNode) HasOffset(o int) bool {
	pos, end := n.Offset()
	return o >= pos && o < end
}

func (n *ParsedNode) Offset() (pos, end int) {
	pos, end = n.fileset.Position(n.Ident.Pos()).Offset, n.fileset.Position(n.Ident.End()).Offset
	return
}

func (n *ParsedNode) String() string {
	return fmt.Sprintf(
		"%s: type %T name %s",
		n.fileset.Position(n.Ident.Pos()),
		n.Ident, n.Name,
	)
}

func find(offset int, f *ast.File, fset *token.FileSet) (ret *ParsedNode) {
	ast.Inspect(f, func(n ast.Node) bool {
		var p *ParsedNode

		switch x := n.(type) {
		case *ast.SelectorExpr:
			p = &ParsedNode{x.Sel, x.Sel.Name, fset}
		case *ast.Ident:
			p = &ParsedNode{x, x.Name, fset}
		}

		if p == nil {
			return true
		}

		if p.HasOffset(offset) {
			ret = p
		}

		return false
	})

	return
}

func TestLookup(t *testing.T) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "src.go", src, 0)
	if err != nil {
		t.Fatal(err)
	}

	p := find(96, f, fset)
	fmt.Println(p.Ident)

	cfg := types.Config{}
	cfg.Importer = importer.Default()
	info := types.Info{}
	info.Defs = make(map[*ast.Ident]types.Object)
	pkg, err := cfg.Check("cmd/src", fset, []*ast.File{f}, &info)
	fmt.Println(pkg, err)
	fmt.Printf("%q\n", info)
}
