package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// SkyBlessedSamurai is the card definition for Sky-Blessed Samurai.
//
// Type: Enchantment Creature — Human Samurai
// Cost: {6}{W}
//
// Oracle text:
//
//	Affinity for enchantments (This spell costs {1} less to cast for each enchantment you control.)
//	Flying
var SkyBlessedSamurai = newSkyBlessedSamurai

func newSkyBlessedSamurai() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Sky-Blessed Samurai",
			ManaCost: opt.Val(cost.Mana{
				cost.O(6),
				cost.W,
			}),
			Colors:    []color.Color{color.White},
			Types:     []types.Card{types.Enchantment, types.Creature},
			Subtypes:  []types.Sub{types.Human, types.Samurai},
			Power:     opt.Val(game.PT{Value: 4}),
			Toughness: opt.Val(game.PT{Value: 4}),
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
				game.FlyingStaticBody,
			},
			OracleText: `
			Affinity for enchantments (This spell costs {1} less to cast for each enchantment you control.)
			Flying
		`,
		},
	}
}
