package m

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// MirranCrusader is the card definition for Mirran Crusader.
//
// Type: Creature — Human Knight
// Cost: {1}{W}{W}
//
// Oracle text:
//
//	Double strike, protection from black and from green
var MirranCrusader = newMirranCrusader

func newMirranCrusader() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Mirran Crusader",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.W,
				cost.W,
			}),
			Colors:    []color.Color{color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Human, types.Knight},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 2}),
			StaticAbilities: []game.StaticAbility{
				game.DoubleStrikeStaticBody,
				game.ProtectionFromColorsStaticAbility(color.Black, color.Green),
			},
			OracleText: `
			Double strike, protection from black and from green
		`,
		},
	}
}
