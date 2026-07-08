package c

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// ChoMannoSBlessing is the card definition for Cho-Manno's Blessing.
//
// Type: Enchantment — Aura
// Cost: {W}{W}
//
// Oracle text:
//
//	Flash
//	Enchant creature
//	As this Aura enters, choose a color.
//	Enchanted creature has protection from the chosen color. This effect doesn't remove this Aura.
var ChoMannoSBlessing = newChoMannoSBlessing

func newChoMannoSBlessing() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Cho-Manno's Blessing",
			ManaCost: opt.Val(cost.Mana{
				cost.W,
				cost.W,
			}),
			Colors:   []color.Color{color.White},
			Types:    []types.Card{types.Enchantment},
			Subtypes: []types.Sub{types.Aura},
			StaticAbilities: []game.StaticAbility{
				game.FlashStaticBody,
				game.EnchantStaticAbility(&game.TargetSpec{
					MinTargets: 1,
					MaxTargets: 1,
					Constraint: "creature",
					Allow:      game.TargetAllowPermanent,
					Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}}),
				}),
				game.StaticAbility{
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Layer: game.LayerAbility,
							Group: game.AttachedObjectGroup(game.SourcePermanentReference()),
							AddAbilities: []game.Ability{
								new(game.ProtectionFromChosenColorStaticAbility()),
							},
						},
					},
				},
			},
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntryColorChoiceReplacement("As this Aura enters, choose a color."),
			},
			OracleText: `
			Flash
			Enchant creature
			As this Aura enters, choose a color.
			Enchanted creature has protection from the chosen color. This effect doesn't remove this Aura.
		`,
		},
	}
}
