package main

import (
	"fmt"
	"go/ast"
	"go/build"
	"go/importer"
	"go/parser"
	"go/token"
	"go/types"
	"path/filepath"
)

// modulePath is the module whose game packages describe the values the renderer
// emits.
const modulePath = "github.com/natefinch/council4"

// gamePackages are the packages the renderer emits literals from. mtg/game is
// the root; the rest are its value sub-packages, plus opt, which supplies the
// generic optional wrapper several game fields are declared with.
var gamePackages = []string{
	"mtg/game",
	"mtg/game/color",
	"mtg/game/compare",
	"mtg/game/cost",
	"mtg/game/counter",
	"mtg/game/mana",
	"mtg/game/types",
	"mtg/game/zone",
	"opt",
}

// loadPackages type-checks every package in gamePackages against the sources
// under root.
//
// It uses the standard library's source importer rather than
// golang.org/x/tools/go/packages because the module has no third-party
// dependencies and must keep it that way. The source importer resolves imports
// through go/build, so the command must run with its working directory inside
// this module.
func loadPackages(root string) (map[string]*types.Package, error) {
	fset := token.NewFileSet()
	imp := importer.ForCompiler(fset, "source", nil)
	loaded := make(map[string]*types.Package, len(gamePackages))
	for _, rel := range gamePackages {
		path := modulePath + "/" + rel
		pkg, err := checkPackage(fset, imp, filepath.Join(root, filepath.FromSlash(rel)), path)
		if err != nil {
			return nil, fmt.Errorf("loading %s: %w", path, err)
		}
		loaded[path] = pkg
	}
	return loaded, nil
}

// checkPackage parses and type-checks the buildable non-test Go files in dir as
// path.
//
// go/build supplies the file list so that build constraints are honoured and the
// files arrive already sorted, which keeps the declaration order recorded for
// each constant reproducible.
func checkPackage(fset *token.FileSet, imp types.Importer, dir, path string) (*types.Package, error) {
	built, err := build.ImportDir(dir, 0)
	if err != nil {
		return nil, err
	}
	files := make([]*ast.File, 0, len(built.GoFiles))
	for _, name := range built.GoFiles {
		file, parseErr := parser.ParseFile(fset, filepath.Join(dir, name), nil, parser.SkipObjectResolution)
		if parseErr != nil {
			return nil, parseErr
		}
		files = append(files, file)
	}
	conf := types.Config{Importer: imp}
	return conf.Check(path, fset, files, nil)
}
