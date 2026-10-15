package w

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// WitherscaleWurm is the card definition for Witherscale Wurm.
//
// Type: Creature — Wurm
// Cost: {4}{G}{G}
//
// Oracle text:
//
//	Whenever this creature blocks or becomes blocked by a creature, that creature gains wither until end of turn. (It deals damage to creatures in the form of -1/-1 counters.)
//	Whenever this creature deals damage to an opponent, remove all -1/-1 counters from it.
var WitherscaleWurm = newWitherscaleWurm

func newWitherscaleWurm() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Witherscale Wurm",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
				cost.G,
				cost.G,
			}),
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Wurm},
			Power:     opt.Val(game.PT{Value: 9}),
			Toughness: opt.Val(game.PT{Value: 9}),
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Text: "Whenever this creature blocks or becomes blocked by a creature, that creature gains wither until end of turn. (It deals damage to creatures in the form of -1/-1 counters.)",
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:                   game.EventBlockerDeclared,
							Source:                  game.TriggerSourceSelf,
							RelatedSubjectSelection: game.Selection{RequiredTypes: []types.Card{types.Creature}},
							StackObjectKind:         game.StackSpell,
							CounterKind:             counter.PlusOnePlusOne,
							UnionEvent:              game.EventAttackerBecameBlocked,
						},
					},
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.ApplyContinuous{
											Object: opt.Val(game.EventRelatedPermanentReference()),
											ContinuousEffects: []game.ContinuousEffect{
												game.ContinuousEffect{
													Duration:    game.DurationPermanent,
													Layer:       game.LayerAbility,
													AddKeywords: []game.Keyword{game.Wither},
												},
											},
											Duration: game.DurationUntilEndOfTurn,
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
					Text: "Whenever this creature deals damage to an opponent, remove all -1/-1 counters from it.",
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:           game.EventDamageDealt,
							Source:          game.TriggerSourceSelf,
							Player:          game.TriggerPlayerOpponent,
							Subject:         game.TriggerSubjectDamageSource,
							StackObjectKind: game.StackSpell,
							DamageRecipient: game.DamageRecipientPlayer,
							CounterKind:     counter.PlusOnePlusOne,
						},
					},
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.RemoveCounter{
											Amount: game.Dynamic(game.DynamicAmount{
												Kind:        game.DynamicAmountObjectCounters,
												CounterKind: counter.MinusOneMinusOne,
												Object:      game.EventPermanentReference(),
											}),
											Object:      game.EventPermanentReference(),
											CounterKind: counter.MinusOneMinusOne,
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
			Whenever this creature blocks or becomes blocked by a creature, that creature gains wither until end of turn. (It deals damage to creatures in the form of -1/-1 counters.)
			Whenever this creature deals damage to an opponent, remove all -1/-1 counters from it.
		`,
		},
	}
}
