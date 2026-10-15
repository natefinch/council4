package k

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// KeenEyedArchers is the card definition for Keen-Eyed Archers.
//
// Type: Creature — Elf Archer
// Cost: {2}{W}
//
// Oracle text:
//
//	Reach (This creature can block creatures with flying.)
var KeenEyedArchers = newKeenEyedArchers

func newKeenEyedArchers() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Keen-Eyed Archers",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.W,
			}),
			Colors:    []color.Color{color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Elf, types.Archer},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 2}),
			StaticAbilities: []game.StaticAbility{
				game.ReachStaticBody,
			},
			OracleText: `
			Reach (This creature can block creatures with flying.)
		`,
		},
	}
}
