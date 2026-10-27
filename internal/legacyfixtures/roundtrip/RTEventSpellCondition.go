// Legacy regression fixture preserved when the Oracle compiler was retired.
package fixtures

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// RTEventSpellCondition is the card definition for RT Event Spell Condition.
//
// Type: Creature
//
// Oracle text:
//
//	Whenever you cast an instant or sorcery spell, you gain 1 life. If that spell has mana value 5 or greater, put a +1/+1 counter on this creature.
var RTEventSpellCondition = newRTEventSpellCondition

func newRTEventSpellCondition() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "RT Event Spell Condition",
			Types:     []types.Card{types.Creature},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 2}),
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Text: "Whenever you cast an instant or sorcery spell, you gain 1 life. If that spell has mana value 5 or greater, put a +1/+1 counter on this creature.",
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:           game.EventSpellCast,
							Controller:      game.TriggerControllerYou,
							CardSelection:   game.Selection{RequiredTypesAny: []types.Card{types.Instant, types.Sorcery}},
							StackObjectKind: game.StackSpell,
							CounterKind:     counter.PlusOnePlusOne,
						},
					},
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.GainLife{
											Amount: game.Fixed(1),
											Player: game.ControllerReference(),
										},
									},
									game.Instruction{
										Primitive: game.AddCounter{
											Amount:      game.Fixed(1),
											Object:      game.SourcePermanentReference(),
											CounterKind: counter.PlusOnePlusOne,
										},
										Condition: opt.Val(game.EffectCondition{
											Condition: opt.Val(game.Condition{
												Text:   "If that spell has mana value 5 or greater",
												Object: opt.Val(game.EventStackObjectReference()),
												ObjectMatches: opt.Val(game.Selection{ManaValue: opt.Val(compare.Int{
													Op:    compare.GreaterOrEqual,
													Value: 5,
												})}),
												SourceCounterKind: counter.PlusOnePlusOne,
											}),
										}),
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
			Whenever you cast an instant or sorcery spell, you gain 1 life. If that spell has mana value 5 or greater, put a +1/+1 counter on this creature.
		`,
		},
	}
}
