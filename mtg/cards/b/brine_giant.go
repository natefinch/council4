package b

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// BrineGiant is the card definition for Brine Giant.
//
// Type: Creature — Giant
// Cost: {6}{U}
//
// Oracle text:
//
//	Affinity for enchantments (This spell costs {1} less to cast for each enchantment you control.)
var BrineGiant = newBrineGiant

func newBrineGiant() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Brine Giant",
			ManaCost: opt.Val(cost.Mana{
				cost.O(6),
				cost.U,
			}),
			Colors:    []color.Color{color.Blue},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Giant},
			Power:     opt.Val(game.PT{Value: 5}),
			Toughness: opt.Val(game.PT{Value: 6}),
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "This spell costs {1} less to cast for each enchantment you control.",
					RuleEffects: []game.RuleEffect{
						game.RuleEffect{
							Kind:           game.RuleEffectCostModifier,
							AffectedSource: true,
							Duration:       game.DurationPermanent,
							CostModifier: game.CostModifier{
								Kind:               game.CostModifierSpell,
								PerObjectReduction: 1,
								CountSelection:     &game.Selection{RequiredTypes: []types.Card{types.Enchantment}, Controller: game.ControllerYou},
							},
						},
					},
				},
			},
			OracleText: `
			Affinity for enchantments (This spell costs {1} less to cast for each enchantment you control.)
		`,
		},
	}
}
