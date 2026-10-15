package f

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
)

// ForgottenCave is the card definition for Forgotten Cave.
//
// Type: Land
//
// Oracle text:
//
//	This land enters tapped.
//	{T}: Add {R}.
//	Cycling {R} ({R}, Discard this card: Draw a card.)
var ForgottenCave = newForgottenCave

func newForgottenCave() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name:  "Forgotten Cave",
			Types: []types.Card{types.Land},
			ActivatedAbilities: []game.ActivatedAbility{
				game.CyclingActivatedAbility(cost.Mana{cost.R}),
			},
			ManaAbilities: []game.ManaAbility{
				game.TapManaAbility(mana.R),
			},
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntersTappedReplacement("This land enters tapped."),
			},
			OracleText: `
			This land enters tapped.
			{T}: Add {R}.
			Cycling {R} ({R}, Discard this card: Draw a card.)
		`,
		},
	}
}
