package c

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// ConsecratedByBlood is the card definition for Consecrated by Blood.
//
// Type: Enchantment — Aura
// Cost: {2}{B}{B}
//
// Oracle text:
//
//	Enchant creature
//	Enchanted creature gets +2/+2 and has flying and "Sacrifice two other creatures: Regenerate this creature." (The next time the creature would be destroyed this turn, instead tap it, remove it from combat, and heal all damage on it.)
var ConsecratedByBlood = newConsecratedByBlood

func newConsecratedByBlood() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Consecrated by Blood",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.B,
				cost.B,
			}),
			Colors:   []color.Color{color.Black},
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
					Text: "Enchanted creature gets +2/+2 and has flying and \"Sacrifice two other creatures: Regenerate this creature.\" (The next time the creature would be destroyed this turn, instead tap it, remove it from combat, and heal all damage on it.)",
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Duration:       game.DurationPermanent,
							Group:          game.AttachedObjectGroup(game.SourcePermanentReference()),
							Layer:          game.LayerPowerToughnessModify,
							PowerDelta:     2,
							ToughnessDelta: 2,
						},
						game.ContinuousEffect{
							Duration:    game.DurationPermanent,
							Group:       game.AttachedObjectGroup(game.SourcePermanentReference()),
							Layer:       game.LayerAbility,
							AddKeywords: []game.Keyword{game.Flying},
						},
						game.ContinuousEffect{
							Duration: game.DurationPermanent,
							Group:    game.AttachedObjectGroup(game.SourcePermanentReference()),
							Layer:    game.LayerAbility,
							AddAbilities: []game.Ability{
								&game.ActivatedAbility{
									Text: "Sacrifice two other creatures: Regenerate this creature.",
									AdditionalCosts: []cost.Additional{
										cost.Additional{
											Kind:               cost.AdditionalSacrifice,
											Text:               "Sacrifice two other creatures",
											Amount:             2,
											MatchPermanentType: true,
											PermanentType:      types.Creature,
											CounterKind:        counter.PlusOnePlusOne,
											ExcludeSource:      true,
										},
									},
									ZoneOfFunction: zone.Battlefield,
									Content: game.AbilityContent{
										Modes: []game.Mode{
											game.Mode{
												Sequence: []game.Instruction{
													game.Instruction{
														Primitive: game.Regenerate{
															Object: game.SourcePermanentReference(),
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
			Enchanted creature gets +2/+2 and has flying and "Sacrifice two other creatures: Regenerate this creature." (The next time the creature would be destroyed this turn, instead tap it, remove it from combat, and heal all damage on it.)
		`,
		},
	}
}
