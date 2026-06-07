package m

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
)

// Mountain is the card definition for Mountain.
//
// Type: Basic Land — Mountain
//
// Oracle text:
//
//	({T}: Add {R}.)
var Mountain = &game.CardDef{
	ColorIdentity: color.NewIdentity(color.Red),
	CardFace: game.CardFace{
		Name:       "Mountain",
		Supertypes: []types.Super{types.Basic},
		Types:      []types.Card{types.Land},
		Subtypes:   []types.Sub{types.Mountain},
		OracleText: `
			({T}: Add {R}.)
		`,
		ManaAbilities: []game.ManaAbility{
			{
				Text: `
					{T}: Add {R}.
				`,
				AdditionalCosts: cost.Tap,
				Content: game.Mode{
					Sequence: []game.Instruction{
						{
							Primitive: game.AddMana{
								Amount:    game.Fixed(1),
								ManaColor: mana.R,
							},
						},
					},
				}.Ability(),
			},
		},
	},
}
