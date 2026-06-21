package cardgen

import (
	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/opt"
)

// lowerInsteadModifyPTSpellAbilities lowers the conditional power/toughness
// replacement cycle: a base "Target creature gets -X/-X until end of turn."
// followed by a larger "That creature gets -Y/-Y until end of turn instead if
// <condition>." alternative (Tragic Slip: "Target creature gets -1/-1 until end
// of turn.\nMorbid — That creature gets -13/-13 until end of turn instead if a
// creature died this turn."). The two paragraphs compile to two separate spell
// abilities; the second references the first's target ("that creature"), carries
// the parser's Instead replacement marker, and gates on a single condition. They
// fuse into one spell over the base target whose base modification resolves only
// when the condition fails and whose larger modification resolves only when it
// holds, so exactly one modification applies (CR 614 replacement).
func lowerInsteadModifyPTSpellAbilities(cardName string, compilation compiler.Compilation) (game.AbilityContent, bool) {
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
	baseEffect, ok := soleTargetModifyPTEffect(base.Content)
	if !ok {
		return game.AbilityContent{}, false
	}
	alternativeEffect, ok := soleInsteadReferencedModifyPTEffect(alternative.Content)
	if !ok {
		return game.AbilityContent{}, false
	}
	targetSpec, ok := permanentTargetSpec(base.Content.Targets[0])
	if !ok {
		return game.AbilityContent{}, false
	}
	condition, ok := lowerCondition(alternative.Content.Conditions[0], conditionContextEffectGate)
	if !ok {
		return game.AbilityContent{}, false
	}
	notCondition := condition
	notCondition.Negate = !notCondition.Negate
	baseInstruction := gatedModifyPTTarget(baseEffect, &notCondition)
	alternativeInstruction := gatedModifyPTTarget(alternativeEffect, &condition)
	return game.Mode{
		Targets:  []game.TargetSpec{targetSpec},
		Sequence: []game.Instruction{baseInstruction, alternativeInstruction},
	}.Ability(), true
}

// soleTargetModifyPTEffect validates that an ability's content is a single exact
// fixed power/toughness change to one target creature until end of turn (the
// base modification) and returns it.
func soleTargetModifyPTEffect(content compiler.AbilityContent) (compiler.CompiledEffect, bool) {
	if len(content.Effects) != 1 ||
		len(content.Conditions) != 0 ||
		len(content.Modes) != 0 ||
		len(content.Keywords) != 0 ||
		len(content.References) != 0 ||
		len(content.Targets) != 1 ||
		content.Targets[0].Cardinality.Min != 1 ||
		content.Targets[0].Cardinality.Max != 1 ||
		content.Targets[0].Selector.Kind != compiler.SelectorCreature {
		return compiler.CompiledEffect{}, false
	}
	effect := content.Effects[0]
	if !fixedModifyPTEffect(&effect) ||
		effect.Context != parser.EffectContextTarget ||
		!effect.Exact ||
		len(effect.Targets) != 1 ||
		effect.Replacement.Kind != parser.EffectReplacementNone {
		return compiler.CompiledEffect{}, false
	}
	return effect, true
}

// soleInsteadReferencedModifyPTEffect validates that an ability's content is a
// single fixed power/toughness change to a referenced creature ("that creature")
// until end of turn, carrying the Instead replacement marker and gated on
// exactly one condition (the alternative modification), and returns it.
func soleInsteadReferencedModifyPTEffect(content compiler.AbilityContent) (compiler.CompiledEffect, bool) {
	if len(content.Effects) != 1 ||
		len(content.Conditions) != 1 ||
		len(content.Modes) != 0 ||
		len(content.Keywords) != 0 ||
		len(content.Targets) != 0 ||
		len(content.References) != 1 ||
		content.References[0].Kind != compiler.ReferenceThatObject {
		return compiler.CompiledEffect{}, false
	}
	effect := content.Effects[0]
	if !fixedModifyPTEffect(&effect) ||
		effect.Context != parser.EffectContextReferencedObject ||
		effect.Replacement.Kind != parser.EffectReplacementInstead ||
		len(effect.References) != 1 {
		return compiler.CompiledEffect{}, false
	}
	return effect, true
}

// fixedModifyPTEffect reports whether an effect is an exact fixed-amount
// power/toughness change until end of turn (no dynamic amount, no negation, no
// optional wrapper, no delayed timing), the shared shape of both the base and
// alternative modifications.
func fixedModifyPTEffect(effect *compiler.CompiledEffect) bool {
	return effect.Kind == compiler.EffectModifyPT &&
		!effect.Negated &&
		!effect.Optional &&
		effect.DelayedTiming == 0 &&
		effect.Duration == compiler.DurationUntilEndOfTurn &&
		effect.Amount.DynamicKind == compiler.DynamicAmountNone &&
		effect.PowerDelta.Known &&
		effect.ToughnessDelta.Known
}

// gatedModifyPTTarget builds a ModifyPT instruction over the spell's single
// target creature, gated on the supplied condition so the modification resolves
// only when the condition holds.
func gatedModifyPTTarget(effect compiler.CompiledEffect, condition *game.Condition) game.Instruction {
	return game.Instruction{
		Primitive: game.ModifyPT{
			Object:         game.TargetPermanentReference(0),
			PowerDelta:     game.Fixed(compiledSignedAmountValue(effect.PowerDelta)),
			ToughnessDelta: game.Fixed(compiledSignedAmountValue(effect.ToughnessDelta)),
			Duration:       game.DurationUntilEndOfTurn,
		},
		Condition: opt.Val(game.EffectCondition{Condition: opt.Val(*condition)}),
	}
}
