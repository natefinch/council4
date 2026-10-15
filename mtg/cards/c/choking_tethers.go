package c

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// ChokingTethers is the card definition for Choking Tethers.
//
// Type: Instant
// Cost: {3}{U}
//
// Oracle text:
//
//	Tap up to four target creatures.
//	Cycling {1}{U} ({1}{U}, Discard this card: Draw a card.)
//	When you cycle this card, you may tap target creature.
var ChokingTethers = newChokingTethers

func newChokingTethers() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Choking Tethers",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.U,
			}),
			Colors: []color.Color{color.Blue},
			Types:  []types.Card{types.Instant},
			ActivatedAbilities: []game.ActivatedAbility{
				game.CyclingActivatedAbility(cost.Mana{cost.O(1), cost.U}),
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Text: "When you cycle this card, you may tap target creature.",
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhen,
						Pattern: game.TriggerPattern{
							Event:           game.EventCycled,
							Source:          game.TriggerSourceSelf,
							Player:          game.TriggerPlayerYou,
							StackObjectKind: game.StackSpell,
							CounterKind:     counter.PlusOnePlusOne,
						},
					},
					Optional: true,
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
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
									game.Instruction{
										Primitive: game.Tap{
											Object: game.TargetPermanentReference(0),
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
			SpellAbility: opt.Val(game.Mode{
				Targets: []game.TargetSpec{
					game.TargetSpec{
						MinTargets: 0,
						MaxTargets: 4,
						Constraint: "up to four target creatures",
						Allow:      game.TargetAllowPermanent,
						Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}}),
					},
				},
				Sequence: []game.Instruction{
					{
						Primitive: game.Tap{
							Object: game.TargetPermanentReference(0),
						},
					},
					{
						Primitive: game.Tap{
							Object: game.TargetPermanentReference(1),
						},
					},
					{
						Primitive: game.Tap{
							Object: game.TargetPermanentReference(2),
						},
					},
					{
						Primitive: game.Tap{
							Object: game.TargetPermanentReference(3),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Tap up to four target creatures.
			Cycling {1}{U} ({1}{U}, Discard this card: Draw a card.)
			When you cycle this card, you may tap target creature.
		`,
		},
	}
}
