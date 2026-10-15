package t

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// TelJiladArchers is the card definition for Tel-Jilad Archers.
//
// Type: Creature — Elf Archer
// Cost: {4}{G}
//
// Oracle text:
//
//	Reach (This creature can block creatures with flying.)
//	Protection from artifacts
var TelJiladArchers = newTelJiladArchers

func newTelJiladArchers() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Tel-Jilad Archers",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
				cost.G,
			}),
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Elf, types.Archer},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 4}),
			StaticAbilities: []game.StaticAbility{
				game.ReachStaticBody,
				game.ProtectionFromTypesStaticAbility(types.Artifact),
			},
			OracleText: `
			Reach (This creature can block creatures with flying.)
			Protection from artifacts
		`,
		},
	}
}
