package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// ShanodinDryads is the card definition for Shanodin Dryads.
//
// Type: Creature — Nymph Dryad
// Cost: {G}
//
// Oracle text:
//
//	Forestwalk (This creature can't be blocked as long as defending player controls a Forest.)
var ShanodinDryads = newShanodinDryads

func newShanodinDryads() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Shanodin Dryads",
			ManaCost: opt.Val(cost.Mana{
				cost.G,
			}),
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Nymph, types.Dryad},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 1}),
			StaticAbilities: []game.StaticAbility{
				game.ForestwalkStaticBody,
			},
			OracleText: `
			Forestwalk (This creature can't be blocked as long as defending player controls a Forest.)
		`,
		},
	}
}
