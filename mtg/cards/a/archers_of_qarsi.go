package a

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// ArchersOfQarsi is the card definition for Archers of Qarsi.
//
// Type: Creature — Snake Archer
// Cost: {3}{G}
//
// Oracle text:
//
//	Defender
//	Reach (This creature can block creatures with flying.)
var ArchersOfQarsi = newArchersOfQarsi

func newArchersOfQarsi() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Archers of Qarsi",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.G,
			}),
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Snake, types.Archer},
			Power:     opt.Val(game.PT{Value: 5}),
			Toughness: opt.Val(game.PT{Value: 2}),
			StaticAbilities: []game.StaticAbility{
				game.DefenderStaticBody,
				game.ReachStaticBody,
			},
			OracleText: `
			Defender
			Reach (This creature can block creatures with flying.)
		`,
		},
	}
}
