package a

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// AdventOfTheWurm is the card definition for Advent of the Wurm.
//
// Type: Instant
// Cost: {1}{G}{G}{W}
//
// Oracle text:
//
//	Create a 5/5 green Wurm creature token with trample.
var AdventOfTheWurm = newAdventOfTheWurm

func newAdventOfTheWurm() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Green),
		CardFace: game.CardFace{
			Name: "Advent of the Wurm",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.G,
				cost.G,
				cost.W,
			}),
			Colors: []color.Color{color.Green, color.White},
			Types:  []types.Card{types.Instant},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.CreateToken{
							Amount: game.Fixed(1),
							Source: game.TokenDef(adventOfTheWurmToken),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Create a 5/5 green Wurm creature token with trample.
		`,
		},
	}
}

var adventOfTheWurmToken = newAdventOfTheWurmToken()

func newAdventOfTheWurmToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Wurm",
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Wurm},
			Power:     opt.Val(game.PT{Value: 5}),
			Toughness: opt.Val(game.PT{Value: 5}),
			StaticAbilities: []game.StaticAbility{
				game.TrampleStaticBody,
			},
		},
	}
}
