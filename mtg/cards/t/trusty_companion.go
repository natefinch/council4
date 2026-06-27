package t

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// TrustyCompanion is the card definition for Trusty Companion.
//
// Type: Creature — Hyena
// Cost: {1}{W}
//
// Oracle text:
//
//	Vigilance
//	This creature can't attack alone.
var TrustyCompanion = newTrustyCompanion()

func newTrustyCompanion() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Trusty Companion",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.W,
			}),
			Colors:    []color.Color{color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Hyena},
			Power:     opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 2}),
			StaticAbilities: []game.StaticAbility{
				game.VigilanceStaticBody,
				game.CantAttackAloneStaticBody,
			},
			OracleText: `
			Vigilance
			This creature can't attack alone.
		`,
		},
	}
}
