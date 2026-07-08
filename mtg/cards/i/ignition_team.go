package i

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// IgnitionTeam is the card definition for Ignition Team.
//
// Type: Creature — Goblin Warrior
// Cost: {5}{R}{R}
//
// Oracle text:
//
//	This creature enters with X +1/+1 counters on it, where X is the number of tapped lands on the battlefield.
//	{2}{R}, Remove a +1/+1 counter from this creature: Target land becomes a 4/4 red Elemental creature until end of turn. It's still a land.
var IgnitionTeam = newIgnitionTeam

func newIgnitionTeam() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Ignition Team",
			ManaCost: opt.Val(cost.Mana{
				cost.O(5),
				cost.R,
				cost.R,
			}),
			Colors:    []color.Color{color.Red},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Goblin, types.Warrior},
			Power:     opt.Val(game.PT{Value: 0}),
			Toughness: opt.Val(game.PT{Value: 0}),
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text:     "{2}{R}, Remove a +1/+1 counter from this creature: Target land becomes a 4/4 red Elemental creature until end of turn. It's still a land.",
					ManaCost: opt.Val(cost.Mana{cost.O(2), cost.R}),
					AdditionalCosts: []cost.Additional{
						{
							Kind:        cost.AdditionalRemoveCounter,
							Text:        "Remove a +1/+1 counter from this creature",
							Amount:      1,
							CounterKind: counter.PlusOnePlusOne,
						},
					},
					ZoneOfFunction: zone.Battlefield,
					Content: game.Mode{
						Targets: []game.TargetSpec{
							game.TargetSpec{
								MinTargets: 1,
								MaxTargets: 1,
								Constraint: "target land",
								Allow:      game.TargetAllowPermanent,
								Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Land}}),
							},
						},
						Sequence: []game.Instruction{
							{
								Primitive: game.ApplyContinuous{
									Object: opt.Val(game.TargetPermanentReference(0)),
									ContinuousEffects: []game.ContinuousEffect{
										game.ContinuousEffect{
											Layer:     game.LayerColor,
											SetColors: []color.Color{color.Red},
										},
										game.ContinuousEffect{
											Layer:       game.LayerType,
											AddTypes:    []types.Card{types.Creature},
											AddSubtypes: []types.Sub{types.Elemental},
										},
										game.ContinuousEffect{
											Layer:        game.LayerPowerToughnessSet,
											SetPower:     opt.Val(game.PT{Value: 4}),
											SetToughness: opt.Val(game.PT{Value: 4}),
										},
									},
									Duration: game.DurationUntilEndOfTurn,
								},
							},
						},
					}.Ability(),
				},
			},
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntersTappedWithCountersReplacement("This creature enters with X +1/+1 counters on it, where X is the number of tapped lands on the battlefield.", game.CounterPlacement{Kind: counter.PlusOnePlusOne, Dynamic: opt.Val(&game.DynamicAmount{
					Kind:       game.DynamicAmountCountSelector,
					Multiplier: 1,
					Group:      game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Land}, Tapped: game.TriTrue}),
				})}),
			},
			OracleText: `
			This creature enters with X +1/+1 counters on it, where X is the number of tapped lands on the battlefield.
			{2}{R}, Remove a +1/+1 counter from this creature: Target land becomes a 4/4 red Elemental creature until end of turn. It's still a land.
		`,
		},
	}
}
