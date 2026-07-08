package c

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// ChampionsOfMinasTirith is the card definition for Champions of Minas Tirith.
//
// Type: Creature — Human Soldier
// Cost: {5}{W}
//
// Oracle text:
//
//	When this creature enters, you become the monarch.
//	At the beginning of combat on each opponent's turn, if you're the monarch, that opponent may pay {X}, where X is the number of cards in their hand. If they don't, they can't attack you this combat.
var ChampionsOfMinasTirith = newChampionsOfMinasTirith

func newChampionsOfMinasTirith() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Champions of Minas Tirith",
			ManaCost: opt.Val(cost.Mana{
				cost.O(5),
				cost.W,
			}),
			Colors:    []color.Color{color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Human, types.Soldier},
			Power:     opt.Val(game.PT{Value: 4}),
			Toughness: opt.Val(game.PT{Value: 6}),
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
							Controller: game.TriggerControllerOpponent,
							Step:       game.StepBeginningOfCombat,
						},
						InterveningIf: "if you're the monarch",
						InterveningCondition: opt.Val(game.Condition{
							ControllerIsMonarch: true,
						}),
					},
					Content: game.Mode{
						Text: "At the beginning of combat on each opponent's turn, if you're the monarch, that opponent may pay {X}, where X is the number of cards in their hand. If they don't, they can't attack you this combat.",
						Sequence: []game.Instruction{
							{
								Primitive: game.PlayerMayPayGenericOrRule{
									Player: game.EventPlayerReference(),
									Amount: game.Dynamic(game.DynamicAmount{
										Kind:      game.DynamicAmountCountCardsInZone,
										Player:    func() *game.PlayerReference { ref := game.EventPlayerReference(); return &ref }(),
										CardZone:  zone.Hand,
										Selection: &game.Selection{},
									}),
									RuleEffects: []game.RuleEffect{
										game.RuleEffect{
											Kind:                      game.RuleEffectCantAttack,
											AffectedPlayerRef:         game.EventPlayerReference(),
											DefendingPlayer:           game.PlayerYou,
											DefendingPlayerDirectOnly: true,
										},
									},
									Duration: game.DurationUntilEndOfCombat,
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			When this creature enters, you become the monarch.
			At the beginning of combat on each opponent's turn, if you're the monarch, that opponent may pay {X}, where X is the number of cards in their hand. If they don't, they can't attack you this combat.
		`,
		},
	}
}
