package c

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// CanopyCover is the card definition for Canopy Cover.
//
// Type: Enchantment — Aura
// Cost: {1}{G}
//
// Oracle text:
//
//	Enchant creature
//	Enchanted creature can't be blocked except by creatures with flying or reach.
//	Enchanted creature can't be the target of spells or abilities your opponents control.
var CanopyCover = newCanopyCover

func newCanopyCover() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Canopy Cover",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
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
					Text: "Enchanted creature can't be blocked except by creatures with flying or reach.",
					RuleEffects: []game.RuleEffect{
						game.RuleEffect{
							Kind:             game.RuleEffectCantBeBlockedExceptBy,
							AffectedAttached: true,
							Duration:         game.DurationPermanent,
							BlockerRestriction: game.BlockerRestriction{
								Kind: game.BlockerRestrictionFlyingOrReach,
							},
						},
					},
				},
				game.StaticAbility{
					Text: "Enchanted creature can't be the target of spells or abilities your opponents control.",
					RuleEffects: []game.RuleEffect{
						game.RuleEffect{
							Kind:             game.RuleEffectCantBeTargetedByControllerOpponents,
							AffectedAttached: true,
							Duration:         game.DurationPermanent,
						},
					},
				},
			},
			OracleText: `
			Enchant creature
			Enchanted creature can't be blocked except by creatures with flying or reach.
			Enchanted creature can't be the target of spells or abilities your opponents control.
		`,
		},
	}
}
