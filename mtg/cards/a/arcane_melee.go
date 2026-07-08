package a

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// ArcaneMelee is the card definition for Arcane Melee.
//
// Type: Enchantment
// Cost: {4}{U}
//
// Oracle text:
//
//	Instant and sorcery spells cost {2} less to cast.
var ArcaneMelee = newArcaneMelee

func newArcaneMelee() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Arcane Melee",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
				cost.U,
			}),
			Colors: []color.Color{color.Blue},
			Types:  []types.Card{types.Enchantment},
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					RuleEffects: []game.RuleEffect{
						game.RuleEffect{
							Kind: game.RuleEffectCostModifier,
							CostModifier: game.CostModifier{
								Kind:             game.CostModifierSpell,
								CardSelection:    game.Selection{RequiredTypes: []types.Card{types.Instant}},
								GenericReduction: 2,
							},
						},
						game.RuleEffect{
							Kind: game.RuleEffectCostModifier,
							CostModifier: game.CostModifier{
								Kind:             game.CostModifierSpell,
								CardSelection:    game.Selection{RequiredTypes: []types.Card{types.Sorcery}},
								GenericReduction: 2,
							},
						},
					},
				},
			},
			OracleText: `
			Instant and sorcery spells cost {2} less to cast.
		`,
		},
	}
}
