package cardgen

import (
	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// lowerOptionalSacrificeReturnWithCounters lowers the "You may sacrifice a
// creature. If you do, return that card to the battlefield under its owner's
// control with N +1/+1 counters on it[ and you become the monarch]." family
// (Heart-Shaped Herb). The optional sacrifice publishes both its success (gating
// the follow-up) and the sacrificed permanent as a linked object; the return
// reads that linked card and puts it onto the battlefield under its owner's
// control with the +1/+1 counters, and the optional trailing "you become the
// monarch" clause runs under the same gate. Any deviation — a non-creature
// sacrifice, an else branch, a control rider other than owner's control, a
// counter kind other than +1/+1, a non-battlefield destination, or an extra
// effect — fails closed.
func lowerOptionalSacrificeReturnWithCounters(ctx contentCtx) (game.AbilityContent, bool) {
	content := ctx.content
	if len(content.Modes) != 0 ||
		len(content.Targets) != 0 ||
		len(content.Keywords) != 0 ||
		len(content.Conditions) != 1 ||
		len(content.Effects) < 2 ||
		len(content.Effects) > 3 {
		return game.AbilityContent{}, false
	}
	// The optionality must be exactly "you may X. If you do, <tail>" with the
	// sacrifice first and every follow-up gated on it, and no else branch.
	plan, ok := planOptionalFlow(content)
	if !ok ||
		!plan.enabled ||
		plan.publishWithoutOptional ||
		plan.independentOptional ||
		plan.optionalIndex != 0 ||
		plan.gateIndex != 1 ||
		plan.elseIndex >= 0 ||
		plan.extraOptionalIndex >= 0 {
		return game.AbilityContent{}, false
	}
	sacrifice := &content.Effects[0]
	if sacrifice.Kind != compiler.EffectSacrifice ||
		!sacrifice.Optional ||
		!sacrifice.Exact ||
		sacrifice.Negated ||
		sacrifice.Context != parser.EffectContextController ||
		!sacrifice.Amount.Known ||
		sacrifice.Amount.Value != 1 ||
		!matchesPlainCreatureCardSelector(sacrifice.Selector, compiler.ControllerAny, zone.None) ||
		len(sacrifice.Targets) != 0 ||
		len(sacrifice.References) != 0 {
		return game.AbilityContent{}, false
	}
	counters, ok := sacrificedReturnEntryCounters(&content.Effects[1])
	if !ok {
		return game.AbilityContent{}, false
	}
	becomeMonarch := false
	if len(content.Effects) == 3 {
		if !isControllerBecomeMonarchEffect(&content.Effects[2]) {
			return game.AbilityContent{}, false
		}
		becomeMonarch = true
	}

	sacrificeInstr, ok := lowerSequenceSacrificeInstruction(ctx)
	if !ok {
		return game.AbilityContent{}, false
	}
	sacrificePrimitive, ok := sacrificeInstr.Primitive.(game.SacrificePermanents)
	if !ok {
		return game.AbilityContent{}, false
	}
	sacrificePrimitive.PublishLinked = sacrificedCreatureLinkKey
	sacrificeInstr.Primitive = sacrificePrimitive
	sacrificeInstr.Optional = true
	sacrificeInstr.PublishResult = optionalIfYouDoResultKey

	gate := opt.Val(game.InstructionResultGate{
		Key:       optionalIfYouDoResultKey,
		Succeeded: game.TriTrue,
	})
	sequence := []game.Instruction{
		sacrificeInstr,
		{
			Primitive: game.PutOnBattlefield{
				Source:        game.LinkedBattlefieldSource(sacrificedCreatureLinkKey),
				EntryCounters: counters,
			},
			ResultGate: gate,
		},
	}
	if becomeMonarch {
		sequence = append(sequence, game.Instruction{
			Primitive:  game.BecomeMonarch{Player: game.ControllerReference()},
			ResultGate: gate,
		})
	}
	return game.Mode{Sequence: sequence}.Ability(), true
}

// sacrificedReturnEntryCounters validates the "return that card to the
// battlefield under its owner's control with N +1/+1 counters on it" consumer of
// an optional-sacrifice flow and returns the entry-counter placement. The
// returned card is the sacrificed creature (read through the linked object the
// sacrifice publishes), so its references may only be the "that card"
// demonstrative and the "its"/"it" pronouns that name it; any other shape fails
// closed.
func sacrificedReturnEntryCounters(effect *compiler.CompiledEffect) ([]game.CounterPlacement, bool) {
	if effect.Kind != compiler.EffectReturn ||
		effect.Negated ||
		effect.Optional ||
		effect.Divided ||
		effect.Context != parser.EffectContextController ||
		effect.DelayedTiming != 0 ||
		effect.Duration != compiler.DurationNone ||
		effect.FromZone != zone.None ||
		effect.ToZone != zone.Battlefield ||
		effect.EntersTapped ||
		effect.UnderYourControl ||
		!effect.UnderOwnersControl ||
		!effect.CounterKindKnown ||
		effect.CounterKind != counter.PlusOnePlusOne {
		return nil, false
	}
	if !effect.Amount.Known ||
		effect.Amount.RangeKnown ||
		effect.Amount.VariableX ||
		effect.Amount.DynamicKind != compiler.DynamicAmountNone ||
		effect.Amount.Value < 1 {
		return nil, false
	}
	for i := range effect.References {
		switch effect.References[i].Kind {
		case compiler.ReferenceThatObject, compiler.ReferencePronoun:
		default:
			return nil, false
		}
	}
	return []game.CounterPlacement{{Kind: counter.PlusOnePlusOne, Amount: effect.Amount.Value}}, true
}

// isControllerBecomeMonarchEffect reports whether effect is a bare "you become
// the monarch" clause with no targets, references, or riders.
func isControllerBecomeMonarchEffect(effect *compiler.CompiledEffect) bool {
	return effect.Kind == compiler.EffectBecomeMonarch &&
		!effect.Negated &&
		!effect.Optional &&
		effect.Context == parser.EffectContextController &&
		len(effect.Targets) == 0 &&
		len(effect.References) == 0
}
