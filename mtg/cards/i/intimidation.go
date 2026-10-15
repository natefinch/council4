package i

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Intimidation is the card definition for Intimidation.
//
// Type: Enchantment
// Cost: {2}{B}{B}{B}
//
// Oracle text:
//
//	Creatures you control have fear. (They can't be blocked except by artifact creatures and/or black creatures.)
var Intimidation = newIntimidation

func newIntimidation() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Intimidation",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.B,
				cost.B,
				cost.B,
			}),
			Colors: []color.Color{color.Black},
			Types:  []types.Card{types.Enchantment},
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "Creatures you control have fear. (They can't be blocked except by artifact creatures and/or black creatures.)",
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Duration:    game.DurationPermanent,
							Group:       game.ObjectControlledGroup(game.SourcePermanentReference(), game.Selection{RequiredTypes: []types.Card{types.Creature}}),
							Layer:       game.LayerAbility,
							AddKeywords: []game.Keyword{game.Fear},
						},
					},
				},
			},
			OracleText: `
			Creatures you control have fear. (They can't be blocked except by artifact creatures and/or black creatures.)
		`,
		},
	}
}
