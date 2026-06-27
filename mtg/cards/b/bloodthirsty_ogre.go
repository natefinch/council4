package b

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// BloodthirstyOgre is the card definition for Bloodthirsty Ogre.
//
// Type: Creature — Ogre Warrior Shaman
// Cost: {2}{B}
//
// Oracle text:
//
//	{T}: Put a devotion counter on this creature.
//	{T}: Target creature gets -X/-X until end of turn, where X is the number of devotion counters on this creature. Activate only if you control a Demon.
var BloodthirstyOgre = newBloodthirstyOgre()

func newBloodthirstyOgre() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Bloodthirsty Ogre",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.B,
			}),
			Colors:    []color.Color{color.Black},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Ogre, types.Warrior, types.Shaman},
			Power:     opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 1}),
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text:            "{T}: Put a devotion counter on this creature.",
					AdditionalCosts: cost.Tap,
					ZoneOfFunction:  zone.Battlefield,
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.AddCounter{
									Amount:      game.Fixed(1),
									Object:      game.SourcePermanentReference(),
									CounterKind: counter.Devotion,
								},
							},
						},
					}.Ability(),
				},
				game.ActivatedAbility{
					Text:            "{T}: Target creature gets -X/-X until end of turn, where X is the number of devotion counters on this creature. Activate only if you control a Demon.",
					AdditionalCosts: cost.Tap,
					ZoneOfFunction:  zone.Battlefield,
					ActivationCondition: opt.Val(game.Condition{
						ControlsMatching: opt.Val(game.SelectionCount{
							Selection: game.Selection{SubtypesAny: []types.Sub{types.Sub("Demon")}},
						}),
					}),
					Content: game.Mode{
						Targets: []game.TargetSpec{
							game.TargetSpec{
								MinTargets: 1,
								MaxTargets: 1,
								Constraint: "target creature",
								Allow:      game.TargetAllowPermanent,
								Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}}),
							},
						},
						Sequence: []game.Instruction{
							{
								Primitive: game.ModifyPT{
									Object: game.TargetPermanentReference(0),
									PowerDelta: game.Dynamic(game.DynamicAmount{
										Kind:        game.DynamicAmountObjectCounters,
										Multiplier:  -1,
										CounterKind: counter.Devotion,
										Object:      game.SourcePermanentReference(),
									}),
									ToughnessDelta: game.Dynamic(game.DynamicAmount{
										Kind:        game.DynamicAmountObjectCounters,
										Multiplier:  -1,
										CounterKind: counter.Devotion,
										Object:      game.SourcePermanentReference(),
									}),
									Duration: game.DurationUntilEndOfTurn,
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			{T}: Put a devotion counter on this creature.
			{T}: Target creature gets -X/-X until end of turn, where X is the number of devotion counters on this creature. Activate only if you control a Demon.
		`,
		},
	}
}
