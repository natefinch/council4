package n

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// NishobaBrawler is the card definition for Nishoba Brawler.
//
// Type: Creature — Cat Warrior
// Cost: {1}{G}
//
// Oracle text:
//
//	Trample
//	Domain — Nishoba Brawler's power is equal to the number of basic land types among lands you control.
var NishobaBrawler = newNishobaBrawler

func newNishobaBrawler() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Nishoba Brawler",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.G,
			}),
			Colors:       []color.Color{color.Green},
			Types:        []types.Card{types.Creature},
			Subtypes:     []types.Sub{types.Cat, types.Warrior},
			Power:        opt.Val(game.PT{IsStar: true}),
			Toughness:    opt.Val(game.PT{Value: 3}),
			DynamicPower: opt.Val(game.DynamicValue{Kind: game.DynamicValueControllerBasicLandTypeCount}),
			StaticAbilities: []game.StaticAbility{
				game.TrampleStaticBody,
			},
			OracleText: `
			Trample
			Domain — Nishoba Brawler's power is equal to the number of basic land types among lands you control.
		`,
		},
	}
}
