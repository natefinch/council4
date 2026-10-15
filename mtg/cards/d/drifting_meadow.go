package d

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
)

// DriftingMeadow is the card definition for Drifting Meadow.
//
// Type: Land
//
// Oracle text:
//
//	This land enters tapped.
//	{T}: Add {W}.
//	Cycling {2} ({2}, Discard this card: Draw a card.)
var DriftingMeadow = newDriftingMeadow

func newDriftingMeadow() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name:  "Drifting Meadow",
			Types: []types.Card{types.Land},
			ActivatedAbilities: []game.ActivatedAbility{
				game.CyclingActivatedAbility(cost.Mana{cost.O(2)}),
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
			Cycling {2} ({2}, Discard this card: Draw a card.)
		`,
		},
	}
}
