package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// SangriteBacklash is the card definition for Sangrite Backlash.
//
// Type: Enchantment — Aura
// Cost: {B/G}{R}
//
// Oracle text:
//
//	Enchant creature
//	Enchanted creature gets +3/-3.
var SangriteBacklash = newSangriteBacklash

func newSangriteBacklash() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black, color.Red, color.Green),
		CardFace: game.CardFace{
			Name: "Sangrite Backlash",
			ManaCost: opt.Val(cost.Mana{
				cost.HybridMana(mana.B, mana.G),
				cost.R,
			}),
			Colors:   []color.Color{color.Black, color.Green, color.Red},
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
					Text: "Enchanted creature gets +3/-3.",
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Duration:       game.DurationPermanent,
							Group:          game.AttachedObjectGroup(game.SourcePermanentReference()),
							Layer:          game.LayerPowerToughnessModify,
							PowerDelta:     3,
							ToughnessDelta: -3,
						},
					},
				},
			},
			OracleText: `
			Enchant creature
			Enchanted creature gets +3/-3.
		`,
		},
	}
}
