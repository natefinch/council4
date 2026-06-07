package f

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
)

// Forest is the card definition for Forest.
//
// Type: Basic Land — Forest
//
// Oracle text:
//
//	({T}: Add {G}.)
var Forest = &game.CardDef{
	ColorIdentity: color.NewIdentity(color.Green),
	CardFace: game.CardFace{
		Name:       "Forest",
		Supertypes: []types.Super{types.Basic},
		Types:      []types.Card{types.Land},
		Subtypes:   []types.Sub{types.Forest},
		OracleText: `
			({T}: Add {G}.)
		`,
		ManaAbilities: []game.ManaAbility{
			{
				Text: `
					{T}: Add {G}.
				`,
				AdditionalCosts: cost.Tap,
				Content: game.Mode{
					Sequence: []game.Instruction{
						{
							Primitive: game.AddMana{
								Amount:    game.Fixed(1),
								ManaColor: mana.G,
							},
						},
					},
				}.Ability(),
			},
		},
	},
}
