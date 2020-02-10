// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package initflagparse defines an Analyzer that checks for invalid
// usages of flag.Parse during package initialization
package initflagparse

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

const Doc = `check for usage of flag.Parse during modules init

The initflagparse analyzer reports incorrect uses of flag.Parse during
the init of the modules.`

var Analyzer = &analysis.Analyzer{
	Name:     "initflagparse",
	Doc:      Doc,
	Requires: []*analysis.Analyzer{inspect.Analyzer},
	Run:      run,
}

func run(pass *analysis.Pass) (interface{}, error) {
	inspect := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)

	nodeFilter := []ast.Node{
		(*ast.FuncDecl)(nil),
	}
	inspect.Preorder(nodeFilter, func(n ast.Node) {
		x := n.(*ast.FuncDecl)
		if x.Recv == nil && x.Name.Name == "init" {
			checkForFlagParse(pass, x)
		}
	})
	return nil, nil
}

// checkForFlagParse check the code inside a node and fail if it finds a
// flag.Parse call
func checkForFlagParse(pass *analysis.Pass, x ast.Node) {
	ast.Inspect(x, func(n ast.Node) bool {
		x, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}
		fun, ok := x.Fun.(*ast.SelectorExpr)
		if !ok {
			return true
		}

		module, ok := fun.X.(*ast.Ident)
		if !ok {
			return true
		}

		if ok && module.Name == "flag" && fun.Sel.Name == "Parse" {
			pass.ReportRangef(x, "flag.Parse during package initialization")
			return false
		}
		return true
	})
}
