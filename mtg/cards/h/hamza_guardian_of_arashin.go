package h

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// HamzaGuardianOfArashin is the card definition for Hamza, Guardian of Arashin.
//
// Type: Legendary Creature — Elephant Warrior
// Cost: {4}{G}{W}
//
// Oracle text:
//
//	This spell costs {1} less to cast for each creature you control with a +1/+1 counter on it.
//	Creature spells you cast cost {1} less to cast for each creature you control with a +1/+1 counter on it.
var HamzaGuardianOfArashin = newHamzaGuardianOfArashin()

func newHamzaGuardianOfArashin() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Green),
		CardFace: game.CardFace{
			Name: "Hamza, Guardian of Arashin",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
				cost.G,
				cost.W,
			}),
			Colors:     []color.Color{color.Green, color.White},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Elephant, types.Warrior},
			Power:      opt.Val(game.PT{Value: 5}),
			Toughness:  opt.Val(game.PT{Value: 5}),
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					RuleEffects: []game.RuleEffect{
						game.RuleEffect{
							Kind:           game.RuleEffectCostModifier,
							AffectedSource: true,
							CostModifier: game.CostModifier{
								Kind:               game.CostModifierSpell,
								PerObjectReduction: 1,
								CountSelection:     &game.Selection{RequiredTypes: []types.Card{types.Creature}, Controller: game.ControllerYou, MatchCounter: true, RequiredCounter: counter.PlusOnePlusOne},
							},
						},
					},
				},
				game.StaticAbility{
					RuleEffects: []game.RuleEffect{
						game.RuleEffect{
							Kind:           game.RuleEffectCostModifier,
							AffectedPlayer: game.PlayerYou,
							CostModifier: game.CostModifier{
								Kind:               game.CostModifierSpell,
								CardSelection:      game.Selection{RequiredTypes: []types.Card{types.Creature}},
								PerObjectReduction: 1,
								CountSelection:     &game.Selection{RequiredTypes: []types.Card{types.Creature}, Controller: game.ControllerYou, MatchCounter: true, RequiredCounter: counter.PlusOnePlusOne},
							},
						},
					},
				},
			},
			OracleText: `
			This spell costs {1} less to cast for each creature you control with a +1/+1 counter on it.
			Creature spells you cast cost {1} less to cast for each creature you control with a +1/+1 counter on it.
		`,
		},
	}
}
