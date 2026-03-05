// Package ssa предоставляет функции для построения SSA представления
package ssa

import (
	"fmt"
	"go/ast"
	"go/importer"
	"go/parser"
	"go/token"
	"go/types"

	"golang.org/x/tools/go/ssa"
)

// Builder отвечает за построение SSA из исходного кода Go
type Builder struct {
	fset *token.FileSet
}

// NewBuilder создаёт новый экземпляр Builder
func NewBuilder() *Builder {
	return &Builder{
		fset: token.NewFileSet(),
	}
}

// ParseAndBuildSSA парсит исходный код Go и создаёт SSA представление
// Возвращает SSA программу и функцию по имени
func (b *Builder) ParseAndBuildSSA(source string, funcName string) (*ssa.Function, error) {
	ssaPkg, err2 := b.ParseAndBuildSSAPkg([]string{source})
	if err2 != nil {
		return nil, err2
	}

	fn := ssaPkg.Func(funcName)
	if fn == nil {
		return nil, fmt.Errorf("func %q not found", funcName)
	}
	return fn, nil
}

func (b *Builder) ParseAndBuildSSAPkg(sources []string) (*ssa.Package, error) {
	if len(sources) == 0 {
		return nil, fmt.Errorf("no sources provided")
	}

	prog := ssa.NewProgram(b.fset, 0)
	conf := types.Config{
		Importer: importer.Default(),
	}
	info := &types.Info{
		Types:      make(map[ast.Expr]types.TypeAndValue),
		Defs:       make(map[*ast.Ident]types.Object),
		Uses:       make(map[*ast.Ident]types.Object),
		Implicits:  make(map[ast.Node]types.Object),
		Instances:  make(map[*ast.Ident]types.Instance),
		Scopes:     make(map[ast.Node]*types.Scope),
		Selections: make(map[*ast.SelectorExpr]*types.Selection),
	}

	var files []*ast.File
	var pkgName string
	for i, source := range sources {
		fileName := fmt.Sprintf("source_%d.go", i)
		file, err := parser.ParseFile(b.fset, fileName, source, parser.AllErrors)
		if err != nil {
			return nil, fmt.Errorf("parse error in %s: %w", fileName, err)
		}

		if pkgName == "" {
			pkgName = file.Name.Name
		} else if file.Name.Name != pkgName {
			return nil, fmt.Errorf("all files must belong to one package: got %q and %q", pkgName, file.Name.Name)
		}

		files = append(files, file)
	}

	pkg, err := conf.Check("main", b.fset, files, info)
	if err != nil {
		return nil, fmt.Errorf("type error: %w", err)
	}

	for _, imp := range pkg.Imports() {
		if prog.ImportedPackage(imp.Path()) == nil {
			prog.CreatePackage(imp, nil, nil, true)
		}
	}

	ssaPkg := prog.CreatePackage(pkg, files, info, true)
	prog.Build()

	return ssaPkg, nil
}
