package t

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// TelJiladOutrider is the card definition for Tel-Jilad Outrider.
//
// Type: Creature — Elf Warrior
// Cost: {3}{G}
//
// Oracle text:
//
//	Protection from artifacts
var TelJiladOutrider = newTelJiladOutrider

func newTelJiladOutrider() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Tel-Jilad Outrider",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.G,
			}),
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Elf, types.Warrior},
			Power:     opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 1}),
			StaticAbilities: []game.StaticAbility{
				game.ProtectionFromTypesStaticAbility(types.Artifact),
			},
			OracleText: `
			Protection from artifacts
		`,
		},
	}
}
