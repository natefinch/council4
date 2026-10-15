package p

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// PriestOfTheHauntedEdge is the card definition for Priest of the Haunted Edge.
//
// Type: Snow Creature — Zombie Cleric
// Cost: {1}{B}
//
// Oracle text:
//
//	{T}, Sacrifice this creature: Target creature gets -X/-X until end of turn, where X is the number of snow lands you control. Activate only as a sorcery.
var PriestOfTheHauntedEdge = newPriestOfTheHauntedEdge

func newPriestOfTheHauntedEdge() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Priest of the Haunted Edge",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.B,
			}),
			Colors:     []color.Color{color.Black},
			Supertypes: []types.Super{types.Snow},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Zombie, types.Cleric},
			Power:      opt.Val(game.PT{Value: 0}),
			Toughness:  opt.Val(game.PT{Value: 4}),
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "{T}, Sacrifice this creature: Target creature gets -X/-X until end of turn, where X is the number of snow lands you control. Activate only as a sorcery.",
					AdditionalCosts: []cost.Additional{
						cost.Additional{
							Kind:        cost.AdditionalTap,
							CounterKind: counter.PlusOnePlusOne,
						},
						cost.Additional{
							Kind:        cost.AdditionalSacrificeSource,
							Text:        "Sacrifice this creature",
							Amount:      1,
							CounterKind: counter.PlusOnePlusOne,
						},
					},
					ZoneOfFunction: zone.Battlefield,
					Timing:         game.SorceryOnly,
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
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
									game.Instruction{
										Primitive: game.ModifyPT{
											Object: game.TargetPermanentReference(0),
											PowerDelta: game.Dynamic(game.DynamicAmount{
												Kind:       game.DynamicAmountCountSelector,
												Multiplier: -1,
												Group:      game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Land}, Supertypes: []types.Super{types.Snow}, Controller: game.ControllerYou}),
											}),
											ToughnessDelta: game.Dynamic(game.DynamicAmount{
												Kind:       game.DynamicAmountCountSelector,
												Multiplier: -1,
												Group:      game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Land}, Supertypes: []types.Super{types.Snow}, Controller: game.ControllerYou}),
											}),
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
			OracleText: `
			{T}, Sacrifice this creature: Target creature gets -X/-X until end of turn, where X is the number of snow lands you control. Activate only as a sorcery.
		`,
		},
	}
}
