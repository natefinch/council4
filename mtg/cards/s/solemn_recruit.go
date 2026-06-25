package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// SolemnRecruit is the card definition for Solemn Recruit.
//
// Type: Creature — Dwarf Warrior
// Cost: {1}{W}{W}
//
// Oracle text:
//
//	Double strike
//	Revolt — At the beginning of your end step, if a permanent left the battlefield under your control this turn, put a +1/+1 counter on this creature.
var SolemnRecruit = &game.CardDef{
	ColorIdentity: color.NewIdentity(color.White),
	CardFace: game.CardFace{
		Name: "Solemn Recruit",
		ManaCost: opt.Val(cost.Mana{
			cost.O(1),
			cost.W,
			cost.W,
		}),
		Colors:    []color.Color{color.White},
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Dwarf, types.Warrior},
		Power:     opt.Val(game.PT{Value: 2}),
		Toughness: opt.Val(game.PT{Value: 2}),
		StaticAbilities: []game.StaticAbility{
			game.DoubleStrikeStaticBody,
		},
		TriggeredAbilities: []game.TriggeredAbility{
			game.TriggeredAbility{
				Trigger: game.TriggerCondition{
					Type: game.TriggerAt,
					Pattern: game.TriggerPattern{
						Event:      game.EventBeginningOfStep,
						Controller: game.TriggerControllerYou,
						Step:       game.StepEnd,
					},
					InterveningIf: "if a permanent left the battlefield under your control this turn",
					InterveningCondition: opt.Val(game.Condition{
						EventHistory: opt.Val(game.EventHistoryCondition{Pattern: game.TriggerPattern{
							Event:         game.EventZoneChanged,
							Controller:    game.TriggerControllerYou,
							MatchFromZone: true,
							FromZone:      zone.Battlefield,
						}, Window: game.EventHistoryCurrentTurn}),
					}),
				},
				Content: game.Mode{
					Sequence: []game.Instruction{
						{
							Primitive: game.AddCounter{
								Amount:      game.Fixed(1),
								Object:      game.SourcePermanentReference(),
								CounterKind: counter.PlusOnePlusOne,
							},
						},
					},
				}.Ability(),
			},
		},
		OracleText: `
			Double strike
			Revolt — At the beginning of your end step, if a permanent left the battlefield under your control this turn, put a +1/+1 counter on this creature.
		`,
	},
}
