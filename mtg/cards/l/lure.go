package l

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Lure is the card definition for Lure.
//
// Type: Enchantment — Aura
// Cost: {1}{G}{G}
//
// Oracle text:
//
//	Enchant creature
//	All creatures able to block enchanted creature do so.
var Lure = newLure

func newLure() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Lure",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.G,
				cost.G,
			}),
			Colors:   []color.Color{color.Green},
			Types:    []types.Card{types.Enchantment},
			Subtypes: []types.Sub{types.Aura},
			StaticAbilities: []game.StaticAbility{
				game.EnchantStaticAbility(&game.TargetSpec{
					MinTargets: 1,
					MaxTargets: 1,
					Constraint: "creature",
					Allow:      game.TargetAllowPermanent,
					Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}}),
				}),
				game.StaticAbility{
					Text: "All creatures able to block enchanted creature do so.",
					RuleEffects: []game.RuleEffect{
						game.RuleEffect{
							Kind:             game.RuleEffectMustBeBlockedByAllAble,
							AffectedAttached: true,
							Duration:         game.DurationPermanent,
						},
					},
				},
			},
			OracleText: `
			Enchant creature
			All creatures able to block enchanted creature do so.
		`,
		},
	}
}
