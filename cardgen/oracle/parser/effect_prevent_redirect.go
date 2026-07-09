package parser

import "strings"

// foldPreventNextSourceRedirect folds the Deflecting Palm redirect rider — "If
// damage is prevented this way, <source> deals that much damage to that source's
// controller." — onto the ability's preceding one-shot
// prevent-next-from-source shield. The redirect is not an independent resolving
// instruction: it fires later, when the shield actually prevents damage, and
// deals the prevented amount back to the prevented source's controller. The
// ordinary effect vocabulary cannot express that deferred, prevention-gated
// dependency as two sibling instructions, so the pair collapses to a single
// annotated prevention effect whose PreventDamageRedirectToSourceController flag
// carries the rider downstream. This is the only place the redirect's Oracle
// wording is inspected.
//
// It runs after emitSemanticAccessors so it can read the re-derived condition
// segments, and clears the redirect sentence's re-derived condition and
// reference semantics in the same pass (mirroring stripAnimateSelfSemantics): the
// "if damage is prevented this way" condition and the "<source> deals ... to that
// source's controller" clause would otherwise resurface as an ability-level
// condition and dangling references, blocking the single-effect prevention
// lowering.
func foldPreventNextSourceRedirect(abilities []Ability) {
	for i := range abilities {
		ability := &abilities[i]
		if len(ability.Sentences) != 2 {
			continue
		}
		shield := lonePreventNextSourceShield(&ability.Sentences[0])
		if shield == nil {
			continue
		}
		if !isPreventedThisWayRedirectSentence(ability, &ability.Sentences[1]) {
			continue
		}
		shield.PreventDamageRedirectToSourceController = true
		// The shield now lowers as a single prevention effect; it must not demand
		// ordered lowering, which the two-instruction parse would otherwise force.
		shield.RequiresOrderedLowering = false
		// Extend the shield's coverage over the consumed redirect sentence so the
		// lowering coverage gate accounts for its tokens (as foldAnimateSelfStill-
		// Sentence does for its trailing confirmation sentence).
		redirect := &ability.Sentences[1]
		shield.Span = spanCover(shield.Span, redirect.Span)
		shield.ClauseSpan = spanCover(shield.ClauseSpan, redirect.Span)
		redirect.Effects = nil
		redirect.Targets = nil
		ability.SemanticReferences = nil
		ability.ConditionBoundaries = nil
		ability.EventHistoryConditions = nil
		ability.ConditionClauses = nil
		ability.ConditionSegments = nil
		ability.TriggerConditionSegments = nil
	}
}

// lonePreventNextSourceShield returns the sentence's sole one-shot
// prevent-next-from-source shield effect, or nil when the sentence is not
// exactly that recognized shield.
func lonePreventNextSourceShield(sentence *Sentence) *EffectSyntax {
	if len(sentence.Effects) != 1 {
		return nil
	}
	effect := &sentence.Effects[0]
	if effect.Kind != EffectPreventDamage ||
		!effect.PreventDamageNextFromSource ||
		!effect.Exact ||
		effect.Negated ||
		effect.Context != EffectContextController {
		return nil
	}
	return effect
}

// isPreventedThisWayRedirectSentence reports whether the sentence is exactly the
// "If damage is prevented this way, <source> deals that much damage to that
// source's controller." redirect that couples to the preceding shield: the
// ability's only condition is the "if damage is prevented this way" link, and
// the sentence's lone effect deals the prevented amount ("that much") to that
// source's controller with no target of its own.
func isPreventedThisWayRedirectSentence(ability *Ability, sentence *Sentence) bool {
	if len(ability.ConditionSegments) != 1 {
		return false
	}
	segment := &ability.ConditionSegments[0]
	if segment.Kind != ConditionIntroIf ||
		!strings.EqualFold(strings.TrimSpace(segment.Text), "if damage is prevented this way") {
		return false
	}
	if len(sentence.Effects) != 1 || len(sentence.Targets) != 0 {
		return false
	}
	effect := &sentence.Effects[0]
	if effect.Kind != EffectDealDamage ||
		effect.Negated ||
		effect.Amount.DynamicKind != EffectDynamicAmountTriggeringCounterCount {
		return false
	}
	selection := tokensWithinParserSpan(sentence.Tokens, effect.Selection.Span)
	return effectWordsAt(selection, 0, "that", "much", "damage", "to", "that", "source's", "controller") &&
		len(normalizedWords(selection)) == 7
}
