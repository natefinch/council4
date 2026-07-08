package c

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// CourtOfEmbereth is the card definition for Court of Embereth.
//
// Type: Enchantment
// Cost: {2}{R}{R}
//
// Oracle text:
//
//	When this enchantment enters, you become the monarch.
//	At the beginning of your upkeep, create a 3/1 red Knight creature token. Then if you're the monarch, this enchantment deals X damage to each opponent, where X is the number of creatures you control.
var CourtOfEmbereth = newCourtOfEmbereth

func newCourtOfEmbereth() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Court of Embereth",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.R,
				cost.R,
			}),
			Colors: []color.Color{color.Red},
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
								Primitive: game.CreateToken{
									Amount: game.Fixed(1),
									Source: game.TokenDef(courtOfEmberethToken),
								},
							},
							{
								Primitive: game.Damage{
									Amount: game.Dynamic(game.DynamicAmount{
										Kind:       game.DynamicAmountCountSelector,
										Multiplier: 1,
										Group:      game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}, Controller: game.ControllerYou}),
									}),
									Recipient:    game.PlayerGroupDamageRecipient(game.OpponentsReference()),
									DamageSource: opt.Val(game.SourcePermanentReference()),
								},
								Condition: opt.Val(game.EffectCondition{
									Condition: opt.Val(game.Condition{
										ControllerIsMonarch: true,
									}),
								}),
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			When this enchantment enters, you become the monarch.
			At the beginning of your upkeep, create a 3/1 red Knight creature token. Then if you're the monarch, this enchantment deals X damage to each opponent, where X is the number of creatures you control.
		`,
		},
	}
}

var courtOfEmberethToken = newCourtOfEmberethToken()

func newCourtOfEmberethToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Knight",
			Colors:    []color.Color{color.Red},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Knight},
			Power:     opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 1}),
		},
	}
}
