package p

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// PsychosisCrawler is the card definition for Psychosis Crawler.
//
// Type: Artifact Creature — Phyrexian Horror
// Cost: {5}
//
// Oracle text:
//
//	Psychosis Crawler's power and toughness are each equal to the number of cards in your hand.
//	Whenever you draw a card, each opponent loses 1 life.
var PsychosisCrawler = newPsychosisCrawler

func newPsychosisCrawler() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Psychosis Crawler",
			ManaCost: opt.Val(cost.Mana{
				cost.O(5),
			}),
			Types:            []types.Card{types.Artifact, types.Creature},
			Subtypes:         []types.Sub{types.Phyrexian, types.Horror},
			Power:            opt.Val(game.PT{IsStar: true}),
			Toughness:        opt.Val(game.PT{IsStar: true}),
			DynamicPower:     opt.Val(game.DynamicValue{Kind: game.DynamicValueControllerHandSize}),
			DynamicToughness: opt.Val(game.DynamicValue{Kind: game.DynamicValueControllerHandSize}),
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Text: "Whenever you draw a card, each opponent loses 1 life.",
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:           game.EventCardDrawn,
							Player:          game.TriggerPlayerYou,
							StackObjectKind: game.StackSpell,
							CounterKind:     counter.PlusOnePlusOne,
						},
					},
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.LoseLife{
											Amount:      game.Fixed(1),
											PlayerGroup: game.OpponentsReference(),
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
			Psychosis Crawler's power and toughness are each equal to the number of cards in your hand.
			Whenever you draw a card, each opponent loses 1 life.
		`,
		},
	}
}
