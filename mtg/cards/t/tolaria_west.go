package t

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
)

// TolariaWest is the card definition for Tolaria West.
//
// Type: Land
//
// Oracle text:
//
//	This land enters tapped.
//	{T}: Add {U}.
//	Transmute {1}{U}{U} ({1}{U}{U}, Discard this card: Search your library for a card with mana value 0, reveal it, put it into your hand, then shuffle. Transmute only as a sorcery.)
var TolariaWest = newTolariaWest

func newTolariaWest() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name:  "Tolaria West",
			Types: []types.Card{types.Land},
			ActivatedAbilities: []game.ActivatedAbility{
				game.TransmuteActivatedAbility(cost.Mana{cost.O(1), cost.U, cost.U}, 0),
			},
			ManaAbilities: []game.ManaAbility{
				game.TapManaAbility(mana.U),
			},
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntersTappedReplacement("This land enters tapped."),
			},
			OracleText: `
			This land enters tapped.
			{T}: Add {U}.
			Transmute {1}{U}{U} ({1}{U}{U}, Discard this card: Search your library for a card with mana value 0, reveal it, put it into your hand, then shuffle. Transmute only as a sorcery.)
		`,
		},
	}
}
