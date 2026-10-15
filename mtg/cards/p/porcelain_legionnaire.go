package p

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// PorcelainLegionnaire is the card definition for Porcelain Legionnaire.
//
// Type: Artifact Creature — Phyrexian Soldier
// Cost: {2}{W/P}
//
// Oracle text:
//
//	({W/P} can be paid with either {W} or 2 life.)
//	First strike
var PorcelainLegionnaire = newPorcelainLegionnaire

func newPorcelainLegionnaire() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Porcelain Legionnaire",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.PhyrexianMana(mana.W),
			}),
			Colors:    []color.Color{color.White},
			Types:     []types.Card{types.Artifact, types.Creature},
			Subtypes:  []types.Sub{types.Phyrexian, types.Soldier},
			Power:     opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 1}),
			StaticAbilities: []game.StaticAbility{
				game.FirstStrikeStaticBody,
			},
			OracleText: `
			({W/P} can be paid with either {W} or 2 life.)
			First strike
		`,
		},
	}
}
