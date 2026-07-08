package t

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// TragicBanshee is the card definition for Tragic Banshee.
//
// Type: Creature — Spirit
// Cost: {4}{B}
//
// Oracle text:
//
//	Morbid — When this creature enters, target creature an opponent controls gets -1/-1 until end of turn. If a creature died this turn, that creature gets -13/-13 until end of turn instead.
var TragicBanshee = newTragicBanshee

func newTragicBanshee() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Tragic Banshee",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
				cost.B,
			}),
			Colors:    []color.Color{color.Black},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Spirit},
			Power:     opt.Val(game.PT{Value: 5}),
			Toughness: opt.Val(game.PT{Value: 3}),
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
								Condition: opt.Val(game.EffectCondition{
									Condition: opt.Val(game.Condition{
										Negate: true,
										EventHistory: opt.Val(game.EventHistoryCondition{Pattern: game.TriggerPattern{
											Event:            game.EventPermanentDied,
											SubjectSelection: game.Selection{RequiredTypes: []types.Card{types.Creature}},
										}, Window: game.EventHistoryCurrentTurn}),
									}),
								}),
							},
							{
								Primitive: game.ModifyPT{
									Object:         game.TargetPermanentReference(0),
									PowerDelta:     game.Fixed(-13),
									ToughnessDelta: game.Fixed(-13),
									Duration:       game.DurationUntilEndOfTurn,
								},
								Condition: opt.Val(game.EffectCondition{
									Condition: opt.Val(game.Condition{
										EventHistory: opt.Val(game.EventHistoryCondition{Pattern: game.TriggerPattern{
											Event:            game.EventPermanentDied,
											SubjectSelection: game.Selection{RequiredTypes: []types.Card{types.Creature}},
										}, Window: game.EventHistoryCurrentTurn}),
									}),
								}),
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Morbid — When this creature enters, target creature an opponent controls gets -1/-1 until end of turn. If a creature died this turn, that creature gets -13/-13 until end of turn instead.
		`,
		},
	}
}
