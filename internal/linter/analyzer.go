package linter

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

var PanicCheckAnalyzer = analysis.Analyzer{
	Name: "panicCheck",
	Doc:  "check for panic using",
	Run:  run,
	Requires: []*analysis.Analyzer{inspect.Analyzer},
}

func run(pass *analysis.Pass) (interface{}, error) {
	ins := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)

	var currentFunc *ast.FuncDecl

	nodeFilter := []ast.Node{
		(*ast.FuncDecl)(nil),
		(*ast.CallExpr)(nil),
	}

	ins.Preorder(nodeFilter, func(n ast.Node) {
		switch x := n.(type) {

		case *ast.FuncDecl:
			currentFunc = x

		case *ast.CallExpr:
			if ident, ok := x.Fun.(*ast.Ident); ok && ident.Name == "panic" {
				pass.Reportf(x.Pos(), "panic usage is forbidden")
				return
			}

			e, ok := x.Fun.(*ast.SelectorExpr)
			if !ok {
				return
			}

			pkgIdent, ok := e.X.(*ast.Ident)
			if !ok {
				return
			}

			isMainPkg := pass.Pkg.Name() == "main"
			isMainFunc := currentFunc != nil && currentFunc.Name.Name == "main"

			if isMainPkg && !isMainFunc {
				switch pkgIdent.Name {
				case "log":
					if e.Sel.Name == "Fatal" {
						pass.Reportf(x.Pos(), "log.Fatal must be called only in main.main")
					}
				case "os":
					if e.Sel.Name == "Exit" {
						pass.Reportf(x.Pos(), "os.Exit must be called only in main.main")
					}
				}
			}
		}
	})

	return nil, nil
}
