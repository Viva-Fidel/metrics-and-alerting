// Package noosexit содержит анализатор, запрещающий прямой вызов os.Exit в функции main пакета main.
package noosexit

import (
	"go/ast"
	"path/filepath"
	"go/types"
	"strings"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

// Doc описывает назначение анализатора.
const Doc = "запрещает прямой вызов os.Exit в функции main пакета main"

// Analyzer проверяет, что в точке входа приложения не вызывается os.Exit напрямую.
var Analyzer = &analysis.Analyzer{
	Name:     "noosexit",
	Doc:      Doc,
	Run:      run,
	Requires: []*analysis.Analyzer{inspect.Analyzer},
}

func run(pass *analysis.Pass) (any, error) {
	if pass.Pkg.Name() != "main" {
		return nil, nil
	}

	insp := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)

	nodeFilter := []ast.Node{(*ast.FuncDecl)(nil)}
	insp.Preorder(nodeFilter, func(n ast.Node) {
		fn, ok := n.(*ast.FuncDecl)
		if !ok || fn.Name.Name != "main" || fn.Body == nil {
			return
		}

		if pos := pass.Fset.Position(fn.Pos()); isGoBuildTempFile(pos.Filename) {
			return
		}

		ast.Inspect(fn.Body, func(node ast.Node) bool {
			call, ok := node.(*ast.CallExpr)
			if !ok {
				return true
			}

			sel, ok := call.Fun.(*ast.SelectorExpr)
			if !ok || sel.Sel.Name != "Exit" {
				return true
			}

			pkgIdent, ok := sel.X.(*ast.Ident)
			if !ok || pkgIdent.Name != "os" {
				return true
			}

			obj := pass.TypesInfo.Uses[pkgIdent]
			if obj == nil {
				return true
			}

			pkgName, ok := obj.(*types.PkgName)
			if !ok || pkgName.Imported().Path() != "os" {
				return true
			}

			pass.Reportf(call.Pos(), "прямой вызов os.Exit в функции main запрещён")
			return true
		})
	})

	return nil, nil
}

func isGoBuildTempFile(filename string) bool {
	for dir := filepath.Dir(filepath.Clean(filename)); ; {
		base := filepath.Base(dir)
		if strings.HasPrefix(base, "go-build") {
			return true
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			return false
		}
		dir = parent
	}
}
