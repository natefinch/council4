package w

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// WarrenWarleader is the card definition for Warren Warleader.
//
// Type: Token Creature — Rabbit Knight
// Cost: {2}{W}{W}
//
// Oracle text:
//   Whenever you attack, choose one —
//   • Create a 1/1 white Rabbit creature token that's tapped and attacking.
//   • Attacking creatures you control get +1/+1 until end of turn.
//   (This token's mana cost is {2}{W}{W}.)

// WarrenWarleaderToken0dd6ba35738741728b40dd4039538a4b is the card definition for Warren Warleader.
var WarrenWarleaderToken0dd6ba35738741728b40dd4039538a4b = newWarrenWarleaderToken0dd6ba35738741728b40dd4039538a4b()

func newWarrenWarleaderToken0dd6ba35738741728b40dd4039538a4b() *game.CardDef {
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
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 1}),
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Text: "Whenever you attack, choose one —\n• Create a 1/1 white Rabbit creature token that's tapped and attacking.\n• Attacking creatures you control get +1/+1 until end of turn.",
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:           game.EventAttackerDeclared,
							Controller:      game.TriggerControllerYou,
							StackObjectKind: game.StackSpell,
							OneOrMore:       true,
							CounterKind:     counter.PlusOnePlusOne,
						},
					},
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Text: "Create a 1/1 white Rabbit creature token that's tapped and attacking.",
								Sequence: []game.Instruction{
									game.Instruction{
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
									game.Instruction{
										Primitive: game.ApplyContinuous{
											ContinuousEffects: []game.ContinuousEffect{
												game.ContinuousEffect{
													Duration:       game.DurationPermanent,
													Group:          game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}, Controller: game.ControllerYou, CombatState: game.CombatStateAttacking}),
													Layer:          game.LayerPowerToughnessModify,
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
			Whenever you attack, choose one —
			• Create a 1/1 white Rabbit creature token that's tapped and attacking.
			• Attacking creatures you control get +1/+1 until end of turn.
			(This token's mana cost is {2}{W}{W}.)
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
