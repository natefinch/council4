package p

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// PowerConduit is the card definition for Power Conduit.
//
// Type: Artifact
// Cost: {2}
//
// Oracle text:
//
//	{T}, Remove a counter from a permanent you control: Choose one —
//	• Put a charge counter on target artifact.
//	• Put a +1/+1 counter on target creature.
var PowerConduit = newPowerConduit

func newPowerConduit() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Power Conduit",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
			}),
			Types: []types.Card{types.Artifact},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "{T}, Remove a counter from a permanent you control: Choose one —\n• Put a charge counter on target artifact.\n• Put a +1/+1 counter on target creature.",
					AdditionalCosts: []cost.Additional{
						cost.Additional{
							Kind:        cost.AdditionalTap,
							CounterKind: counter.PlusOnePlusOne,
						},
						cost.Additional{
							Kind:           cost.AdditionalRemoveCounterAmong,
							Text:           "Remove a counter from a permanent you control",
							Amount:         1,
							CounterKind:    counter.PlusOnePlusOne,
							AnyCounterKind: true,
						},
					},
					ZoneOfFunction: zone.Battlefield,
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Text: "Put a charge counter on target artifact.",
								Targets: []game.TargetSpec{
									game.TargetSpec{
										MinTargets: 1,
										MaxTargets: 1,
										Constraint: "target artifact",
										Allow:      game.TargetAllowPermanent,
										Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Artifact}}),
									},
								},
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.AddCounter{
											Amount:      game.Fixed(1),
											Object:      game.TargetPermanentReference(0),
											CounterKind: counter.Charge,
										},
									},
								},
							},
							game.Mode{
								Text: "Put a +1/+1 counter on target creature.",
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
										Primitive: game.AddCounter{
											Amount:      game.Fixed(1),
											Object:      game.TargetPermanentReference(0),
											CounterKind: counter.PlusOnePlusOne,
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
			{T}, Remove a counter from a permanent you control: Choose one —
			• Put a charge counter on target artifact.
			• Put a +1/+1 counter on target creature.
		`,
		},
	}
}
