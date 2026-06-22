package cardgen

import (
	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/opt"
)

// lowerInsteadDamageSpellAbilities lowers the conditional damage-boost cycle: a
// base "<Spell> deals A damage to <target>." followed by a larger "<Ability
// word> — <Spell> deals B damage instead if <condition>." alternative
// (Brimstone Volley: "Brimstone Volley deals 3 damage to any target.\nMorbid —
// Brimstone Volley deals 5 damage instead if a creature died this turn.";
// Galvanic Blast, Cackling Flames, Thermal Blast, ...). The two paragraphs
// compile to two separate single-target damage spell abilities; the second
// restates no target, carries the parser's Instead replacement marker, and
// gates on a single condition. They fuse into one spell over the base target
// whose base amount resolves only when the condition fails and whose larger
// amount resolves only when it holds, so exactly one damage event happens
// (CR 614 replacement). It fails closed for every other shape — riders, dynamic
// amounts, restated targets, or a condition the effect-gate path cannot lower —
// leaving those spells rejected.
func lowerInsteadDamageSpellAbilities(cardName string, compilation compiler.Compilation) (game.AbilityContent, bool) {
	if len(compilation.Abilities) != 2 ||
		len(compilation.Syntax.Abilities) != 2 {
		return game.AbilityContent{}, false
	}
	base := compilation.Abilities[0]
	alternative := compilation.Abilities[1]
	if !isPlainSpellShell(base) || !isPlainSpellShell(alternative) {
		return game.AbilityContent{}, false
	}
	if !isSoleBaseFixedDamageContent(base.Content) {
		return game.AbilityContent{}, false
	}
	alternativeValue, ok := soleInsteadFixedDamageValue(alternative.Content)
	if !ok {
		return game.AbilityContent{}, false
	}
	condition, ok := lowerCondition(alternative.Content.Conditions[0], conditionContextEffectGate)
	if !ok {
		return game.AbilityContent{}, false
	}
	baseDamage, baseSpell, ok := loweredSingleTargetDamageInstruction(cardName, base, &compilation.Syntax.Abilities[0])
	if !ok {
		return game.AbilityContent{}, false
	}
	notCondition := condition
	notCondition.Negate = !notCondition.Negate
	baseSpell.Modes[0].Sequence[0].Condition = opt.Val(game.EffectCondition{Condition: opt.Val(notCondition)})
	alternativeDamage := baseDamage
	alternativeDamage.Amount = game.Fixed(alternativeValue)
	alternativeInstruction := game.Instruction{
		Primitive: alternativeDamage,
		Condition: opt.Val(game.EffectCondition{Condition: opt.Val(condition)}),
	}
	baseSpell.Modes[0].Sequence = append(baseSpell.Modes[0].Sequence, alternativeInstruction)
	return baseSpell, true
}

// loweredSingleTargetDamageInstruction lowers the base damage spell through the
// standard executable pipeline and confirms it produced exactly one spell mode
// whose sole instruction is an ungated Damage primitive. Reusing the standard
// path keeps the recipient and damage source resolution shared with every other
// damage spell rather than reconstructed here.
func loweredSingleTargetDamageInstruction(
	cardName string,
	base compiler.CompiledAbility,
	syntax *parser.Ability,
) (game.Damage, game.AbilityContent, bool) {
	lowered, diagnostic := lowerExecutableAbility(cardName, false, nil, base, syntax)
	if diagnostic != nil || !lowered.complete(base, syntax) {
		return game.Damage{}, game.AbilityContent{}, false
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
		return game.Damage{}, game.AbilityContent{}, false
	}
	spell := lowered.spellAbility.Val
	if spell.IsModal() ||
		len(spell.SharedTargets) != 0 ||
		len(spell.Modes) != 1 ||
		len(spell.Modes[0].Sequence) != 1 {
		return game.Damage{}, game.AbilityContent{}, false
	}
	damage, ok := spell.Modes[0].Sequence[0].Primitive.(game.Damage)
	if !ok || spell.Modes[0].Sequence[0].Condition.Exists {
		return game.Damage{}, game.AbilityContent{}, false
	}
	return damage, spell, true
}

// isSoleBaseFixedDamageContent validates that the base ability is a single exact
// fixed-amount deal-damage effect to one target with no replacement marker,
// rider, or dynamic amount — the unconditional half of the conditional-boost
// cycle.
func isSoleBaseFixedDamageContent(content compiler.AbilityContent) bool {
	if len(content.Effects) != 1 ||
		len(content.Conditions) != 0 ||
		len(content.Modes) != 0 ||
		len(content.Keywords) != 0 ||
		len(content.Targets) != 1 ||
		content.Targets[0].Cardinality.Min != 1 ||
		content.Targets[0].Cardinality.Max != 1 {
		return false
	}
	effect := content.Effects[0]
	return fixedSingleTargetDamageEffect(&effect) &&
		effect.Exact &&
		effect.Replacement.Kind == parser.EffectReplacementNone
}

// soleInsteadFixedDamageValue validates that the alternative ability is a single
// fixed-amount deal-damage effect carrying the Instead replacement marker, gated
// on exactly one condition, restating no target of its own, and returns its
// fixed damage amount.
func soleInsteadFixedDamageValue(content compiler.AbilityContent) (int, bool) {
	if len(content.Effects) != 1 ||
		len(content.Conditions) != 1 ||
		len(content.Modes) != 0 ||
		len(content.Keywords) != 0 ||
		len(content.Targets) != 0 {
		return 0, false
	}
	effect := content.Effects[0]
	if !fixedSingleTargetDamageEffect(&effect) ||
		effect.Replacement.Kind != parser.EffectReplacementInstead ||
		len(effect.Targets) != 0 {
		return 0, false
	}
	return effect.Amount.Value, true
}

// fixedSingleTargetDamageEffect reports whether an effect is an exact fixed
// deal-damage effect with a known positive amount, no negation, no optional
// wrapper, no delayed timing, and none of the multi-recipient riders. It is the
// shared shape of both the base and alternative damage of the conditional-boost
// cycle; both deal a flat amount to the spell's single chosen target.
func fixedSingleTargetDamageEffect(effect *compiler.CompiledEffect) bool {
	return effect.Kind == compiler.EffectDealDamage &&
		!effect.Negated &&
		!effect.Optional &&
		effect.DelayedTiming == 0 &&
		effect.Amount.Known &&
		effect.Amount.Value >= 1 &&
		effect.Amount.DynamicKind == compiler.DynamicAmountNone &&
		!effect.HasSelfDamageRider &&
		!effect.HasSecondTargetDamageRider &&
		effect.TargetControllerDamageRiderRecipient == parser.DamageRecipientReferenceNone &&
		effect.DamageRecipientReference == parser.DamageRecipientReferenceNone &&
		effect.EachSourceDamageRecipient == parser.DamageRecipientReferenceNone &&
		len(effect.DamageRecipientSelectors) == 0
}
