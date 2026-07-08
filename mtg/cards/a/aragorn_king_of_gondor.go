package a

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// AragornKingOfGondor is the card definition for Aragorn, King of Gondor.
//
// Type: Legendary Creature — Human Noble
// Cost: {1}{U}{R}{W}
//
// Oracle text:
//
//	Vigilance, lifelink
//	When Aragorn enters, you become the monarch.
//	Whenever Aragorn attacks, up to one target creature can't block this turn. If you're the monarch, creatures can't block this turn.
var AragornKingOfGondor = newAragornKingOfGondor

func newAragornKingOfGondor() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Blue, color.Red),
		CardFace: game.CardFace{
			Name: "Aragorn, King of Gondor",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.U,
				cost.R,
				cost.W,
			}),
			Colors:     []color.Color{color.Red, color.Blue, color.White},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Human, types.Noble},
			Power:      opt.Val(game.PT{Value: 4}),
			Toughness:  opt.Val(game.PT{Value: 4}),
			StaticAbilities: []game.StaticAbility{
				game.VigilanceStaticBody,
				game.LifelinkStaticBody,
			},
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
							Event:  game.EventAttackerDeclared,
							Source: game.TriggerSourceSelf,
						},
					},
					Content: game.Mode{
						Targets: []game.TargetSpec{
							game.TargetSpec{
								MinTargets: 0,
								MaxTargets: 1,
								Constraint: "up to one target creature",
								Allow:      game.TargetAllowPermanent,
								Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}}),
							},
						},
						Sequence: []game.Instruction{
							{
								Primitive: game.ApplyRule{
									Object: opt.Val(game.TargetPermanentReference(0)),
									RuleEffects: []game.RuleEffect{
										game.RuleEffect{
											Kind: game.RuleEffectCantBlock,
										},
									},
									Duration: game.DurationThisTurn,
								},
							},
							{
								Primitive: game.ApplyRule{
									RuleEffects: []game.RuleEffect{
										game.RuleEffect{
											Kind:           game.RuleEffectCantBlock,
											PermanentTypes: []types.Card{types.Creature},
										},
									},
									Duration: game.DurationThisTurn,
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
			Vigilance, lifelink
			When Aragorn enters, you become the monarch.
			Whenever Aragorn attacks, up to one target creature can't block this turn. If you're the monarch, creatures can't block this turn.
		`,
		},
	}
}
