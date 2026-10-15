package k

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// KreeCommandos is the card definition for Kree Commandos.
//
// Type: Creature — Kree Soldier Villain
// Cost: {2}{W}
//
// Oracle text:
//
//	Flying, vigilance
//	Prowess (Whenever you cast a noncreature spell, this creature gets +1/+1 until end of turn.)
var KreeCommandos = newKreeCommandos

func newKreeCommandos() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Kree Commandos",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.W,
			}),
			Colors:    []color.Color{color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Sub("Kree"), types.Soldier, types.Villain},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 1}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
				game.VigilanceStaticBody,
				game.ProwessStaticBody,
			},
			OracleText: `
			Flying, vigilance
			Prowess (Whenever you cast a noncreature spell, this creature gets +1/+1 until end of turn.)
		`,
		},
	}
}
