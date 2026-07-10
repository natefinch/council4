package t

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// ThunderhawkGunship is the card definition for Thunderhawk Gunship.
//
// Type: Artifact — Vehicle
// Cost: {6}
//
// Oracle text:
//
//	Flying
//	When this Vehicle enters, create two 2/2 white Astartes Warrior creature tokens with vigilance.
//	Whenever this Vehicle attacks, attacking creatures you control gain flying until end of turn.
//	Crew 2
var ThunderhawkGunship = newThunderhawkGunship

func newThunderhawkGunship() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Thunderhawk Gunship",
			ManaCost: opt.Val(cost.Mana{
				cost.O(6),
			}),
			Types:     []types.Card{types.Artifact},
			Subtypes:  []types.Sub{types.Vehicle},
			Power:     opt.Val(game.PT{Value: 6}),
			Toughness: opt.Val(game.PT{Value: 6}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
			},
			ActivatedAbilities: []game.ActivatedAbility{
				game.CrewActivatedAbility(2),
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
								Primitive: game.CreateToken{
									Amount: game.Fixed(2),
									Source: game.TokenDef(thunderhawkGunshipToken),
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
						Sequence: []game.Instruction{
							{
								Primitive: game.ApplyContinuous{
									ContinuousEffects: []game.ContinuousEffect{
										game.ContinuousEffect{
											Layer: game.LayerAbility,
											Group: game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}, Controller: game.ControllerYou, CombatState: game.CombatStateAttacking}),
											AddKeywords: []game.Keyword{
												game.Flying,
											},
										},
									},
									Duration: game.DurationUntilEndOfTurn,
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Flying
			When this Vehicle enters, create two 2/2 white Astartes Warrior creature tokens with vigilance.
			Whenever this Vehicle attacks, attacking creatures you control gain flying until end of turn.
			Crew 2
		`,
		},
	}
}

var thunderhawkGunshipToken = newThunderhawkGunshipToken()

func newThunderhawkGunshipToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Astartes Warrior",
			Colors:    []color.Color{color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Astartes, types.Warrior},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 2}),
			StaticAbilities: []game.StaticAbility{
				game.VigilanceStaticBody,
			},
		},
	}
}
