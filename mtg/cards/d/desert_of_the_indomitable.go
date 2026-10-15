package d

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
)

// DesertOfTheIndomitable is the card definition for Desert of the Indomitable.
//
// Type: Land — Desert
//
// Oracle text:
//
//	This land enters tapped.
//	{T}: Add {G}.
//	Cycling {1}{G} ({1}{G}, Discard this card: Draw a card.)
var DesertOfTheIndomitable = newDesertOfTheIndomitable

func newDesertOfTheIndomitable() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name:     "Desert of the Indomitable",
			Types:    []types.Card{types.Land},
			Subtypes: []types.Sub{types.Desert},
			ActivatedAbilities: []game.ActivatedAbility{
				game.CyclingActivatedAbility(cost.Mana{cost.O(1), cost.G}),
			},
			ManaAbilities: []game.ManaAbility{
				game.TapManaAbility(mana.G),
			},
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntersTappedReplacement("This land enters tapped."),
			},
			OracleText: `
			This land enters tapped.
			{T}: Add {G}.
			Cycling {1}{G} ({1}{G}, Discard this card: Draw a card.)
		`,
		},
	}
}
