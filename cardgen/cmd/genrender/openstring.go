package main

import (
	"fmt"
	"go/types"
	"slices"
	"strings"
)

// openStringTypes are named string types that carry exported constants but are
// not closed sets: lowering synthesizes values for them, such as the choice key
// "oracle-mana-color", which no constant names.
//
// They render as a conversion of the string value rather than as a constant
// name, matching what the hand-written renderers emitted. The list is explicit
// because openness cannot be read off the type: game.ChoiceKey and types.Sub are
// both named strings with constants, but an undeclared types.Sub is a bug that
// must fail the render, while an undeclared game.ChoiceKey is routine.
var openStringTypes = []string{
	"game.ChoiceKey",
	"game.LinkedKey",
	"game.ResultKey",
}

func isOpenString(ref string) bool { return slices.Contains(openStringTypes, ref) }

// checkOpenStringTypes fails if a listed type is missing, is not a string, or is
// not reachable, so the list cannot silently rot as mtg/game changes.
func checkOpenStringTypes(graph *reachable, pkgs map[string]*types.Package) error {
	var problems []string
	for _, ref := range openStringTypes {
		named, ok := lookupNamed(pkgs, ref)
		if !ok {
			problems = append(problems, ref+" does not exist")
			continue
		}
		basic, ok := named.Underlying().(*types.Basic)
		if !ok || basicKind(basic) != kindString {
			problems = append(problems, ref+" is not a named string type")
			continue
		}
		if !containsNamed(graph.Structs, ref) && !reachableAsField(graph, ref) {
			problems = append(problems, ref+" is not reachable from a rendered value")
		}
	}
	if len(problems) > 0 {
		return fmt.Errorf("openStringTypes is stale: %s", strings.Join(problems, "; "))
	}
	return nil
}

// lookupNamed resolves a qualified name such as game.ChoiceKey against the
// loaded packages.
func lookupNamed(pkgs map[string]*types.Package, ref string) (*types.Named, bool) {
	pkgName, typeName, ok := strings.Cut(ref, ".")
	if !ok {
		return nil, false
	}
	for _, pkg := range pkgs {
		if pkg.Name() != pkgName {
			continue
		}
		obj := pkg.Scope().Lookup(typeName)
		if obj == nil {
			continue
		}
		named, ok := obj.Type().(*types.Named)
		return named, ok
	}
	return nil, false
}

// reachableAsField reports whether any reachable struct has a field of the named
// type, directly or inside a container.
func reachableAsField(graph *reachable, ref string) bool {
	c := newClassifier(graph)
	for _, named := range graph.Structs {
		for _, f := range c.fieldsOf(named) {
			for t := f.Type; t != nil; t = t.Elem {
				if t.Ref() == ref {
					return true
				}
			}
		}
	}
	return false
}
