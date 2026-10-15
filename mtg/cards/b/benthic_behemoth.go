package b

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// BenthicBehemoth is the card definition for Benthic Behemoth.
//
// Type: Creature — Serpent
// Cost: {5}{U}{U}{U}
//
// Oracle text:
//
//	Islandwalk (This creature can't be blocked as long as defending player controls an Island.)
var BenthicBehemoth = newBenthicBehemoth

func newBenthicBehemoth() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Benthic Behemoth",
			ManaCost: opt.Val(cost.Mana{
				cost.O(5),
				cost.U,
				cost.U,
				cost.U,
			}),
			Colors:    []color.Color{color.Blue},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Serpent},
			Power:     opt.Val(game.PT{Value: 7}),
			Toughness: opt.Val(game.PT{Value: 6}),
			StaticAbilities: []game.StaticAbility{
				game.IslandwalkStaticBody,
			},
			OracleText: `
			Islandwalk (This creature can't be blocked as long as defending player controls an Island.)
		`,
		},
	}
}
