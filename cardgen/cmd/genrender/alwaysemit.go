package main

import (
	"fmt"
	"go/types"
	"slices"
	"strings"
)

// alwaysEmitEnums are enum types whose zero value is a real value rather than an
// "unset" sentinel, so the generated literal states it on every field.
//
// Omitting a zero field is behavior-preserving, but it is not meaning-preserving
// when zero names something. Enums that reserve zero for a default, such as the
// ContinuousLayer set that starts at iota + 1, are deliberately absent: for
// those, omission is the clearer rendering.
var alwaysEmitEnums = map[string]bool{
	// EffectDuration(0) is DurationPermanent. How long an effect lasts is the
	// distinction between "gets +1/+1" and "gets +1/+1 until end of turn", so
	// every effect that has a duration says which one it is.
	"game.EffectDuration": true,
}

// alwaysEmitFields are individual struct fields whose zero names a real value,
// for enums where that is true only in some positions.
//
// counter.Kind is the motivating case. Kind(0) is a +1/+1 counter, so the kind
// an effect adds or removes must be written even when it is the zero value:
// AddCounter with no CounterKind would read as "no counter kind" when it means
// the most common counter in the game. The same type is also used for filters
// such as Selection.RequiredCounter, where zero really does mean "no counter
// requirement" and is gated by a sibling field; writing counter.PlusOnePlusOne
// there would assert a restriction the selection does not have.
var alwaysEmitFields = map[string]bool{
	"cost.Additional.CounterKind":                   true,
	"game.AddCounter.CounterKind":                   true,
	"game.AddPlayerCounter.CounterKind":             true,
	"game.Condition.SourceCounterKind":              true,
	"game.ConditionalCounterPlacement.Kind":         true,
	"game.CounterPlacement.Kind":                    true,
	"game.DynamicAmount.CounterKind":                true,
	"game.MoveCounters.CounterKind":                 true,
	"game.OptionalCounterForEachPlayer.CounterKind": true,
	"game.RemoveCounter.CounterKind":                true,
	"game.RevealPutOntoBattlefield.CounterKind":     true,
	"game.TriggerPattern.CounterKind":               true,
}

// checkAlwaysEmit fails when a listed type or field is missing, is not an
// integer enum, or has no constant naming zero. Without this, renaming a field
// or renumbering an enum would quietly turn the rule off and start omitting a
// value-bearing field from every card.
func checkAlwaysEmit(graph *reachable, pkgs map[string]*types.Package) error {
	var problems []string
	for ref := range alwaysEmitEnums {
		named, ok := lookupNamed(pkgs, ref)
		if !ok {
			problems = append(problems, ref+" does not exist")
			continue
		}
		problems = append(problems, enumZeroProblems(pkgs, named, ref)...)
		if !containsNamed(graph.Structs, ref) && !reachableAsField(graph, ref) {
			problems = append(problems, ref+" is not reachable from a rendered value")
		}
	}

	c := newClassifier(graph)
	for key := range alwaysEmitFields {
		structRef, fieldName, ok := cutLast(key, ".")
		if !ok {
			problems = append(problems, key+" is not a qualified field name")
			continue
		}
		named, ok := lookupNamed(pkgs, structRef)
		if !ok || !containsNamed(graph.Structs, structRef) {
			problems = append(problems, structRef+" is not a reachable struct")
			continue
		}
		idx := slices.IndexFunc(c.fieldsOf(named), func(f structField) bool { return f.Name == fieldName })
		if idx < 0 {
			problems = append(problems, key+" names no rendered field")
			continue
		}
		t := c.fieldsOf(named)[idx].Type
		if t.Kind != kindEnum {
			problems = append(problems, key+" is not an enum field")
			continue
		}
		// Resolve the field's type through the package scope rather than
		// reusing the field's own *types.Named: the constants are declared
		// against the scope's instance.
		enum, ok := lookupNamed(pkgs, t.Ref())
		if !ok {
			problems = append(problems, key+" has an unresolvable enum type "+t.Ref())
			continue
		}
		problems = append(problems, enumZeroProblems(pkgs, enum, key+" ("+t.Ref()+")")...)
	}

	if len(problems) > 0 {
		return fmt.Errorf("always-emit tables are stale: %s", strings.Join(problems, "; "))
	}
	return nil
}

// enumZeroProblems reports why named is not an integer enum with a constant
// naming zero, which is the property that makes always emitting it worthwhile.
func enumZeroProblems(pkgs map[string]*types.Package, named *types.Named, label string) []string {
	basic, ok := named.Underlying().(*types.Basic)
	if !ok || basicKind(basic) != kindInt {
		return []string{label + " is not a named integer type"}
	}
	if !hasZeroConstant(pkgs, named) {
		return []string{label + " has no constant equal to zero"}
	}
	return nil
}

// hasZeroConstant reports whether the enum declares a constant whose value is
// zero.
func hasZeroConstant(pkgs map[string]*types.Package, named *types.Named) bool {
	for _, pkg := range pkgs {
		scope := pkg.Scope()
		for _, name := range scope.Names() {
			con, ok := scope.Lookup(name).(*types.Const)
			if !ok || !types.Identical(con.Type(), named) {
				continue
			}
			if con.Val().String() == "0" {
				return true
			}
		}
	}
	return false
}

// cutLast splits key around the final instance of sep.
func cutLast(key, sep string) (before, after string, found bool) {
	i := strings.LastIndex(key, sep)
	if i < 0 {
		return "", "", false
	}
	return key[:i], key[i+len(sep):], true
}
