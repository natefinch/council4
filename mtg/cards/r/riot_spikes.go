package r

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// RiotSpikes is the card definition for Riot Spikes.
//
// Type: Enchantment — Aura
// Cost: {B/R}
//
// Oracle text:
//
//	({B/R} can be paid with either {B} or {R}.)
//	Enchant creature
//	Enchanted creature gets +2/-1.
var RiotSpikes = newRiotSpikes

func newRiotSpikes() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black, color.Red),
		CardFace: game.CardFace{
			Name: "Riot Spikes",
			ManaCost: opt.Val(cost.Mana{
				cost.HybridMana(mana.B, mana.R),
			}),
			Colors:   []color.Color{color.Black, color.Red},
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
					Text: "Enchanted creature gets +2/-1.",
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Duration:       game.DurationPermanent,
							Group:          game.AttachedObjectGroup(game.SourcePermanentReference()),
							Layer:          game.LayerPowerToughnessModify,
							PowerDelta:     2,
							ToughnessDelta: -1,
						},
					},
				},
			},
			OracleText: `
			({B/R} can be paid with either {B} or {R}.)
			Enchant creature
			Enchanted creature gets +2/-1.
		`,
		},
	}
}
