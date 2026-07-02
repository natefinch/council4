package c

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// CourtOfBounty is the card definition for Court of Bounty.
//
// Type: Enchantment
// Cost: {2}{G}{G}
//
// Oracle text:
//
//	When this enchantment enters, you become the monarch.
//	At the beginning of your upkeep, you may put a land card from your hand onto the battlefield. If you're the monarch, instead you may put a creature or land card from your hand onto the battlefield.
var CourtOfBounty = newCourtOfBounty()

func newCourtOfBounty() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Court of Bounty",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.G,
				cost.G,
			}),
			Colors: []color.Color{color.Green},
			Types:  []types.Card{types.Enchantment},
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
								Primitive: game.BecomeMonarch{
									Player: game.ControllerReference(),
								},
							},
						},
					}.Ability(),
				},
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerAt,
						Pattern: game.TriggerPattern{
							Event:      game.EventBeginningOfStep,
							Controller: game.TriggerControllerYou,
							Step:       game.StepUpkeep,
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
									Prompt: "Choose a card to put onto the battlefield",
								},
								Condition: opt.Val(game.EffectCondition{
									Condition: opt.Val(game.Condition{
										Negate:              true,
										ControllerIsMonarch: true,
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
									Prompt: "Choose a card to put onto the battlefield",
								},
								Condition: opt.Val(game.EffectCondition{
									Condition: opt.Val(game.Condition{
										ControllerIsMonarch: true,
									}),
								}),
								Optional: true,
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			When this enchantment enters, you become the monarch.
			At the beginning of your upkeep, you may put a land card from your hand onto the battlefield. If you're the monarch, instead you may put a creature or land card from your hand onto the battlefield.
		`,
		},
	}
}
