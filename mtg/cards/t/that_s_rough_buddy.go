package t

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// ThatSRoughBuddy is the card definition for That's Rough Buddy.
//
// Type: Instant — Lesson
// Cost: {1}{W}
//
// Oracle text:
//
//	Put a +1/+1 counter on target creature. Put two +1/+1 counters on that creature instead if a creature left the battlefield under your control this turn.
//	Draw a card.
var ThatSRoughBuddy = newThatSRoughBuddy()

func newThatSRoughBuddy() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "That's Rough Buddy",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.W,
			}),
			Colors:   []color.Color{color.White},
			Types:    []types.Card{types.Instant},
			Subtypes: []types.Sub{types.Lesson},
			SpellAbility: opt.Val(game.Mode{
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
							CounterKind: counter.PlusOnePlusOne,
						},
						Condition: opt.Val(game.EffectCondition{
							Condition: opt.Val(game.Condition{
								Negate: true,
								EventHistory: opt.Val(game.EventHistoryCondition{Pattern: game.TriggerPattern{
									Event:            game.EventZoneChanged,
									Controller:       game.TriggerControllerYou,
									MatchFromZone:    true,
									FromZone:         zone.Battlefield,
									SubjectSelection: game.Selection{RequiredTypes: []types.Card{types.Creature}},
								}, Window: game.EventHistoryCurrentTurn}),
							}),
						}),
					},
					{
						Primitive: game.AddCounter{
							Amount:      game.Fixed(2),
							Object:      game.TargetPermanentReference(0),
							CounterKind: counter.PlusOnePlusOne,
						},
						Condition: opt.Val(game.EffectCondition{
							Condition: opt.Val(game.Condition{
								EventHistory: opt.Val(game.EventHistoryCondition{Pattern: game.TriggerPattern{
									Event:            game.EventZoneChanged,
									Controller:       game.TriggerControllerYou,
									MatchFromZone:    true,
									FromZone:         zone.Battlefield,
									SubjectSelection: game.Selection{RequiredTypes: []types.Card{types.Creature}},
								}, Window: game.EventHistoryCurrentTurn}),
							}),
						}),
					},
					{
						Primitive: game.Draw{
							Amount: game.Fixed(1),
							Player: game.ControllerReference(),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Put a +1/+1 counter on target creature. Put two +1/+1 counters on that creature instead if a creature left the battlefield under your control this turn.
			Draw a card.
		`,
		},
	}
}
