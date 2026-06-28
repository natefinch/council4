package g

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// GrakmawSkyclaveRavager is the card definition for Grakmaw, Skyclave Ravager.
//
// Type: Legendary Creature — Hydra Horror
// Cost: {1}{B}{G}
//
// Oracle text:
//
//	Grakmaw enters with three +1/+1 counters on it.
//	Whenever another creature you control dies, if it had a +1/+1 counter on it, put a +1/+1 counter on Grakmaw.
//	When Grakmaw dies, create an X/X black and green Hydra creature token, where X is the number of +1/+1 counters on Grakmaw.
var GrakmawSkyclaveRavager = newGrakmawSkyclaveRavager()

func newGrakmawSkyclaveRavager() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black, color.Green),
		CardFace: game.CardFace{
			Name: "Grakmaw, Skyclave Ravager",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.B,
				cost.G,
			}),
			Colors:     []color.Color{color.Black, color.Green},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Hydra, types.Horror},
			Power:      opt.Val(game.PT{Value: 0}),
			Toughness:  opt.Val(game.PT{Value: 0}),
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:            game.EventPermanentDied,
							Controller:       game.TriggerControllerYou,
							ExcludeSelf:      true,
							SubjectSelection: game.Selection{RequiredTypes: []types.Card{types.Creature}},
						},
						InterveningIf: "if it had a +1/+1 counter on it",
						InterveningIfEventPermanentHadCounterKind: opt.Val(counter.PlusOnePlusOne),
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.AddCounter{
									Amount:      game.Fixed(1),
									Object:      game.SourcePermanentReference(),
									CounterKind: counter.PlusOnePlusOne,
								},
							},
						},
					}.Ability(),
				},
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
						Sequence: []game.Instruction{
							{
								Primitive: game.CreateToken{
									Amount: game.Fixed(1),
									Source: game.TokenDef(grakmawSkyclaveRavagerToken),
									Power: opt.Val(game.Dynamic(game.DynamicAmount{
										Kind:        game.DynamicAmountObjectCounters,
										Multiplier:  1,
										CounterKind: counter.PlusOnePlusOne,
										Object:      game.SourcePermanentReference(),
									})),
									Toughness: opt.Val(game.Dynamic(game.DynamicAmount{
										Kind:        game.DynamicAmountObjectCounters,
										Multiplier:  1,
										CounterKind: counter.PlusOnePlusOne,
										Object:      game.SourcePermanentReference(),
									})),
								},
							},
						},
					}.Ability(),
				},
			},
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntersWithCountersReplacement("Grakmaw enters with three +1/+1 counters on it.", game.CounterPlacement{Kind: counter.PlusOnePlusOne, Amount: 3}),
			},
			OracleText: `
			Grakmaw enters with three +1/+1 counters on it.
			Whenever another creature you control dies, if it had a +1/+1 counter on it, put a +1/+1 counter on Grakmaw.
			When Grakmaw dies, create an X/X black and green Hydra creature token, where X is the number of +1/+1 counters on Grakmaw.
		`,
		},
	}
}

var grakmawSkyclaveRavagerToken = newGrakmawSkyclaveRavagerToken()

func newGrakmawSkyclaveRavagerToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:     "Hydra",
			Colors:   []color.Color{color.Black, color.Green},
			Types:    []types.Card{types.Creature},
			Subtypes: []types.Sub{types.Hydra},
		},
	}
}
