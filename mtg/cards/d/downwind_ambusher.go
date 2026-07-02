package d

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// DownwindAmbusher is the card definition for Downwind Ambusher.
//
// Type: Creature — Skunk Assassin
// Cost: {3}{B}
//
// Oracle text:
//
//	Flash
//	When this creature enters, choose one —
//	• Target creature an opponent controls gets -1/-1 until end of turn.
//	• Destroy target creature an opponent controls that was dealt damage this turn.
var DownwindAmbusher = newDownwindAmbusher()

func newDownwindAmbusher() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Downwind Ambusher",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.B,
			}),
			Colors:    []color.Color{color.Black},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Skunk, types.Assassin},
			Power:     opt.Val(game.PT{Value: 4}),
			Toughness: opt.Val(game.PT{Value: 2}),
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
								Text: "Target creature an opponent controls gets -1/-1 until end of turn.",
								Targets: []game.TargetSpec{
									game.TargetSpec{
										MinTargets: 1,
										MaxTargets: 1,
										Constraint: "target creature an opponent controls",
										Allow:      game.TargetAllowPermanent,
										Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}, Controller: game.ControllerOpponent}),
									},
								},
								Sequence: []game.Instruction{
									{
										Primitive: game.ModifyPT{
											Object:         game.TargetPermanentReference(0),
											PowerDelta:     game.Fixed(-1),
											ToughnessDelta: game.Fixed(-1),
											Duration:       game.DurationUntilEndOfTurn,
										},
									},
								},
							},
							game.Mode{
								Text: "Destroy target creature an opponent controls that was dealt damage this turn.",
								Targets: []game.TargetSpec{
									game.TargetSpec{
										MinTargets: 1,
										MaxTargets: 1,
										Constraint: "target creature an opponent controls that was dealt damage this turn",
										Allow:      game.TargetAllowPermanent,
										Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}, Controller: game.ControllerOpponent, DealtDamageThisTurn: true}),
									},
								},
								Sequence: []game.Instruction{
									{
										Primitive: game.Destroy{
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
			Flash
			When this creature enters, choose one —
			• Target creature an opponent controls gets -1/-1 until end of turn.
			• Destroy target creature an opponent controls that was dealt damage this turn.
		`,
		},
	}
}
