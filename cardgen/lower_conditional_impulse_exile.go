package cardgen

import (
	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/opt"
)

// lowerConditionalImpulseExileSequence lowers two complementary atomic impulse
// branches: a condition-gated free play and an otherwise normal play. Both
// branches include the top-card exile, so exactly one instruction performs the
// move and grants the matching permission.
func lowerConditionalImpulseExileSequence(ctx contentCtx) (game.AbilityContent, bool) {
	if len(ctx.content.Effects) != 2 ||
		len(ctx.content.Conditions) != 1 ||
		len(ctx.content.Targets) != 0 ||
		(len(ctx.content.References) != 3 && len(ctx.content.References) != 4) ||
		len(ctx.content.Keywords) != 0 ||
		len(ctx.content.Modes) != 0 {
		return game.AbilityContent{}, false
	}
	free := ctx.content.Effects[0]
	normal := ctx.content.Effects[1]
	if !conditionalImpulseEffect(free, true) ||
		!conditionalImpulseEffect(normal, false) ||
		normal.Connection != parser.EffectConnectionOtherwise {
		return game.AbilityContent{}, false
	}
	sourceReferences := 0
	cardReferences := 0
	possessiveReferences := 0
	for i := range ctx.content.References {
		reference := ctx.content.References[i]
		switch {
		case reference.Kind == compiler.ReferenceThisObject &&
			reference.Binding == compiler.ReferenceBindingSource:
			sourceReferences++
		case reference.Kind == compiler.ReferenceThatObject:
			cardReferences++
		case reference.Kind == compiler.ReferencePronoun &&
			reference.Pronoun == compiler.ReferencePronounIts:
			possessiveReferences++
		default:
			return game.AbilityContent{}, false
		}
	}
	if sourceReferences > 1 || cardReferences != 2 || possessiveReferences != 1 {
		return game.AbilityContent{}, false
	}
	condition, ok := lowerCondition(ctx.content.Conditions[0], conditionContextEffectGate)
	if !ok {
		return game.AbilityContent{}, false
	}
	player, ok := lowerImpulseExileLibraryOwner(free.Context)
	if !ok || normal.Context != free.Context {
		return game.AbilityContent{}, false
	}
	duration, ok := lowerImpulseExileDuration(free.Duration)
	if !ok || normal.Duration != free.Duration {
		return game.AbilityContent{}, false
	}
	freePrimitive := game.ImpulseExile{
		Player:                player,
		Amount:                game.Fixed(1),
		Duration:              duration,
		WithoutPayingManaCost: true,
	}
	normalPrimitive := freePrimitive
	normalPrimitive.WithoutPayingManaCost = false
	normalCondition := condition
	normalCondition.Negate = !normalCondition.Negate
	return game.Mode{Sequence: []game.Instruction{
		{
			Primitive: freePrimitive,
			Condition: opt.Val(game.EffectCondition{Condition: opt.Val(condition)}),
		},
		{
			Primitive: normalPrimitive,
			Condition: opt.Val(game.EffectCondition{Condition: opt.Val(normalCondition)}),
		},
	}}.Ability(), true
}

func conditionalImpulseEffect(effect compiler.CompiledEffect, free bool) bool {
	return effect.Kind == compiler.EffectImpulseExile &&
		effect.Exact &&
		!effect.Negated &&
		!effect.Optional &&
		effect.Context == parser.EffectContextController &&
		effect.Amount.Known &&
		effect.Amount.Value == 1 &&
		effect.Duration == compiler.DurationThisTurn &&
		effect.ImpulseWithoutPayingManaCost == free &&
		!effect.ImpulseCast &&
		!effect.ImpulseSpendAnyColor
}
