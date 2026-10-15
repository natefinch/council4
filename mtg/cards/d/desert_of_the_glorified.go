package d

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
)

// DesertOfTheGlorified is the card definition for Desert of the Glorified.
//
// Type: Land — Desert
//
// Oracle text:
//
//	This land enters tapped.
//	{T}: Add {B}.
//	Cycling {1}{B} ({1}{B}, Discard this card: Draw a card.)
var DesertOfTheGlorified = newDesertOfTheGlorified

func newDesertOfTheGlorified() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name:     "Desert of the Glorified",
			Types:    []types.Card{types.Land},
			Subtypes: []types.Sub{types.Desert},
			ActivatedAbilities: []game.ActivatedAbility{
				game.CyclingActivatedAbility(cost.Mana{cost.O(1), cost.B}),
			},
			ManaAbilities: []game.ManaAbility{
				game.TapManaAbility(mana.B),
			},
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntersTappedReplacement("This land enters tapped."),
			},
			OracleText: `
			This land enters tapped.
			{T}: Add {B}.
			Cycling {1}{B} ({1}{B}, Discard this card: Draw a card.)
		`,
		},
	}
}
