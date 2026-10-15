package i

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// IriniSengir is the card definition for Irini Sengir.
//
// Type: Legendary Creature — Vampire Dwarf
// Cost: {2}{B}{B}
//
// Oracle text:
//
//	Green enchantment spells and white enchantment spells cost {2} more to cast.
var IriniSengir = newIriniSengir

func newIriniSengir() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Irini Sengir",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.B,
				cost.B,
			}),
			Colors:     []color.Color{color.Black},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Vampire, types.Dwarf},
			Power:      opt.Val(game.PT{Value: 2}),
			Toughness:  opt.Val(game.PT{Value: 2}),
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "Green enchantment spells and white enchantment spells cost {2} more to cast.",
					RuleEffects: []game.RuleEffect{
						game.RuleEffect{
							Kind:     game.RuleEffectCostModifier,
							Duration: game.DurationPermanent,
							CostModifier: game.CostModifier{
								Kind:            game.CostModifierSpell,
								GenericIncrease: 2,
								CardSelection:   game.Selection{RequiredTypes: []types.Card{types.Enchantment}, ColorsAny: []color.Color{color.Green, color.White}},
							},
						},
					},
				},
			},
			OracleText: `
			Green enchantment spells and white enchantment spells cost {2} more to cast.
		`,
		},
	}
}
