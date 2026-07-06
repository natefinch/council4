package c

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// CanalCourier is the card definition for Canal Courier.
//
// Type: Creature — Human Rogue
// Cost: {5}{U}
//
// Oracle text:
//
//	When this creature enters, you become the monarch.
//	Whenever this creature and another creature attack different players, this creature can't be blocked this combat.
var CanalCourier = newCanalCourier()

func newCanalCourier() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Canal Courier",
			ManaCost: opt.Val(cost.Mana{
				cost.O(5),
				cost.U,
			}),
			Colors:    []color.Color{color.Blue},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Human, types.Rogue},
			Power:     opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 5}),
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
							Event:                             game.EventAttackerDeclared,
							Source:                            game.TriggerSourceSelf,
							AttacksDifferentPlayerThanAnother: true,
						},
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.ApplyRule{
									Object: opt.Val(game.SourcePermanentReference()),
									RuleEffects: []game.RuleEffect{
										game.RuleEffect{
											Kind: game.RuleEffectCantBeBlocked,
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
			Whenever this creature and another creature attack different players, this creature can't be blocked this combat.
		`,
		},
	}
}
