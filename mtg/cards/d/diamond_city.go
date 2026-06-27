package d

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// DiamondCity is the card definition for Diamond City.
//
// Type: Land
//
// Oracle text:
//
//	This land enters with a shield counter on it. (If it would be dealt damage or destroyed, remove a shield counter from it instead.)
//	{T}: Add {C}.
//	{T}: Move a shield counter from this land onto target creature. Activate only if two or more creatures entered the battlefield under your control this turn.
var DiamondCity = newDiamondCity()

func newDiamondCity() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:  "Diamond City",
			Types: []types.Card{types.Land},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text:            "{T}: Move a shield counter from this land onto target creature. Activate only if two or more creatures entered the battlefield under your control this turn.",
					AdditionalCosts: cost.Tap,
					ZoneOfFunction:  zone.Battlefield,
					ActivationCondition: opt.Val(game.Condition{
						EventHistory: opt.Val(game.EventHistoryCondition{Pattern: game.TriggerPattern{
							Event:            game.EventPermanentEnteredBattlefield,
							Controller:       game.TriggerControllerYou,
							SubjectSelection: game.Selection{RequiredTypes: []types.Card{types.Creature}},
						}, Window: game.EventHistoryCurrentTurn, MinCount: 2}),
					}),
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
								Primitive: game.MoveCounters{
									Amount:      game.Fixed(1),
									Object:      game.TargetPermanentReference(0),
									CounterKind: counter.Shield,
									Source:      game.CounterSourceSpec{Kind: game.CounterSourceSelf},
								},
							},
						},
					}.Ability(),
				},
			},
			ManaAbilities: []game.ManaAbility{
				game.TapManaAbility(mana.C),
			},
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntersWithCountersReplacement("This land enters with a shield counter on it. (If it would be dealt damage or destroyed, remove a shield counter from it instead.)", game.CounterPlacement{Kind: counter.Shield, Amount: 1}),
			},
			OracleText: `
			This land enters with a shield counter on it. (If it would be dealt damage or destroyed, remove a shield counter from it instead.)
			{T}: Add {C}.
			{T}: Move a shield counter from this land onto target creature. Activate only if two or more creatures entered the battlefield under your control this turn.
		`,
		},
	}
}
