package d

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// DeceiverExarch is the card definition for Deceiver Exarch.
//
// Type: Creature — Phyrexian Cleric
// Cost: {2}{U}
//
// Oracle text:
//
//	Flash (You may cast this spell any time you could cast an instant.)
//	When this creature enters, choose one —
//	• Untap target permanent you control.
//	• Tap target permanent an opponent controls.
var DeceiverExarch = newDeceiverExarch()

func newDeceiverExarch() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Deceiver Exarch",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.U,
			}),
			Colors:    []color.Color{color.Blue},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Phyrexian, types.Cleric},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 4}),
			StaticAbilities: []game.StaticAbility{
				game.FlashStaticBody,
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
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Text: "Untap target permanent you control.",
								Targets: []game.TargetSpec{
									game.TargetSpec{
										MinTargets: 1,
										MaxTargets: 1,
										Constraint: "target permanent you control",
										Allow:      game.TargetAllowPermanent,
										Selection:  opt.Val(game.Selection{Controller: game.ControllerYou}),
									},
								},
								Sequence: []game.Instruction{
									{
										Primitive: game.Untap{
											Object: game.TargetPermanentReference(0),
										},
									},
								},
							},
							game.Mode{
								Text: "Tap target permanent an opponent controls.",
								Targets: []game.TargetSpec{
									game.TargetSpec{
										MinTargets: 1,
										MaxTargets: 1,
										Constraint: "target permanent an opponent controls",
										Allow:      game.TargetAllowPermanent,
										Selection:  opt.Val(game.Selection{Controller: game.ControllerOpponent}),
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
			Flash (You may cast this spell any time you could cast an instant.)
			When this creature enters, choose one —
			• Untap target permanent you control.
			• Tap target permanent an opponent controls.
		`,
		},
	}
}
