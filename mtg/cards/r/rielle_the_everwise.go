package r

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// RielleTheEverwise is the card definition for Rielle, the Everwise.
//
// Type: Legendary Creature — Human Wizard
// Cost: {1}{U}{R}
//
// Oracle text:
//
//	Rielle gets +1/+0 for each instant and sorcery card in your graveyard.
//	Whenever you discard one or more cards for the first time each turn, draw that many cards.
var RielleTheEverwise = newRielleTheEverwise

func newRielleTheEverwise() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue, color.Red),
		CardFace: game.CardFace{
			Name: "Rielle, the Everwise",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.U,
				cost.R,
			}),
			Colors:     []color.Color{color.Red, color.Blue},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Human, types.Wizard},
			Power:      opt.Val(game.PT{Value: 0}),
			Toughness:  opt.Val(game.PT{Value: 3}),
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "Rielle gets +1/+0 for each instant and sorcery card in your graveyard.",
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Duration:       game.DurationPermanent,
							AffectedSource: true,
							Layer:          game.LayerPowerToughnessModify,
							PowerDeltaDynamic: opt.Val(game.DynamicAmount{
								Kind:        game.DynamicAmountCountCardsInZone,
								Multiplier:  1,
								CounterKind: counter.PlusOnePlusOne,
								Player:      func() *game.PlayerReference { ref := game.ControllerReference(); return &ref }(),
								CardZone:    zone.Graveyard,
								Selection:   &game.Selection{RequiredTypesAny: []types.Card{types.Instant, types.Sorcery}},
							}),
						},
					},
				},
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Text: "Whenever you discard one or more cards for the first time each turn, draw that many cards.",
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:                      game.EventCardDiscarded,
							Player:                     game.TriggerPlayerYou,
							StackObjectKind:            game.StackSpell,
							PlayerEventOrdinalThisTurn: 1,
							OneOrMore:                  true,
							CounterKind:                counter.PlusOnePlusOne,
						},
					},
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.Draw{
											Amount: game.Dynamic(game.DynamicAmount{
												Kind:       game.DynamicAmountEventCardCount,
												Multiplier: 1,
											}),
											Player: game.ControllerReference(),
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
			Rielle gets +1/+0 for each instant and sorcery card in your graveyard.
			Whenever you discard one or more cards for the first time each turn, draw that many cards.
		`,
		},
	}
}
