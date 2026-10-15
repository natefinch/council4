package j

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// JiangYangguWildcrafter is the card definition for Jiang Yanggu, Wildcrafter.
//
// Type: Legendary Planeswalker — Yanggu
// Cost: {2}{G}
//
// Oracle text:
//
//	Each creature you control with a +1/+1 counter on it has "{T}: Add one mana of any color."
//	−1: Put a +1/+1 counter on target creature.
var JiangYangguWildcrafter = newJiangYangguWildcrafter

func newJiangYangguWildcrafter() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Jiang Yanggu, Wildcrafter",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.G,
			}),
			Colors:     []color.Color{color.Green},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Planeswalker},
			Subtypes:   []types.Sub{types.Yanggu},
			Loyalty:    opt.Val(3),
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "Each creature you control with a +1/+1 counter on it has \"{T}: Add one mana of any color.\"",
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Duration: game.DurationPermanent,
							Group:    game.ObjectControlledGroup(game.SourcePermanentReference(), game.Selection{RequiredTypes: []types.Card{types.Creature}, MatchCounter: true, RequiredCounter: counter.PlusOnePlusOne}),
							Layer:    game.LayerAbility,
							AddAbilities: []game.Ability{
								new(game.TapManaChoiceAbility(mana.W, mana.U, mana.B, mana.R, mana.G)),
							},
						},
					},
				},
			},
			LoyaltyAbilities: []game.LoyaltyAbility{
				game.LoyaltyAbility{
					Text:        "−1: Put a +1/+1 counter on target creature.",
					LoyaltyCost: -1,
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
										Primitive: game.AddCounter{
											Amount:      game.Fixed(1),
											Object:      game.TargetPermanentReference(0),
											CounterKind: counter.PlusOnePlusOne,
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
			Each creature you control with a +1/+1 counter on it has "{T}: Add one mana of any color."
			−1: Put a +1/+1 counter on target creature.
		`,
		},
	}
}
