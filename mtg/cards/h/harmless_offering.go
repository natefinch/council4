package h

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// HarmlessOffering is the card definition for Harmless Offering.
//
// Type: Sorcery
// Cost: {2}{R}
//
// Oracle text:
//
//	Target opponent gains control of target permanent you control.
var HarmlessOffering = newHarmlessOffering()

func newHarmlessOffering() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Harmless Offering",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.R,
			}),
			Colors: []color.Color{color.Red},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.Mode{
				Targets: []game.TargetSpec{
					game.TargetSpec{
						MinTargets: 1,
						MaxTargets: 1,
						Constraint: "Target opponent",
						Allow:      game.TargetAllowPlayer,
						Selection:  opt.Val(game.Selection{Player: game.PlayerOpponent}),
					},
					game.TargetSpec{
						MinTargets: 1,
						MaxTargets: 1,
						Constraint: "target permanent you control",
						Allow:      game.TargetAllowPermanent,
						Selection:  opt.Val(game.Selection{Controller: game.ControllerYou}),
					},
				},
				Sequence: []game.Instruction{
					{
						Primitive: game.ApplyContinuous{
							Object: opt.Val(game.TargetPermanentReference(1)),
							ContinuousEffects: []game.ContinuousEffect{
								game.ContinuousEffect{
									Layer:            game.LayerControl,
									NewControllerRef: opt.Val(game.TargetPlayerReference(0)),
								},
							},
							Duration: game.DurationPermanent,
						},
					},
				},
			}.Ability()),
			OracleText: `
			Target opponent gains control of target permanent you control.
		`,
		},
	}
}
