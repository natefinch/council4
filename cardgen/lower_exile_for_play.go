package cardgen

import (
	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/zone"
)

// lowerExileForPlay lowers a discard trigger's reflexive "you may exile that
// card from your graveyard. If you do, you may play (or cast) that card <this
// turn>." body into a single ExileForPlay primitive.
//
// Containment Construct ("Whenever you discard a card, you may exile that card
// from your graveyard. If you do, you may play that card this turn.") is the
// canonical shape: the discarded card rests in the graveyard, the optional
// exile moves it to exile, and the trailing permission lets its controller play
// it for the remainder of the turn. Conspiracy Theorist's "you may cast it this
// turn" is the cast-permission sibling.
//
// The compiler models this as two optional effects gated by a
// PriorInstructionAccepted condition: effect[0] exiles the back-referenced
// ("that card") just-discarded object from the graveyard, and effect[1] grants
// the play/cast permission. Because exiling the card advances its zone version,
// the move and the permission grant cannot be two instructions sharing one
// event reference; the combined ExileForPlay primitive captures the card
// identity once and performs both atomically.
//
// The lowerer accepts two back-reference shapes. The single "that card"
// (ReferenceThatObject) shape selects the lone discarded card the runtime
// resolves through CardReferenceEvent. The plural "one of them"
// (ReferencePronoun/ReferencePronounThem, Amount 1) shape over a "discard one
// or more cards" batch sets SelectFromBatch, letting the runtime reconstruct
// the batch and have the controller choose which card to exile.
func lowerExileForPlay(ctx contentCtx) (game.AbilityContent, bool) {
	if ctx.enclosingKind != compiler.AbilityTriggered ||
		ctx.triggerEvent != game.EventCardDiscarded ||
		len(ctx.content.Effects) != 2 ||
		len(ctx.content.Targets) != 0 ||
		len(ctx.content.Modes) != 0 ||
		len(ctx.content.Keywords) != 0 {
		return game.AbilityContent{}, false
	}
	exile := ctx.content.Effects[0]
	if exile.Kind != compiler.EffectExile ||
		!exile.Optional ||
		exile.Negated ||
		exile.Context != parser.EffectContextController ||
		exile.FromZone != zone.Graveyard {
		return game.AbilityContent{}, false
	}
	selectFromBatch := false
	switch {
	case exileBackReferencesSingleObject(exile):
	case exileBackReferencesBatchSelection(exile, ctx):
		selectFromBatch = true
	default:
		return game.AbilityContent{}, false
	}
	grant := ctx.content.Effects[1]
	cast := grant.Kind == compiler.EffectCast
	if (grant.Kind != compiler.EffectPlay && grant.Kind != compiler.EffectCast) ||
		grant.Negated ||
		grant.Context != parser.EffectContextController {
		return game.AbilityContent{}, false
	}
	duration, ok := lowerImpulseExileDuration(grant.Duration)
	if !ok {
		return game.AbilityContent{}, false
	}
	if !exileForPlayConditions(ctx.content.Conditions) {
		return game.AbilityContent{}, false
	}
	return game.Mode{Sequence: []game.Instruction{{
		Optional: true,
		Primitive: game.ExileForPlay{
			Card:            game.CardReference{Kind: game.CardReferenceEvent},
			FromZone:        zone.Graveyard,
			Duration:        duration,
			Cast:            cast,
			SelectFromBatch: selectFromBatch,
		},
	}}}.Ability(), true
}

// exileBackReferencesBatchSelection reports whether an exile effect selects
// "one of them" from a coalesced discard batch: a single plural "them" pronoun
// (ReferencePronoun/ReferencePronounThem) that exiles exactly one card
// (Amount 1) inside a "discard one or more cards" trigger (triggerOneOrMore).
// The runtime reconstructs the batch from the triggering events and has the
// controller choose which card to exile, so the back-reference need not resolve
// to a single CardReferenceEvent card.
func exileBackReferencesBatchSelection(effect compiler.CompiledEffect, ctx contentCtx) bool {
	if !ctx.triggerOneOrMore {
		return false
	}
	if !effect.Amount.Known || effect.Amount.Value != 1 {
		return false
	}
	if len(effect.References) != 1 {
		return false
	}
	ref := effect.References[0]
	return ref.Kind == compiler.ReferencePronoun && ref.Pronoun == compiler.ReferencePronounThem
}

// exileBackReferencesSingleObject reports whether an exile effect's references
// are exactly one "that card" back-reference (ReferenceThatObject), the lone
// just-discarded object the runtime resolves through CardReferenceEvent. A
// plural pronoun ("one of them") over several discarded cards is rejected.
func exileBackReferencesSingleObject(effect compiler.CompiledEffect) bool {
	count := 0
	for i := range effect.References {
		if effect.References[i].Kind != compiler.ReferenceThatObject {
			return false
		}
		count++
	}
	return count == 1
}

// exileForPlayConditions reports whether the reflexive body carries only its
// implied PriorInstructionAccepted gate ("If you do, ..."), which the combined
// ExileForPlay primitive subsumes by performing the exile and permission grant
// atomically. Any other condition is rejected.
func exileForPlayConditions(conditions []compiler.CompiledCondition) bool {
	switch len(conditions) {
	case 0:
		return true
	case 1:
		return conditions[0].Predicate == compiler.ConditionPredicatePriorInstructionAccepted
	default:
		return false
	}
}
