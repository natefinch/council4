package p

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// PopularEntertainer is the card definition for Popular Entertainer.
//
// Type: Legendary Enchantment — Background
// Cost: {1}{R}
//
// Oracle text:
//
//	Commander creatures you own have "Whenever one or more creatures you control deal combat damage to a player, goad target creature that player controls." (Until your next turn, that creature attacks each combat if able and attacks a player other than you if able.)
var PopularEntertainer = newPopularEntertainer

func newPopularEntertainer() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Popular Entertainer",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.R,
			}),
			Colors:     []color.Color{color.Red},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Enchantment},
			Subtypes:   []types.Sub{types.Background},
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Layer: game.LayerAbility,
							Group: game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}, Owner: game.OwnerYou, MatchCommander: true}),
							AddAbilities: []game.Ability{
								new(game.TriggeredAbility{
									Trigger: game.TriggerCondition{
										Type: game.TriggerWhenever,
										Pattern: game.TriggerPattern{
											Event:                 game.EventDamageDealt,
											Controller:            game.TriggerControllerYou,
											Subject:               game.TriggerSubjectDamageSource,
											OneOrMore:             true,
											RequireCombatDamage:   true,
											DamageRecipient:       game.DamageRecipientPlayer,
											DamageSourceSelection: game.Selection{RequiredTypes: []types.Card{types.Creature}},
										},
									},
									Content: game.Mode{
										Targets: []game.TargetSpec{
											game.TargetSpec{
												MinTargets: 1,
												MaxTargets: 1,
												Constraint: "target creature that player controls",
												Allow:      game.TargetAllowPermanent,
												Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}, ControlledByEventPlayer: true}),
											},
										},
										Sequence: []game.Instruction{
											{
												Primitive: game.Goad{
													Object: game.TargetPermanentReference(0),
												},
											},
										},
									}.Ability(),
								}),
							},
						},
					},
				},
			},
			OracleText: `
			Commander creatures you own have "Whenever one or more creatures you control deal combat damage to a player, goad target creature that player controls." (Until your next turn, that creature attacks each combat if able and attacks a player other than you if able.)
		`,
		},
	}
}
