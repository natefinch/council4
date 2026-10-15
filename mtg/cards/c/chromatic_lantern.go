package c

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// ChromaticLantern is the card definition for Chromatic Lantern.
//
// Type: Artifact
// Cost: {3}
//
// Oracle text:
//
//	Lands you control have "{T}: Add one mana of any color."
//	{T}: Add one mana of any color.
var ChromaticLantern = newChromaticLantern

func newChromaticLantern() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Chromatic Lantern",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
			}),
			Types: []types.Card{types.Artifact},
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "Lands you control have \"{T}: Add one mana of any color.\"",
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Duration: game.DurationPermanent,
							Group:    game.ObjectControlledGroup(game.SourcePermanentReference(), game.Selection{RequiredTypes: []types.Card{types.Land}}),
							Layer:    game.LayerAbility,
							AddAbilities: []game.Ability{
								new(game.TapManaChoiceAbility(mana.W, mana.U, mana.B, mana.R, mana.G)),
							},
						},
					},
				},
			},
			ManaAbilities: []game.ManaAbility{
				game.TapManaChoiceAbility(mana.W, mana.U, mana.B, mana.R, mana.G),
			},
			OracleText: `
			Lands you control have "{T}: Add one mana of any color."
			{T}: Add one mana of any color.
		`,
		},
	}
}
