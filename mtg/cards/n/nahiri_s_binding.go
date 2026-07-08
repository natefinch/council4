package n

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// NahiriSBinding is the card definition for Nahiri's Binding.
//
// Type: Enchantment — Aura
// Cost: {1}{W}{W}
//
// Oracle text:
//
//	Enchant creature or planeswalker
//	Enchanted permanent can't attack or block, and its activated abilities can't be activated.
var NahiriSBinding = newNahiriSBinding

func newNahiriSBinding() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Nahiri's Binding",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.W,
				cost.W,
			}),
			Colors:   []color.Color{color.White},
			Types:    []types.Card{types.Enchantment},
			Subtypes: []types.Sub{types.Aura},
			StaticAbilities: []game.StaticAbility{
				game.EnchantStaticAbility(&game.TargetSpec{
					MinTargets: 1,
					MaxTargets: 1,
					Constraint: "creature or planeswalker",
					Allow:      game.TargetAllowPermanent,
					Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature, types.Planeswalker}}),
				}),
				game.StaticAbility{
					RuleEffects: []game.RuleEffect{
						game.RuleEffect{
							Kind:             game.RuleEffectCantAttack,
							AffectedAttached: true,
						},
						game.RuleEffect{
							Kind:             game.RuleEffectCantBlock,
							AffectedAttached: true,
						},
						game.RuleEffect{
							Kind:             game.RuleEffectCantActivateAbilitiesOfPermanent,
							AffectedAttached: true,
						},
					},
				},
			},
			OracleText: `
			Enchant creature or planeswalker
			Enchanted permanent can't attack or block, and its activated abilities can't be activated.
		`,
		},
	}
}
