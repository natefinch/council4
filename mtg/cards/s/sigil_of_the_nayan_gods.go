package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// SigilOfTheNayanGods is the card definition for Sigil of the Nayan Gods.
//
// Type: Enchantment — Aura
// Cost: {1}{G}{W}
//
// Oracle text:
//
//	Enchant creature
//	Enchanted creature gets +1/+1 for each creature you control.
//	Cycling {G/W} ({G/W}, Discard this card: Draw a card.)
var SigilOfTheNayanGods = newSigilOfTheNayanGods

func newSigilOfTheNayanGods() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Green),
		CardFace: game.CardFace{
			Name: "Sigil of the Nayan Gods",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.G,
				cost.W,
			}),
			Colors:   []color.Color{color.Green, color.White},
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
					Text: "Enchanted creature gets +1/+1 for each creature you control.",
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Duration: game.DurationPermanent,
							Group:    game.AttachedObjectGroup(game.SourcePermanentReference()),
							Layer:    game.LayerPowerToughnessModify,
							PowerDeltaDynamic: opt.Val(game.DynamicAmount{
								Kind:        game.DynamicAmountCountSelector,
								Multiplier:  1,
								CounterKind: counter.PlusOnePlusOne,
								Group:       game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}, Controller: game.ControllerYou}),
							}),
							ToughnessDeltaDynamic: opt.Val(game.DynamicAmount{
								Kind:        game.DynamicAmountCountSelector,
								Multiplier:  1,
								CounterKind: counter.PlusOnePlusOne,
								Group:       game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}, Controller: game.ControllerYou}),
							}),
						},
					},
				},
			},
			ActivatedAbilities: []game.ActivatedAbility{
				game.CyclingActivatedAbility(cost.Mana{cost.HybridMana(mana.G, mana.W)}),
			},
			OracleText: `
			Enchant creature
			Enchanted creature gets +1/+1 for each creature you control.
			Cycling {G/W} ({G/W}, Discard this card: Draw a card.)
		`,
		},
	}
}
