package main

import (
	"bytes"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"pulp"

	"github.com/kr/pretty"
	"golang.org/x/tools/go/ast/astutil"
)

func replace(sourceName, source string) ([]byte, error) {
	fset := token.NewFileSet()
	expr, err := parser.ParseFile(fset, sourceName, source, parser.AllErrors)
	if err != nil {
		return nil, err
	}

	pretty.Print(expr)

	result := astutil.Apply(expr, func(cr *astutil.Cursor) bool {
		if source := detect(cr.Node()); source != nil {
			g := &pulp.Generator{}
			pulp.NewParser(*source).Parse().Gen(g)
			cr.Replace(&ast.BasicLit{Value: g.Out()})
			return false
		}
		return true
	}, nil)

	retBuf := &bytes.Buffer{}
	format.Node(retBuf, fset, result)

	return retBuf.Bytes(), nil
}

func detect(node ast.Node) *string {
	if r, ok := node.(*ast.CallExpr); ok {
		if t, ok1 := r.Fun.(*ast.SelectorExpr); ok1 {
			if t.Sel.Name == "L" {
				if sourceLit, ok2 := r.Args[0].(*ast.BasicLit); ok2 {
					return &sourceLit.Value
				}
			}
		}
	}

	return nil
}
