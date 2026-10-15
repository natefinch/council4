package n

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// NathSBuffoon is the card definition for Nath's Buffoon.
//
// Type: Creature — Goblin Rogue
// Cost: {1}{B}
//
// Oracle text:
//
//	Protection from Elves
var NathSBuffoon = newNathSBuffoon

func newNathSBuffoon() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Nath's Buffoon",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.B,
			}),
			Colors:    []color.Color{color.Black},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Goblin, types.Rogue},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 1}),
			StaticAbilities: []game.StaticAbility{
				game.ProtectionFromSubtypesStaticAbility(types.Elf),
			},
			OracleText: `
			Protection from Elves
		`,
		},
	}
}
