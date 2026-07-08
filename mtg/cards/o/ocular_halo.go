package o

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// OcularHalo is the card definition for Ocular Halo.
//
// Type: Enchantment — Aura
// Cost: {3}{U}
//
// Oracle text:
//
//	Enchant creature
//	Enchanted creature has "{T}: Draw a card."
//	{W}: Enchanted creature gains vigilance until end of turn.
var OcularHalo = newOcularHalo

func newOcularHalo() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Blue),
		CardFace: game.CardFace{
			Name: "Ocular Halo",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.U,
			}),
			Colors:   []color.Color{color.Blue},
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
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Layer: game.LayerAbility,
							Group: game.AttachedObjectGroup(game.SourcePermanentReference()),
							AddAbilities: []game.Ability{
								new(game.ActivatedAbility{
									Text:            "{T}: Draw a card.",
									AdditionalCosts: cost.Tap,
									ZoneOfFunction:  zone.Battlefield,
									Content: game.Mode{
										Sequence: []game.Instruction{
											{
												Primitive: game.Draw{
													Amount: game.Fixed(1),
													Player: game.ControllerReference(),
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
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text:           "{W}: Enchanted creature gains vigilance until end of turn.",
					ManaCost:       opt.Val(cost.Mana{cost.W}),
					ZoneOfFunction: zone.Battlefield,
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.ApplyContinuous{
									ContinuousEffects: []game.ContinuousEffect{
										game.ContinuousEffect{
											Layer: game.LayerAbility,
											Group: game.AttachedObjectGroup(game.SourcePermanentReference()),
											AddKeywords: []game.Keyword{
												game.Vigilance,
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
			Enchant creature
			Enchanted creature has "{T}: Draw a card."
			{W}: Enchanted creature gains vigilance until end of turn.
		`,
		},
	}
}
