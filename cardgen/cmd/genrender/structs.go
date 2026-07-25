package main

import (
	"fmt"
	"go/types"
	"strings"
)

// opaqueRenderer describes how to call the hand-written renderer for a type the
// generator cannot emit itself, because the type carries unexported state that a
// composite literal in another package cannot set.
//
// These are the only value emitters that stay hand-written. Every opaque type
// must have an entry; checkOpaqueRenderers fails generation otherwise, so a new
// opaque type in mtg/game cannot silently lose its rendering.
type opaqueRenderer struct {
	// Func is the Go function name.
	Func string
	// Method reports whether Func is called on the Renderer receiver.
	Method bool
	// Ctx reports whether Func takes the render context as its first
	// argument.
	Ctx bool
	// Empty is a Go expression template, with one %s for the value, that
	// reports whether the value is the omitted zero value.
	//
	// It is required for types Go cannot compare with ==, because their state
	// is unexported: the generator cannot synthesize a correct test, and
	// assuming such a value is always zero silently drops it from every card.
	Empty string
}

// emptyTest returns the condition that holds when expr carries a value worth
// emitting.
func (o opaqueRenderer) emptyTest(expr, typeName string) string {
	if o.Empty != "" {
		return "!(" + fmt.Sprintf(o.Empty, expr) + ")"
	}
	return expr + " != (" + typeName + "{})"
}

var opaqueRenderers = map[string]opaqueRenderer{
	"color.Identity":         {Func: "renderColorIdentity", Method: true, Ctx: true, Empty: "len(%s.Colors()) == 0"},
	"game.BattlefieldSource": {Func: "renderBattlefieldSource"},
	"game.DamageRecipient":   {Func: "renderDamageRecipient", Method: true, Ctx: true},
	"game.GroupReference":    {Func: "renderGroupReference", Method: true, Ctx: true, Empty: "%s.Empty()"},
	"game.ObjectReference":   {Func: "renderObjectReference", Method: true},
	"game.PlayerReference":   {Func: "renderPlayerReference", Method: true},
	// PlayerGroupReference is not opaque, but rendering it as a composite
	// literal loses the OpponentsReference()/AllPlayersReference() constructors
	// that make the generated cards readable.
	"game.PlayerGroupReference": {Func: "renderPlayerGroupReferenceWithContext", Method: true, Ctx: true},
	// Symbol is likewise renderable as a literal, but mana costs are far
	// easier to read as cost.O(2) and cost.U than as a four-field struct
	// repeated for every symbol in every mana cost in the corpus.
	"cost.Symbol":      {Func: "renderManaSymbol", Ctx: true},
	"game.Quantity":    {Func: "renderQuantity", Method: true, Ctx: true},
	"game.TokenSource": {Func: "renderTokenSource", Method: true, Ctx: true, Empty: "!%s.Valid()"},
}

// call returns the Go expression rendering value expr through this renderer.
func (o opaqueRenderer) call(expr string) string {
	var b strings.Builder
	if o.Method {
		_, _ = b.WriteString("r.")
	}
	_, _ = b.WriteString(o.Func)
	_, _ = b.WriteString("(")
	if o.Ctx {
		_, _ = b.WriteString("ctx, ")
	}
	_, _ = b.WriteString(expr)
	_, _ = b.WriteString(")")
	return b.String()
}

// checkOpaqueRenderers fails when a reachable opaque type has no hand-written
// emitter, or when a type Go cannot compare has no emptiness test. Either gap
// would make the generated renderer drop values silently rather than fail.
func checkOpaqueRenderers(graph *reachable) error {
	var problems []string
	for _, named := range graph.Opaque {
		ref := qualified(named)
		o, ok := opaqueRenderers[ref]
		if !ok {
			problems = append(problems, ref+" has no hand-written renderer")
			continue
		}
		if !types.Comparable(named) && o.Empty == "" {
			problems = append(problems, ref+" is not comparable and has no Empty test")
		}
	}
	if len(problems) > 0 {
		return fmt.Errorf("opaqueRenderers is incomplete: %s", strings.Join(problems, "; "))
	}
	return nil
}

// importPaths maps a package name to the renderer's import path constant, so
// generated code can register the imports the literals it emits require.
var importPaths = map[string]string{
	"game":    "importGame",
	"color":   "importColor",
	"compare": "importCompare",
	"counter": "importCounter",
	"cost":    "importCost",
	"mana":    "importMana",
	"types":   "importTypes",
	"zone":    "importZone",
	"opt":     "importOpt",
}

// generatedName is the generated renderer's method name for a named type, for
// example game.Amass becomes renderGameAmass.
func generatedName(named *types.Named) string {
	obj := named.Obj()
	pkg := ""
	if obj.Pkg() != nil {
		pkg = obj.Pkg().Name()
	}
	return "render" + exportName(pkg) + obj.Name()
}

// exportName upper-cases the first rune so a package name can be spliced into a
// Go identifier.
func exportName(s string) string {
	if s == "" {
		return ""
	}
	return strings.ToUpper(s[:1]) + s[1:]
}
