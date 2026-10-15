package o

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// OblivionCrown is the card definition for Oblivion Crown.
//
// Type: Enchantment — Aura
// Cost: {1}{B}
//
// Oracle text:
//
//	Flash (You may cast this spell any time you could cast an instant.)
//	Enchant creature
//	Enchanted creature has "Discard a card: This creature gets +1/+1 until end of turn."
var OblivionCrown = newOblivionCrown

func newOblivionCrown() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Oblivion Crown",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.B,
			}),
			Colors:   []color.Color{color.Black},
			Types:    []types.Card{types.Enchantment},
			Subtypes: []types.Sub{types.Aura},
			StaticAbilities: []game.StaticAbility{
				game.FlashStaticBody,
				game.EnchantStaticAbility(&game.TargetSpec{
					MinTargets: 1,
					MaxTargets: 1,
					Constraint: "creature",
					Allow:      game.TargetAllowPermanent,
					Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}}),
				}),
				game.StaticAbility{
					Text: "Enchanted creature has \"Discard a card: This creature gets +1/+1 until end of turn.\"",
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Duration: game.DurationPermanent,
							Group:    game.AttachedObjectGroup(game.SourcePermanentReference()),
							Layer:    game.LayerAbility,
							AddAbilities: []game.Ability{
								&game.ActivatedAbility{
									Text: "Discard a card: This creature gets +1/+1 until end of turn.",
									AdditionalCosts: []cost.Additional{
										cost.Additional{
											Kind:        cost.AdditionalDiscard,
											Text:        "Discard a card",
											Amount:      1,
											Source:      zone.Hand,
											CounterKind: counter.PlusOnePlusOne,
										},
									},
									ZoneOfFunction: zone.Battlefield,
									Content: game.AbilityContent{
										Modes: []game.Mode{
											game.Mode{
												Sequence: []game.Instruction{
													game.Instruction{
														Primitive: game.ModifyPT{
															Object:         game.SourcePermanentReference(),
															PowerDelta:     game.Fixed(1),
															ToughnessDelta: game.Fixed(1),
															Duration:       game.DurationUntilEndOfTurn,
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
			Flash (You may cast this spell any time you could cast an instant.)
			Enchant creature
			Enchanted creature has "Discard a card: This creature gets +1/+1 until end of turn."
		`,
		},
	}
}
