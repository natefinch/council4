package g

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// GlacialPlating is the card definition for Glacial Plating.
//
// Type: Snow Enchantment — Aura
// Cost: {2}{W}{W}
//
// Oracle text:
//
//	Enchant creature
//	Cumulative upkeep {S} (At the beginning of your upkeep, put an age counter on this permanent, then sacrifice it unless you pay its upkeep cost for each age counter on it. {S} can be paid with one mana from a snow source.)
//	Enchanted creature gets +3/+3 for each age counter on this Aura.
var GlacialPlating = newGlacialPlating

func newGlacialPlating() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Glacial Plating",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.W,
				cost.W,
			}),
			Colors:     []color.Color{color.White},
			Supertypes: []types.Super{types.Snow},
			Types:      []types.Card{types.Enchantment},
			Subtypes:   []types.Sub{types.Aura},
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
							Layer: game.LayerPowerToughnessModify,
							Group: game.AttachedObjectGroup(game.SourcePermanentReference()),
							PowerDeltaDynamic: opt.Val(game.DynamicAmount{
								Kind:        game.DynamicAmountObjectCounters,
								Multiplier:  3,
								CounterKind: counter.Age,
								Object:      game.SourcePermanentReference(),
							}),
							ToughnessDeltaDynamic: opt.Val(game.DynamicAmount{
								Kind:        game.DynamicAmountObjectCounters,
								Multiplier:  3,
								CounterKind: counter.Age,
								Object:      game.SourcePermanentReference(),
							}),
						},
					},
				},
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.CumulativeUpkeepTriggeredAbility(cost.Mana{cost.S}),
			},
			OracleText: `
			Enchant creature
			Cumulative upkeep {S} (At the beginning of your upkeep, put an age counter on this permanent, then sacrifice it unless you pay its upkeep cost for each age counter on it. {S} can be paid with one mana from a snow source.)
			Enchanted creature gets +3/+3 for each age counter on this Aura.
		`,
		},
	}
}
