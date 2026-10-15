package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// SpinedThopter is the card definition for Spined Thopter.
//
// Type: Artifact Creature — Phyrexian Thopter
// Cost: {2}{U/P}
//
// Oracle text:
//
//	({U/P} can be paid with either {U} or 2 life.)
//	Flying
var SpinedThopter = newSpinedThopter

func newSpinedThopter() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Spined Thopter",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.PhyrexianMana(mana.U),
			}),
			Colors:    []color.Color{color.Blue},
			Types:     []types.Card{types.Artifact, types.Creature},
			Subtypes:  []types.Sub{types.Phyrexian, types.Thopter},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 1}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
			},
			OracleText: `
			({U/P} can be paid with either {U} or 2 life.)
			Flying
		`,
		},
	}
}
