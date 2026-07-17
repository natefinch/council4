package t

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// TillerEngine is the card definition for Tiller Engine.
//
// Type: Artifact Creature — Construct
// Cost: {2}
//
// Oracle text:
//
//	Whenever a land you control enters tapped, choose one —
//	• Untap that land.
//	• Tap target nonland permanent an opponent controls.
var TillerEngine = newTillerEngine

func newTillerEngine() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Tiller Engine",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
			}),
			Types:     []types.Card{types.Artifact, types.Creature},
			Subtypes:  []types.Sub{types.Construct},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 3}),
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:            game.EventPermanentEnteredBattlefield,
							Controller:       game.TriggerControllerYou,
							SubjectSelection: game.Selection{RequiredTypes: []types.Card{types.Land}, Tapped: game.TriTrue},
						},
					},
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Text: "Untap that land.",
								Sequence: []game.Instruction{
									{
										Primitive: game.Untap{
											Object: game.EventPermanentReference(),
										},
									},
								},
							},
							game.Mode{
								Text: "Tap target nonland permanent an opponent controls.",
								Targets: []game.TargetSpec{
									game.TargetSpec{
										MinTargets: 1,
										MaxTargets: 1,
										Constraint: "target nonland permanent an opponent controls",
										Allow:      game.TargetAllowPermanent,
										Selection:  opt.Val(game.Selection{ExcludedTypes: []types.Card{types.Land}, Controller: game.ControllerOpponent}),
									},
								},
								Sequence: []game.Instruction{
									{
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
			OracleText: `
			Whenever a land you control enters tapped, choose one —
			• Untap that land.
			• Tap target nonland permanent an opponent controls.
		`,
		},
	}
}
