package d

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
)

// DesertOfTheFervent is the card definition for Desert of the Fervent.
//
// Type: Land — Desert
//
// Oracle text:
//
//	This land enters tapped.
//	{T}: Add {R}.
//	Cycling {1}{R} ({1}{R}, Discard this card: Draw a card.)
var DesertOfTheFervent = newDesertOfTheFervent

func newDesertOfTheFervent() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name:     "Desert of the Fervent",
			Types:    []types.Card{types.Land},
			Subtypes: []types.Sub{types.Desert},
			ActivatedAbilities: []game.ActivatedAbility{
				game.CyclingActivatedAbility(cost.Mana{cost.O(1), cost.R}),
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
			Cycling {1}{R} ({1}{R}, Discard this card: Draw a card.)
		`,
		},
	}
}
