package g

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// GlimpseTheSunGod is the card definition for Glimpse the Sun God.
//
// Type: Instant
// Cost: {X}{W}
//
// Oracle text:
//
//	Tap X target creatures. Scry 1. (Look at the top card of your library. You may put that card on the bottom.)
var GlimpseTheSunGod = newGlimpseTheSunGod

func newGlimpseTheSunGod() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Glimpse the Sun God",
			ManaCost: opt.Val(cost.Mana{
				cost.X,
				cost.W,
			}),
			Colors: []color.Color{color.White},
			Types:  []types.Card{types.Instant},
			SpellAbility: opt.Val(game.Mode{
				Targets: []game.TargetSpec{
					game.TargetSpec{
						MinTargets:   0,
						MaxTargets:   20,
						Constraint:   "x target creatures",
						Allow:        game.TargetAllowPermanent,
						Selection:    opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}}),
						CountEqualsX: true,
					},
				},
				Sequence: []game.Instruction{
					{
						Primitive: game.Tap{
							Object: game.AllTargetPermanentsReference(0),
						},
					},
					{
						Primitive: game.Scry{
							Amount: game.Fixed(1),
							Player: game.ControllerReference(),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Tap X target creatures. Scry 1. (Look at the top card of your library. You may put that card on the bottom.)
		`,
		},
	}
}
