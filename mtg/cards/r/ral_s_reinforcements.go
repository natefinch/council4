package r

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// RalSReinforcements is the card definition for Ral's Reinforcements.
//
// Type: Sorcery
// Cost: {1}{R}
//
// Oracle text:
//
//	Create two 1/1 blue and red Elemental creature tokens.
var RalSReinforcements = newRalSReinforcements

func newRalSReinforcements() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Ral's Reinforcements",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.R,
			}),
			Colors: []color.Color{color.Red},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.CreateToken{
							Amount: game.Fixed(2),
							Source: game.TokenDef(ralSReinforcementsToken),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Create two 1/1 blue and red Elemental creature tokens.
		`,
		},
	}
}

var ralSReinforcementsToken = newRalSReinforcementsToken()

func newRalSReinforcementsToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Elemental",
			Colors:    []color.Color{color.Blue, color.Red},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Elemental},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 1}),
		},
	}
}
