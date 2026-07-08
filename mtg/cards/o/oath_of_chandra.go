package o

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// OathOfChandra is the card definition for Oath of Chandra.
//
// Type: Legendary Enchantment
// Cost: {1}{R}
//
// Oracle text:
//
//	When Oath of Chandra enters, it deals 3 damage to target creature an opponent controls.
//	At the beginning of each end step, if a planeswalker entered the battlefield under your control this turn, Oath of Chandra deals 2 damage to each opponent.
var OathOfChandra = newOathOfChandra

func newOathOfChandra() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Oath of Chandra",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.R,
			}),
			Colors:     []color.Color{color.Red},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Enchantment},
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
						Targets: []game.TargetSpec{
							game.TargetSpec{
								MinTargets: 1,
								MaxTargets: 1,
								Constraint: "target creature an opponent controls",
								Allow:      game.TargetAllowPermanent,
								Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}, Controller: game.ControllerOpponent}),
							},
						},
						Sequence: []game.Instruction{
							{
								Primitive: game.Damage{
									Amount:       game.Fixed(3),
									Recipient:    game.AnyTargetDamageRecipient(0),
									DamageSource: opt.Val(game.EventPermanentReference()),
								},
							},
						},
					}.Ability(),
				},
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerAt,
						Pattern: game.TriggerPattern{
							Event: game.EventBeginningOfStep,
							Step:  game.StepEnd,
						},
						InterveningIf: "if a planeswalker entered the battlefield under your control this turn",
						InterveningCondition: opt.Val(game.Condition{
							EventHistory: opt.Val(game.EventHistoryCondition{Pattern: game.TriggerPattern{
								Event:            game.EventPermanentEnteredBattlefield,
								Controller:       game.TriggerControllerYou,
								SubjectSelection: game.Selection{RequiredTypes: []types.Card{types.Planeswalker}},
							}, Window: game.EventHistoryCurrentTurn}),
						}),
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.Damage{
									Amount:    game.Fixed(2),
									Recipient: game.PlayerGroupDamageRecipient(game.OpponentsReference()),
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			When Oath of Chandra enters, it deals 3 damage to target creature an opponent controls.
			At the beginning of each end step, if a planeswalker entered the battlefield under your control this turn, Oath of Chandra deals 2 damage to each opponent.
		`,
		},
	}
}
