package cardgen

import (
	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/opt"
)

// lowerConditionalDamageAmountReplacementSequence lowers the single-paragraph
// conditional damage-amount replacement: "<Spell> deals A damage to <target>. If
// <condition>, it deals B damage instead." (Shivan Fire, Burst Lightning, Roil
// Eruption, Magma Burst, Frost Bite, Invasive Maneuvers, Cinder Strike,
// Lithomantic Barrage, Firebending Lesson, ...). Unlike the two-paragraph
// conditional-boost cycle handled by lowerInsteadDamageSpellAbilities, both
// halves live in one ability whose content carries two deal-damage effects, one
// chosen target, and one gating condition. The "instead" clause restates the
// damage with the pronoun "it" (the spell) and the same target ("it"/"that
// creature"), so it deals to the base target rather than a new one.
//
// They fuse into one spell over the single base target: the base amount resolves
// only when the condition fails and the larger amount resolves only when it
// holds, so exactly one damage event happens (CR 614 replacement). It fails
// closed for every other shape — riders, dynamic amounts, a restated independent
// target, a recipient that is not the base target, more than one condition, or a
// condition the effect-gate path cannot lower — leaving those spells rejected.
func lowerConditionalDamageAmountReplacementSequence(ctx contentCtx) (game.AbilityContent, bool) {
	content := ctx.content
	if len(content.Effects) != 2 ||
		len(content.Conditions) != 1 ||
		len(content.Modes) != 0 ||
		len(content.Targets) != 1 ||
		len(abilityKeywordsExcludingSelectorPredicates(content)) != 0 {
		return game.AbilityContent{}, false
	}
	base := content.Effects[0]
	alternative := content.Effects[1]
	if !fixedSingleTargetDamageEffect(&base) ||
		!base.Exact ||
		base.Replacement.Kind != parser.EffectReplacementNone ||
		base.Context != parser.EffectContextSource ||
		len(base.Targets) != 1 ||
		!exactDamageSourceSyntax(base.References) {
		return game.AbilityContent{}, false
	}
	if !fixedSingleTargetDamageEffect(&alternative) ||
		alternative.Replacement.Kind != parser.EffectReplacementInstead ||
		alternative.Context != parser.EffectContextReferencedObject ||
		len(alternative.Targets) != 0 {
		return game.AbilityContent{}, false
	}
	target, ok := damageTargetSpec(content.Targets[0])
	if !ok {
		return game.AbilityContent{}, false
	}
	// Match the single condition to the effect it gates. The "instead" clause
	// must be the gated effect; the base clause must carry no independent gate.
	effectConditions, _, ok := matchSequenceEffectConditions(content.Effects, content.Conditions)
	if !ok {
		return game.AbilityContent{}, false
	}
	if _, gated := effectConditions[0]; gated {
		return game.AbilityContent{}, false
	}
	alternativeGate, gated := effectConditions[1]
	if !gated {
		return game.AbilityContent{}, false
	}
	baseGate, ok := negatedEffectCondition(&alternativeGate)
	if !ok {
		return game.AbilityContent{}, false
	}
	recipient := game.AnyTargetDamageRecipient(0)
	damageSource := primaryDamageSource(base.References)
	baseDamage := game.Damage{
		Amount:       game.Fixed(base.Amount.Value),
		Recipient:    recipient,
		DamageSource: damageSource,
	}
	alternativeDamage := game.Damage{
		Amount:       game.Fixed(alternative.Amount.Value),
		Recipient:    recipient,
		DamageSource: damageSource,
	}
	sequence := []game.Instruction{
		{Primitive: baseDamage, Condition: opt.Val(baseGate)},
		{Primitive: alternativeDamage, Condition: opt.Val(alternativeGate)},
	}
	return game.Mode{Targets: []game.TargetSpec{target}, Sequence: sequence}.Ability(), true
}
