package main

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"os"
	"pulp"

	"github.com/kr/pretty"
	"golang.org/x/tools/go/ast/astutil"
)

func vPrintf(format string, args ...interface{}) {
	if !*verbose {
		return
	}
	fmt.Printf(format, args...)
}

func replace(sourceName, source string) ([]byte, error) {
	fset := token.NewFileSet()
	expr, err := parser.ParseFile(fset, sourceName, source, parser.AllErrors)
	if err != nil {
		return nil, err
	}

	shouldReturnOuter := false
	var outerErr error

	result := astutil.Apply(expr, func(cr *astutil.Cursor) bool {
		if source := detect(cr.Node()); source != nil {
			*source = (*source)[1 : len(*source)-1] // removes the backticks or the " from the string literal

			g := pulp.NewGenerator()
			parser := pulp.NewParser(*source)
			tree, err := parser.Parse()
			if err != nil {
				shouldReturnOuter = true
				outerErr = err
				return false
			}
			vPrintf("ast: %v\n", pretty.Sprint(tree))
			if parser.Error != nil {
				fmt.Fprint(os.Stderr, parser.Error)
				os.Exit(-1)
			}
			tree.Gen(g)
			vPrintf("gen: %v\n", g.Out())
			cr.Replace(&ast.BasicLit{Value: g.Out()})
			return false
		}
		return true
	}, nil)

	if shouldReturnOuter {
		return nil, fmt.Errorf("parser error: %v", outerErr)
	}

	retBuf := &bytes.Buffer{}
	if err := format.Node(retBuf, fset, result); err != nil {
		return nil, err
	}

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
