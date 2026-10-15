package g

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// GalewindMoose is the card definition for Galewind Moose.
//
// Type: Creature — Elemental Elk
// Cost: {4}{G}{G}
//
// Oracle text:
//
//	Flash
//	Reach, vigilance, trample
var GalewindMoose = newGalewindMoose

func newGalewindMoose() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Galewind Moose",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
				cost.G,
				cost.G,
			}),
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Elemental, types.Elk},
			Power:     opt.Val(game.PT{Value: 6}),
			Toughness: opt.Val(game.PT{Value: 6}),
			StaticAbilities: []game.StaticAbility{
				game.FlashStaticBody,
				game.ReachStaticBody,
				game.VigilanceStaticBody,
				game.TrampleStaticBody,
			},
			OracleText: `
			Flash
			Reach, vigilance, trample
		`,
		},
	}
}
