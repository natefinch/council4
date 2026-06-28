package p

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Petrify is the card definition for Petrify.
//
// Type: Enchantment — Aura
// Cost: {1}{W}
//
// Oracle text:
//
//	Enchant artifact or creature
//	Enchanted permanent can't attack or block, and its activated abilities can't be activated.
var Petrify = newPetrify()

func newPetrify() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Petrify",
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
					Constraint: "artifact or creature",
					Allow:      game.TargetAllowPermanent,
					Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Artifact, types.Creature}}),
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
			Enchant artifact or creature
			Enchanted permanent can't attack or block, and its activated abilities can't be activated.
		`,
		},
	}
}
