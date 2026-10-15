package a

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// AshenmoorGouger is the card definition for Ashenmoor Gouger.
//
// Type: Creature — Elemental Warrior
// Cost: {B/R}{B/R}{B/R}
//
// Oracle text:
//
//	This creature can't block.
var AshenmoorGouger = newAshenmoorGouger

func newAshenmoorGouger() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black, color.Red),
		CardFace: game.CardFace{
			Name: "Ashenmoor Gouger",
			ManaCost: opt.Val(cost.Mana{
				cost.HybridMana(mana.B, mana.R),
				cost.HybridMana(mana.B, mana.R),
				cost.HybridMana(mana.B, mana.R),
			}),
			Colors:    []color.Color{color.Black, color.Red},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Elemental, types.Warrior},
			Power:     opt.Val(game.PT{Value: 4}),
			Toughness: opt.Val(game.PT{Value: 4}),
			StaticAbilities: []game.StaticAbility{
				game.CantBlockStaticBody,
			},
			OracleText: `
			This creature can't block.
		`,
		},
	}
}
