package c

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// ContagionEngine is the card definition for Contagion Engine.
//
// Type: Artifact
// Cost: {6}
//
// Oracle text:
//
//	When this artifact enters, put a -1/-1 counter on each creature target player controls.
//	{4}, {T}: Proliferate twice. (Choose any number of permanents and/or players, then give each another counter of each kind already there. Then do it again.)
var ContagionEngine = newContagionEngine

func newContagionEngine() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Contagion Engine",
			ManaCost: opt.Val(cost.Mana{
				cost.O(6),
			}),
			Types: []types.Card{types.Artifact},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text:            "{4}, {T}: Proliferate twice. (Choose any number of permanents and/or players, then give each another counter of each kind already there. Then do it again.)",
					ManaCost:        opt.Val(cost.Mana{cost.O(4)}),
					AdditionalCosts: cost.Tap,
					ZoneOfFunction:  zone.Battlefield,
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.Proliferate{
									Amount: game.Fixed(2),
								},
							},
						},
					}.Ability(),
				},
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
					Content: game.Mode{
						Targets: []game.TargetSpec{
							game.TargetSpec{
								MinTargets: 1,
								MaxTargets: 1,
								Constraint: "target player",
								Allow:      game.TargetAllowPlayer,
							},
						},
						Sequence: []game.Instruction{
							{
								Primitive: game.AddCounter{
									Amount:      game.Fixed(1),
									Group:       game.PlayerControlledGroup(game.TargetPlayerReference(0), game.Selection{RequiredTypes: []types.Card{types.Creature}}),
									CounterKind: counter.MinusOneMinusOne,
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			When this artifact enters, put a -1/-1 counter on each creature target player controls.
			{4}, {T}: Proliferate twice. (Choose any number of permanents and/or players, then give each another counter of each kind already there. Then do it again.)
		`,
		},
	}
}
