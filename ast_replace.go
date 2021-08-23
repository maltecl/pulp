package pulp

import (
	"fmt"
	"go/ast"
	"go/format"
	goparser "go/parser"
	gotoken "go/token"
	"os"

	"github.com/kr/pretty"
	"golang.org/x/tools/go/ast/astutil"
)

var testGoSource = `
package p1

func render() {
	return pulp.L{"hello world!"}
}

`

var testGoSource2 = ""

func init() {
	// file, err := os.Open("./test/test1.go")
	// if err != nil {
	// 	panic(err)
	// }

	// bs, err := ioutil.ReadAll(file)
	// if err != nil {
	// 	panic(err)
	// }
	// testGoSource2 = string(bs)
}

func replace() {
	fmt.Println("testGoSource2: ", testGoSource2)

	fset := gotoken.NewFileSet()

	expr, err := goparser.ParseFile(fset, "/test/test1.go", testGoSource2, goparser.AllErrors)

	if err != nil {
		panic(err)
	}

	pretty.Print(expr)

	astutil.Apply(expr, func(cr *astutil.Cursor) bool {
		if source := detect(cr.Node()); source != nil {
			g := &Generator{}
			NewParser(*source).Parse().Gen(g)
			cr.Replace(&ast.BasicLit{Value: g.sourceWriter.String()})
			return false
		}
		return true
	}, nil) // Print result

	format.Node(os.Stderr, fset, expr)

}

func detect(node ast.Node) *string {
	if r, ok := node.(*ast.CompositeLit); ok {
		if t, ok1 := r.Type.(*ast.SelectorExpr); ok1 {
			if t.Sel.Name == "L" {
				if sourceLit, ok2 := r.Elts[0].(*ast.BasicLit); ok2 {
					return &sourceLit.Value
				}
			}
		}
	}

	return nil
}
