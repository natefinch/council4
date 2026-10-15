package c

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// CourtArchers is the card definition for Court Archers.
//
// Type: Creature — Human Archer
// Cost: {2}{G}
//
// Oracle text:
//
//	Reach (This creature can block creatures with flying.)
//	Exalted (Whenever a creature you control attacks alone, that creature gets +1/+1 until end of turn.)
var CourtArchers = newCourtArchers

func newCourtArchers() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Court Archers",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.G,
			}),
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Human, types.Archer},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 3}),
			StaticAbilities: []game.StaticAbility{
				game.ReachStaticBody,
				game.ExaltedStaticBody,
			},
			OracleText: `
			Reach (This creature can block creatures with flying.)
			Exalted (Whenever a creature you control attacks alone, that creature gets +1/+1 until end of turn.)
		`,
		},
	}
}
