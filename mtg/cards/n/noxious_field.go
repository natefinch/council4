package n

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// NoxiousField is the card definition for Noxious Field.
//
// Type: Enchantment — Aura
// Cost: {1}{B}{B}
//
// Oracle text:
//
//	Enchant land
//	Enchanted land has "{T}: This land deals 1 damage to each creature and each player."
var NoxiousField = newNoxiousField

func newNoxiousField() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Noxious Field",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
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
					Constraint: "land",
					Allow:      game.TargetAllowPermanent,
					Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Land}}),
				}),
				game.StaticAbility{
					Text: "Enchanted land has \"{T}: This land deals 1 damage to each creature and each player.\"",
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Duration: game.DurationPermanent,
							Group:    game.AttachedObjectGroup(game.SourcePermanentReference()),
							Layer:    game.LayerAbility,
							AddAbilities: []game.Ability{
								&game.ActivatedAbility{
									Text: "{T}: This land deals 1 damage to each creature and each player.",
									AdditionalCosts: []cost.Additional{
										cost.Additional{
											Kind:        cost.AdditionalTap,
											CounterKind: counter.PlusOnePlusOne,
										},
									},
									ZoneOfFunction: zone.Battlefield,
									Content: game.AbilityContent{
										Modes: []game.Mode{
											game.Mode{
												Sequence: []game.Instruction{
													game.Instruction{
														Primitive: game.Damage{
															Amount:       game.Fixed(1),
															Recipient:    game.GroupDamageRecipient(game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}})),
															DamageSource: opt.Val(game.SourcePermanentReference()),
														},
													},
													game.Instruction{
														Primitive: game.Damage{
															Amount:       game.Fixed(1),
															Recipient:    game.PlayerGroupDamageRecipient(game.AllPlayersReference()),
															DamageSource: opt.Val(game.SourcePermanentReference()),
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
			Enchant land
			Enchanted land has "{T}: This land deals 1 damage to each creature and each player."
		`,
		},
	}
}
