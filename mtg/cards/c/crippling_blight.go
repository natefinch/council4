package c

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// CripplingBlight is the card definition for Crippling Blight.
//
// Type: Enchantment — Aura
// Cost: {B}
//
// Oracle text:
//
//	Enchant creature
//	Enchanted creature gets -1/-1 and can't block.
var CripplingBlight = newCripplingBlight

func newCripplingBlight() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Crippling Blight",
			ManaCost: opt.Val(cost.Mana{
				cost.B,
			}),
			Colors:   []color.Color{color.Black},
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
					Text: "Enchanted creature gets -1/-1 and can't block.",
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Duration:       game.DurationPermanent,
							Group:          game.AttachedObjectGroup(game.SourcePermanentReference()),
							Layer:          game.LayerPowerToughnessModify,
							PowerDelta:     -1,
							ToughnessDelta: -1,
						},
					},
					RuleEffects: []game.RuleEffect{
						game.RuleEffect{
							Kind:             game.RuleEffectCantBlock,
							AffectedAttached: true,
							Duration:         game.DurationPermanent,
						},
					},
				},
			},
			OracleText: `
			Enchant creature
			Enchanted creature gets -1/-1 and can't block.
		`,
		},
	}
}
