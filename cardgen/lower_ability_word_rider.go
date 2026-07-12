package cardgen

import (
	"slices"

	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/mtg/game"
)

// lowerAbilityWordConditionalRiderSpell lowers the conditional-rider cycle: a
// base spell paragraph followed by a rules-free ability-word paragraph whose
// whole effect is gated on a single intervening condition and added to the base
// resolution (the Adamant ability word's additive form: "Create two 1/1 white
// Human creature tokens.\nAdamant — If at least three white mana was spent to
// cast this spell, you gain 1 life for each creature you control."; Rally for
// the Throne, Unexplained Vision, Foreboding Fruit, Turn into a Pumpkin).
//
// The two paragraphs compile to two separate spell abilities. The rider gates
// its own untargeted effect on the mana-spent (or other effect-gate) condition,
// so its lowered instructions are already condition-bearing. This combiner
// lowers each paragraph through the standard executable pipeline, confirms the
// rider adds only gated, untargeted instructions, and appends them to the base
// spell's single mode so the base resolves unconditionally and the rider
// resolves only when its condition holds (CR 603.7 / intervening-if).
//
// It fails closed for every other shape: a base or rider that is modal, shares
// targets, or lowers to anything but one plain spell mode; a rider that targets,
// that carries an Instead replacement marker (handled by the instead-* fusions),
// or whose instructions are not all gated; and any rider whose ability word is
// not a recognized rules-free label.
func lowerAbilityWordConditionalRiderSpell(
	cardName string,
	compilation compiler.Compilation,
) (game.AbilityContent, bool) {
	if len(compilation.Abilities) != 2 || len(compilation.Syntax.Abilities) != 2 {
		return game.AbilityContent{}, false
	}
	base := compilation.Abilities[0]
	rider := compilation.Abilities[1]
	if !isPlainSpellShell(base) || base.AbilityWord != "" {
		return game.AbilityContent{}, false
	}
	if !isPlainSpellShell(rider) ||
		rider.AbilityWord == "" ||
		!rulesFreeAbilityWordLabel(rider.AbilityWord) {
		return game.AbilityContent{}, false
	}
	if !additiveConditionalRiderContent(rider.Content) {
		return game.AbilityContent{}, false
	}
	baseSpell, ok := loweredPlainSpellAbility(cardName, base, &compilation.Syntax.Abilities[0])
	if !ok ||
		baseSpell.IsModal() ||
		len(baseSpell.SharedTargets) != 0 ||
		len(baseSpell.Modes) != 1 {
		return game.AbilityContent{}, false
	}
	riderSpell, ok := loweredPlainSpellAbility(cardName, rider, &compilation.Syntax.Abilities[1])
	if !ok ||
		riderSpell.IsModal() ||
		len(riderSpell.SharedTargets) != 0 ||
		len(riderSpell.Modes) != 1 ||
		len(riderSpell.Modes[0].Targets) != 0 {
		return game.AbilityContent{}, false
	}
	riderSequence := riderSpell.Modes[0].Sequence
	if len(riderSequence) == 0 {
		return game.AbilityContent{}, false
	}
	for i := range riderSequence {
		if !riderSequence[i].Condition.Exists {
			return game.AbilityContent{}, false
		}
	}
	fused := baseSpell
	fused.Modes = slices.Clone(baseSpell.Modes)
	fused.Modes[0].Sequence = append(slices.Clone(baseSpell.Modes[0].Sequence), riderSequence...)
	return fused, true
}

// additiveConditionalRiderContent reports whether a rider paragraph's compiled
// content is exactly one intervening condition gating one or more additive
// effects: no modes, no targets of its own, and no Instead replacement marker on
// any effect. The Instead exclusion keeps this combiner disjoint from the
// instead-damage and instead-power/toughness fusions, which replace the base
// amount rather than adding to it.
func additiveConditionalRiderContent(content compiler.AbilityContent) bool {
	if len(content.Conditions) != 1 ||
		len(content.Effects) == 0 ||
		len(content.Modes) != 0 ||
		len(content.Targets) != 0 {
		return false
	}
	for i := range content.Effects {
		if content.Effects[i].Replacement.Kind != parser.EffectReplacementNone {
			return false
		}
	}
	return true
}

// loweredPlainSpellAbility lowers a single ability through the standard
// executable pipeline and returns its spell content only when the ability
// produced exactly one spell ability and nothing else (no cost, trigger, static,
// activated, or other ability kind). Combiners reuse it so each fused paragraph
// shares the recipient, target, and effect resolution of every other spell
// rather than reconstructing it.
func loweredPlainSpellAbility(
	cardName string,
	ability compiler.CompiledAbility,
	syntax *parser.Ability,
) (game.AbilityContent, bool) {
	lowered, diagnostic := lowerExecutableAbility(cardName, false, nil, -1, ability, syntax)
	if diagnostic != nil || !lowered.complete(ability, syntax) {
		return game.AbilityContent{}, false
	}
	if !lowered.spellAbility.Exists ||
		lowered.activatedAbility.Exists ||
		lowered.triggeredAbility.Exists ||
		lowered.manaAbility.Exists ||
		lowered.loyaltyAbility.Exists ||
		lowered.chapterAbility.Exists ||
		lowered.replacementAbility.Exists ||
		len(lowered.staticAbilities) != 0 ||
		lowered.overloadCost.Exists ||
		len(lowered.additionalCosts) != 0 ||
		len(lowered.alternativeCosts) != 0 {
		return game.AbilityContent{}, false
	}
	return lowered.spellAbility.Val, true
}
