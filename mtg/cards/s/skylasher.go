package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Skylasher is the card definition for Skylasher.
//
// Type: Creature — Insect
// Cost: {1}{G}
//
// Oracle text:
//
//	Flash
//	This spell can't be countered.
//	Reach, protection from blue
var Skylasher = newSkylasher

func newSkylasher() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Skylasher",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.G,
			}),
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Insect},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 2}),
			StaticAbilities: []game.StaticAbility{
				game.FlashStaticBody,
				game.CantBeCounteredStaticBody,
				game.ReachStaticBody,
				game.ProtectionFromColorsStaticAbility(color.Blue),
			},
			OracleText: `
			Flash
			This spell can't be countered.
			Reach, protection from blue
		`,
		},
	}
}
