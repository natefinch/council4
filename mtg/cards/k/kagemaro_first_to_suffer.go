package k

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// KagemaroFirstToSuffer is the card definition for Kagemaro, First to Suffer.
//
// Type: Legendary Creature — Demon Spirit
// Cost: {3}{B}{B}
//
// Oracle text:
//
//	Kagemaro's power and toughness are each equal to the number of cards in your hand.
//	{B}, Sacrifice Kagemaro: All creatures get -X/-X until end of turn, where X is the number of cards in your hand.
var KagemaroFirstToSuffer = newKagemaroFirstToSuffer

func newKagemaroFirstToSuffer() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Kagemaro, First to Suffer",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.B,
				cost.B,
			}),
			Colors:           []color.Color{color.Black},
			Supertypes:       []types.Super{types.Legendary},
			Types:            []types.Card{types.Creature},
			Subtypes:         []types.Sub{types.Demon, types.Spirit},
			Power:            opt.Val(game.PT{IsStar: true}),
			Toughness:        opt.Val(game.PT{IsStar: true}),
			DynamicPower:     opt.Val(game.DynamicValue{Kind: game.DynamicValueControllerHandSize}),
			DynamicToughness: opt.Val(game.DynamicValue{Kind: game.DynamicValueControllerHandSize}),
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "{B}, Sacrifice Kagemaro: All creatures get -X/-X until end of turn, where X is the number of cards in your hand.",
					ManaCost: opt.Val(cost.Mana{
						cost.B,
					}),
					AdditionalCosts: []cost.Additional{
						cost.Additional{
							Kind:        cost.AdditionalSacrificeSource,
							Text:        "Sacrifice Kagemaro",
							Amount:      1,
							CounterKind: counter.PlusOnePlusOne,
						},
					},
					ZoneOfFunction: zone.Battlefield,
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.ApplyContinuous{
											ContinuousEffects: []game.ContinuousEffect{
												game.ContinuousEffect{
													Duration: game.DurationPermanent,
													Group:    game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}}),
													Layer:    game.LayerPowerToughnessModify,
													PowerDeltaDynamic: opt.Val(game.DynamicAmount{
														Kind:        game.DynamicAmountCountCardsInZone,
														Multiplier:  -1,
														CounterKind: counter.PlusOnePlusOne,
														Player:      func() *game.PlayerReference { ref := game.ControllerReference(); return &ref }(),
														CardZone:    zone.Hand,
														Selection:   &game.Selection{},
													}),
													ToughnessDeltaDynamic: opt.Val(game.DynamicAmount{
														Kind:        game.DynamicAmountCountCardsInZone,
														Multiplier:  -1,
														CounterKind: counter.PlusOnePlusOne,
														Player:      func() *game.PlayerReference { ref := game.ControllerReference(); return &ref }(),
														CardZone:    zone.Hand,
														Selection:   &game.Selection{},
													}),
												},
											},
											Duration: game.DurationUntilEndOfTurn,
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
			Kagemaro's power and toughness are each equal to the number of cards in your hand.
			{B}, Sacrifice Kagemaro: All creatures get -X/-X until end of turn, where X is the number of cards in your hand.
		`,
		},
	}
}
