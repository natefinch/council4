package e

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// EldraziDevastator is the card definition for Eldrazi Devastator.
//
// Type: Creature — Eldrazi
// Cost: {8}
//
// Oracle text:
//
//	Trample
var EldraziDevastator = newEldraziDevastator

func newEldraziDevastator() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Eldrazi Devastator",
			ManaCost: opt.Val(cost.Mana{
				cost.O(8),
			}),
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Eldrazi},
			Power:     opt.Val(game.PT{Value: 8}),
			Toughness: opt.Val(game.PT{Value: 9}),
			StaticAbilities: []game.StaticAbility{
				game.TrampleStaticBody,
			},
			OracleText: `
			Trample
		`,
		},
	}
}
