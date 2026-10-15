package l

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// LionSash is the card definition for Lion Sash.
//
// Type: Artifact Creature — Equipment Cat
// Cost: {1}{W}
//
// Oracle text:
//
//	{W}: Exile target card from a graveyard. If it was a permanent card, put a +1/+1 counter on this permanent.
//	Equipped creature gets +1/+1 for each +1/+1 counter on this Equipment.
//	Reconfigure {2} ({2}: Attach to target creature you control; or unattach from a creature. Reconfigure only as a sorcery. While attached, this isn't a creature.)
var LionSash = newLionSash

func newLionSash() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Lion Sash",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.W,
			}),
			Colors:    []color.Color{color.White},
			Types:     []types.Card{types.Artifact, types.Creature},
			Subtypes:  []types.Sub{types.Equipment, types.Cat},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 1}),
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "Equipped creature gets +1/+1 for each +1/+1 counter on this Equipment.",
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Duration: game.DurationPermanent,
							Group:    game.AttachedObjectGroup(game.SourcePermanentReference()),
							Layer:    game.LayerPowerToughnessModify,
							PowerDeltaDynamic: opt.Val(game.DynamicAmount{
								Kind:        game.DynamicAmountObjectCounters,
								Multiplier:  1,
								CounterKind: counter.PlusOnePlusOne,
								Object:      game.SourcePermanentReference(),
							}),
							ToughnessDeltaDynamic: opt.Val(game.DynamicAmount{
								Kind:        game.DynamicAmountObjectCounters,
								Multiplier:  1,
								CounterKind: counter.PlusOnePlusOne,
								Object:      game.SourcePermanentReference(),
							}),
						},
					},
				},
			},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "{W}: Exile target card from a graveyard. If it was a permanent card, put a +1/+1 counter on this permanent.",
					ManaCost: opt.Val(cost.Mana{
						cost.W,
					}),
					ZoneOfFunction: zone.Battlefield,
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Targets: []game.TargetSpec{
									game.TargetSpec{
										MinTargets: 1,
										MaxTargets: 1,
										Constraint: "target card from a graveyard",
										Allow:      game.TargetAllowCard,
										Selection:  opt.Val(game.Selection{}),
										TargetZone: zone.Graveyard,
									},
								},
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.MoveCard{
											Card: game.CardReference{
												Kind: game.CardReferenceTarget,
											},
											FromZone:    zone.Graveyard,
											Destination: zone.Exile,
										},
										PublishResult: game.ResultKey("condition-target-card-1"),
									},
									game.Instruction{
										Primitive: game.AddCounter{
											Amount:      game.Fixed(1),
											Object:      game.SourcePermanentReference(),
											CounterKind: counter.PlusOnePlusOne,
										},
										Condition: opt.Val(game.EffectCondition{
											Condition: opt.Val(game.Condition{
												Text:                "If it was a permanent card",
												Object:              opt.Val(game.TargetCardReference(0)),
												ObjectMatches:       opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Artifact, types.Battle, types.Creature, types.Enchantment, types.Land, types.Planeswalker}}),
												TargetCardResultKey: game.ResultKey("condition-target-card-1"),
												SourceCounterKind:   counter.PlusOnePlusOne,
											}),
										}),
									},
								},
							},
						},
						MinModes: 1,
						MaxModes: 1,
					},
				},
				game.ReconfigureActivatedAbility(cost.Mana{cost.O(2)}),
			},
			OracleText: `
			{W}: Exile target card from a graveyard. If it was a permanent card, put a +1/+1 counter on this permanent.
			Equipped creature gets +1/+1 for each +1/+1 counter on this Equipment.
			Reconfigure {2} ({2}: Attach to target creature you control; or unattach from a creature. Reconfigure only as a sorcery. While attached, this isn't a creature.)
		`,
		},
	}
}
