package cardgen

import (
	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// beseechExiledLinkID is the body-scoped link key under which the search half of
// the bargain search/exile/conditional-cast payoff publishes the exiled card, so
// the free-cast and move-to-hand instructions reference the exact same card.
const beseechExiledLinkID = "beseech-exiled"

// lowerBargainSearchCastPayoffSequence lowers the parser-recognized spell body
// "Search your library for a card, exile it face down, then shuffle. If this
// spell was bargained, you may cast the exiled card without paying its mana cost
// if that spell's mana value is N or less. Put the exiled card into your hand if
// it wasn't cast this way." (Beseech the Mirror) into its fixed three-instruction
// template. The compiler marks the body with a text-blind exact-sequence kind and
// carries the mana-value bound as a typed parameter; this lowering reads only
// that typed value, so it never inspects Oracle words.
//
// The sequence composes existing primitives: a search that exiles one found card
// face down and publishes it under a linked key; an optional free cast of the
// linked card, gated on the resolving spell being bargained and the linked card's
// mana value being at most the bound; and an ungated move of the linked card from
// exile to hand. The move is a no-op when the card was cast (it has left exile),
// so it fires only when the card was not cast this way.
func lowerBargainSearchCastPayoffSequence(ability compiler.CompiledAbility) game.AbilityContent {
	linked := game.CardReference{Kind: game.CardReferenceLinked, LinkID: beseechExiledLinkID}
	sequence := []game.Instruction{
		{
			Primitive: game.Search{
				Player:        game.ControllerReference(),
				Amount:        game.Fixed(1),
				Spec:          game.SearchSpec{SourceZone: zone.Library, Destination: zone.Exile, ExileFaceDown: true},
				PublishLinked: game.LinkedKey(beseechExiledLinkID),
			},
		},
		{
			Primitive: game.CastForFree{Player: game.ControllerReference(), Zone: zone.Exile, Card: linked},
			Optional:  true,
			Condition: opt.Val(game.EffectCondition{
				Condition: opt.Val(game.Condition{SpellWasBargained: true}),
			}),
			CardCondition: opt.Val(game.CardSelection{
				Card: linked,
				Selection: game.Selection{
					ManaValue: opt.Val(compare.Int{Op: compare.LessOrEqual, Value: int(ability.ExactSequenceMaxManaValue)}),
				},
			}),
		},
		{Primitive: game.MoveCard{Card: linked, FromZone: zone.Exile, Destination: zone.Hand}},
	}
	return game.Mode{Text: ability.Text, Sequence: sequence}.Ability()
}
