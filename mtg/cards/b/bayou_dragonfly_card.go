package b

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// BayouDragonfly is the card definition for Bayou Dragonfly.
//
// Type: Creature — Insect
// Cost: {1}{G}
//
// Oracle text:
//
//	Flying; swampwalk (This creature can't be blocked as long as defending player controls a Swamp.)
var BayouDragonfly = newBayouDragonfly

func newBayouDragonfly() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Bayou Dragonfly",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.G,
			}),
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Insect},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 1}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
				game.SwampwalkStaticBody,
			},
			OracleText: `
			Flying; swampwalk (This creature can't be blocked as long as defending player controls a Swamp.)
		`,
		},
	}
}
