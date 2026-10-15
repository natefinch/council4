package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// SigiledBehemoth is the card definition for Sigiled Behemoth.
//
// Type: Creature — Beast
// Cost: {4}{G}{W}
//
// Oracle text:
//
//	Exalted (Whenever a creature you control attacks alone, that creature gets +1/+1 until end of turn.)
var SigiledBehemoth = newSigiledBehemoth

func newSigiledBehemoth() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Green),
		CardFace: game.CardFace{
			Name: "Sigiled Behemoth",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
				cost.G,
				cost.W,
			}),
			Colors:    []color.Color{color.Green, color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Beast},
			Power:     opt.Val(game.PT{Value: 5}),
			Toughness: opt.Val(game.PT{Value: 4}),
			StaticAbilities: []game.StaticAbility{
				game.ExaltedStaticBody,
			},
			OracleText: `
			Exalted (Whenever a creature you control attacks alone, that creature gets +1/+1 until end of turn.)
		`,
		},
	}
}
