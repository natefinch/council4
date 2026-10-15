package p

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// PiranhaFly is the card definition for Piranha Fly.
//
// Type: Creature — Fish Insect
// Cost: {1}{U}
//
// Oracle text:
//
//	Flying
//	This creature enters tapped.
var PiranhaFly = newPiranhaFly

func newPiranhaFly() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Piranha Fly",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.U,
			}),
			Colors:    []color.Color{color.Blue},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Fish, types.Insect},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 1}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
			},
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntersTappedReplacement("This creature enters tapped."),
			},
			OracleText: `
			Flying
			This creature enters tapped.
		`,
		},
	}
}
