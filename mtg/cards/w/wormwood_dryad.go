package w

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// WormwoodDryad is the card definition for Wormwood Dryad.
//
// Type: Creature — Dryad
// Cost: {2}{G}
//
// Oracle text:
//
//	{G}: This creature gains forestwalk until end of turn and deals 1 damage to you. (It can't be blocked as long as defending player controls a Forest.)
//	{B}: This creature gains swampwalk until end of turn and deals 1 damage to you. (It can't be blocked as long as defending player controls a Swamp.)
var WormwoodDryad = newWormwoodDryad

func newWormwoodDryad() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black, color.Green),
		CardFace: game.CardFace{
			Name: "Wormwood Dryad",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.G,
			}),
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Dryad},
			Power:     opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 1}),
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "{G}: This creature gains forestwalk until end of turn and deals 1 damage to you. (It can't be blocked as long as defending player controls a Forest.)",
					ManaCost: opt.Val(cost.Mana{
						cost.G,
					}),
					ZoneOfFunction: zone.Battlefield,
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.ApplyContinuous{
											Object: opt.Val(game.SourcePermanentReference()),
											ContinuousEffects: []game.ContinuousEffect{
												game.ContinuousEffect{
													Duration: game.DurationPermanent,
													Layer:    game.LayerAbility,
													AddAbilities: []game.Ability{
														&game.StaticAbility{
															Text: "Forestwalk",
															KeywordAbilities: []game.KeywordAbility{
																game.LandwalkKeyword{
																	Subtype: types.Forest,
																},
															},
														},
													},
												},
											},
											Duration: game.DurationUntilEndOfTurn,
										},
									},
									game.Instruction{
										Primitive: game.Damage{
											Amount:       game.Fixed(1),
											Recipient:    game.PlayerDamageRecipient(game.ControllerReference()),
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
				game.ActivatedAbility{
					Text: "{B}: This creature gains swampwalk until end of turn and deals 1 damage to you. (It can't be blocked as long as defending player controls a Swamp.)",
					ManaCost: opt.Val(cost.Mana{
						cost.B,
					}),
					ZoneOfFunction: zone.Battlefield,
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.ApplyContinuous{
											Object: opt.Val(game.SourcePermanentReference()),
											ContinuousEffects: []game.ContinuousEffect{
												game.ContinuousEffect{
													Duration: game.DurationPermanent,
													Layer:    game.LayerAbility,
													AddAbilities: []game.Ability{
														&game.StaticAbility{
															Text: "Swampwalk",
															KeywordAbilities: []game.KeywordAbility{
																game.LandwalkKeyword{
																	Subtype: types.Swamp,
																},
															},
														},
													},
												},
											},
											Duration: game.DurationUntilEndOfTurn,
										},
									},
									game.Instruction{
										Primitive: game.Damage{
											Amount:       game.Fixed(1),
											Recipient:    game.PlayerDamageRecipient(game.ControllerReference()),
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
			OracleText: `
			{G}: This creature gains forestwalk until end of turn and deals 1 damage to you. (It can't be blocked as long as defending player controls a Forest.)
			{B}: This creature gains swampwalk until end of turn and deals 1 damage to you. (It can't be blocked as long as defending player controls a Swamp.)
		`,
		},
	}
}
