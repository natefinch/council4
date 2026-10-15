package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// SpitefulMotives is the card definition for Spiteful Motives.
//
// Type: Enchantment — Aura
// Cost: {3}{R}
//
// Oracle text:
//
//	Flash (You may cast this spell any time you could cast an instant.)
//	Enchant creature
//	Enchanted creature gets +3/+0 and has first strike.
var SpitefulMotives = newSpitefulMotives

func newSpitefulMotives() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Spiteful Motives",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.R,
			}),
			Colors:   []color.Color{color.Red},
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
					Text: "Enchanted creature gets +3/+0 and has first strike.",
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Duration:   game.DurationPermanent,
							Group:      game.AttachedObjectGroup(game.SourcePermanentReference()),
							Layer:      game.LayerPowerToughnessModify,
							PowerDelta: 3,
						},
						game.ContinuousEffect{
							Duration:    game.DurationPermanent,
							Group:       game.AttachedObjectGroup(game.SourcePermanentReference()),
							Layer:       game.LayerAbility,
							AddKeywords: []game.Keyword{game.FirstStrike},
						},
					},
				},
			},
			OracleText: `
			Flash (You may cast this spell any time you could cast an instant.)
			Enchant creature
			Enchanted creature gets +3/+0 and has first strike.
		`,
		},
	}
}
