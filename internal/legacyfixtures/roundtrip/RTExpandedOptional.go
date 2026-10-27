// Legacy regression fixture preserved when the Oracle compiler was retired.
package fixtures

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// RTExpandedOptional is the card definition for RT Expanded Optional.
//
// Type: Artifact
//
// Oracle text:
//
//	{1}: You may put a +1/+1 counter on each of up to two target creatures.
var RTExpandedOptional = newRTExpandedOptional

func newRTExpandedOptional() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:  "RT Expanded Optional",
			Types: []types.Card{types.Artifact},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "{1}: You may put a +1/+1 counter on each of up to two target creatures.",
					ManaCost: opt.Val(cost.Mana{
						cost.O(1),
					}),
					ZoneOfFunction: zone.Battlefield,
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Targets: []game.TargetSpec{
									game.TargetSpec{
										MaxTargets: 2,
										Constraint: "up to two target creatures",
										Allow:      game.TargetAllowPermanent,
										Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}}),
									},
								},
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.AddCounter{
											Amount:      game.Fixed(1),
											Object:      game.TargetPermanentReference(0),
											CounterKind: counter.PlusOnePlusOne,
										},
										Optional:                true,
										PublishOptionalDecision: game.OptionalDecisionKey("optional-clause-1"),
									},
									game.Instruction{
										Primitive: game.AddCounter{
											Amount:      game.Fixed(1),
											Object:      game.TargetPermanentReference(1),
											CounterKind: counter.PlusOnePlusOne,
										},
										OptionalDecisionGate: game.OptionalDecisionKey("optional-clause-1"),
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
			{1}: You may put a +1/+1 counter on each of up to two target creatures.
		`,
		},
	}
}
