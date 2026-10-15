package k

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// KinjalliSSunwing is the card definition for Kinjalli's Sunwing.
//
// Type: Creature — Dinosaur
// Cost: {2}{W}
//
// Oracle text:
//
//	Flying
//	Creatures your opponents control enter tapped.
var KinjalliSSunwing = newKinjalliSSunwing

func newKinjalliSSunwing() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Kinjalli's Sunwing",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.W,
			}),
			Colors:    []color.Color{color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Dinosaur},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 3}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
			},
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntersTappedGroupReplacement("Creatures your opponents control enter tapped.", game.TriggerControllerOpponent, types.Creature),
			},
			OracleText: `
			Flying
			Creatures your opponents control enter tapped.
		`,
		},
	}
}
