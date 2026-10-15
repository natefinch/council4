package g

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Gridlock is the card definition for Gridlock.
//
// Type: Instant
// Cost: {X}{U}
//
// Oracle text:
//
//	Tap X target nonland permanents.
var Gridlock = newGridlock

func newGridlock() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Gridlock",
			ManaCost: opt.Val(cost.Mana{
				cost.X,
				cost.U,
			}),
			Colors: []color.Color{color.Blue},
			Types:  []types.Card{types.Instant},
			SpellAbility: opt.Val(game.Mode{
				Targets: []game.TargetSpec{
					game.TargetSpec{
						MinTargets:   0,
						MaxTargets:   20,
						Constraint:   "x target nonland permanents",
						Allow:        game.TargetAllowPermanent,
						Selection:    opt.Val(game.Selection{ExcludedTypes: []types.Card{types.Land}}),
						CountEqualsX: true,
					},
				},
				Sequence: []game.Instruction{
					{
						Primitive: game.Tap{
							Object: game.AllTargetPermanentsReference(0),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Tap X target nonland permanents.
		`,
		},
	}
}
