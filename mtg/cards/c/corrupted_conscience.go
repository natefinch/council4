package c

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// CorruptedConscience is the card definition for Corrupted Conscience.
//
// Type: Enchantment — Aura
// Cost: {3}{U}{U}
//
// Oracle text:
//
//	Enchant creature
//	You control enchanted creature.
//	Enchanted creature has infect. (It deals damage to creatures in the form of -1/-1 counters and to players in the form of poison counters.)
var CorruptedConscience = newCorruptedConscience

func newCorruptedConscience() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Corrupted Conscience",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.U,
				cost.U,
			}),
			Colors:   []color.Color{color.Blue},
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
					Text: "You control enchanted creature.",
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Duration:      game.DurationPermanent,
							Group:         game.AttachedObjectGroup(game.SourcePermanentReference()),
							Layer:         game.LayerControl,
							NewController: opt.Val(game.Player1),
						},
					},
				},
				game.StaticAbility{
					Text: "Enchanted creature has infect. (It deals damage to creatures in the form of -1/-1 counters and to players in the form of poison counters.)",
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Duration:    game.DurationPermanent,
							Group:       game.AttachedObjectGroup(game.SourcePermanentReference()),
							Layer:       game.LayerAbility,
							AddKeywords: []game.Keyword{game.Infect},
						},
					},
				},
			},
			OracleText: `
			Enchant creature
			You control enchanted creature.
			Enchanted creature has infect. (It deals damage to creatures in the form of -1/-1 counters and to players in the form of poison counters.)
		`,
		},
	}
}
