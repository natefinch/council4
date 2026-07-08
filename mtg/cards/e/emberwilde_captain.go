package e

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// EmberwildeCaptain is the card definition for Emberwilde Captain.
//
// Type: Creature — Djinn Pirate
// Cost: {3}{R}
//
// Oracle text:
//
//	When this creature enters, you become the monarch.
//	Whenever an opponent attacks you while you're the monarch, this creature deals damage to that player equal to the number of cards in their hand.
var EmberwildeCaptain = newEmberwildeCaptain

func newEmberwildeCaptain() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Emberwilde Captain",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.R,
			}),
			Colors:    []color.Color{color.Red},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Djinn, types.Pirate},
			Power:     opt.Val(game.PT{Value: 4}),
			Toughness: opt.Val(game.PT{Value: 2}),
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
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:                    game.EventAttackerDeclared,
							Controller:               game.TriggerControllerOpponent,
							Player:                   game.TriggerPlayerYou,
							OneOrMore:                true,
							OneOrMorePerAttackTarget: true,
							AttackRecipient:          game.AttackRecipientPlayer,
						},
						InterveningIf: "while you're the monarch",
						InterveningCondition: opt.Val(game.Condition{
							ControllerIsMonarch: true,
						}),
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.Damage{
									Amount: game.Dynamic(game.DynamicAmount{
										Kind:       game.DynamicAmountCountCardsInZone,
										Multiplier: 1,
										Player:     func() *game.PlayerReference { ref := game.EventPlayerReference(); return &ref }(),
										CardZone:   zone.Hand,
										Selection:  &game.Selection{},
									}),
									Recipient:    game.PlayerDamageRecipient(game.EventPlayerReference()),
									DamageSource: opt.Val(game.SourcePermanentReference()),
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			When this creature enters, you become the monarch.
			Whenever an opponent attacks you while you're the monarch, this creature deals damage to that player equal to the number of cards in their hand.
		`,
		},
	}
}
