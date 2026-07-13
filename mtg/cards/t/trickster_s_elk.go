package t

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// TricksterSElk is the card definition for Trickster's Elk.
//
// Type: Enchantment Creature — Elk
// Cost: {2}{G}
//
// Oracle text:
//
//	Bestow {1}{G} (If you cast this card for its bestow cost, it's an Aura spell with enchant creature. It becomes a creature again if it's not attached.)
//	Enchanted creature loses all abilities and is a green Elk creature with base power and toughness 3/3.
var TricksterSElk = newTricksterSElk

func newTricksterSElk() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Trickster's Elk",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.G,
			}),
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Enchantment, types.Creature},
			Subtypes:  []types.Sub{types.Elk},
			Power:     opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 3}),
			StaticAbilities: []game.StaticAbility{
				game.BestowStaticAbility(cost.Mana{cost.O(1), cost.G}, &game.TargetSpec{
					MinTargets: 1,
					MaxTargets: 1,
					Constraint: "creature",
					Allow:      game.TargetAllowPermanent,
					Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}}),
				}),
				game.StaticAbility{
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Layer:              game.LayerAbility,
							Group:              game.AttachedObjectGroup(game.SourcePermanentReference()),
							RemoveAllAbilities: true,
						},
						game.ContinuousEffect{
							Layer:     game.LayerColor,
							Group:     game.AttachedObjectGroup(game.SourcePermanentReference()),
							SetColors: []color.Color{color.Green},
						},
						game.ContinuousEffect{
							Layer:       game.LayerType,
							Group:       game.AttachedObjectGroup(game.SourcePermanentReference()),
							SetTypes:    []types.Card{types.Creature},
							SetSubtypes: []types.Sub{types.Elk},
						},
						game.ContinuousEffect{
							Layer:        game.LayerPowerToughnessSet,
							Group:        game.AttachedObjectGroup(game.SourcePermanentReference()),
							SetPower:     opt.Val(game.PT{Value: 3}),
							SetToughness: opt.Val(game.PT{Value: 3}),
						},
					},
				},
			},
			OracleText: `
			Bestow {1}{G} (If you cast this card for its bestow cost, it's an Aura spell with enchant creature. It becomes a creature again if it's not attached.)
			Enchanted creature loses all abilities and is a green Elk creature with base power and toughness 3/3.
		`,
		},
	}
}
