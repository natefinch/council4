package t

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// TreetopBracers is the card definition for Treetop Bracers.
//
// Type: Enchantment — Aura
// Cost: {1}{G}
//
// Oracle text:
//
//	Enchant creature (Target a creature as you cast this. This card enters attached to that creature.)
//	Enchanted creature gets +1/+1 and can't be blocked except by creatures with flying.
var TreetopBracers = newTreetopBracers()

func newTreetopBracers() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Treetop Bracers",
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
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Layer:          game.LayerPowerToughnessModify,
							Group:          game.AttachedObjectGroup(game.SourcePermanentReference()),
							PowerDelta:     1,
							ToughnessDelta: 1,
						},
					},
					RuleEffects: []game.RuleEffect{
						game.RuleEffect{
							Kind:             game.RuleEffectCantBeBlockedExceptBy,
							AffectedAttached: true,
							BlockerRestriction: game.BlockerRestriction{
								Kind: game.BlockerRestrictionFlying,
							},
						},
					},
				},
			},
			OracleText: `
			Enchant creature (Target a creature as you cast this. This card enters attached to that creature.)
			Enchanted creature gets +1/+1 and can't be blocked except by creatures with flying.
		`,
		},
	}
}
