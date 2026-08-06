package analyzer

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"
)

// Analyzer сообщает о вызовах panic, а также log.Fatal / os.Exit
// вне функции main пакета main.
var Analyzer = &analysis.Analyzer{
	Name: "criticalcheck",
	Doc:  "reports usage of panic, and log.Fatal/os.Exit outside main.main",
	Run:  run,
}

func run(pass *analysis.Pass) (interface{}, error) {
	for _, file := range pass.Files {
		// имя текущей функции; nil если не внутри функции
		var currentFunc *ast.FuncDecl

		ast.Inspect(file, func(n ast.Node) bool {
			switch node := n.(type) {
			case *ast.FuncDecl:
				currentFunc = node
				return true

			case *ast.CallExpr:
				checkCall(pass, node, file, currentFunc)
			}
			return true
		})
	}
	return nil, nil
}

func checkCall(pass *analysis.Pass, call *ast.CallExpr, file *ast.File, fn *ast.FuncDecl) {
	switch f := call.Fun.(type) {

	case *ast.Ident:
		if f.Name == "panic" {
			pass.Reportf(call.Pos(), "usage of panic is forbidden")
		}

	case *ast.SelectorExpr:
		pkgIdent, ok := f.X.(*ast.Ident)
		if !ok {
			return
		}
		name := pkgIdent.Name + "." + f.Sel.Name

		if name == "log.Fatal" || name == "os.Exit" {
			if !isMainMain(pass, file, fn) {
				pass.Reportf(call.Pos(),
					"calling %s is forbidden outside of main.main", name)
			}
		}
	}
}

func isMainMain(pass *analysis.Pass, file *ast.File, fn *ast.FuncDecl) bool {
	if fn == nil {
		return false
	}
	return pass.Pkg.Name() == "main" &&
		fn.Name.Name == "main" &&
		fn.Recv == nil
}
