package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// SongOfTheDryads is the card definition for Song of the Dryads.
//
// Type: Enchantment — Aura
// Cost: {2}{G}
//
// Oracle text:
//
//	Enchant permanent
//	Enchanted permanent is a colorless Forest land.
var SongOfTheDryads = newSongOfTheDryads

func newSongOfTheDryads() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Song of the Dryads",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.G,
			}),
			Colors:   []color.Color{color.Green},
			Types:    []types.Card{types.Enchantment},
			Subtypes: []types.Sub{types.Aura},
			StaticAbilities: []game.StaticAbility{
				game.EnchantStaticAbility(&game.TargetSpec{
					MinTargets: 1,
					MaxTargets: 1,
					Constraint: "permanent",
					Allow:      game.TargetAllowPermanent,
				}),
				game.StaticAbility{
					Text: "Enchanted permanent is a colorless Forest land.",
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Duration:     game.DurationPermanent,
							Group:        game.AttachedObjectGroup(game.SourcePermanentReference()),
							Layer:        game.LayerColor,
							SetColorless: true,
						},
						game.ContinuousEffect{
							Duration:    game.DurationPermanent,
							Group:       game.AttachedObjectGroup(game.SourcePermanentReference()),
							Layer:       game.LayerType,
							SetTypes:    []types.Card{types.Land},
							SetSubtypes: []types.Sub{types.Forest},
						},
					},
				},
			},
			OracleText: `
			Enchant permanent
			Enchanted permanent is a colorless Forest land.
		`,
		},
	}
}
