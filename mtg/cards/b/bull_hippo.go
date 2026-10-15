package b

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// BullHippo is the card definition for Bull Hippo.
//
// Type: Creature — Hippo
// Cost: {3}{G}
//
// Oracle text:
//
//	Islandwalk (This creature can't be blocked as long as defending player controls an Island.)
var BullHippo = newBullHippo

func newBullHippo() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Bull Hippo",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.G,
			}),
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Hippo},
			Power:     opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 3}),
			StaticAbilities: []game.StaticAbility{
				game.IslandwalkStaticBody,
			},
			OracleText: `
			Islandwalk (This creature can't be blocked as long as defending player controls an Island.)
		`,
		},
	}
}
