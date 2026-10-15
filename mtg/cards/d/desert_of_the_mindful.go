package d

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
)

// DesertOfTheMindful is the card definition for Desert of the Mindful.
//
// Type: Land — Desert
//
// Oracle text:
//
//	This land enters tapped.
//	{T}: Add {U}.
//	Cycling {1}{U} ({1}{U}, Discard this card: Draw a card.)
var DesertOfTheMindful = newDesertOfTheMindful

func newDesertOfTheMindful() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name:     "Desert of the Mindful",
			Types:    []types.Card{types.Land},
			Subtypes: []types.Sub{types.Desert},
			ActivatedAbilities: []game.ActivatedAbility{
				game.CyclingActivatedAbility(cost.Mana{cost.O(1), cost.U}),
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
			Cycling {1}{U} ({1}{U}, Discard this card: Draw a card.)
		`,
		},
	}
}
