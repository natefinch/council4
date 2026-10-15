package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// SparkhunterMasticore is the card definition for Sparkhunter Masticore.
//
// Type: Artifact Creature — Masticore
// Cost: {3}
//
// Oracle text:
//
//	As an additional cost to cast this spell, discard a card.
//	Protection from planeswalkers
//	{1}: This creature deals 1 damage to target planeswalker.
//	{3}: This creature gains indestructible until end of turn.
var SparkhunterMasticore = newSparkhunterMasticore

func newSparkhunterMasticore() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Sparkhunter Masticore",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
			}),
			Types:     []types.Card{types.Artifact, types.Creature},
			Subtypes:  []types.Sub{types.Masticore},
			Power:     opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 4}),
			StaticAbilities: []game.StaticAbility{
				game.ProtectionFromTypesStaticAbility(types.Planeswalker),
			},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "{1}: This creature deals 1 damage to target planeswalker.",
					ManaCost: opt.Val(cost.Mana{
						cost.O(1),
					}),
					ZoneOfFunction: zone.Battlefield,
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Targets: []game.TargetSpec{
									game.TargetSpec{
										MinTargets: 1,
										MaxTargets: 1,
										Constraint: "target planeswalker",
										Allow:      game.TargetAllowPermanent,
										Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Planeswalker}}),
									},
								},
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.Damage{
											Amount:       game.Fixed(1),
											Recipient:    game.AnyTargetDamageRecipient(0),
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
					Text: "{3}: This creature gains indestructible until end of turn.",
					ManaCost: opt.Val(cost.Mana{
						cost.O(3),
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
													Duration:    game.DurationPermanent,
													Layer:       game.LayerAbility,
													AddKeywords: []game.Keyword{game.Indestructible},
												},
											},
											Duration: game.DurationUntilEndOfTurn,
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
			AdditionalCosts: []cost.Additional{
				{
					Kind:   cost.AdditionalDiscard,
					Text:   "discard a card",
					Amount: 1,
					Source: zone.Hand,
				},
			},
			OracleText: `
			As an additional cost to cast this spell, discard a card.
			Protection from planeswalkers
			{1}: This creature deals 1 damage to target planeswalker.
			{3}: This creature gains indestructible until end of turn.
		`,
		},
	}
}
