package w

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// WoodlotCrawler is the card definition for Woodlot Crawler.
//
// Type: Creature — Insect
// Cost: {U}{B}
//
// Oracle text:
//
//	Forestwalk (This creature can't be blocked as long as defending player controls a Forest.)
//	Protection from green
var WoodlotCrawler = newWoodlotCrawler

func newWoodlotCrawler() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue, color.Black),
		CardFace: game.CardFace{
			Name: "Woodlot Crawler",
			ManaCost: opt.Val(cost.Mana{
				cost.U,
				cost.B,
			}),
			Colors:    []color.Color{color.Black, color.Blue},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Insect},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 1}),
			StaticAbilities: []game.StaticAbility{
				game.ForestwalkStaticBody,
				game.ProtectionFromColorsStaticAbility(color.Green),
			},
			OracleText: `
			Forestwalk (This creature can't be blocked as long as defending player controls a Forest.)
			Protection from green
		`,
		},
	}
}
