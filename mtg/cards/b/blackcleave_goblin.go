package b

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// BlackcleaveGoblin is the card definition for Blackcleave Goblin.
//
// Type: Creature — Phyrexian Goblin Zombie
// Cost: {3}{B}
//
// Oracle text:
//
//	Haste
//	Infect (This creature deals damage to creatures in the form of -1/-1 counters and to players in the form of poison counters.)
var BlackcleaveGoblin = newBlackcleaveGoblin

func newBlackcleaveGoblin() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Blackcleave Goblin",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.B,
			}),
			Colors:    []color.Color{color.Black},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Phyrexian, types.Goblin, types.Zombie},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 1}),
			StaticAbilities: []game.StaticAbility{
				game.HasteStaticBody,
				game.InfectStaticBody,
			},
			OracleText: `
			Haste
			Infect (This creature deals damage to creatures in the form of -1/-1 counters and to players in the form of poison counters.)
		`,
		},
	}
}
