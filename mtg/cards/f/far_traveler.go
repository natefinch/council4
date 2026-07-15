package f

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// FarTraveler is the card definition for Far Traveler.
//
// Type: Legendary Enchantment — Background
// Cost: {2}{W}
//
// Oracle text:
//
//	Commander creatures you own have "At the beginning of your end step, exile up to one target tapped creature you control, then return it to the battlefield under its owner's control."
var FarTraveler = newFarTraveler

func newFarTraveler() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Far Traveler",
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
							Group: game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}, Owner: game.OwnerYou, MatchCommander: true}),
							AddAbilities: []game.Ability{
								new(game.TriggeredAbility{
									Trigger: game.TriggerCondition{
										Type: game.TriggerAt,
										Pattern: game.TriggerPattern{
											Event:      game.EventBeginningOfStep,
											Controller: game.TriggerControllerYou,
											Step:       game.StepEnd,
										},
									},
									Content: game.Mode{
										Targets: []game.TargetSpec{
											game.TargetSpec{
												MinTargets: 0,
												MaxTargets: 1,
												Constraint: "up to one target tapped creature you control",
												Allow:      game.TargetAllowPermanent,
												Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}, Controller: game.ControllerYou, Tapped: game.TriTrue}),
											},
										},
										Sequence: []game.Instruction{
											{
												Primitive: game.Exile{
													Object:         game.TargetPermanentReference(0),
													ExileLinkedKey: game.LinkedKey("blink-1"),
												},
											},
											{
												Primitive: game.PutOnBattlefield{
													Source: game.LinkedBattlefieldSource(game.LinkedKey("blink-1")),
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
			Commander creatures you own have "At the beginning of your end step, exile up to one target tapped creature you control, then return it to the battlefield under its owner's control."
		`,
		},
	}
}
