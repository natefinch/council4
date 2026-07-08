package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// ServantOfTheScale is the card definition for Servant of the Scale.
//
// Type: Creature — Human Soldier
// Cost: {G}
//
// Oracle text:
//
//	This creature enters with a +1/+1 counter on it.
//	When this creature dies, put X +1/+1 counters on target creature you control, where X is the number of +1/+1 counters on this creature.
var ServantOfTheScale = newServantOfTheScale

func newServantOfTheScale() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Servant of the Scale",
			ManaCost: opt.Val(cost.Mana{
				cost.G,
			}),
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Human, types.Soldier},
			Power:     opt.Val(game.PT{Value: 0}),
			Toughness: opt.Val(game.PT{Value: 0}),
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhen,
						Pattern: game.TriggerPattern{
							Event:            game.EventPermanentDied,
							Source:           game.TriggerSourceSelf,
							SubjectSelection: game.Selection{RequiredTypes: []types.Card{types.Creature}},
						},
					},
					Content: game.Mode{
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
									Amount: game.Dynamic(game.DynamicAmount{
										Kind:        game.DynamicAmountObjectCounters,
										Multiplier:  1,
										CounterKind: counter.PlusOnePlusOne,
										Object:      game.SourcePermanentReference(),
									}),
									Object:      game.TargetPermanentReference(0),
									CounterKind: counter.PlusOnePlusOne,
								},
							},
						},
					}.Ability(),
				},
			},
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntersWithCountersReplacement("This creature enters with a +1/+1 counter on it.", game.CounterPlacement{Kind: counter.PlusOnePlusOne, Amount: 1}),
			},
			OracleText: `
			This creature enters with a +1/+1 counter on it.
			When this creature dies, put X +1/+1 counters on target creature you control, where X is the number of +1/+1 counters on this creature.
		`,
		},
	}
}
