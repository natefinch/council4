package a

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// ActiveVolcano is the card definition for Active Volcano.
//
// Type: Instant
// Cost: {R}
//
// Oracle text:
//
//	Choose one —
//	• Destroy target blue permanent.
//	• Return target Island to its owner's hand.
var ActiveVolcano = newActiveVolcano

func newActiveVolcano() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Active Volcano",
			ManaCost: opt.Val(cost.Mana{
				cost.R,
			}),
			Colors: []color.Color{color.Red},
			Types:  []types.Card{types.Instant},
			SpellAbility: opt.Val(game.AbilityContent{
				Modes: []game.Mode{
					game.Mode{
						Text: "Destroy target blue permanent.",
						Targets: []game.TargetSpec{
							game.TargetSpec{
								MinTargets: 1,
								MaxTargets: 1,
								Constraint: "target blue permanent",
								Allow:      game.TargetAllowPermanent,
								Selection:  opt.Val(game.Selection{ColorsAny: []color.Color{color.Blue}}),
							},
						},
						Sequence: []game.Instruction{
							{
								Primitive: game.Destroy{
									Object: game.TargetPermanentReference(0),
								},
							},
						},
					},
					game.Mode{
						Text: "Return target Island to its owner's hand.",
						Targets: []game.TargetSpec{
							game.TargetSpec{
								MinTargets: 1,
								MaxTargets: 1,
								Constraint: "target Island",
								Allow:      game.TargetAllowPermanent,
								Selection:  opt.Val(game.Selection{SubtypesAny: []types.Sub{types.Sub("Island")}}),
							},
						},
						Sequence: []game.Instruction{
							{
								Primitive: game.MovePermanent{
									Object:      game.TargetPermanentReference(0),
									Destination: zone.Hand,
								},
							},
						},
					},
				},
				MinModes: 1,
				MaxModes: 1,
			}),
			OracleText: `
			Choose one —
			• Destroy target blue permanent.
			• Return target Island to its owner's hand.
		`,
		},
	}
}
