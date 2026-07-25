package main

import (
	"fmt"
	"go/types"
	"strings"
)

// dispatchName is the generated method that renders any value of an interface
// type by switching on its concrete type.
func dispatchName(named *types.Named) string {
	return "render" + exportName(qualifiedIdent(named)) + "Value"
}

// generationTargets are the reachable structs the generator emits a renderer
// for: every reachable struct except the opaque ones, which keep hand-written
// emitters, and the optional container, whose value renders through its type
// argument.
func generationTargets(graph *reachable) []*types.Named {
	var out []*types.Named
	seen := map[string]bool{}
	for _, named := range graph.Structs {
		ref := qualified(named)
		if ref == optionalType || containsNamed(graph.Opaque, ref) || seen[ref] {
			continue
		}
		seen[ref] = true
		out = append(out, named)
	}
	return out
}

// emitStructRenderer writes the renderer for one struct type. Fields are emitted
// in declaration order, and a field holding its zero value is omitted.
func (e *emitter) emitStructRenderer(b *strings.Builder, named *types.Named) error {
	// Locals are numbered per function so an unrelated change elsewhere does
	// not renumber, and so rediff, the whole generated file.
	e.temp = 0
	ref := qualified(named)
	e.compact = compactLiteralTypes[ref]
	defer func() { e.compact = false }()
	fields := e.class.fieldsOf(named)
	_, _ = fmt.Fprintf(b, "\n// %s renders a %s value as a Go composite literal.\n", generatedName(named), ref)
	_, _ = fmt.Fprintf(b, "func (r Renderer) %s(ctx *renderCtx, v %s) (string, error) {\n", generatedName(named), e.ref(named))
	if len(fields) == 0 {
		_, _ = fmt.Fprintf(b, "\t_ = ctx\n\t_ = v\n\treturn %q, nil\n}\n", ref+"{}")
		return nil
	}
	_, _ = fmt.Fprint(b, "\tvar fields []string\n")
	for _, f := range fields {
		expr := "v." + f.Name
		cond, err := e.emitCond(expr, ref+"."+f.Name, f.Type)
		if err != nil {
			return fmt.Errorf("%s.%s: %w", ref, f.Name, err)
		}
		_, _ = fmt.Fprintf(b, "\tif %s {\n", cond)
		name, err := e.emitValue(b, "\t\t", expr, f.Type, ref+"."+f.Name)
		if err != nil {
			return err
		}
		if e.compact {
			_, _ = fmt.Fprintf(b, "\t\tfields = append(fields, %q+%s)\n", f.Name+": ", name)
		} else {
			_, _ = fmt.Fprintf(b, "\t\tfields = append(fields, %q+%s+\",\")\n", f.Name+": ", name)
		}
		_, _ = fmt.Fprint(b, "\t}\n")
	}
	_, _ = fmt.Fprintf(b, "\treturn %s(%q, fields), nil\n}\n", literalFunc(ref), ref)
	return nil
}

// compactLiteralTypes render on a single line so gofmt keeps them inline. This
// is a presentation choice the type graph cannot express, so it is listed here.
var compactLiteralTypes = map[string]bool{
	"game.Selection": true,
}

func literalFunc(ref string) string {
	if compactLiteralTypes[ref] {
		return "compactStructLit"
	}
	return "structLit"
}

// emitZeroFunc writes the zero predicate for a struct that cannot be compared
// with ==, which is any struct holding a slice, map, or function field.
func (e *emitter) emitZeroFunc(b *strings.Builder, named *types.Named) error {
	ref := qualified(named)
	_, _ = fmt.Fprintf(b, "\n// %s reports whether every field of a %s holds its zero value.\n",
		zeroFuncName(named), ref)
	_, _ = fmt.Fprintf(b, "func %s(v %s) bool {\n", zeroFuncName(named), e.ref(named))
	fields := e.class.fieldsOf(named)
	if len(fields) == 0 {
		_, _ = fmt.Fprint(b, "\t_ = v\n\treturn true\n}\n")
		return nil
	}
	var conds []string
	for _, f := range fields {
		cond, err := e.nonZero("v."+f.Name, f.Type)
		if err != nil {
			return fmt.Errorf("%s.%s: %w", ref, f.Name, err)
		}
		conds = append(conds, "!("+cond+")")
	}
	_, _ = fmt.Fprintf(b, "\treturn %s\n}\n", strings.Join(conds, " &&\n\t\t"))
	return nil
}

// nonComparableTargets are the generation targets that need a generated zero
// predicate, because Go cannot compare them with ==.
func nonComparableTargets(graph *reachable, targets []*types.Named) []*types.Named {
	need := map[string]bool{}
	var out []*types.Named
	add := func(named *types.Named) {
		ref := qualified(named)
		// Opaque types never get a generated predicate: their state is
		// unexported, so a predicate built from exported fields would report
		// every value zero and drop it. They use their Empty test instead.
		if need[ref] || types.Comparable(named) || containsNamed(graph.Opaque, ref) {
			return
		}
		need[ref] = true
		out = append(out, named)
	}
	// A predicate is needed for any struct that appears as a field, so walk
	// every target's fields rather than the targets themselves.
	c := newClassifier(graph)
	for _, named := range targets {
		for _, f := range c.fieldsOf(named) {
			for t := f.Type; t != nil; t = t.Elem {
				if (t.Kind == kindStruct || t.Kind == kindOpaque || t.Kind == kindArray) && t.Named != nil {
					add(t.Named)
				}
			}
		}
	}
	return out
}

// emitDispatch writes the type switch that renders any value of an interface
// type, one case per recorded implementation.
//
// An implementation is matched in whichever form actually satisfies the
// interface: value receivers give a value case, pointer receivers a pointer
// case whose payload is dereferenced before rendering.
func (e *emitter) emitDispatch(b *strings.Builder, iface *types.Named, impls []*types.Named) error {
	ref := qualified(iface)
	underlying, ok := iface.Underlying().(*types.Interface)
	if !ok {
		return fmt.Errorf("%s is not an interface", ref)
	}
	_, _ = fmt.Fprintf(b, "\n// %s renders any %s implementation by switching on its concrete type.\n",
		dispatchName(iface), ref)
	_, _ = fmt.Fprint(b, "//\n// Every implementation in mtg/game has a case, so adding one and forgetting to\n")
	_, _ = fmt.Fprint(b, "// teach the renderer about it is impossible: the case appears when this file\n")
	_, _ = fmt.Fprint(b, "// is regenerated, and the generator fails if the value cannot be emitted.\n")
	_, _ = fmt.Fprintf(b, "func (r Renderer) %s(ctx *renderCtx, v %s) (string, error) {\n", dispatchName(iface), e.ref(iface))
	_, _ = fmt.Fprintf(b, "\tif v == nil {\n\t\treturn \"\", errors.New(%q)\n\t}\n", "render: nil "+ref)
	_, _ = fmt.Fprint(b, "\tswitch value := v.(type) {\n")
	for _, impl := range impls {
		switch {
		case types.Implements(impl, underlying):
			_, _ = fmt.Fprintf(b, "\tcase %s:\n", e.ref(impl))
			_, _ = fmt.Fprintf(b, "\t\treturn r.%s(ctx, value)\n", generatedName(impl))
		case types.Implements(types.NewPointer(impl), underlying):
			_, _ = fmt.Fprintf(b, "\tcase *%s:\n", e.ref(impl))
			_, _ = fmt.Fprint(b, "\t\tif value == nil {\n")
			_, _ = fmt.Fprintf(b, "\t\t\treturn \"\", errors.New(%q)\n\t\t}\n", "render: nil *"+qualified(impl))
			_, _ = fmt.Fprintf(b, "\t\tlit, err := r.%s(ctx, *value)\n", generatedName(impl))
			_, _ = fmt.Fprint(b, "\t\tif err != nil {\n\t\t\treturn \"\", err\n\t\t}\n")
			_, _ = fmt.Fprint(b, "\t\treturn \"&\" + lit, nil\n")
		default:
			return fmt.Errorf("%s does not implement %s in value or pointer form", qualified(impl), ref)
		}
	}
	_, _ = fmt.Fprint(b, "\tdefault:\n")
	_, _ = fmt.Fprintf(b, "\t\treturn \"\", fmt.Errorf(%q, v)\n", "render: unhandled "+ref+": %T")
	_, _ = fmt.Fprint(b, "\t}\n}\n")
	return nil
}
