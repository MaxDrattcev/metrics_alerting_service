package analyzer

import (
	"go/ast"
	"go/types"
	"path/filepath"
	"strings"

	"golang.org/x/tools/go/analysis"
)

// Analyzer проверяет:
// 1) использование встроенной panic
// 2) вызовы log.Fatal / os.Exit вне func main() пакета main.
var Analyzer = &analysis.Analyzer{
	Name: "projectlint",
	Doc:  "reports panic usage and log.Fatal/os.Exit outside main.main",
	Run:  run,
}

func run(pass *analysis.Pass) (any, error) {
	isMainPkg := pass.Pkg.Name() == "main"

	for _, file := range pass.Files {
		if shouldSkipFile(pass, file) {
			continue
		}
		// 1) Проверяем все верхнеуровневые объявления.
		for _, decl := range file.Decls {
			switch d := decl.(type) {
			case *ast.FuncDecl:
				allowExit := isMainPkg && d.Recv == nil && d.Name != nil && d.Name.Name == "main"
				inspectNode(pass, d, allowExit)
			default:
				inspectNode(pass, d, false)
			}
		}
	}

	return nil, nil
}

func inspectNode(pass *analysis.Pass, node ast.Node, allowExit bool) {
	if node == nil {
		return
	}

	ast.Inspect(node, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}

		// panic(...)
		if ident, ok := call.Fun.(*ast.Ident); ok && isBuiltinPanic(pass, ident) {
			pass.Reportf(call.Pos(), "avoid using built-in panic")
			return true
		}

		// log.Fatal(...)
		// os.Exit(...)
		if sel, ok := call.Fun.(*ast.SelectorExpr); ok {
			if isPkgSelector(pass, sel, "log", "Fatal") && !allowExit {
				pass.Reportf(call.Pos(), "log.Fatal is forbidden outside main.main")
			}
			if isPkgSelector(pass, sel, "os", "Exit") && !allowExit {
				pass.Reportf(call.Pos(), "os.Exit is forbidden outside main.main")
			}
		}

		return true
	})
}

func isBuiltinPanic(pass *analysis.Pass, ident *ast.Ident) bool {
	obj := pass.TypesInfo.Uses[ident]
	if obj == nil {
		return false
	}
	b, ok := obj.(*types.Builtin)
	return ok && b.Name() == "panic"
}

func isPkgSelector(pass *analysis.Pass, sel *ast.SelectorExpr, pkgName, funcName string) bool {
	if sel.Sel == nil || sel.Sel.Name != funcName {
		return false
	}

	x, ok := sel.X.(*ast.Ident)
	if !ok {
		return false
	}

	obj := pass.TypesInfo.Uses[x]
	pkgObj, ok := obj.(*types.PkgName)
	if !ok || pkgObj.Imported() == nil {
		return false
	}

	imported := pkgObj.Imported()
	return imported.Name() == pkgName && imported.Path() == pkgName
}

func shouldSkipFile(pass *analysis.Pass, f *ast.File) bool {
	tf := pass.Fset.File(f.Pos())
	if tf == nil {
		return false
	}
	name := filepath.ToSlash(tf.Name())

	if strings.Contains(name, "/internal/mocks/") {
		return true
	}

	for _, cg := range f.Comments {
		for _, c := range cg.List {
			if strings.Contains(c.Text, "Code generated") {
				return true
			}
		}
	}
	return false
}
