package cardgen

import (
	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/zone"
)

const exiledTopCardsLinkKey game.LinkedKey = "exiled-top-cards"

// lowerExileTopThenPutAnyAmongToBattlefield lowers the typed sequence:
//
//	Exile the top N cards of your library. You may put any number of <filter>
//	cards from among them onto the battlefield.
//
// The exile publishes exactly the cards that reach exile, and ChooseFromZone
// restricts the optional any-number battlefield choice to that linked set. The
// shape is filter-generic, so creature/land unions and single card types share
// the same lowering without card-name or Oracle-text checks.
func lowerExileTopThenPutAnyAmongToBattlefield(ctx contentCtx) (game.AbilityContent, bool) {
	if ctx.optional ||
		len(ctx.content.Effects) != 2 ||
		len(ctx.content.Conditions) != 0 ||
		len(ctx.content.Targets) != 0 ||
		len(ctx.content.Keywords) != 0 ||
		len(ctx.content.Modes) != 0 {
		return game.AbilityContent{}, false
	}
	exile := ctx.content.Effects[0]
	put := ctx.content.Effects[1]
	if exile.Kind != compiler.EffectExile ||
		exile.CardSource != parser.EffectCardSourceTopOfPlayerLibrary ||
		exile.Context != parser.EffectContextController ||
		!exile.Exact ||
		exile.Negated ||
		exile.Optional ||
		exile.FaceDown ||
		exile.CounterKindKnown ||
		len(exile.Targets) != 0 ||
		len(exile.References) != 0 {
		return game.AbilityContent{}, false
	}
	amount, ok := cardCountQuantityForContext(ctx, exile.Amount, true)
	if !ok {
		return game.AbilityContent{}, false
	}
	if put.Kind != compiler.EffectPut ||
		put.Context != parser.EffectContextController ||
		!put.Optional ||
		put.Negated ||
		put.DelayedTiming != 0 ||
		put.ToZone != zone.Battlefield ||
		put.EntersTapped ||
		put.EntersAttacking ||
		put.EntersTransformed ||
		put.UnderYourControl ||
		put.Payment.Form != parser.EffectPaymentFormUnknown ||
		!put.Amount.AnyNumber ||
		len(put.Targets) != 0 ||
		len(put.References) != 1 ||
		put.References[0].Pronoun != compiler.ReferencePronounThem ||
		put.References[0].Binding != compiler.ReferenceBindingPriorInstructionResult ||
		put.References[0].PriorInstruction != 0 {
		return game.AbilityContent{}, false
	}
	selection, ok := cardSelectionForSelector(put.Selector)
	if !ok {
		return game.AbilityContent{}, false
	}
	return game.Mode{Sequence: []game.Instruction{
		{Primitive: game.ExileTopOfLibrary{
			Amount:        amount,
			Player:        game.ControllerReference(),
			PublishLinked: exiledTopCardsLinkKey,
		}},
		{Primitive: game.ChooseFromZone{
			Player:      game.ControllerReference(),
			SourceZone:  zone.Exile,
			Filter:      selection,
			Count:       game.ChooseAnyNumber,
			Destination: game.ChooseDestination{Zone: zone.Battlefield},
			Riders: game.ChooseRiders{
				FromLinked: exiledTopCardsLinkKey,
			},
		}},
	}}.Ability(), true
}
