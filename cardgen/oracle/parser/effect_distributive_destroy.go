package parser

import (
	"slices"
	"strings"

	"github.com/natefinch/council4/cardgen/oracle/shared"
)

// exactCreateTokenForEachDestroyedThisWayEffectSyntax recognizes the per-
// controller payoff clause "For each creature destroyed this way, its controller
// creates a <token>." (The Curse of Fenric, chapter I). The "destroyed this way"
// set is the creatures a sibling distributive destroy removed; each one's
// controller creates a single token, so the count is one per destroyed creature
// rather than a multiplier on the create. The "its controller" subject is the
// referenced-object-controller form already modeled by the create token
// machinery; only the "For each creature destroyed this way," distribution
// prefix is new.
//
// effectSubjectStart drops that prefix from the reconstructed clause text, so the
// recognizer confirms the prefix on the full effect text and rebuilds the
// remainder from the referenced-controller subject and the canonical token spec.
// Any other create shape leaves the clause non-exact so lowering fails closed.
func exactCreateTokenForEachDestroyedThisWayEffectSyntax(effect *EffectSyntax) bool {
	if effect.Kind != EffectCreate || effect.Negated || effect.Optional {
		return false
	}
	if effect.Context != EffectContextReferencedObjectController {
		return false
	}
	if len(effect.Targets) != 0 || !effect.TokenPTKnown {
		return false
	}
	if effect.Amount.DynamicForm != EffectDynamicAmountFormNone ||
		!effect.Amount.Known || effect.Amount.Value != 1 {
		return false
	}
	specBody, ok := creatureTokenSpecBody(effect)
	if !ok {
		return false
	}
	if !strings.HasPrefix(strings.ToLower(strings.TrimSpace(fullEffectClauseText(effect))),
		"for each creature destroyed this way, ") {
		return false
	}
	subject := referencedControllerSubject(effect)
	if subject == "" {
		return false
	}
	if !strings.EqualFold(exactEffectClauseText(effect), subject+" creates "+specBody("a", "token")+".") {
		return false
	}
	effect.CreateTokenForEachDestroyedThisWay = true
	return true
}

// exactCreateTokenForEachExiledThisWayEffectSyntax recognizes the per-controller
// payoff clause "For each creature exiled this way, its controller creates a
// <token>." (Curse of the Swine). It mirrors the destroyed-this-way recognizer
// exactly except for the "exiled" distribution prefix: the "exiled this way" set
// is the creatures a sibling variable-target exile removed, and each one's
// controller creates a single token. Any other create shape leaves the clause
// non-exact so lowering fails closed.
func exactCreateTokenForEachExiledThisWayEffectSyntax(effect *EffectSyntax) bool {
	if effect.Kind != EffectCreate || effect.Negated || effect.Optional {
		return false
	}
	if effect.Context != EffectContextReferencedObjectController {
		return false
	}
	if len(effect.Targets) != 0 || !effect.TokenPTKnown {
		return false
	}
	if effect.Amount.DynamicForm != EffectDynamicAmountFormNone ||
		!effect.Amount.Known || effect.Amount.Value != 1 {
		return false
	}
	specBody, ok := creatureTokenSpecBody(effect)
	if !ok {
		return false
	}
	if !strings.HasPrefix(strings.ToLower(strings.TrimSpace(fullEffectClauseText(effect))),
		"for each creature exiled this way, ") {
		return false
	}
	subject := referencedControllerSubject(effect)
	if subject == "" {
		return false
	}
	if !strings.EqualFold(exactEffectClauseText(effect), subject+" creates "+specBody("a", "token")+".") {
		return false
	}
	effect.CreateTokenForEachExiledThisWay = true
	return true
}

// referencedControllerSubject returns the clause subject between the last
// clause-internal boundary and the verb ("its controller"), the creating player
// of a referenced-object-controller create. It mirrors how exactEffectClauseText
// rebuilds the clause body, so the two agree on where the subject begins after a
// leading distribution prefix such as "For each creature destroyed this way,".
func referencedControllerSubject(effect *EffectSyntax) string {
	verb := slices.IndexFunc(effect.Tokens, func(token shared.Token) bool {
		return token.Span == effect.VerbSpan
	})
	if verb <= 0 {
		return ""
	}
	start := effectSubjectStart(effect.Tokens, verb, effectSelfNameSpans(effect))
	if start >= verb {
		return ""
	}
	return joinedEffectText(effect.Tokens[start:verb])
}
