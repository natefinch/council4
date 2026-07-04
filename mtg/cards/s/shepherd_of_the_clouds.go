package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// ShepherdOfTheClouds is the card definition for Shepherd of the Clouds.
//
// Type: Creature — Pegasus
// Cost: {4}{W}
//
// Oracle text:
//
//	Flying, vigilance
//	When this creature enters, return target permanent card with mana value 3 or less from your graveyard to your hand. Return that card to the battlefield instead if you control a Mount.
var ShepherdOfTheClouds = newShepherdOfTheClouds()

func newShepherdOfTheClouds() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Shepherd of the Clouds",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
				cost.W,
			}),
			Colors:    []color.Color{color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Pegasus},
			Power:     opt.Val(game.PT{Value: 4}),
			Toughness: opt.Val(game.PT{Value: 3}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
				game.VigilanceStaticBody,
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
						Targets: []game.TargetSpec{
							game.TargetSpec{
								MinTargets: 1,
								MaxTargets: 1,
								Constraint: "target permanent card with mana value 3 or less from your graveyard",
								Allow:      game.TargetAllowCard,
								TargetZone: zone.Graveyard,
								Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Artifact, types.Creature, types.Enchantment, types.Land, types.Planeswalker, types.Battle}, Controller: game.ControllerYou, ManaValue: opt.Val(compare.Int{Op: compare.LessOrEqual, Value: 3})}),
							},
						},
						Sequence: []game.Instruction{
							{
								Primitive: game.MoveCard{
									Card:        game.CardReference{Kind: game.CardReferenceTarget},
									FromZone:    zone.Graveyard,
									Destination: zone.Hand,
								},
								Condition: opt.Val(game.EffectCondition{
									Condition: opt.Val(game.Condition{
										Negate: true,
										ControlsMatching: opt.Val(game.SelectionCount{
											Selection: game.Selection{SubtypesAny: []types.Sub{types.Sub("Mount")}},
										}),
									}),
								}),
							},
							{
								Primitive: game.PutOnBattlefield{
									Source: game.CardBattlefieldSource(game.CardReference{Kind: game.CardReferenceTarget}),
								},
								Condition: opt.Val(game.EffectCondition{
									Condition: opt.Val(game.Condition{
										ControlsMatching: opt.Val(game.SelectionCount{
											Selection: game.Selection{SubtypesAny: []types.Sub{types.Sub("Mount")}},
										}),
									}),
								}),
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Flying, vigilance
			When this creature enters, return target permanent card with mana value 3 or less from your graveyard to your hand. Return that card to the battlefield instead if you control a Mount.
		`,
		},
	}
}
