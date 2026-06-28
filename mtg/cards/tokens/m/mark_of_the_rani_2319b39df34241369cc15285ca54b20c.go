package m

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Mark of the Rani
//
// Type: Token Enchantment — Aura
//
// Oracle text:
//   Enchant creature
//   Enchanted creature gets +2/+2 and is goaded.

// MarkOfTheRaniToken2319b39df34241369cc15285ca54b20c is the card definition for Mark of the Rani.
var MarkOfTheRaniToken2319b39df34241369cc15285ca54b20c = newMarkOfTheRaniToken2319b39df34241369cc15285ca54b20c()

func newMarkOfTheRaniToken2319b39df34241369cc15285ca54b20c() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name:     "Mark of the Rani",
			Colors:   []color.Color{color.Red},
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
							PowerDelta:     2,
							ToughnessDelta: 2,
						},
					},
					RuleEffects: []game.RuleEffect{
						game.RuleEffect{
							Kind:             game.RuleEffectGoaded,
							AffectedAttached: true,
						},
					},
				},
			},
			OracleText: `
			Enchant creature
			Enchanted creature gets +2/+2 and is goaded.
		`,
		},
	}
}
