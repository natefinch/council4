package main

import (
	"fmt"
	"go/constant"
	"go/token"
	"go/types"
	"slices"
	"strings"
)

// enumType is a named type whose exported constants the renderer emits by name.
type enumType struct {
	// Pkg is the package declaring the type.
	Pkg *types.Package
	// Name is the type's name within Pkg.
	Name string
	// Values are the type's exported constants, ordered by declaration.
	Values []enumValue
}

// Ref is the type's qualified name as written in generated source.
func (e enumType) Ref() string { return e.Pkg.Name() + "." + e.Name }

// TableName is the generated map variable holding this type's constant names.
func (e enumType) TableName() string {
	return lowerFirst(exportedPrefix(e.Pkg.Name())) + e.Name + "Literals"
}

// enumValue is one exported constant of an enumType.
type enumValue struct {
	// Name is the constant's name within its package.
	Name string
	// Value is the constant's value, used to detect constants that share a
	// value and so cannot both be map keys.
	Value constant.Value
	// Pos is the declaration position, which orders constants that share a
	// value so the emitted choice is stable.
	Pos token.Pos
	// Aliases are other exported constants with the same value, which the
	// generated table records in a comment rather than as duplicate keys.
	Aliases []string
}

// Ref is the constant's qualified name as written in generated source.
func (v enumValue) Ref(pkg *types.Package) string { return pkg.Name() + "." + v.Name }

// collectEnums returns every named type in pkgs that has at least one exported
// constant, keyed by qualified type name and ordered deterministically.
//
// Types whose constants are unexported are skipped: the renderer emits Go source
// compiled outside these packages, so it can only name exported constants.
func collectEnums(pkgs map[string]*types.Package) []enumType {
	byType := map[*types.TypeName][]enumValue{}
	for _, pkg := range pkgs {
		scope := pkg.Scope()
		for _, name := range scope.Names() {
			obj, ok := scope.Lookup(name).(*types.Const)
			if !ok || !obj.Exported() {
				continue
			}
			named, ok := obj.Type().(*types.Named)
			if !ok || named.Obj().Pkg() == nil || !named.Obj().Exported() {
				continue
			}
			if _, ok := named.Underlying().(*types.Basic); !ok {
				continue
			}
			tn := named.Obj()
			byType[tn] = append(byType[tn], enumValue{Name: name, Value: obj.Val(), Pos: obj.Pos()})
		}
	}

	enums := make([]enumType, 0, len(byType))
	for tn, values := range byType {
		enums = append(enums, enumType{Pkg: tn.Pkg(), Name: tn.Name(), Values: dedupeValues(values)})
	}
	slices.SortFunc(enums, func(a, b enumType) int { return strings.Compare(a.Ref(), b.Ref()) })
	return enums
}

// dedupeValues orders constants by declaration and folds constants that share a
// value into the first one's Aliases.
//
// Two constants with the same value cannot both key the generated map, and
// picking between them by declaration order keeps the emitted name stable as
// unrelated constants are added.
func dedupeValues(values []enumValue) []enumValue {
	slices.SortFunc(values, func(a, b enumValue) int {
		if a.Pos != b.Pos {
			return int(a.Pos - b.Pos)
		}
		return strings.Compare(a.Name, b.Name)
	})
	var unique []enumValue
	index := map[string]int{}
	for _, v := range values {
		key := v.Value.ExactString()
		if at, ok := index[key]; ok {
			unique[at].Aliases = append(unique[at].Aliases, v.Name)
			continue
		}
		index[key] = len(unique)
		unique = append(unique, v)
	}
	return unique
}

// literal renders a constant value as a Go map key comment suffix.
func (v enumValue) comment() string {
	if len(v.Aliases) == 0 {
		return ""
	}
	return " // also " + strings.Join(v.Aliases, ", ")
}

// exportedPrefix converts a package name into the prefix used for that
// package's generated identifiers.
func exportedPrefix(pkgName string) string {
	if pkgName == "" {
		return ""
	}
	return strings.ToUpper(pkgName[:1]) + pkgName[1:]
}

func lowerFirst(s string) string {
	if s == "" {
		return ""
	}
	return strings.ToLower(s[:1]) + s[1:]
}

// describe summarises an enum for the command's dry-run output.
func (e enumType) describe() string {
	aliases := 0
	for _, v := range e.Values {
		aliases += len(v.Aliases)
	}
	return fmt.Sprintf("%-40s %4d values, %d aliased", e.Ref(), len(e.Values), aliases)
}
