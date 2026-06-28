package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// StasisCocoon is the card definition for Stasis Cocoon.
//
// Type: Enchantment — Aura
// Cost: {1}{W}
//
// Oracle text:
//
//	Enchant artifact
//	Enchanted artifact can't attack or block, and its activated abilities can't be activated.
var StasisCocoon = newStasisCocoon()

func newStasisCocoon() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Stasis Cocoon",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.W,
			}),
			Colors:   []color.Color{color.White},
			Types:    []types.Card{types.Enchantment},
			Subtypes: []types.Sub{types.Aura},
			StaticAbilities: []game.StaticAbility{
				game.EnchantStaticAbility(&game.TargetSpec{
					MinTargets: 1,
					MaxTargets: 1,
					Constraint: "artifact",
					Allow:      game.TargetAllowPermanent,
					Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Artifact}}),
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
			Enchant artifact
			Enchanted artifact can't attack or block, and its activated abilities can't be activated.
		`,
		},
	}
}
