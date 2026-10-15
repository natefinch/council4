package d

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// DistortingWake is the card definition for Distorting Wake.
//
// Type: Sorcery
// Cost: {X}{U}{U}{U}
//
// Oracle text:
//
//	Return X target nonland permanents to their owners' hands.
var DistortingWake = newDistortingWake

func newDistortingWake() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Distorting Wake",
			ManaCost: opt.Val(cost.Mana{
				cost.X,
				cost.U,
				cost.U,
				cost.U,
			}),
			Colors: []color.Color{color.Blue},
			Types:  []types.Card{types.Sorcery},
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
						Primitive: game.MovePermanent{
							Object:      game.AllTargetPermanentsReference(0),
							Destination: zone.Hand,
						},
					},
				},
			}.Ability()),
			OracleText: `
			Return X target nonland permanents to their owners' hands.
		`,
		},
	}
}
