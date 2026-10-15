package h

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// HullbreakerHorror is the card definition for Hullbreaker Horror.
//
// Type: Creature — Kraken Horror
// Cost: {5}{U}{U}
//
// Oracle text:
//
//	Flash
//	This spell can't be countered.
//	Whenever you cast a spell, choose up to one —
//	• Return target spell you don't control to its owner's hand.
//	• Return target nonland permanent to its owner's hand.
var HullbreakerHorror = newHullbreakerHorror

func newHullbreakerHorror() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Hullbreaker Horror",
			ManaCost: opt.Val(cost.Mana{
				cost.O(5),
				cost.U,
				cost.U,
			}),
			Colors:    []color.Color{color.Blue},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Kraken, types.Horror},
			Power:     opt.Val(game.PT{Value: 7}),
			Toughness: opt.Val(game.PT{Value: 8}),
			StaticAbilities: []game.StaticAbility{
				game.FlashStaticBody,
				game.CantBeCounteredStaticBody,
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Text: "Whenever you cast a spell, choose up to one —\n• Return target spell you don't control to its owner's hand.\n• Return target nonland permanent to its owner's hand.",
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:           game.EventSpellCast,
							Controller:      game.TriggerControllerYou,
							StackObjectKind: game.StackSpell,
							CounterKind:     counter.PlusOnePlusOne,
						},
					},
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Text: "Return target spell you don't control to its owner's hand.",
								Targets: []game.TargetSpec{
									game.TargetSpec{
										MinTargets: 1,
										MaxTargets: 1,
										Constraint: "target spell you don't control",
										Allow:      game.TargetAllowStackObject,
										Predicate: game.TargetPredicate{
											StackObjectKinds: []game.StackObjectKind{game.StackSpell},
											Controller:       game.ControllerNotYou,
										},
									},
								},
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.MovePermanent{
											Object:      game.TargetObjectReference(0),
											Destination: zone.Hand,
										},
									},
								},
							},
							game.Mode{
								Text: "Return target nonland permanent to its owner's hand.",
								Targets: []game.TargetSpec{
									game.TargetSpec{
										MinTargets: 1,
										MaxTargets: 1,
										Constraint: "target nonland permanent",
										Allow:      game.TargetAllowPermanent,
										Selection:  opt.Val(game.Selection{ExcludedTypes: []types.Card{types.Land}}),
									},
								},
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.MovePermanent{
											Object:      game.TargetPermanentReference(0),
											Destination: zone.Hand,
										},
									},
								},
							},
						},
						MaxModes: 1,
					},
				},
			},
			OracleText: `
			Flash
			This spell can't be countered.
			Whenever you cast a spell, choose up to one —
			• Return target spell you don't control to its owner's hand.
			• Return target nonland permanent to its owner's hand.
		`,
		},
	}
}
