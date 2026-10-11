package compiler

import "github.com/natefinch/council4/cardgen/oracle/parser"

func projectPlayerActorReferenceOrders(effects []CompiledEffect, references []CompiledReference) {
	for _, effect := range effects {
		if !hasPlayerActorContext(effect) {
			continue
		}
		for j := range effect.References {
			owned := &effect.References[j]
			if !isPlayerActorReference(*owned) || owned.Order.End > owned.Order.Start {
				continue
			}
			if canonical, ok := referenceForNodeID(references, owned.NodeID); ok &&
				canonical.Order.End > canonical.Order.Start {
				owned.Order = canonical.Order
			}
		}
	}
}

// A nearer explicit player actor cannot be the object/card result of an older
// clause. Object pronouns and unowned player-looking references keep their
// ordinary result/target precedence.
func playerActorTargetAntecedent(
	reference CompiledReference,
	targets []CompiledTarget,
	effects []CompiledEffect,
	prior int,
) (int, bool) {
	if reference.Order.End <= reference.Order.Start {
		return 0, false
	}
	if !isPlayerActorReference(reference) {
		return 0, false
	}
	occurrence, ok, ambiguous := targetAntecedent(reference, targets)
	if !ok || ambiguous || !targets[occurrence].Exact ||
		targets[occurrence].Order.Start <= effects[prior].VerbOrder.Start ||
		(targets[occurrence].Selector.Kind != SelectorPlayer && targets[occurrence].Selector.Kind != SelectorOpponent) {
		return 0, false
	}
	for _, effect := range effects {
		if !hasPlayerActorContext(effect) {
			continue
		}
		for _, owned := range effect.References {
			if owned.NodeID == reference.NodeID && owned.Order == reference.Order {
				return occurrence, true
			}
		}
	}
	return 0, false
}

func hasPlayerActorContext(effect CompiledEffect) bool {
	return effect.ClauseID > 0 &&
		(effect.Context == parser.EffectContextTarget ||
			effect.Context == parser.EffectContextEventPlayer ||
			effect.Context == parser.EffectContextReferencedPlayer)
}

func isPlayerActorReference(reference CompiledReference) bool {
	return reference.Kind == ReferenceThatPlayer ||
		reference.Kind == ReferencePronoun &&
			(reference.Pronoun == ReferencePronounThey || reference.Pronoun == ReferencePronounTheir)
}
