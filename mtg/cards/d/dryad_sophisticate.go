package d

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// DryadSophisticate is the card definition for Dryad Sophisticate.
//
// Type: Creature — Dryad
// Cost: {1}{G}
//
// Oracle text:
//
//	Nonbasic landwalk (This creature can't be blocked as long as defending player controls a nonbasic land.)
var DryadSophisticate = newDryadSophisticate

func newDryadSophisticate() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Dryad Sophisticate",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.G,
			}),
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Dryad},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 1}),
			StaticAbilities: []game.StaticAbility{
				game.NonbasicLandwalkStaticBody,
			},
			OracleText: `
			Nonbasic landwalk (This creature can't be blocked as long as defending player controls a nonbasic land.)
		`,
		},
	}
}
