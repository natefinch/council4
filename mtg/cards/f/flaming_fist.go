package f

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// FlamingFist is the card definition for Flaming Fist.
//
// Type: Legendary Enchantment — Background
// Cost: {2}{W}
//
// Oracle text:
//
//	Commander creatures you own have "Whenever this creature attacks, it gains double strike until end of turn."
var FlamingFist = newFlamingFist

func newFlamingFist() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Flaming Fist",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.W,
			}),
			Colors:     []color.Color{color.White},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Enchantment},
			Subtypes:   []types.Sub{types.Background},
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Layer: game.LayerAbility,
							Group: game.ObjectControlledGroup(game.SourcePermanentReference(), game.Selection{RequiredTypes: []types.Card{types.Creature}, MatchCommander: true}),
							AddAbilities: []game.Ability{
								new(game.TriggeredAbility{
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
													Object: opt.Val(game.EventPermanentReference()),
													ContinuousEffects: []game.ContinuousEffect{
														game.ContinuousEffect{
															Layer: game.LayerAbility,
															AddKeywords: []game.Keyword{
																game.DoubleStrike,
															},
														},
													},
													Duration: game.DurationUntilEndOfTurn,
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
			Commander creatures you own have "Whenever this creature attacks, it gains double strike until end of turn."
		`,
		},
	}
}
