package d

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// DevouringDeep is the card definition for Devouring Deep.
//
// Type: Creature — Fish
// Cost: {2}{U}
//
// Oracle text:
//
//	Islandwalk (This creature can't be blocked as long as defending player controls an Island.)
var DevouringDeep = newDevouringDeep

func newDevouringDeep() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Devouring Deep",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.U,
			}),
			Colors:    []color.Color{color.Blue},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Fish},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 2}),
			StaticAbilities: []game.StaticAbility{
				game.IslandwalkStaticBody,
			},
			OracleText: `
			Islandwalk (This creature can't be blocked as long as defending player controls an Island.)
		`,
		},
	}
}
