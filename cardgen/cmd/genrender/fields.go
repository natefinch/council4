package main

import (
	"fmt"
	"go/types"
	"io"
	"slices"
)

// optionalType is the generic container mtg/game uses for "field may be absent".
// Its type argument carries the renderable value.
const optionalType = "opt.V"

// fieldKind classifies how a value renders into Go source. Classification is
// driven entirely by the type graph, so a field added to mtg/game is either
// classified into an existing kind or reported as unsupported at generation
// time. It is never silently skipped.
type fieldKind int

const (
	// kindUnsupported is a type the generator has no emission rule for. It is
	// an error, not a fallback.
	kindUnsupported fieldKind = iota
	kindBool
	kindString
	kindInt
	kindFloat
	kindEnum
	kindBitmask
	kindStruct
	kindOpaque
	kindInterface
	kindOptional
	kindSlice
	kindArray
	kindMap
	kindPointer
)

var fieldKindNames = map[fieldKind]string{
	kindUnsupported: "unsupported",
	kindBool:        "bool",
	kindString:      "string",
	kindInt:         "int",
	kindFloat:       "float",
	kindEnum:        "enum",
	kindBitmask:     "bitmask",
	kindStruct:      "struct",
	kindOpaque:      "opaque",
	kindInterface:   "interface",
	kindOptional:    "optional",
	kindSlice:       "slice",
	kindArray:       "array",
	kindMap:         "map",
	kindPointer:     "pointer",
}

func (k fieldKind) String() string {
	if name, ok := fieldKindNames[k]; ok {
		return name
	}
	return fmt.Sprintf("fieldKind(%d)", int(k))
}

// renderType is a field type resolved into the shape the emitter needs: what
// kind of value it is, the named type it came from when that matters for the
// literal's spelling, and the element types of containers.
type renderType struct {
	Kind fieldKind
	// Named is the declared type, set for every kind whose literal is
	// spelled with a type name. A named scalar such as game.LinkedKey needs
	// it for the conversion, and a plain string does not.
	Named *types.Named
	// Elem is the contained type for optional, slice, and pointer kinds, and
	// the value type for maps.
	Elem *renderType
	// Key is the key type for maps.
	Key *renderType
	// Len is the element count for arrays, whose literal spells its length.
	Len int64
}

// Ref is the qualified name of the underlying named type, or the empty string
// for unnamed types.
func (t *renderType) Ref() string {
	if t.Named == nil {
		return ""
	}
	return qualified(t.Named)
}

// describe renders the type for the -fields report.
func (t *renderType) describe() string {
	switch t.Kind {
	case kindOptional, kindSlice, kindPointer:
		return fmt.Sprintf("%s<%s>", t.Kind, t.Elem.describe())
	case kindArray:
		return fmt.Sprintf("array%d<%s>", t.Len, t.Elem.describe())
	case kindMap:
		return fmt.Sprintf("map<%s,%s>", t.Key.describe(), t.Elem.describe())
	case kindBool, kindString, kindInt, kindFloat:
		if t.Named != nil {
			return fmt.Sprintf("%s(%s)", t.Kind, t.Ref())
		}
		return t.Kind.String()
	default:
		return fmt.Sprintf("%s(%s)", t.Kind, t.Ref())
	}
}

// classifier resolves field types against a discovered type graph.
type classifier struct {
	enums     map[string]enumType
	opaque    map[string]bool
	structs   map[string]bool
	ifaces    map[string]bool
	bitmasks  map[string]bool
	unhandled []string
}

func newClassifier(graph *reachable) *classifier {
	c := &classifier{
		enums:    map[string]enumType{},
		opaque:   map[string]bool{},
		structs:  map[string]bool{},
		ifaces:   map[string]bool{},
		bitmasks: map[string]bool{},
	}
	for _, e := range graph.Enums {
		c.enums[e.Ref()] = e
		if isBitmask(e) {
			c.bitmasks[e.Ref()] = true
		}
	}
	for _, n := range graph.Opaque {
		c.opaque[qualified(n)] = true
	}
	for _, n := range graph.Structs {
		c.structs[qualified(n)] = true
	}
	for _, n := range graph.Interfaces {
		c.ifaces[qualified(n)] = true
	}
	return c
}

// classify resolves a field type into its renderable shape. An unsupported type
// yields kindUnsupported rather than an error so a report can list every gap at
// once; callers that emit code must reject it.
func (c *classifier) classify(t types.Type) *renderType {
	switch t := t.(type) {
	case *types.Named:
		return c.classifyNamed(t)
	case *types.Basic:
		return &renderType{Kind: basicKind(t)}
	case *types.Slice:
		return &renderType{Kind: kindSlice, Elem: c.classify(t.Elem())}
	case *types.Array:
		return &renderType{Kind: kindArray, Elem: c.classify(t.Elem()), Len: t.Len()}
	case *types.Pointer:
		return &renderType{Kind: kindPointer, Elem: c.classify(t.Elem())}
	case *types.Map:
		return &renderType{Kind: kindMap, Key: c.classify(t.Key()), Elem: c.classify(t.Elem())}
	case *types.Alias:
		return c.classify(types.Unalias(t))
	default:
		return &renderType{Kind: kindUnsupported}
	}
}

func (c *classifier) classifyNamed(named *types.Named) *renderType {
	ref := qualified(named)
	// The optional container is recognized by name because its type argument,
	// not the container, is what renders.
	if ref == optionalType {
		args := named.TypeArgs()
		if args == nil || args.Len() != 1 {
			return &renderType{Kind: kindUnsupported, Named: named}
		}
		return &renderType{Kind: kindOptional, Named: named, Elem: c.classify(args.At(0))}
	}
	switch under := named.Underlying().(type) {
	case *types.Basic:
		if _, ok := c.enums[ref]; ok {
			kind := kindEnum
			if c.bitmasks[ref] {
				kind = kindBitmask
			}
			return &renderType{Kind: kind, Named: named}
		}
		return &renderType{Kind: basicKind(under), Named: named}
	case *types.Struct:
		// A hand-written entry wins even when the generator could emit the
		// struct itself: some types have constructors whose spelling is far
		// more readable in the generated cards than a composite literal.
		if _, hand := opaqueRenderers[ref]; c.opaque[ref] || hand {
			return &renderType{Kind: kindOpaque, Named: named}
		}
		if c.structs[ref] {
			return &renderType{Kind: kindStruct, Named: named}
		}
		return &renderType{Kind: kindUnsupported, Named: named}
	case *types.Interface:
		if c.ifaces[ref] {
			return &renderType{Kind: kindInterface, Named: named}
		}
		return &renderType{Kind: kindUnsupported, Named: named}
	case *types.Slice:
		return &renderType{Kind: kindSlice, Named: named, Elem: c.classify(under.Elem())}
	case *types.Array:
		return &renderType{Kind: kindArray, Named: named, Elem: c.classify(under.Elem()), Len: under.Len()}
	case *types.Map:
		return &renderType{Kind: kindMap, Named: named, Key: c.classify(under.Key()), Elem: c.classify(under.Elem())}
	case *types.Pointer:
		return &renderType{Kind: kindPointer, Named: named, Elem: c.classify(under.Elem())}
	default:
		return &renderType{Kind: kindUnsupported, Named: named}
	}
}

func basicKind(b *types.Basic) fieldKind {
	switch {
	case b.Info()&types.IsBoolean != 0:
		return kindBool
	case b.Info()&types.IsString != 0:
		return kindString
	case b.Info()&types.IsInteger != 0:
		return kindInt
	case b.Info()&types.IsFloat != 0:
		return kindFloat
	default:
		return kindUnsupported
	}
}

// structField is one exported field of a reachable struct, resolved.
type structField struct {
	Name string
	Type *renderType
}

// fieldsOf returns the exported fields of a named struct in declaration order.
// Unexported fields never appear: a struct carrying one is opaque and is not a
// generation target.
func (c *classifier) fieldsOf(named *types.Named) []structField {
	st, ok := named.Underlying().(*types.Struct)
	if !ok {
		return nil
	}
	var out []structField
	for f := range st.Fields() {
		if !f.Exported() {
			continue
		}
		out = append(out, structField{Name: f.Name(), Type: c.classify(f.Type())})
	}
	return out
}

// writeFieldReport prints every reachable struct's fields with their resolved
// kinds, and summarizes the kinds in use. It exists to size and audit the
// emission rules the generator needs; it writes no files.
func writeFieldReport(w io.Writer, graph *reachable) {
	c := newClassifier(graph)
	counts := map[string]int{}
	var unsupported []string
	for _, named := range graph.Structs {
		if containsNamed(graph.Opaque, qualified(named)) {
			continue
		}
		fields := c.fieldsOf(named)
		_, _ = fmt.Fprintf(w, "%s (%d fields)\n", qualified(named), len(fields))
		for _, f := range fields {
			_, _ = fmt.Fprintf(w, "    %-28s %s\n", f.Name, f.Type.describe())
			counts[f.Type.Kind.String()]++
			if hasUnsupported(f.Type) {
				unsupported = append(unsupported, fmt.Sprintf("%s.%s %s", qualified(named), f.Name, f.Type.describe()))
			}
		}
	}
	_, _ = fmt.Fprint(w, "\nfield kinds in use:\n")
	for _, kind := range sortedKeys(counts) {
		_, _ = fmt.Fprintf(w, "    %-12s %d\n", kind, counts[kind])
	}
	_, _ = fmt.Fprintf(w, "\nunsupported fields: %d\n", len(unsupported))
	for _, u := range unsupported {
		_, _ = fmt.Fprintf(w, "    %s\n", u)
	}
}

// hasUnsupported reports whether a type or any type it contains is unsupported.
func hasUnsupported(t *renderType) bool {
	if t == nil {
		return false
	}
	return t.Kind == kindUnsupported || hasUnsupported(t.Elem) || hasUnsupported(t.Key)
}

func sortedKeys(m map[string]int) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	slices.Sort(out)
	return out
}
