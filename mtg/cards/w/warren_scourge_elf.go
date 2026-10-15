package w

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// WarrenScourgeElf is the card definition for Warren-Scourge Elf.
//
// Type: Creature — Elf Warrior
// Cost: {1}{G}
//
// Oracle text:
//
//	Protection from Goblins
var WarrenScourgeElf = newWarrenScourgeElf

func newWarrenScourgeElf() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Warren-Scourge Elf",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.G,
			}),
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Elf, types.Warrior},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 1}),
			StaticAbilities: []game.StaticAbility{
				game.ProtectionFromSubtypesStaticAbility(types.Goblin),
			},
			OracleText: `
			Protection from Goblins
		`,
		},
	}
}
