package p

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// PoisonerSApprentice is the card definition for Poisoner's Apprentice.
//
// Type: Creature — Orc Warlock
// Cost: {2}{B}
//
// Oracle text:
//
//	Infusion — When this creature enters, target creature an opponent controls gets -4/-4 until end of turn if you gained life this turn.
var PoisonerSApprentice = newPoisonerSApprentice

func newPoisonerSApprentice() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Poisoner's Apprentice",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.B,
			}),
			Colors:    []color.Color{color.Black},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Orc, types.Warlock},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 2}),
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Text: "Infusion — When this creature enters, target creature an opponent controls gets -4/-4 until end of turn if you gained life this turn.",
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhen,
						Pattern: game.TriggerPattern{
							Event:           game.EventPermanentEnteredBattlefield,
							Source:          game.TriggerSourceSelf,
							StackObjectKind: game.StackSpell,
							CounterKind:     counter.PlusOnePlusOne,
						},
					},
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
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
									game.Instruction{
										Primitive: game.ModifyPT{
											Object:         game.TargetPermanentReference(0),
											PowerDelta:     game.Fixed(-4),
											ToughnessDelta: game.Fixed(-4),
											Duration:       game.DurationUntilEndOfTurn,
										},
										Condition: opt.Val(game.EffectCondition{
											Condition: opt.Val(game.Condition{
												Text:              "if you gained life this turn",
												SourceCounterKind: counter.PlusOnePlusOne,
												EventHistory: opt.Val(game.EventHistoryCondition{
													Pattern: game.TriggerPattern{
														Event:           game.EventLifeGained,
														Player:          game.TriggerPlayerYou,
														StackObjectKind: game.StackSpell,
														CounterKind:     counter.PlusOnePlusOne,
													},
													Window: game.EventHistoryCurrentTurn,
												}),
											}),
										}),
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
			Infusion — When this creature enters, target creature an opponent controls gets -4/-4 until end of turn if you gained life this turn.
		`,
		},
	}
}
