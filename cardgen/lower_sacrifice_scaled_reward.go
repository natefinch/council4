package cardgen

import (
	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/opt"
)

// sacrificedCreatureLinkKey is the linked-object key under which an optional
// resolving sacrifice records the permanent it sacrificed, so a gated reward
// effect can read that creature's power/toughness/mana value through last-known
// information once it has left the battlefield.
const sacrificedCreatureLinkKey = game.LinkedKey("sacrificed-creature")

// lowerOptionalSacrificeScaledReward lowers the "you may sacrifice another
// creature. If you do, <rewards>" family where each reward is a controller life
// gain or card draw scaled by the sacrificed creature's power, toughness, or
// mana value (Disciple of Freyalise's "you gain X life and draw X cards, where X
// is that creature's power."). The optional sacrifice publishes both its
// success (gating the rewards) and the sacrificed permanent as a linked object;
// each reward reads that permanent's characteristic via DynamicAmountObjectPower
// and friends. Any other shape — a non-creature sacrifice, an else branch, a
// reward that is neither a life gain nor a draw, or an unmodeled reference —
// fails closed.
func lowerOptionalSacrificeScaledReward(ctx contentCtx) (game.AbilityContent, bool) {
	content := ctx.content
	if len(content.Modes) != 0 ||
		len(content.Targets) != 0 ||
		len(content.Keywords) != 0 ||
		len(content.Effects) < 2 ||
		len(content.Conditions) != 1 {
		return game.AbilityContent{}, false
	}
	sacrifice := &content.Effects[0]
	if sacrifice.Kind != compiler.EffectSacrifice ||
		!sacrifice.Optional ||
		!sacrifice.Exact ||
		sacrifice.Negated ||
		sacrifice.Context != parser.EffectContextController ||
		!sacrifice.Amount.Known ||
		sacrifice.Amount.Value != 1 {
		return game.AbilityContent{}, false
	}
	// The optionality must be exactly "you may X. If you do, <tail>" with no else
	// branch, so every reward is gated on the sacrifice having succeeded.
	plan, ok := planOptionalFlow(content)
	if !ok ||
		!plan.enabled ||
		plan.publishWithoutOptional ||
		plan.optionalIndex != 0 ||
		plan.gateIndex != 1 ||
		plan.elseIndex >= 0 {
		return game.AbilityContent{}, false
	}
	// Only the source self-reference and the "that creature" demonstratives that
	// name the sacrificed creature are permitted; any other reference denotes an
	// unmodeled object and fails closed.
	for i := range content.References {
		switch content.References[i].Kind {
		case compiler.ReferenceThisObject, compiler.ReferenceSelfName, compiler.ReferenceThatObject:
		default:
			return game.AbilityContent{}, false
		}
	}

	sharedAmount, hasShared := sacrificeScaledSharedAmount(content.Effects[1:])
	rewards := make([]game.Instruction, 0, len(content.Effects)-1)
	for i := 1; i < len(content.Effects); i++ {
		effect := &content.Effects[i]
		amount, ok := sacrificeScaledRewardAmount(effect, sharedAmount, hasShared)
		if !ok {
			return game.AbilityContent{}, false
		}
		primitive, ok := sacrificeScaledRewardPrimitive(effect, amount)
		if !ok {
			return game.AbilityContent{}, false
		}
		rewards = append(rewards, game.Instruction{
			Primitive: primitive,
			ResultGate: opt.Val(game.InstructionResultGate{
				Key:       optionalIfYouDoResultKey,
				Succeeded: game.TriTrue,
			}),
		})
	}

	sacrificeInstr, ok := lowerSequenceSacrificeInstruction(ctx)
	if !ok {
		return game.AbilityContent{}, false
	}
	sacrifice0, ok := sacrificeInstr.Primitive.(game.SacrificePermanents)
	if !ok {
		return game.AbilityContent{}, false
	}
	sacrifice0.PublishLinked = sacrificedCreatureLinkKey
	sacrificeInstr.Primitive = sacrifice0
	sacrificeInstr.Optional = true
	sacrificeInstr.PublishResult = optionalIfYouDoResultKey

	sequence := make([]game.Instruction, 0, len(rewards)+1)
	sequence = append(sequence, sacrificeInstr)
	sequence = append(sequence, rewards...)
	return game.Mode{Sequence: sequence}.Ability(), true
}

// sacrificeScaledSharedAmount resolves the lone "where X is that creature's
// <characteristic>" reward definition so sibling bare-X reward clauses can adopt
// it. It returns false when no reward clause carries such a definition.
func sacrificeScaledSharedAmount(rewards []compiler.CompiledEffect) (game.Quantity, bool) {
	for i := range rewards {
		amount := rewards[i].Amount
		if amount.DynamicForm != compiler.DynamicAmountWhereX {
			continue
		}
		if quantity, ok := sacrificedCharacteristicAmount(amount); ok {
			return quantity, true
		}
	}
	return game.Quantity{}, false
}

// sacrificeScaledRewardAmount resolves one reward clause's quantity: its own
// "that creature's <characteristic>" definition, or the shared definition when
// the clause carries only the bare variable X.
func sacrificeScaledRewardAmount(
	effect *compiler.CompiledEffect,
	shared game.Quantity,
	hasShared bool,
) (game.Quantity, bool) {
	amount := effect.Amount
	if sacrificedCharacteristicKind(amount.DynamicKind) {
		return sacrificedCharacteristicAmount(amount)
	}
	if amount.VariableX && amount.DynamicKind == compiler.DynamicAmountNone && hasShared {
		return shared, true
	}
	return game.Quantity{}, false
}

// sacrificedCharacteristicAmount lowers a "that creature's power/toughness/mana
// value" amount that reads the sacrificed creature through the linked object the
// sacrifice published. It fails closed for any other dynamic kind.
func sacrificedCharacteristicAmount(amount compiler.CompiledAmount) (game.Quantity, bool) {
	if !sacrificedCharacteristicKind(amount.DynamicKind) {
		return game.Quantity{}, false
	}
	dynamic, ok := lowerDynamicAmount(amount, game.LinkedObjectReference(string(sacrificedCreatureLinkKey)))
	if !ok {
		return game.Quantity{}, false
	}
	return game.Dynamic(dynamic), true
}

// sacrificedCharacteristicKind reports whether a dynamic-amount kind reads a
// referenced object's power, toughness, or mana value — the characteristics a
// sacrificed creature can scale a reward by.
func sacrificedCharacteristicKind(kind compiler.DynamicAmountKind) bool {
	switch kind {
	case compiler.DynamicAmountSourcePower,
		compiler.DynamicAmountSourceToughness,
		compiler.DynamicAmountSourceManaValue:
		return true
	default:
		return false
	}
}

// sacrificeScaledRewardPrimitive builds a reward clause's primitive: a
// controller life gain or card draw of the resolved quantity. Any other effect
// kind, recipient, or modifier fails closed.
func sacrificeScaledRewardPrimitive(
	effect *compiler.CompiledEffect,
	amount game.Quantity,
) (game.Primitive, bool) {
	if effect.Context != parser.EffectContextController ||
		!effect.Exact ||
		effect.Negated ||
		effect.Optional {
		return nil, false
	}
	switch effect.Kind {
	case compiler.EffectGain:
		if !effect.LifeObject {
			return nil, false
		}
		return game.GainLife{Player: game.ControllerReference(), Amount: amount}, true
	case compiler.EffectDraw:
		return game.Draw{Player: game.ControllerReference(), Amount: amount}, true
	default:
		return nil, false
	}
}
