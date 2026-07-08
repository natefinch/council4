package a

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// ApothecaryStomper is the card definition for Apothecary Stomper.
//
// Type: Creature — Elephant
// Cost: {4}{G}{G}
//
// Oracle text:
//
//	Vigilance (Attacking doesn't cause this creature to tap.)
//	When this creature enters, choose one —
//	• Put two +1/+1 counters on target creature you control.
//	• You gain 4 life.
var ApothecaryStomper = newApothecaryStomper

func newApothecaryStomper() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Apothecary Stomper",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
				cost.G,
				cost.G,
			}),
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Elephant},
			Power:     opt.Val(game.PT{Value: 4}),
			Toughness: opt.Val(game.PT{Value: 4}),
			StaticAbilities: []game.StaticAbility{
				game.VigilanceStaticBody,
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhen,
						Pattern: game.TriggerPattern{
							Event:  game.EventPermanentEnteredBattlefield,
							Source: game.TriggerSourceSelf,
						},
					},
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Text: "Put two +1/+1 counters on target creature you control.",
								Targets: []game.TargetSpec{
									game.TargetSpec{
										MinTargets: 1,
										MaxTargets: 1,
										Constraint: "target creature you control",
										Allow:      game.TargetAllowPermanent,
										Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}, Controller: game.ControllerYou}),
									},
								},
								Sequence: []game.Instruction{
									{
										Primitive: game.AddCounter{
											Amount:      game.Fixed(2),
											Object:      game.TargetPermanentReference(0),
											CounterKind: counter.PlusOnePlusOne,
										},
									},
								},
							},
							game.Mode{
								Text: "You gain 4 life.",
								Sequence: []game.Instruction{
									{
										Primitive: game.GainLife{
											Amount: game.Fixed(4),
											Player: game.ControllerReference(),
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
			Vigilance (Attacking doesn't cause this creature to tap.)
			When this creature enters, choose one —
			• Put two +1/+1 counters on target creature you control.
			• You gain 4 life.
		`,
		},
	}
}
