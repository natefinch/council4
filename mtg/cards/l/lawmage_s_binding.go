package l

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// LawmageSBinding is the card definition for Lawmage's Binding.
//
// Type: Enchantment — Aura
// Cost: {1}{W}{U}
//
// Oracle text:
//
//	Flash
//	Enchant creature
//	Enchanted creature can't attack or block, and its activated abilities can't be activated.
var LawmageSBinding = newLawmageSBinding()

func newLawmageSBinding() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Blue),
		CardFace: game.CardFace{
			Name: "Lawmage's Binding",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.W,
				cost.U,
			}),
			Colors:   []color.Color{color.Blue, color.White},
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
			Flash
			Enchant creature
			Enchanted creature can't attack or block, and its activated abilities can't be activated.
		`,
		},
	}
}
