package d

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// DeadlyInsect is the card definition for Deadly Insect.
//
// Type: Creature — Insect
// Cost: {4}{G}
//
// Oracle text:
//
//	Shroud (This creature can't be the target of spells or abilities.)
var DeadlyInsect = newDeadlyInsect

func newDeadlyInsect() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Deadly Insect",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
				cost.G,
			}),
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Insect},
			Power:     opt.Val(game.PT{Value: 6}),
			Toughness: opt.Val(game.PT{Value: 1}),
			StaticAbilities: []game.StaticAbility{
				game.ShroudStaticBody,
			},
			OracleText: `
			Shroud (This creature can't be the target of spells or abilities.)
		`,
		},
	}
}
