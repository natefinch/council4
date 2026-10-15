package p

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// ParishBladeTrainee is the card definition for Parish-Blade Trainee.
//
// Type: Creature — Human Soldier
// Cost: {1}{W}
//
// Oracle text:
//
//	Training (Whenever this creature attacks with another creature with greater power, put a +1/+1 counter on this creature.)
//	When this creature dies, put its counters on target creature you control.
var ParishBladeTrainee = newParishBladeTrainee

func newParishBladeTrainee() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Parish-Blade Trainee",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.W,
			}),
			Colors:    []color.Color{color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Human, types.Soldier},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 2}),
			TriggeredAbilities: []game.TriggeredAbility{
				game.TrainingTriggeredBody,
				game.TriggeredAbility{
					Text: "When this creature dies, put its counters on target creature you control.",
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhen,
						Pattern: game.TriggerPattern{
							Event:            game.EventPermanentDied,
							Source:           game.TriggerSourceSelf,
							SubjectSelection: game.Selection{RequiredTypes: []types.Card{types.Creature}},
							StackObjectKind:  game.StackSpell,
							CounterKind:      counter.PlusOnePlusOne,
						},
					},
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
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
									game.Instruction{
										Primitive: game.MoveCounters{
											Object:      game.TargetPermanentReference(0),
											CounterKind: counter.PlusOnePlusOne,
											Source: game.CounterSourceSpec{
												Kind: game.CounterSourceEventPermanent,
											},
											AllKinds: true,
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
			Training (Whenever this creature attacks with another creature with greater power, put a +1/+1 counter on this creature.)
			When this creature dies, put its counters on target creature you control.
		`,
		},
	}
}
