package o

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// OtterPenguin is the card definition for Otter-Penguin.
//
// Type: Creature — Otter Bird
// Cost: {1}{U}
//
// Oracle text:
//
//	Whenever you draw your second card each turn, this creature gets +1/+2 until end of turn and can't be blocked this turn.
var OtterPenguin = newOtterPenguin

func newOtterPenguin() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Otter-Penguin",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.U,
			}),
			Colors:    []color.Color{color.Blue},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Otter, types.Bird},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 1}),
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Text: "Whenever you draw your second card each turn, this creature gets +1/+2 until end of turn and can't be blocked this turn.",
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:                      game.EventCardDrawn,
							Player:                     game.TriggerPlayerYou,
							StackObjectKind:            game.StackSpell,
							PlayerEventOrdinalThisTurn: 2,
							CounterKind:                counter.PlusOnePlusOne,
						},
					},
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.ModifyPT{
											Object:         game.SourcePermanentReference(),
											PowerDelta:     game.Fixed(1),
											ToughnessDelta: game.Fixed(2),
											Duration:       game.DurationUntilEndOfTurn,
										},
									},
									game.Instruction{
										Primitive: game.ApplyRule{
											Object: opt.Val(game.SourcePermanentReference()),
											RuleEffects: []game.RuleEffect{
												game.RuleEffect{
													Kind:     game.RuleEffectCantBeBlocked,
													Duration: game.DurationPermanent,
												},
											},
											Duration: game.DurationThisTurn,
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
			Whenever you draw your second card each turn, this creature gets +1/+2 until end of turn and can't be blocked this turn.
		`,
		},
	}
}
