package c

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// CommanderSAuthority is the card definition for Commander's Authority.
//
// Type: Enchantment — Aura
// Cost: {4}{W}
//
// Oracle text:
//
//	Enchant creature
//	Enchanted creature has "At the beginning of your upkeep, create a 1/1 white Human creature token."
var CommanderSAuthority = newCommanderSAuthority

func newCommanderSAuthority() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Commander's Authority",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
				cost.W,
			}),
			Colors:   []color.Color{color.White},
			Types:    []types.Card{types.Enchantment},
			Subtypes: []types.Sub{types.Aura},
			StaticAbilities: []game.StaticAbility{
				game.EnchantStaticAbility(&game.TargetSpec{
					MinTargets: 1,
					MaxTargets: 1,
					Constraint: "creature",
					Allow:      game.TargetAllowPermanent,
					Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}}),
				}),
				game.StaticAbility{
					Text: "Enchanted creature has \"At the beginning of your upkeep, create a 1/1 white Human creature token.\"",
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Duration: game.DurationPermanent,
							Group:    game.AttachedObjectGroup(game.SourcePermanentReference()),
							Layer:    game.LayerAbility,
							AddAbilities: []game.Ability{
								&game.TriggeredAbility{
									Text: "At the beginning of your upkeep, create a 1/1 white Human creature token.",
									Trigger: game.TriggerCondition{
										Type: game.TriggerAt,
										Pattern: game.TriggerPattern{
											Event:           game.EventBeginningOfStep,
											Controller:      game.TriggerControllerYou,
											StackObjectKind: game.StackSpell,
											CounterKind:     counter.PlusOnePlusOne,
											Step:            game.StepUpkeep,
										},
									},
									Content: game.AbilityContent{
										Modes: []game.Mode{
											game.Mode{
												Sequence: []game.Instruction{
													game.Instruction{
														Primitive: game.CreateToken{
															Amount: game.Fixed(1),
															Source: game.TokenDef(commanderSAuthorityToken),
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
						},
					},
				},
			},
			OracleText: `
			Enchant creature
			Enchanted creature has "At the beginning of your upkeep, create a 1/1 white Human creature token."
		`,
		},
	}
}

var commanderSAuthorityToken = newCommanderSAuthorityToken()

func newCommanderSAuthorityToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Human",
			Colors:    []color.Color{color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Human},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 1}),
		},
	}
}
