package j

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// JackalFamiliar is the card definition for Jackal Familiar.
//
// Type: Creature — Jackal
// Cost: {R}
//
// Oracle text:
//
//	This creature can't attack or block alone.
var JackalFamiliar = newJackalFamiliar()

func newJackalFamiliar() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Jackal Familiar",
			ManaCost: opt.Val(cost.Mana{
				cost.R,
			}),
			Colors:    []color.Color{color.Red},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Jackal},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 2}),
			StaticAbilities: []game.StaticAbility{
				game.CantAttackOrBlockAloneStaticBody,
			},
			OracleText: `
			This creature can't attack or block alone.
		`,
		},
	}
}
