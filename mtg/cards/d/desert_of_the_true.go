package d

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
)

// DesertOfTheTrue is the card definition for Desert of the True.
//
// Type: Land — Desert
//
// Oracle text:
//
//	This land enters tapped.
//	{T}: Add {W}.
//	Cycling {1}{W} ({1}{W}, Discard this card: Draw a card.)
var DesertOfTheTrue = newDesertOfTheTrue

func newDesertOfTheTrue() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name:     "Desert of the True",
			Types:    []types.Card{types.Land},
			Subtypes: []types.Sub{types.Desert},
			ActivatedAbilities: []game.ActivatedAbility{
				game.CyclingActivatedAbility(cost.Mana{cost.O(1), cost.W}),
			},
			ManaAbilities: []game.ManaAbility{
				game.TapManaAbility(mana.W),
			},
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntersTappedReplacement("This land enters tapped."),
			},
			OracleText: `
			This land enters tapped.
			{T}: Add {W}.
			Cycling {1}{W} ({1}{W}, Discard this card: Draw a card.)
		`,
		},
	}
}
