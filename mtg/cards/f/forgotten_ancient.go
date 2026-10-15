package f

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// ForgottenAncient is the card definition for Forgotten Ancient.
//
// Type: Creature — Elemental
// Cost: {3}{G}
//
// Oracle text:
//
//	Whenever a player casts a spell, you may put a +1/+1 counter on this creature.
//	At the beginning of your upkeep, you may move any number of +1/+1 counters from this creature onto other creatures.
var ForgottenAncient = newForgottenAncient

func newForgottenAncient() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Forgotten Ancient",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.G,
			}),
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Elemental},
			Power:     opt.Val(game.PT{Value: 0}),
			Toughness: opt.Val(game.PT{Value: 3}),
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Text: "Whenever a player casts a spell, you may put a +1/+1 counter on this creature.",
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:           game.EventSpellCast,
							StackObjectKind: game.StackSpell,
							CounterKind:     counter.PlusOnePlusOne,
						},
					},
					Optional: true,
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.AddCounter{
											Amount:      game.Fixed(1),
											Object:      game.SourcePermanentReference(),
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
				game.TriggeredAbility{
					Text: "At the beginning of your upkeep, you may move any number of +1/+1 counters from this creature onto other creatures.",
					Trigger: game.TriggerCondition{
						Type: game.TriggerAt,
						Pattern: game.TriggerPattern{
							Event:           game.EventBeginningOfStep,
							Controller:      game.TriggerControllerYou,
							StackObjectKind: game.StackSpell,
							CounterKind:     counter.PlusOnePlusOne,
							Step:            game.StepUpkeep,
						},
					},
					Optional: true,
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.MoveCounters{
											CounterKind: counter.PlusOnePlusOne,
											Source: game.CounterSourceSpec{
												Kind: game.CounterSourceSelf,
											},
											Group: func() *game.GroupReference {
												ref := game.BattlefieldGroupExcluding(game.Selection{RequiredTypes: []types.Card{types.Creature}, ExcludeSource: true}, game.SourcePermanentReference())
												return &ref
											}(),
											Distribute: true,
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
			Whenever a player casts a spell, you may put a +1/+1 counter on this creature.
			At the beginning of your upkeep, you may move any number of +1/+1 counters from this creature onto other creatures.
		`,
		},
	}
}
