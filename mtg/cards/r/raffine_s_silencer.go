package r

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// RaffineSSilencer is the card definition for Raffine's Silencer.
//
// Type: Creature — Human Assassin
// Cost: {2}{B}
//
// Oracle text:
//
//	When this creature enters, it connives. (Draw a card, then discard a card. If you discarded a nonland card, put a +1/+1 counter on this creature.)
//	When this creature dies, target creature an opponent controls gets -X/-X until end of turn, where X is this creature's power.
var RaffineSSilencer = newRaffineSSilencer()

func newRaffineSSilencer() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Raffine's Silencer",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.B,
			}),
			Colors:    []color.Color{color.Black},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Human, types.Assassin},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 1}),
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
						Sequence: []game.Instruction{
							{
								Primitive: game.Connive{
									Object: game.EventPermanentReference(),
									Player: game.ObjectControllerReference(game.EventPermanentReference()),
									Amount: game.Fixed(1),
								},
							},
						},
					}.Ability(),
				},
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhen,
						Pattern: game.TriggerPattern{
							Event:            game.EventPermanentDied,
							Source:           game.TriggerSourceSelf,
							SubjectSelection: game.Selection{RequiredTypes: []types.Card{types.Creature}},
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
									Object: game.TargetPermanentReference(0),
									PowerDelta: game.Dynamic(game.DynamicAmount{
										Kind:       game.DynamicAmountObjectPower,
										Multiplier: -1,
										Object:     game.SourcePermanentReference(),
									}),
									ToughnessDelta: game.Dynamic(game.DynamicAmount{
										Kind:       game.DynamicAmountObjectPower,
										Multiplier: -1,
										Object:     game.SourcePermanentReference(),
									}),
									Duration: game.DurationUntilEndOfTurn,
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			When this creature enters, it connives. (Draw a card, then discard a card. If you discarded a nonland card, put a +1/+1 counter on this creature.)
			When this creature dies, target creature an opponent controls gets -X/-X until end of turn, where X is this creature's power.
		`,
		},
	}
}
