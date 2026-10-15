package r

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// ResistanceSkywarden is the card definition for Resistance Skywarden.
//
// Type: Creature — Ogre Rebel
// Cost: {3}{R}{R}
//
// Oracle text:
//
//	Reach (This creature can block creatures with flying.)
//	Menace (This creature can't be blocked except by two or more creatures.)
var ResistanceSkywarden = newResistanceSkywarden

func newResistanceSkywarden() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Resistance Skywarden",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.R,
				cost.R,
			}),
			Colors:    []color.Color{color.Red},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Ogre, types.Rebel},
			Power:     opt.Val(game.PT{Value: 5}),
			Toughness: opt.Val(game.PT{Value: 5}),
			StaticAbilities: []game.StaticAbility{
				game.ReachStaticBody,
				game.MenaceStaticBody,
			},
			OracleText: `
			Reach (This creature can block creatures with flying.)
			Menace (This creature can't be blocked except by two or more creatures.)
		`,
		},
	}
}
