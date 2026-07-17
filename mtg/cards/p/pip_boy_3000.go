package p

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// PipBoy3000 is the card definition for Pip-Boy 3000.
//
// Type: Artifact — Equipment
// Cost: {1}
//
// Oracle text:
//
//	Whenever equipped creature attacks, choose one —
//	• Sort Inventory — Draw a card, then discard a card.
//	• Pick a Perk — Put a +1/+1 counter on that creature.
//	• Check Map — Untap up to two target lands.
//	Equip {2}
var PipBoy3000 = newPipBoy3000

func newPipBoy3000() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Pip-Boy 3000",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
			}),
			Types:    []types.Card{types.Artifact},
			Subtypes: []types.Sub{types.Equipment},
			ActivatedAbilities: []game.ActivatedAbility{
				game.EquipActivatedAbility(cost.Mana{cost.O(2)}),
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:            game.EventAttackerDeclared,
							Source:           game.TriggerSourceAttachedPermanent,
							SubjectSelection: game.Selection{RequiredTypes: []types.Card{types.Creature}},
						},
					},
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Text: "Sort Inventory — Draw a card, then discard a card.",
								Sequence: []game.Instruction{
									{
										Primitive: game.Draw{
											Amount: game.Fixed(1),
											Player: game.ControllerReference(),
										},
									},
									{
										Primitive: game.Discard{
											Amount: game.Fixed(1),
											Player: game.ControllerReference(),
										},
									},
								},
							},
							game.Mode{
								Text: "Pick a Perk — Put a +1/+1 counter on that creature.",
								Sequence: []game.Instruction{
									{
										Primitive: game.AddCounter{
											Amount:      game.Fixed(1),
											Object:      game.EventPermanentReference(),
											CounterKind: counter.PlusOnePlusOne,
										},
									},
								},
							},
							game.Mode{
								Text: "Check Map — Untap up to two target lands.",
								Targets: []game.TargetSpec{
									game.TargetSpec{
										MinTargets: 0,
										MaxTargets: 2,
										Constraint: "up to two target lands",
										Allow:      game.TargetAllowPermanent,
										Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Land}}),
									},
								},
								Sequence: []game.Instruction{
									{
										Primitive: game.Untap{
											Object: game.TargetPermanentReference(0),
										},
									},
									{
										Primitive: game.Untap{
											Object: game.TargetPermanentReference(1),
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
			Whenever equipped creature attacks, choose one —
			• Sort Inventory — Draw a card, then discard a card.
			• Pick a Perk — Put a +1/+1 counter on that creature.
			• Check Map — Untap up to two target lands.
			Equip {2}
		`,
		},
	}
}
