package u

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// UurgSpawnOfTurg is the card definition for Uurg, Spawn of Turg.
//
// Type: Legendary Creature — Frog Beast
// Cost: {B}{B}{G}
//
// Oracle text:
//
//	Uurg's power is equal to the number of land cards in your graveyard.
//	At the beginning of your upkeep, surveil 1. (Look at the top card of your library. You may put that card into your graveyard.)
//	{B}{G}, Sacrifice a land: You gain 2 life.
var UurgSpawnOfTurg = newUurgSpawnOfTurg

func newUurgSpawnOfTurg() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black, color.Green),
		CardFace: game.CardFace{
			Name: "Uurg, Spawn of Turg",
			ManaCost: opt.Val(cost.Mana{
				cost.B,
				cost.B,
				cost.G,
			}),
			Colors:       []color.Color{color.Black, color.Green},
			Supertypes:   []types.Super{types.Legendary},
			Types:        []types.Card{types.Creature},
			Subtypes:     []types.Sub{types.Frog, types.Beast},
			Power:        opt.Val(game.PT{IsStar: true}),
			Toughness:    opt.Val(game.PT{Value: 5}),
			DynamicPower: opt.Val(game.DynamicValue{Kind: game.DynamicValueControllerLandCardsInGraveyard}),
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "{B}{G}, Sacrifice a land: You gain 2 life.",
					ManaCost: opt.Val(cost.Mana{
						cost.B,
						cost.G,
					}),
					AdditionalCosts: []cost.Additional{
						cost.Additional{
							Kind:               cost.AdditionalSacrifice,
							Text:               "Sacrifice a land",
							Amount:             1,
							MatchPermanentType: true,
							PermanentType:      types.Land,
							CounterKind:        counter.PlusOnePlusOne,
						},
					},
					ZoneOfFunction: zone.Battlefield,
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.GainLife{
											Amount: game.Fixed(2),
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
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Text: "At the beginning of your upkeep, surveil 1. (Look at the top card of your library. You may put that card into your graveyard.)",
					Trigger: game.TriggerCondition{
						Type: game.TriggerAt,
						Pattern: game.TriggerPattern{
							Event:           game.EventBeginningOfStep,
							Controller:      game.TriggerControllerYou,
							StackObjectKind: game.StackSpell,
							CounterKind:     counter.PlusOnePlusOne,
							Step:            game.StepUpkeep,
						},
					},
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.Surveil{
											Amount: game.Fixed(1),
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
			Uurg's power is equal to the number of land cards in your graveyard.
			At the beginning of your upkeep, surveil 1. (Look at the top card of your library. You may put that card into your graveyard.)
			{B}{G}, Sacrifice a land: You gain 2 life.
		`,
		},
	}
}
