package b

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Blightwidow is the card definition for Blightwidow.
//
// Type: Creature — Phyrexian Spider
// Cost: {3}{G}
//
// Oracle text:
//
//	Reach (This creature can block creatures with flying.)
//	Infect (This creature deals damage to creatures in the form of -1/-1 counters and to players in the form of poison counters.)
var Blightwidow = newBlightwidow

func newBlightwidow() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Blightwidow",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.G,
			}),
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Phyrexian, types.Spider},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 4}),
			StaticAbilities: []game.StaticAbility{
				game.ReachStaticBody,
				game.InfectStaticBody,
			},
			OracleText: `
			Reach (This creature can block creatures with flying.)
			Infect (This creature deals damage to creatures in the form of -1/-1 counters and to players in the form of poison counters.)
		`,
		},
	}
}
