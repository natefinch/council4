package c

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// CliffThreader is the card definition for Cliff Threader.
//
// Type: Creature — Kor Scout
// Cost: {1}{W}
//
// Oracle text:
//
//	Mountainwalk (This creature can't be blocked as long as defending player controls a Mountain.)
var CliffThreader = newCliffThreader

func newCliffThreader() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Cliff Threader",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.W,
			}),
			Colors:    []color.Color{color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Kor, types.Scout},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 1}),
			StaticAbilities: []game.StaticAbility{
				game.MountainwalkStaticBody,
			},
			OracleText: `
			Mountainwalk (This creature can't be blocked as long as defending player controls a Mountain.)
		`,
		},
	}
}
