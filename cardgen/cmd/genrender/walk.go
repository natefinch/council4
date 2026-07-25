package main

import (
	"fmt"
	"go/types"
	"slices"
	"strings"
)

// rootTypes are the entry points the renderer emits values of. Everything the
// renderer can encounter is reachable from these through fields, elements, and
// interface implementations.
var rootTypes = []string{"CardDef", "CardFace"}

// reachable is the set of named types the renderer can encounter, discovered by
// walking the type graph from rootTypes.
//
// Restricting generation to this set keeps the generated tables to constants the
// renderer can actually emit, so an unused table never appears and a type added
// to a CardDef field is picked up automatically.
type reachable struct {
	// Enums are named basic types with exported constants.
	Enums []enumType
	// Structs are named struct types, ordered by qualified name.
	Structs []*types.Named
	// Interfaces are named interface types, ordered by qualified name.
	Interfaces []*types.Named
	// Implementations maps a qualified interface name to the named types
	// implementing it, ordered by qualified name.
	Implementations map[string][]*types.Named
	// Opaque are reachable struct types with unexported fields, which cannot
	// be emitted as a composite literal from outside their package.
	Opaque []*types.Named
}

// walk discovers every named type reachable from rootTypes in the game package.
func walk(pkgs map[string]*types.Package, enums []enumType) (*reachable, error) {
	game, ok := pkgs[modulePath+"/mtg/game"]
	if !ok {
		return nil, fmt.Errorf("game package %s not loaded", modulePath+"/mtg/game")
	}
	w := &walker{
		pkgs:  pkgs,
		seen:  map[*types.Named]bool{},
		enums: map[string]enumType{},
	}
	for _, e := range enums {
		w.enums[e.Ref()] = e
	}
	for _, name := range rootTypes {
		obj := game.Scope().Lookup(name)
		if obj == nil {
			return nil, fmt.Errorf("root type game.%s not found", name)
		}
		named, ok := obj.Type().(*types.Named)
		if !ok {
			return nil, fmt.Errorf("root type game.%s is not a named type", name)
		}
		w.visitNamed(named)
	}
	return w.result(), nil
}

type walker struct {
	pkgs       map[string]*types.Package
	seen       map[*types.Named]bool
	enums      map[string]enumType
	foundEnum  []enumType
	structs    []*types.Named
	interfaces []*types.Named
	impls      map[string][]*types.Named
}

func (w *walker) visitNamed(named *types.Named) {
	if w.seen[named] {
		return
	}
	w.seen[named] = true
	// Type arguments are walked before the tracked-package check because a
	// generic container declared elsewhere still carries tracked types: the
	// renderer reaches game.ManaSpendRider only through opt.V[ManaSpendRider].
	if args := named.TypeArgs(); args != nil {
		for arg := range args.Types() {
			w.visitType(arg)
		}
	}
	obj := named.Obj()
	if obj.Pkg() == nil || !w.tracked(obj.Pkg()) {
		return
	}
	switch under := named.Underlying().(type) {
	case *types.Basic:
		if e, ok := w.enums[qualified(named)]; ok {
			w.foundEnum = append(w.foundEnum, e)
		}
	case *types.Struct:
		w.structs = append(w.structs, named)
		for field := range under.Fields() {
			w.visitType(field.Type())
		}
	case *types.Interface:
		w.interfaces = append(w.interfaces, named)
		w.visitImplementations(named)
	case *types.Slice:
		w.visitType(under.Elem())
	case *types.Array:
		w.visitType(under.Elem())
	case *types.Pointer:
		w.visitType(under.Elem())
	case *types.Map:
		w.visitType(under.Key())
		w.visitType(under.Elem())
	default:
		// Signatures, channels, and type parameters carry no renderable
		// value, so nothing further is reachable through them.
	}
}

func (w *walker) visitType(t types.Type) {
	switch t := t.(type) {
	case *types.Named:
		w.visitNamed(t)
	case *types.Pointer:
		w.visitType(t.Elem())
	case *types.Slice:
		w.visitType(t.Elem())
	case *types.Array:
		w.visitType(t.Elem())
	case *types.Map:
		w.visitType(t.Key())
		w.visitType(t.Elem())
	case *types.Struct:
		for field := range t.Fields() {
			w.visitType(field.Type())
		}
	default:
		// Unnamed interfaces cannot be dispatched on by name, and the
		// remaining types carry no renderable value. Every game interface is
		// named, so the renderer never encounters an unnamed one.
	}
}

// visitImplementations records every named type in the tracked packages that
// implements iface, and walks into each so their fields are reachable too.
func (w *walker) visitImplementations(iface *types.Named) {
	underlying, ok := iface.Underlying().(*types.Interface)
	if !ok {
		return
	}
	if w.impls == nil {
		w.impls = map[string][]*types.Named{}
	}
	key := qualified(iface)
	var found []*types.Named
	for _, pkg := range w.sortedPkgs() {
		scope := pkg.Scope()
		for _, name := range scope.Names() {
			tn, ok := scope.Lookup(name).(*types.TypeName)
			if !ok || !tn.Exported() {
				continue
			}
			named, ok := tn.Type().(*types.Named)
			if !ok || named == iface {
				continue
			}
			if !implements(named, underlying) {
				continue
			}
			found = append(found, named)
		}
	}
	slices.SortFunc(found, func(a, b *types.Named) int {
		return strings.Compare(qualified(a), qualified(b))
	})
	w.impls[key] = found
	for _, named := range found {
		w.visitNamed(named)
	}
}

// implements reports whether named, or a pointer to it, satisfies iface.
// Interfaces themselves are excluded: only concrete types are dispatch targets.
func implements(named *types.Named, iface *types.Interface) bool {
	if _, ok := named.Underlying().(*types.Interface); ok {
		return false
	}
	return types.Implements(named, iface) || types.Implements(types.NewPointer(named), iface)
}

func (w *walker) sortedPkgs() []*types.Package {
	paths := make([]string, 0, len(w.pkgs))
	for p := range w.pkgs {
		paths = append(paths, p)
	}
	slices.Sort(paths)
	pkgs := make([]*types.Package, 0, len(paths))
	for _, p := range paths {
		pkgs = append(pkgs, w.pkgs[p])
	}
	return pkgs
}

func (w *walker) tracked(pkg *types.Package) bool {
	_, ok := w.pkgs[pkg.Path()]
	return ok
}

func (w *walker) result() *reachable {
	byName := func(a, b *types.Named) int { return strings.Compare(qualified(a), qualified(b)) }
	slices.SortFunc(w.structs, byName)
	slices.SortFunc(w.interfaces, byName)
	slices.SortFunc(w.foundEnum, func(a, b enumType) int { return strings.Compare(a.Ref(), b.Ref()) })

	var opaque []*types.Named
	for _, named := range w.structs {
		if hasUnexportedField(named) {
			opaque = append(opaque, named)
		}
	}
	impls := w.impls
	if impls == nil {
		impls = map[string][]*types.Named{}
	}
	return &reachable{
		Enums:           w.foundEnum,
		Structs:         w.structs,
		Interfaces:      w.interfaces,
		Implementations: impls,
		Opaque:          opaque,
	}
}

// hasUnexportedField reports whether a struct type carries state the renderer
// cannot set from a composite literal written in another package.
func hasUnexportedField(named *types.Named) bool {
	st, ok := named.Underlying().(*types.Struct)
	if !ok {
		return false
	}
	for field := range st.Fields() {
		if !field.Exported() {
			return true
		}
	}
	return false
}

// qualified is a named type's package-qualified name, used for stable ordering
// and map keys.
func qualified(named *types.Named) string {
	obj := named.Obj()
	if obj.Pkg() == nil {
		return obj.Name()
	}
	return obj.Pkg().Name() + "." + obj.Name()
}

// containsNamed reports whether names contains a type with the given qualified
// name.
func containsNamed(named []*types.Named, name string) bool {
	return slices.ContainsFunc(named, func(n *types.Named) bool { return qualified(n) == name })
}

// containsEnum reports whether enums contains a type with the given qualified
// name.
func containsEnum(enums []enumType, name string) bool {
	return slices.ContainsFunc(enums, func(e enumType) bool { return e.Ref() == name })
}
