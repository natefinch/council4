package p

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// PhyrexianCrusader is the card definition for Phyrexian Crusader.
//
// Type: Creature — Phyrexian Zombie Knight
// Cost: {1}{B}{B}
//
// Oracle text:
//
//	First strike, protection from red and from white
//	Infect (This creature deals damage to creatures in the form of -1/-1 counters and to players in the form of poison counters.)
var PhyrexianCrusader = newPhyrexianCrusader

func newPhyrexianCrusader() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Phyrexian Crusader",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.B,
				cost.B,
			}),
			Colors:    []color.Color{color.Black},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Phyrexian, types.Zombie, types.Knight},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 2}),
			StaticAbilities: []game.StaticAbility{
				game.FirstStrikeStaticBody,
				game.ProtectionFromColorsStaticAbility(color.Red, color.White),
				game.InfectStaticBody,
			},
			OracleText: `
			First strike, protection from red and from white
			Infect (This creature deals damage to creatures in the form of -1/-1 counters and to players in the form of poison counters.)
		`,
		},
	}
}
