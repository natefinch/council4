package t

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// TitaniaSCommand is the card definition for Titania's Command.
//
// Type: Sorcery
// Cost: {4}{G}{G}
//
// Oracle text:
//
//	Choose two —
//	• Exile target player's graveyard. You gain 1 life for each card exiled this way.
//	• Search your library for up to two land cards, put them onto the battlefield tapped, then shuffle.
//	• Create two 2/2 green Bear creature tokens.
//	• Put two +1/+1 counters on each creature you control.
var TitaniaSCommand = newTitaniaSCommand

func newTitaniaSCommand() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Titania's Command",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
				cost.G,
				cost.G,
			}),
			Colors: []color.Color{color.Green},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.AbilityContent{
				Modes: []game.Mode{
					game.Mode{
						Text: "Exile target player's graveyard. You gain 1 life for each card exiled this way.",
						Targets: []game.TargetSpec{
							game.TargetSpec{
								MinTargets: 1,
								MaxTargets: 1,
								Constraint: "target player's graveyard",
								Allow:      game.TargetAllowPlayer,
							},
						},
						Sequence: []game.Instruction{
							{
								Primitive: game.MoveCard{
									Player:      game.TargetPlayerReference(0),
									FromZone:    zone.Graveyard,
									Destination: zone.Exile,
								},
								PublishResult: game.ResultKey("actual-card-quantity-1"),
								LocalProducts: game.LocalProducts{
									Results: []game.ResultKey{game.ResultKey("actual-card-quantity-1")},
								},
							},
							{
								Primitive: game.GainLife{
									Amount: game.Dynamic(game.DynamicAmount{
										Kind:       game.DynamicAmountPreviousEffectResult,
										Multiplier: 1,
										Selection:  &game.Selection{},
										ResultKey:  game.ResultKey("actual-card-quantity-1"),
									}),
									Player: game.ControllerReference(),
								},
								ResultGate: opt.Val(game.InstructionResultGate{
									Key:      "actual-card-quantity-1",
									Accepted: game.TriTrue,
								}),
							},
						},
					},
					game.Mode{
						Text: "Search your library for up to two land cards, put them onto the battlefield tapped, then shuffle.",
						Sequence: []game.Instruction{
							{
								Primitive: game.Search{
									Player: game.ControllerReference(),
									Spec: game.SearchSpec{
										SourceZone:   zone.Library,
										Destination:  zone.Battlefield,
										Filter:       game.Selection{RequiredTypes: []types.Card{types.Land}},
										EntersTapped: true,
									},
									Amount: game.Fixed(2),
								},
							},
						},
					},
					game.Mode{
						Text: "Create two 2/2 green Bear creature tokens.",
						Sequence: []game.Instruction{
							{
								Primitive: game.CreateToken{
									Amount: game.Fixed(2),
									Source: game.TokenDef(titaniaSCommandToken),
								},
							},
						},
					},
					game.Mode{
						Text: "Put two +1/+1 counters on each creature you control.",
						Sequence: []game.Instruction{
							{
								Primitive: game.AddCounter{
									Amount:      game.Fixed(2),
									Group:       game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}, Controller: game.ControllerYou}),
									CounterKind: counter.PlusOnePlusOne,
								},
							},
						},
					},
				},
				MinModes: 2,
				MaxModes: 2,
			}),
			OracleText: `
			Choose two —
			• Exile target player's graveyard. You gain 1 life for each card exiled this way.
			• Search your library for up to two land cards, put them onto the battlefield tapped, then shuffle.
			• Create two 2/2 green Bear creature tokens.
			• Put two +1/+1 counters on each creature you control.
		`,
		},
	}
}

var titaniaSCommandToken = newTitaniaSCommandToken()

func newTitaniaSCommandToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Bear",
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Bear},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 2}),
		},
	}
}
