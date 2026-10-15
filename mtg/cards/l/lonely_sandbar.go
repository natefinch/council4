package l

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
)

// LonelySandbar is the card definition for Lonely Sandbar.
//
// Type: Land
//
// Oracle text:
//
//	This land enters tapped.
//	{T}: Add {U}.
//	Cycling {U} ({U}, Discard this card: Draw a card.)
var LonelySandbar = newLonelySandbar

func newLonelySandbar() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name:  "Lonely Sandbar",
			Types: []types.Card{types.Land},
			ActivatedAbilities: []game.ActivatedAbility{
				game.CyclingActivatedAbility(cost.Mana{cost.U}),
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
			Cycling {U} ({U}, Discard this card: Draw a card.)
		`,
		},
	}
}
