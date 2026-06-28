package c

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// ComposerOfSpring is the card definition for Composer of Spring.
//
// Type: Creature — Satyr Bard
// Cost: {1}{G}
//
// Oracle text:
//
//	Constellation — Whenever an enchantment you control enters, you may put a land card from your hand onto the battlefield tapped. If you control six or more enchantments, instead you may put a creature or land card from your hand onto the battlefield tapped.
var ComposerOfSpring = newComposerOfSpring()

func newComposerOfSpring() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Composer of Spring",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.G,
			}),
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Satyr, types.Bard},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 3}),
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:            game.EventPermanentEnteredBattlefield,
							Controller:       game.TriggerControllerYou,
							SubjectSelection: game.Selection{RequiredTypes: []types.Card{types.Enchantment}},
						},
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.ChooseFromZone{
									Player:     game.ControllerReference(),
									SourceZone: zone.Hand,
									Filter:     game.Selection{RequiredTypes: []types.Card{types.Land}},
									Quantity:   game.Fixed(1),
									Destination: game.ChooseDestination{
										Zone: zone.Battlefield,
									},
									Riders: game.ChooseRiders{
										EntersTapped: true,
									},
									Prompt: "Choose a card to put onto the battlefield",
								},
								Condition: opt.Val(game.EffectCondition{
									Condition: opt.Val(game.Condition{
										Negate: true,
										ControlsMatching: opt.Val(game.SelectionCount{
											Selection: game.Selection{RequiredTypes: []types.Card{types.Enchantment}},
											MinCount:  6,
										}),
									}),
								}),
								Optional: true,
							},
							{
								Primitive: game.ChooseFromZone{
									Player:     game.ControllerReference(),
									SourceZone: zone.Hand,
									Filter:     game.Selection{RequiredTypesAny: []types.Card{types.Creature, types.Land}},
									Quantity:   game.Fixed(1),
									Destination: game.ChooseDestination{
										Zone: zone.Battlefield,
									},
									Riders: game.ChooseRiders{
										EntersTapped: true,
									},
									Prompt: "Choose a card to put onto the battlefield",
								},
								Condition: opt.Val(game.EffectCondition{
									Condition: opt.Val(game.Condition{
										ControlsMatching: opt.Val(game.SelectionCount{
											Selection: game.Selection{RequiredTypes: []types.Card{types.Enchantment}},
											MinCount:  6,
										}),
									}),
								}),
								Optional: true,
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Constellation — Whenever an enchantment you control enters, you may put a land card from your hand onto the battlefield tapped. If you control six or more enchantments, instead you may put a creature or land card from your hand onto the battlefield tapped.
		`,
		},
	}
}
