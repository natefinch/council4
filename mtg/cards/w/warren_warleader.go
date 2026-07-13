package w

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// WarrenWarleader is the card definition for WarrenWarleader.
//
// Type: Creature — Rabbit Knight
// Cost: {2}{W}{W}
//
// Oracle text:
//
//	Offspring {2} (You may pay an additional {2} as you cast this spell. If you do, when this creature enters, create a 1/1 token copy of it.)
//	Whenever you attack, choose one —
//	• Create a 1/1 white Rabbit creature token that's tapped and attacking.
//	• Attacking creatures you control get +1/+1 until end of turn.
var WarrenWarleader = newWarrenWarleader

func newWarrenWarleader() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Warren Warleader",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.W,
				cost.W,
			}),
			Colors:    []color.Color{color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Rabbit, types.Knight},
			Power:     opt.Val(game.PT{Value: 4}),
			Toughness: opt.Val(game.PT{Value: 4}),
			StaticAbilities: []game.StaticAbility{
				game.OffspringStaticAbility(cost.Mana{cost.O(2)}),
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.OffspringEnterTriggeredAbility(),
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:      game.EventAttackerDeclared,
							Controller: game.TriggerControllerYou,
							OneOrMore:  true,
						},
					},
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Text: "Create a 1/1 white Rabbit creature token that's tapped and attacking.",
								Sequence: []game.Instruction{
									{
										Primitive: game.CreateToken{
											Amount:         game.Fixed(1),
											Source:         game.TokenDef(warrenWarleaderToken),
											EntryTapped:    true,
											EntryAttacking: true,
										},
									},
								},
							},
							game.Mode{
								Text: "Attacking creatures you control get +1/+1 until end of turn.",
								Sequence: []game.Instruction{
									{
										Primitive: game.ApplyContinuous{
											ContinuousEffects: []game.ContinuousEffect{
												game.ContinuousEffect{
													Layer:          game.LayerPowerToughnessModify,
													Group:          game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}, Controller: game.ControllerYou, CombatState: game.CombatStateAttacking}),
													PowerDelta:     1,
													ToughnessDelta: 1,
												},
											},
											Duration: game.DurationUntilEndOfTurn,
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
			Offspring {2} (You may pay an additional {2} as you cast this spell. If you do, when this creature enters, create a 1/1 token copy of it.)
			Whenever you attack, choose one —
			• Create a 1/1 white Rabbit creature token that's tapped and attacking.
			• Attacking creatures you control get +1/+1 until end of turn.
		`,
		},
	}
}

var warrenWarleaderToken = newWarrenWarleaderToken()

func newWarrenWarleaderToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Rabbit",
			Colors:    []color.Color{color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Rabbit},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 1}),
		},
	}
}
