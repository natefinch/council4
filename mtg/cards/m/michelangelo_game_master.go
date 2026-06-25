package m

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// MichelangeloGameMaster is the card definition for Michelangelo, Game Master.
//
// Type: Legendary Creature — Mutant Ninja Turtle
// Cost: {2}{G}
//
// Oracle text:
//
//	Disappear — At the beginning of your end step, if a permanent left the battlefield under your control this turn, put a +1/+1 counter on Michelangelo.
var MichelangeloGameMaster = &game.CardDef{
	ColorIdentity: color.NewIdentity(color.Green),
	CardFace: game.CardFace{
		Name: "Michelangelo, Game Master",
		ManaCost: opt.Val(cost.Mana{
			cost.O(2),
			cost.G,
		}),
		Colors:     []color.Color{color.Green},
		Supertypes: []types.Super{types.Legendary},
		Types:      []types.Card{types.Creature},
		Subtypes:   []types.Sub{types.Mutant, types.Ninja, types.Turtle},
		Power:      opt.Val(game.PT{Value: 3}),
		Toughness:  opt.Val(game.PT{Value: 3}),
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
			Disappear — At the beginning of your end step, if a permanent left the battlefield under your control this turn, put a +1/+1 counter on Michelangelo.
		`,
	},
}
