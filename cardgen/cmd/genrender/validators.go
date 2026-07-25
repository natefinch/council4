package main

import (
	"fmt"
	"slices"
	"strings"
)

// preRenderValidators maps a struct type to a hand-written cardgen function,
// with signature func(T) error, that the generated renderer calls before it
// builds the literal.
//
// These exist because rendering and support-gating are separate concerns that
// the hand-written renderers had fused. A hand-written renderer would validate
// a value on the way to building its literal, so deleting it in favour of the
// generated renderer would silently delete the gate too — and because the
// generated renderer reaches nested values directly rather than through the
// hand-written wrapper, keeping the gate at the outermost call site would not
// have restored it.
//
// Only *runtime-capability* gates belong here: rules the executable backend
// genuinely cannot run, so a card carrying them must be reported unsupported
// rather than compiled into a card the engine will misplay. Gates that only
// said "no one wrote a render arm for this" are deliberately not preserved —
// removing those is the point of generating the renderer, and the corpus
// support count is the check that the distinction was drawn correctly.
var preRenderValidators = map[string]string{
	"game.ContinuousEffect": "validateContinuousEffect",
	"game.CostModifier":     "validateCostModifier",
	"game.RuleEffect":       "validateRuleEffect",
	"game.TriggerCondition": "validateTriggerCondition",
	"game.TriggerPattern":   "validateTriggerPattern",
}

// constructorRenderers maps a struct type to a hand-written cardgen method,
// with signature func(*renderCtx, T) (string, bool, error), that the generated
// renderer consults before it builds the literal. Reporting false means "no
// constructor matches"; the generated literal is then used.
//
// These are pure spelling: a value that a game constructor produces reads far
// better as game.WardStaticAbility(cost.Mana{cost.G(2)}) than as the twenty-line
// literal it expands to, and every arm proves the spelling is faithful by
// reflect.DeepEqual-ing the value against the constructor's actual output.
//
// Like validators, they have to hang off the generated renderer rather than off
// a hand-written wrapper: the generated renderer for a parent struct calls the
// generated renderer for a nested field type directly, so a constructor check
// placed only at the outermost call site would be skipped for every ability
// reached through a card face, a mode, or an effect.
var constructorRenderers = map[string]string{
	"game.ActivatedAbility": "constructorActivatedAbility",
	"game.ManaAbility":      "constructorManaAbility",
	"game.StaticAbility":    "constructorStaticAbility",
	"game.TriggeredAbility": "constructorTriggeredAbility",
}

// checkPreRenderValidators reports validator entries that no longer name a
// reachable struct, so a type graph change cannot leave a support gate wired to
// nothing. A validator whose name or signature is wrong is caught by compiling
// the generated file.
func checkPreRenderValidators(graph *reachable) error {
	var problems []string
	for ref := range preRenderValidators {
		if !containsNamed(graph.Structs, ref) {
			problems = append(problems, "pre-render validator "+ref+" is not a reachable struct")
		}
	}
	for ref := range constructorRenderers {
		if !containsNamed(graph.Structs, ref) {
			problems = append(problems, "constructor renderer "+ref+" is not a reachable struct")
		}
	}
	if len(problems) == 0 {
		return nil
	}
	slices.Sort(problems)
	return fmt.Errorf("generated-renderer hook tables are stale: %s", strings.Join(problems, "; "))
}
