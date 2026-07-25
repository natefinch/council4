package main

import (
	"fmt"
	"go/types"
	"strconv"
	"strings"
)

// emitter writes the generated struct renderers. It tracks the enum tables so a
// field can be mapped to the table that renders it.
type emitter struct {
	class *classifier
	enums map[string]enumType
	// temp numbers the locals a nested emission needs, so recursive value
	// emission never reuses a name within one function body.
	temp int
	// used records the packages the emitted source actually names, so the
	// generated file imports exactly what it references.
	used map[string]string
	// compact is set while emitting a type listed in compactLiteralTypes.
	// Such a literal must stay on one line for gofmt to keep it inline, so
	// the containers inside it render on one line too.
	compact bool
}

// ref is a named type's Go source spelling, recording its package as used.
//
// It is called only where the spelling is emitted as Go code. A spelling that
// lands inside a generated string literal, such as the element type sliceLit
// receives, names no package in the generated file and must not be recorded, or
// the file would import packages it never references.
func (e *emitter) ref(named *types.Named) string {
	if obj := named.Obj(); obj.Pkg() != nil {
		e.used[obj.Pkg().Name()] = obj.Pkg().Path()
	}
	return qualified(named)
}

// spellCode is spell for a spelling emitted as Go code, recording every package
// the spelling names.
func (e *emitter) spellCode(t *renderType) (string, error) {
	if t.Named != nil && t.Kind != kindPointer {
		return e.ref(t.Named), nil
	}
	// Container spellings name their element's package but not their own, so
	// recurse purely to record the element's import.
	if t.Kind == kindPointer || t.Kind == kindSlice || t.Kind == kindArray {
		if _, err := e.spellCode(t.Elem); err != nil {
			return "", err
		}
	}
	return e.spell(t)
}

func newEmitter(graph *reachable) *emitter {
	e := &emitter{class: newClassifier(graph), enums: map[string]enumType{}, used: map[string]string{}}
	for _, en := range graph.Enums {
		e.enums[en.Ref()] = en
	}
	return e
}

func (e *emitter) local(prefix string) string {
	e.temp++
	return fmt.Sprintf("%s%d", prefix, e.temp)
}

// nonZero returns the Go condition that holds when expr carries a value worth
// emitting.
//
// Fields equal to their zero value are omitted from the literal. That is
// behavior-preserving by definition: a field absent from a Go composite literal
// takes exactly the zero value the omitted field held.
// emitCond returns the condition under which a struct field is written to the
// literal. It is nonZero except for enums whose zero names a real value, which
// are always written: omitting a zero field preserves behavior, but a bare
// AddCounters literal would read as "no counter kind" when it means a +1/+1
// counter. Struct-level zero tests must not use this, or a type would never
// compare equal to its own zero value.
func (e *emitter) emitCond(expr, key string, t *renderType) (string, error) {
	if t.Kind == kindEnum && (alwaysEmitEnums[t.Ref()] || alwaysEmitFields[key]) {
		return "true", nil
	}
	return e.nonZero(expr, t)
}

func (e *emitter) nonZero(expr string, t *renderType) (string, error) {
	switch t.Kind {
	case kindBool:
		return expr, nil
	case kindString:
		return expr + ` != ""`, nil
	case kindInt, kindFloat:
		return expr + " != 0", nil
	case kindEnum, kindBitmask:
		// An enum's zero test follows its underlying type: mtg/game has
		// both integer enums and string enums such as types.Sub.
		basic, ok := t.Named.Underlying().(*types.Basic)
		if !ok {
			return "", fmt.Errorf("enum %s has no basic underlying type", t.Ref())
		}
		if basicKind(basic) == kindString {
			return expr + ` != ""`, nil
		}
		return expr + " != 0", nil
	case kindInterface, kindPointer:
		return expr + " != nil", nil
	case kindOptional:
		return expr + ".Exists", nil
	case kindSlice:
		return "len(" + expr + ") > 0", nil
	case kindOpaque:
		o, ok := opaqueRenderers[t.Ref()]
		if !ok {
			return "", fmt.Errorf("no hand-written renderer for opaque type %s", t.Ref())
		}
		if o.Empty != "" {
			return o.emptyTest(expr, ""), nil
		}
		return o.emptyTest(expr, e.ref(t.Named)), nil
	case kindStruct, kindArray:
		if t.Named != nil && !types.Comparable(t.Named) {
			return "!" + zeroFuncName(t.Named) + "(" + expr + ")", nil
		}
		spelling, err := e.spellCode(t)
		if err != nil {
			return "", err
		}
		return expr + " != (" + spelling + "{})", nil
	default:
		return "", fmt.Errorf("no zero comparison for %s", t.Kind)
	}
}

// zeroFuncName is the generated predicate reporting whether a non-comparable
// struct holds only zero values.
func zeroFuncName(named *types.Named) string {
	return "isZero" + exportName(qualifiedIdent(named))
}

// qualifiedIdent is a named type's qualified name with the separator removed, so
// it can be spliced into a Go identifier.
func qualifiedIdent(named *types.Named) string {
	obj := named.Obj()
	pkg := ""
	if obj.Pkg() != nil {
		pkg = exportName(obj.Pkg().Name())
	}
	return pkg + obj.Name()
}

// emitValue writes statements that assign the Go source literal for expr to a
// newly declared string variable, and returns that variable's name.
//
// path names the field being rendered and is woven into error messages so a
// failure deep in a card points at the field that caused it.
func (e *emitter) emitValue(b *strings.Builder, indent, expr string, t *renderType, path string) (string, error) {
	name := e.local("lit")
	switch t.Kind {
	case kindBool:
		_, _ = fmt.Fprintf(b, "%s%s := strconv.FormatBool(bool(%s))\n", indent, name, expr)
	case kindString:
		_, _ = fmt.Fprintf(b, "%s%s := %s\n", indent, name, e.scalarLiteral(expr, t, "strconv.Quote(string(%s))"))
	case kindInt:
		_, _ = fmt.Fprintf(b, "%s%s := %s\n", indent, name, e.scalarLiteral(expr, t, "strconv.FormatInt(int64(%s), 10)"))
	case kindFloat:
		_, _ = fmt.Fprintf(b, "%s%s := %s\n", indent, name, e.scalarLiteral(expr, t, "strconv.FormatFloat(float64(%s), 'g', -1, 64)"))
	case kindEnum:
		en, ok := e.enums[t.Ref()]
		if !ok {
			return "", fmt.Errorf("%s: no literal table for enum %s", path, t.Ref())
		}
		fn := "enumLiteral"
		if isOpenString(t.Ref()) {
			fn = "openStringLiteral"
		}
		e.emitChecked(b, indent, name, fmt.Sprintf("%s(%s, %q, %s)", fn, en.TableName(), t.Ref(), expr), path)
	case kindBitmask:
		en, ok := e.enums[t.Ref()]
		if !ok {
			return "", fmt.Errorf("%s: no flag table for bitmask %s", path, t.Ref())
		}
		e.emitChecked(b, indent, name, fmt.Sprintf("bitmaskLiteral(%s, %q, %s)", en.FlagsName(), t.Ref(), expr), path)
	case kindStruct:
		e.emitChecked(b, indent, name, fmt.Sprintf("r.%s(ctx, %s)", generatedName(t.Named), expr), path)
	case kindOpaque:
		o, ok := opaqueRenderers[t.Ref()]
		if !ok {
			return "", fmt.Errorf("%s: no hand-written renderer for opaque type %s", path, t.Ref())
		}
		e.emitChecked(b, indent, name, o.call(expr), path)
	case kindInterface:
		e.emitChecked(b, indent, name, fmt.Sprintf("r.%s(ctx, %s)", dispatchName(t.Named), expr), path)
	case kindOptional:
		inner, err := e.emitValue(b, indent, expr+".Val", t.Elem, path)
		if err != nil {
			return "", err
		}
		_, _ = fmt.Fprintf(b, "%sctx.need(importOpt)\n", indent)
		_, _ = fmt.Fprintf(b, "%s%s := \"opt.Val(\" + %s + \")\"\n", indent, name, inner)
	case kindPointer:
		inner, err := e.emitValue(b, indent, "*"+expr, t.Elem, path)
		if err != nil {
			return "", err
		}
		switch t.Elem.Kind {
		case kindStruct:
			// A generated struct renderer always produces a composite
			// literal, which is addressable.
			_, _ = fmt.Fprintf(b, "%s%s := \"&\" + %s\n", indent, name, inner)
		case kindOpaque:
			// A hand-written renderer for an opaque type produces a
			// constructor call, whose result has no address. Bind it to a
			// local first, matching what the renderer emitted by hand.
			elemName, err := e.spellCode(t.Elem)
			if err != nil {
				return "", fmt.Errorf("%s: %w", path, err)
			}
			_, _ = fmt.Fprintf(b, "%s%s := pointerLit(%q, %s)\n", indent, name, elemName, inner)
		default:
			return "", fmt.Errorf("%s: cannot take the address of a %s value", path, t.Elem.Kind)
		}
	case kindSlice:
		elemName, err := e.elemTypeName(t.Elem)
		if err != nil {
			return "", fmt.Errorf("%s: %w", path, err)
		}
		// A named slice type spells its literal with the name, not with the
		// underlying element type: cost.Mana{...}, never []cost.Symbol{...},
		// which would not be assignable to a cost.Mana field.
		named := ""
		if t.Named != nil {
			named = e.ref(t.Named)
		}
		items := e.local("items")
		item := e.local("item")
		_, _ = fmt.Fprintf(b, "%svar %s []string\n", indent, items)
		_, _ = fmt.Fprintf(b, "%sfor _, %s := range %s {\n", indent, item, expr)
		inner, err := e.emitValue(b, indent+"\t", item, t.Elem, path+"[]")
		if err != nil {
			return "", err
		}
		if e.compact || inlineElements(t.Elem) {
			_, _ = fmt.Fprintf(b, "%s\t%s = append(%s, %s)\n", indent, items, items, inner)
			_, _ = fmt.Fprintf(b, "%s}\n", indent)
			_, _ = fmt.Fprintf(b, "%s%s := compactNamedSliceLit(%q, %q, %s)\n", indent, name, named, elemName, items)
			break
		}
		_, _ = fmt.Fprintf(b, "%s\t%s = append(%s, %s+\",\")\n", indent, items, items, inner)
		_, _ = fmt.Fprintf(b, "%s}\n", indent)
		_, _ = fmt.Fprintf(b, "%s%s := namedSliceLit(%q, %q, %s)\n", indent, name, named, elemName, items)
	case kindArray:
		elemName, err := e.spell(t)
		if err != nil {
			return "", fmt.Errorf("%s: %w", path, err)
		}
		items := e.local("items")
		idx := e.local("i")
		elemZero, err := e.nonZero(fmt.Sprintf("%s[%s]", expr, idx), t.Elem)
		if err != nil {
			return "", fmt.Errorf("%s: %w", path, err)
		}
		_, _ = fmt.Fprintf(b, "%svar %s []string\n", indent, items)
		_, _ = fmt.Fprintf(b, "%sfor %s := range %s {\n", indent, idx, expr)
		// Elements holding the zero value are omitted, and the rest carry
		// their index, so an array with a gap keeps every element in place.
		_, _ = fmt.Fprintf(b, "%s\tif !(%s) {\n%s\t\tcontinue\n%s\t}\n", indent, elemZero, indent, indent)
		inner, err := e.emitValue(b, indent+"\t", fmt.Sprintf("%s[%s]", expr, idx), t.Elem, path+"[]")
		if err != nil {
			return "", err
		}
		_, _ = fmt.Fprintf(b, "%s\t%s = append(%s, strconv.Itoa(%s)+\": \"+%s)\n", indent, items, items, idx, inner)
		_, _ = fmt.Fprintf(b, "%s}\n", indent)
		_, _ = fmt.Fprintf(b, "%s%s := %q + \"{\" + strings.Join(%s, \", \") + \"}\"\n",
			indent, name, elemName, items)
	default:
		return "", fmt.Errorf("%s: cannot emit %s", path, t.Kind)
	}
	if imp := e.importFor(t); imp != "" {
		_, _ = fmt.Fprintf(b, "%sctx.need(%s)\n", indent, imp)
	}
	return name, nil
}

// scalarLiteral wraps a scalar conversion in the declared named type when the
// field's type is not the plain builtin, so the emitted literal keeps its type.
func (*emitter) scalarLiteral(expr string, t *renderType, conv string) string {
	inner := fmt.Sprintf(conv, expr)
	if t.Named == nil {
		return inner
	}
	return strconv.Quote(qualified(t.Named)+"(") + " + " + inner + " + " + strconv.Quote(")")
}

// emitChecked writes a call returning (string, error) and the error check that
// annotates a failure with the field path.
func (e *emitter) emitChecked(b *strings.Builder, indent, name, call, path string) {
	errName := e.local("err")
	_, _ = fmt.Fprintf(b, "%s%s, %s := %s\n", indent, name, errName, call)
	_, _ = fmt.Fprintf(b, "%sif %s != nil {\n", indent, errName)
	_, _ = fmt.Fprintf(b, "%s\treturn \"\", fmt.Errorf(%q, %s)\n", indent, path+": %w", errName)
	_, _ = fmt.Fprintf(b, "%s}\n", indent)
}

// elemTypeName is the Go source spelling of a container's element type, used in
// the slice and array literal prefixes.
func (e *emitter) elemTypeName(t *renderType) (string, error) { return e.spell(t) }

// spell is the Go source spelling of a type as written from the cardgen
// package. Named types spell as their qualified name; the anonymous containers
// mtg/game uses are built up from their element spelling.
func (e *emitter) spell(t *renderType) (string, error) {
	if t.Named != nil && t.Kind != kindPointer {
		return qualified(t.Named), nil
	}
	switch t.Kind {
	case kindBool:
		return "bool", nil
	case kindString:
		return "string", nil
	case kindInt:
		return "int", nil
	case kindFloat:
		return "float64", nil
	case kindPointer:
		inner, err := e.spell(t.Elem)
		if err != nil {
			return "", err
		}
		return "*" + inner, nil
	case kindSlice:
		inner, err := e.spell(t.Elem)
		if err != nil {
			return "", err
		}
		return "[]" + inner, nil
	case kindArray:
		inner, err := e.spell(t.Elem)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("[%d]%s", t.Len, inner), nil
	default:
		return "", fmt.Errorf("unnamed %s has no type spelling", t.Kind)
	}
}

// importFor is the renderer import constant a value of this type requires, or
// the empty string when the value's spelling names no package.
func (*emitter) importFor(t *renderType) string {
	if t.Named == nil {
		return ""
	}
	// The optional container's own package is registered by the emitted
	// opt.Val wrapper, and containers register their element's package
	// through the recursive emission of that element.
	if t.Kind == kindOptional || t.Kind == kindPointer || t.Kind == kindSlice || t.Kind == kindArray {
		return ""
	}
	obj := t.Named.Obj()
	if obj.Pkg() == nil {
		return ""
	}
	return importPaths[obj.Pkg().Name()]
}

// inlineElements reports whether a slice of this element type is written on one
// line, as []types.Card{types.Creature, types.Artifact}.
//
// Scalars and enums stay inline because one element per line buries a short list
// in punctuation. Structs, interfaces and constructor-rendered values are broken
// across lines, matching how the generated cards were written by hand.
func inlineElements(elem *renderType) bool {
	if elem == nil {
		return false
	}
	switch elem.Kind {
	case kindBool, kindInt, kindFloat, kindString, kindEnum, kindBitmask:
		return true
	default:
		return false
	}
}
