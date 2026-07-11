package cardgen

import (
	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/opt"
)

// lowerDiesThisWayCopySequence lowers the "Destroy target creature. If that
// creature dies this way, its controller creates <n> tokens that are copies of
// that creature, except their power is half that creature's power and their
// toughness is half that creature's toughness. Round up each time." family (Saw
// in Half): an exact single-target destruction followed by a linked "if that
// creature dies this way" copy-token payoff whose copies halve the destroyed
// creature's last-known power and toughness, rounding up.
//
// The generic per-effect ordered-sequence path cannot model a condition linked
// to the destroyed object's death, so this dedicated lowerer emits a Destroy
// that publishes its success and a CreateToken gated on that success. The copy
// source and the recipient's controller both resolve through the destroyed
// creature's last-known information, matching the ruling that the tokens copy
// the creature as it last existed on the battlefield. It fails closed
// (ok=false) for any shape it cannot model exactly.
func lowerDiesThisWayCopySequence(ctx contentCtx) (game.AbilityContent, bool) {
	content := ctx.content
	if ctx.optional ||
		len(content.Modes) != 0 ||
		len(content.Keywords) != 0 ||
		len(content.Effects) != 2 ||
		len(content.Targets) != 1 ||
		len(content.Conditions) != 1 {
		return game.AbilityContent{}, false
	}
	destroy := &content.Effects[0]
	create := &content.Effects[1]
	condition := &content.Conditions[0]
	if destroy.Kind != compiler.EffectDestroy ||
		!destroy.Exact ||
		destroy.Negated ||
		destroy.Optional ||
		destroy.Selector.All ||
		destroy.Context != parser.EffectContextController ||
		destroy.Order.Contains(condition.Order) {
		return game.AbilityContent{}, false
	}
	if create.Kind != compiler.EffectCreate ||
		!create.Exact ||
		!create.TokenCopyOfReferenceHalvedPT ||
		!create.TokenCopyHalvePTRoundUp ||
		create.Negated ||
		create.Optional ||
		create.DelayedTiming != 0 ||
		create.Duration != compiler.DurationNone ||
		create.Context != parser.EffectContextReferencedObjectController ||
		!create.Order.Contains(condition.Order) {
		return game.AbilityContent{}, false
	}
	if condition.Kind != compiler.ConditionIf ||
		condition.Predicate != compiler.ConditionPredicateDiesThisWay ||
		condition.Negated ||
		condition.Intervening {
		return game.AbilityContent{}, false
	}
	if content.Targets[0].Cardinality.Min != 1 ||
		content.Targets[0].Cardinality.Max != 1 ||
		!tokenCopyAuxiliaryReferencesOK(content.References) {
		return game.AbilityContent{}, false
	}
	targetSpec, ok := permanentTargetSpec(content.Targets[0])
	if !ok {
		return game.AbilityContent{}, false
	}
	spec, ok := tokenCopyModifiers(create, game.TargetPermanentReference(0))
	if !ok {
		return game.AbilityContent{}, false
	}
	spec.HalvePowerToughnessRoundUp = true
	amount, ok := createTokenAmount(ctx, create, game.ObjectReference{})
	if !ok {
		return game.AbilityContent{}, false
	}
	const resultKey = game.ResultKey("dies-this-way-copy")
	return game.Mode{
		Targets: []game.TargetSpec{targetSpec},
		Sequence: []game.Instruction{
			{
				Primitive: game.Destroy{
					Object:              game.TargetPermanentReference(0),
					PreventRegeneration: destroy.PreventRegeneration,
				},
				PublishResult: resultKey,
			},
			{
				Primitive: game.CreateToken{
					Amount:    amount,
					Source:    game.TokenCopyOf(spec),
					Recipient: opt.Val(game.AffectedTargetControllerReference(0)),
				},
				ResultGate: opt.Val(game.InstructionResultGate{Key: resultKey, Succeeded: game.TriTrue}),
			},
		},
	}.Ability(), true
}
