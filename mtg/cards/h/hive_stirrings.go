package h

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// HiveStirrings is the card definition for Hive Stirrings.
//
// Type: Sorcery
// Cost: {2}{W}
//
// Oracle text:
//
//	Create two 1/1 colorless Sliver creature tokens.
var HiveStirrings = newHiveStirrings

func newHiveStirrings() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Hive Stirrings",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.W,
			}),
			Colors: []color.Color{color.White},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.CreateToken{
							Amount: game.Fixed(2),
							Source: game.TokenDef(hiveStirringsToken),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Create two 1/1 colorless Sliver creature tokens.
		`,
		},
	}
}

var hiveStirringsToken = newHiveStirringsToken()

func newHiveStirringsToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Sliver",
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Sliver},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 1}),
		},
	}
}
