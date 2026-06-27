package b

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// BondedHorncrest is the card definition for Bonded Horncrest.
//
// Type: Creature — Dinosaur
// Cost: {3}{R}
//
// Oracle text:
//
//	This creature can't attack or block alone.
var BondedHorncrest = newBondedHorncrest()

func newBondedHorncrest() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Bonded Horncrest",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.R,
			}),
			Colors:    []color.Color{color.Red},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Dinosaur},
			Power:     opt.Val(game.PT{Value: 5}),
			Toughness: opt.Val(game.PT{Value: 5}),
			StaticAbilities: []game.StaticAbility{
				game.CantAttackOrBlockAloneStaticBody,
			},
			OracleText: `
			This creature can't attack or block alone.
		`,
		},
	}
}
