package g

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// GriefTyrant is the card definition for Grief Tyrant.
//
// Type: Creature — Horror
// Cost: {5}{B/R}
//
// Oracle text:
//
//	This creature enters with four -1/-1 counters on it.
//	When this creature dies, put a -1/-1 counter on target creature for each -1/-1 counter on this creature.
var GriefTyrant = newGriefTyrant()

func newGriefTyrant() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black, color.Red),
		CardFace: game.CardFace{
			Name: "Grief Tyrant",
			ManaCost: opt.Val(cost.Mana{
				cost.O(5),
				cost.HybridMana(mana.B, mana.R),
			}),
			Colors:    []color.Color{color.Black, color.Red},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Horror},
			Power:     opt.Val(game.PT{Value: 8}),
			Toughness: opt.Val(game.PT{Value: 8}),
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
								Constraint: "target creature",
								Allow:      game.TargetAllowPermanent,
								Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}}),
							},
						},
						Sequence: []game.Instruction{
							{
								Primitive: game.AddCounter{
									Amount: game.Dynamic(game.DynamicAmount{
										Kind:        game.DynamicAmountObjectCounters,
										Multiplier:  1,
										CounterKind: counter.MinusOneMinusOne,
										Object:      game.SourcePermanentReference(),
									}),
									Object:      game.TargetPermanentReference(0),
									CounterKind: counter.MinusOneMinusOne,
								},
							},
						},
					}.Ability(),
				},
			},
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntersWithCountersReplacement("This creature enters with four -1/-1 counters on it.", game.CounterPlacement{Kind: counter.MinusOneMinusOne, Amount: 4}),
			},
			OracleText: `
			This creature enters with four -1/-1 counters on it.
			When this creature dies, put a -1/-1 counter on target creature for each -1/-1 counter on this creature.
		`,
		},
	}
}
