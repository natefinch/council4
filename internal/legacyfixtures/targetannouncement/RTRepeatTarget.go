// Legacy regression fixture preserved when the Oracle compiler was retired.
package fixtures

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// RTRepeatTarget is the card definition for RT Repeat Target.
//
// Type: Artifact
//
// Oracle text:
//
//	{0}: Target creature gets +1/+1 until end of turn.
var RTRepeatTarget = newRTRepeatTarget

func newRTRepeatTarget() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:  "RT Repeat Target",
			Types: []types.Card{types.Artifact},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "{0}: Target creature gets +1/+1 until end of turn.",
					ManaCost: opt.Val(cost.Mana{
						cost.O(0),
					}),
					ZoneOfFunction: zone.Battlefield,
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Targets: []game.TargetSpec{
									game.TargetSpec{
										MinTargets: 1,
										MaxTargets: 1,
										Constraint: "target creature",
										Allow:      game.TargetAllowPermanent,
										Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}}),
									},
								},
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.ModifyPT{
											Object:         game.TargetPermanentReference(0),
											PowerDelta:     game.Fixed(1),
											ToughnessDelta: game.Fixed(1),
											Duration:       game.DurationUntilEndOfTurn,
										},
									},
								},
							},
						},
						MinModes: 1,
						MaxModes: 1,
					},
				},
			},
			OracleText: `
			{0}: Target creature gets +1/+1 until end of turn.
		`,
		},
	}
}
