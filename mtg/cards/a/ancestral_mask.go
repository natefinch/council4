package a

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// AncestralMask is the card definition for Ancestral Mask.
//
// Type: Enchantment — Aura
// Cost: {2}{G}
//
// Oracle text:
//
//	Enchant creature
//	Enchanted creature gets +2/+2 for each other enchantment on the battlefield.
var AncestralMask = newAncestralMask

func newAncestralMask() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Ancestral Mask",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.G,
			}),
			Colors:   []color.Color{color.Green},
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
					Text: "Enchanted creature gets +2/+2 for each other enchantment on the battlefield.",
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Duration: game.DurationPermanent,
							Group:    game.AttachedObjectGroup(game.SourcePermanentReference()),
							Layer:    game.LayerPowerToughnessModify,
							PowerDeltaDynamic: opt.Val(game.DynamicAmount{
								Kind:        game.DynamicAmountCountSelector,
								Multiplier:  2,
								CounterKind: counter.PlusOnePlusOne,
								Group:       game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Enchantment}, ExcludeSource: true}),
							}),
							ToughnessDeltaDynamic: opt.Val(game.DynamicAmount{
								Kind:        game.DynamicAmountCountSelector,
								Multiplier:  2,
								CounterKind: counter.PlusOnePlusOne,
								Group:       game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Enchantment}, ExcludeSource: true}),
							}),
						},
					},
				},
			},
			OracleText: `
			Enchant creature
			Enchanted creature gets +2/+2 for each other enchantment on the battlefield.
		`,
		},
	}
}
