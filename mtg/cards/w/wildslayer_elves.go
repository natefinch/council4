package w

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// WildslayerElves is the card definition for Wildslayer Elves.
//
// Type: Creature — Elf Warrior
// Cost: {3}{G}
//
// Oracle text:
//
//	Wither (This deals damage to creatures in the form of -1/-1 counters.)
var WildslayerElves = newWildslayerElves

func newWildslayerElves() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Wildslayer Elves",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.G,
			}),
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Elf, types.Warrior},
			Power:     opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 3}),
			StaticAbilities: []game.StaticAbility{
				game.WitherStaticBody,
			},
			OracleText: `
			Wither (This deals damage to creatures in the form of -1/-1 counters.)
		`,
		},
	}
}
