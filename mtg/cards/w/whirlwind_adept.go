package w

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// WhirlwindAdept is the card definition for Whirlwind Adept.
//
// Type: Creature — Djinn Monk
// Cost: {4}{U}
//
// Oracle text:
//
//	Hexproof (This creature can't be the target of spells or abilities your opponents control.)
//	Prowess (Whenever you cast a noncreature spell, this creature gets +1/+1 until end of turn.)
var WhirlwindAdept = newWhirlwindAdept

func newWhirlwindAdept() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Whirlwind Adept",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
				cost.U,
			}),
			Colors:    []color.Color{color.Blue},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Djinn, types.Monk},
			Power:     opt.Val(game.PT{Value: 4}),
			Toughness: opt.Val(game.PT{Value: 2}),
			StaticAbilities: []game.StaticAbility{
				game.HexproofStaticBody,
				game.ProwessStaticBody,
			},
			OracleText: `
			Hexproof (This creature can't be the target of spells or abilities your opponents control.)
			Prowess (Whenever you cast a noncreature spell, this creature gets +1/+1 until end of turn.)
		`,
		},
	}
}
