package cardgen

import (
	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/cardgen/oracle/shared"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/opt"
)

// lowerCantBeBlockedSpell lowers the temporary combat-evasion effect "<subject>
// can't be blocked this turn." into ApplyRule instructions that place a
// RuleEffectCantBeBlocked restriction on each affected creature for the turn
// (game.DurationThisTurn, removed during cleanup). It accepts the three subject
// shapes the parser recognizes: a target noun phrase with single, plural, or
// optional cardinality ("Up to one target creature can't be blocked this
// turn."), the source itself ("This creature can't be blocked this turn."), a
// prior-subject sequence clause that inherits the source as its subject ("...
// and can't be blocked this turn."), and the compound "source and up to one
// other target creature" subject (Martha Jones), where the source and each
// chosen target each gain the restriction. Every other recipient, duration,
// condition, mode, or reference fails closed so the broader "can't be blocked
// this turn" family stays faithful and bounded.
func lowerCantBeBlockedSpell(ctx contentCtx) (game.AbilityContent, *shared.Diagnostic) {
	unsupported := func() (game.AbilityContent, *shared.Diagnostic) {
		return game.AbilityContent{}, contentDiagnostic(
			ctx,
			"unsupported can't-be-blocked effect",
			"the executable source backend supports only exact \"<subject> can't be blocked this turn.\"",
		)
	}
	effect := ctx.content.Effects[0]
	if !effect.Exact ||
		effect.Negated ||
		effect.Optional ||
		effect.Duration != compiler.DurationThisTurn ||
		ctx.optional ||
		len(ctx.content.Conditions) != 0 ||
		len(ctx.content.Keywords) != 0 ||
		len(ctx.content.Modes) != 0 {
		return unsupported()
	}
	targetSubject := effect.Context == parser.EffectContextTarget &&
		len(ctx.content.Targets) == 1 &&
		len(ctx.content.References) == 0 &&
		ctx.content.Targets[0].Selector.Kind == compiler.SelectorCreature
	sourceSubject := (effect.Context == parser.EffectContextSource ||
		effect.Context == parser.EffectContextPriorSubject) &&
		len(ctx.content.Targets) == 0 &&
		len(ctx.content.References) == 1 &&
		ctx.content.References[0].Binding == compiler.ReferenceBindingSource
	sourceAndTargetSubject := effect.Context == parser.EffectContextTarget &&
		len(ctx.content.Targets) == 1 &&
		ctx.content.Targets[0].Selector.Kind == compiler.SelectorCreature &&
		len(ctx.content.References) == 1 &&
		ctx.content.References[0].Binding == compiler.ReferenceBindingSource
	switch {
	case sourceAndTargetSubject:
		// "<source> and up to one other target creature can't be blocked this
		// turn." (Martha Jones): the source itself plus the chosen target(s) each
		// gain the restriction. The source reference resolves to the source
		// permanent; the target slots resolve to the chosen creatures, and the
		// "other" qualifier excludes the source from being chosen twice.
		object, ok := lowerObjectReference(ctx.content.References[0], referenceLoweringContext{AllowSource: true})
		if !ok {
			return unsupported()
		}
		targetSpec, ok := permanentTargetSpecWithCardinality(ctx.content.Targets[0])
		if !ok {
			return unsupported()
		}
		sequence := make([]game.Instruction, 0, targetSpec.MaxTargets+1)
		sequence = append(sequence, cantBeBlockedInstruction(object))
		for i := range targetSpec.MaxTargets {
			sequence = append(sequence, cantBeBlockedInstruction(game.TargetPermanentReference(i)))
		}
		return game.Mode{
			Targets:  []game.TargetSpec{targetSpec},
			Sequence: sequence,
		}.Ability(), nil
	case targetSubject:
		targetSpec, ok := permanentTargetSpecWithCardinality(ctx.content.Targets[0])
		if !ok {
			return unsupported()
		}
		sequence := make([]game.Instruction, 0, targetSpec.MaxTargets)
		for i := range targetSpec.MaxTargets {
			sequence = append(sequence, cantBeBlockedInstruction(game.TargetPermanentReference(i)))
		}
		return game.Mode{
			Targets:  []game.TargetSpec{targetSpec},
			Sequence: sequence,
		}.Ability(), nil
	case sourceSubject:
		object, ok := lowerObjectReference(ctx.content.References[0], referenceLoweringContext{AllowSource: true})
		if !ok {
			return unsupported()
		}
		return game.Mode{
			Sequence: []game.Instruction{cantBeBlockedInstruction(object)},
		}.Ability(), nil
	default:
		return unsupported()
	}
}

// cantBeBlockedInstruction builds the ApplyRule instruction that grants the
// given object a can't-be-blocked restriction for the turn.
func cantBeBlockedInstruction(object game.ObjectReference) game.Instruction {
	return game.Instruction{
		Primitive: game.ApplyRule{
			Object: opt.Val(object),
			RuleEffects: []game.RuleEffect{
				{Kind: game.RuleEffectCantBeBlocked},
			},
			Duration: game.DurationThisTurn,
		},
	}
}
