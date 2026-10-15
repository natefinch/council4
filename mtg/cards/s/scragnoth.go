package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Scragnoth is the card definition for Scragnoth.
//
// Type: Creature — Beast
// Cost: {4}{G}
//
// Oracle text:
//
//	This spell can't be countered.
//	Protection from blue
var Scragnoth = newScragnoth

func newScragnoth() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Scragnoth",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
				cost.G,
			}),
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Beast},
			Power:     opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 4}),
			StaticAbilities: []game.StaticAbility{
				game.CantBeCounteredStaticBody,
				game.ProtectionFromColorsStaticAbility(color.Blue),
			},
			OracleText: `
			This spell can't be countered.
			Protection from blue
		`,
		},
	}
}
