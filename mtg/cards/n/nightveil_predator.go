package n

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// NightveilPredator is the card definition for Nightveil Predator.
//
// Type: Creature — Vampire
// Cost: {U}{U}{B}{B}
//
// Oracle text:
//
//	Flying, deathtouch
//	Hexproof (This creature can't be the target of spells or abilities your opponents control.)
var NightveilPredator = newNightveilPredator

func newNightveilPredator() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue, color.Black),
		CardFace: game.CardFace{
			Name: "Nightveil Predator",
			ManaCost: opt.Val(cost.Mana{
				cost.U,
				cost.U,
				cost.B,
				cost.B,
			}),
			Colors:    []color.Color{color.Black, color.Blue},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Vampire},
			Power:     opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 3}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
				game.DeathtouchStaticBody,
				game.HexproofStaticBody,
			},
			OracleText: `
			Flying, deathtouch
			Hexproof (This creature can't be the target of spells or abilities your opponents control.)
		`,
		},
	}
}
