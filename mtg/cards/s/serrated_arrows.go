package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// SerratedArrows is the card definition for Serrated Arrows.
//
// Type: Artifact
// Cost: {4}
//
// Oracle text:
//
//	This artifact enters with three arrowhead counters on it.
//	At the beginning of your upkeep, if there are no arrowhead counters on this artifact, sacrifice it.
//	{T}, Remove an arrowhead counter from this artifact: Put a -1/-1 counter on target creature.
var SerratedArrows = newSerratedArrows()

func newSerratedArrows() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Serrated Arrows",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
			}),
			Types: []types.Card{types.Artifact},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "{T}, Remove an arrowhead counter from this artifact: Put a -1/-1 counter on target creature.",
					AdditionalCosts: []cost.Additional{
						{
							Kind: cost.AdditionalTap,
						},
						{
							Kind:        cost.AdditionalRemoveCounter,
							Text:        "Remove an arrowhead counter from this artifact",
							Amount:      1,
							CounterKind: counter.Arrowhead,
						},
					},
					ZoneOfFunction: zone.Battlefield,
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
									Amount:      game.Fixed(1),
									Object:      game.TargetPermanentReference(0),
									CounterKind: counter.MinusOneMinusOne,
								},
							},
						},
					}.Ability(),
				},
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerAt,
						Pattern: game.TriggerPattern{
							Event:      game.EventBeginningOfStep,
							Controller: game.TriggerControllerYou,
							Step:       game.StepUpkeep,
						},
						InterveningIf: "if there are no arrowhead counters on this artifact",
						InterveningCondition: opt.Val(game.Condition{
							Negate:        true,
							Object:        opt.Val(game.SourcePermanentReference()),
							ObjectMatches: opt.Val(game.Selection{RequiredCounterCount: opt.Val(compare.Int{Op: compare.GreaterOrEqual, Value: 1}), RequiredCounter: counter.Arrowhead}),
						}),
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.Sacrifice{
									Object: game.SourcePermanentReference(),
								},
							},
						},
					}.Ability(),
				},
			},
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntersWithCountersReplacement("This artifact enters with three arrowhead counters on it.", game.CounterPlacement{Kind: counter.Arrowhead, Amount: 3}),
			},
			OracleText: `
			This artifact enters with three arrowhead counters on it.
			At the beginning of your upkeep, if there are no arrowhead counters on this artifact, sacrifice it.
			{T}, Remove an arrowhead counter from this artifact: Put a -1/-1 counter on target creature.
		`,
		},
	}
}
