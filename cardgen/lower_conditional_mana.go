package cardgen

import (
	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/opt"
)

// lowerThresholdInsteadManaSpellAbilities lowers the Threshold conditional-mana
// cycle: a base mana production followed by a larger "Add <more> instead if
// <condition>" alternative (Cabal Ritual: "Add {B}{B}{B}.\nThreshold — Add
// {B}{B}{B}{B}{B} instead if there are seven or more cards in your
// graveyard."). The two paragraphs compile to two separate add-mana spell
// abilities; the second carries the parser's Instead flag and a single
// effect-gate condition. They fuse into one spell whose base output resolves
// only when the condition fails and whose larger output resolves only when it
// holds, so exactly one production happens.
func lowerThresholdInsteadManaSpellAbilities(cardName string, compilation compiler.Compilation) (game.AbilityContent, bool) {
	_ = cardName
	if len(compilation.Abilities) != 2 ||
		len(compilation.Syntax.Abilities) != 2 {
		return game.AbilityContent{}, false
	}
	base := compilation.Abilities[0]
	alternative := compilation.Abilities[1]
	if !isPlainSpellShell(base) || !isPlainSpellShell(alternative) {
		return game.AbilityContent{}, false
	}
	baseEffect, ok := soleFixedAddManaEffect(base.Content, false)
	if !ok {
		return game.AbilityContent{}, false
	}
	alternativeEffect, ok := soleFixedAddManaEffect(alternative.Content, true)
	if !ok {
		return game.AbilityContent{}, false
	}
	condition, ok := lowerCondition(alternative.Content.Conditions[0], conditionContextEffectGate)
	if !ok {
		return game.AbilityContent{}, false
	}
	notCondition := condition
	notCondition.Negate = !notCondition.Negate
	baseSeq := gatedFixedAddMana(baseEffect.Mana.Colors, &notCondition)
	alternativeSeq := gatedFixedAddMana(alternativeEffect.Mana.Colors, &condition)
	return game.Mode{Sequence: append(baseSeq, alternativeSeq...)}.Ability(), true
}

// isPlainSpellShell reports whether an ability is an unconditional spell with no
// trigger, cost, static rider, or optional wrapper (an ability word such as
// "Threshold" is allowed and carries no semantics).
func isPlainSpellShell(ability compiler.CompiledAbility) bool {
	return ability.Kind == compiler.AbilitySpell &&
		ability.Trigger == nil &&
		ability.Cost == nil &&
		ability.Static == nil &&
		ability.AlternativeCost == nil &&
		!ability.Optional
}

// soleFixedAddManaEffect validates that an ability's content is a single
// fixed-color add-mana effect to the controller and returns it. wantInstead
// selects between the base production (no Instead flag, no conditions) and the
// conditional alternative (Instead flag set, exactly one condition).
func soleFixedAddManaEffect(content compiler.AbilityContent, wantInstead bool) (compiler.CompiledEffect, bool) {
	wantConditions := 0
	if wantInstead {
		wantConditions = 1
	}
	if len(content.Effects) != 1 ||
		len(content.Conditions) != wantConditions ||
		len(content.Targets) != 0 ||
		len(content.Keywords) != 0 ||
		len(content.Modes) != 0 ||
		len(content.References) != 0 {
		return compiler.CompiledEffect{}, false
	}
	effect := content.Effects[0]
	if !fixedAddManaToController(effect, wantInstead) {
		return compiler.CompiledEffect{}, false
	}
	return effect, true
}

// fixedAddManaToController reports whether an effect adds a fixed, known-color
// amount of mana to its controller with no targets or references. wantInstead
// selects the parser's Instead flag (the larger conditional alternative) versus
// the base production.
func fixedAddManaToController(effect compiler.CompiledEffect, wantInstead bool) bool {
	if effect.Kind != compiler.EffectAddMana ||
		effect.Negated ||
		effect.Optional ||
		effect.DelayedTiming != 0 ||
		effect.Duration != compiler.DurationNone ||
		effect.Context != parser.EffectContextController ||
		effect.Payment.Form != parser.EffectPaymentFormUnknown ||
		len(effect.Targets) != 0 ||
		len(effect.References) != 0 {
		return false
	}
	manaEffect := effect.Mana
	return manaEffect.Instead == wantInstead &&
		manaEffect.ColorsKnown &&
		!manaEffect.Choice &&
		!manaEffect.AnyColor &&
		len(manaEffect.Colors) != 0
}

// tronConditionalManaContent detects a single activated mana ability that adds a
// base fixed amount of mana and, gated on an effect-condition, a different fixed
// amount "instead" (the Urza tron lands: "{T}: Add {C}. If you control an
// Urza's Power-Plant and an Urza's Tower, add {C}{C} instead."). It returns a
// stripped ability that produces only the base mana — for ordinary activation
// shell lowering — and the gated content that resolves exactly one production
// depending on whether the condition holds.
func tronConditionalManaContent(ability compiler.CompiledAbility) (compiler.CompiledAbility, game.AbilityContent, bool) {
	content := ability.Content
	if len(content.Effects) != 2 ||
		len(content.Conditions) != 1 ||
		len(content.Targets) != 0 ||
		len(content.Keywords) != 0 ||
		len(content.Modes) != 0 ||
		len(content.References) != 0 {
		return compiler.CompiledAbility{}, game.AbilityContent{}, false
	}
	baseEffect := content.Effects[0]
	altEffect := content.Effects[1]
	if !fixedAddManaToController(baseEffect, false) || !fixedAddManaToController(altEffect, true) {
		return compiler.CompiledAbility{}, game.AbilityContent{}, false
	}
	condition, ok := lowerCondition(content.Conditions[0], conditionContextEffectGate)
	if !ok {
		return compiler.CompiledAbility{}, game.AbilityContent{}, false
	}
	notCondition := condition
	notCondition.Negate = !notCondition.Negate
	baseSeq := gatedFixedAddMana(baseEffect.Mana.Colors, &notCondition)
	altSeq := gatedFixedAddMana(altEffect.Mana.Colors, &condition)
	gated := game.Mode{Sequence: append(baseSeq, altSeq...)}.Ability()

	stripped := ability
	stripped.Content = content
	strippedBase := baseEffect
	strippedBase.RequiresOrderedLowering = false
	stripped.Content.Effects = []compiler.CompiledEffect{strippedBase}
	stripped.Content.Conditions = nil
	return stripped, gated, true
}

// gatedFixedAddMana builds one fixed-color AddMana instruction per color, each
// gated on the supplied condition so the production resolves only when the
// condition holds.
func gatedFixedAddMana(colors []mana.Color, condition *game.Condition) []game.Instruction {
	gate := opt.Val(game.EffectCondition{Condition: opt.Val(*condition)})
	seq := make([]game.Instruction, 0, len(colors))
	for _, c := range colors {
		seq = append(seq, game.Instruction{
			Primitive: game.AddMana{Amount: game.Fixed(1), ManaColor: c},
			Condition: gate,
		})
	}
	return seq
}
