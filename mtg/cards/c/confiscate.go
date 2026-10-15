package c

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Confiscate is the card definition for Confiscate.
//
// Type: Enchantment — Aura
// Cost: {4}{U}{U}
//
// Oracle text:
//
//	Enchant permanent
//	You control enchanted permanent.
var Confiscate = newConfiscate

func newConfiscate() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Confiscate",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
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
					Constraint: "permanent",
					Allow:      game.TargetAllowPermanent,
				}),
				game.StaticAbility{
					Text: "You control enchanted permanent.",
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Duration:      game.DurationPermanent,
							Group:         game.AttachedObjectGroup(game.SourcePermanentReference()),
							Layer:         game.LayerControl,
							NewController: opt.Val(game.Player1),
						},
					},
				},
			},
			OracleText: `
			Enchant permanent
			You control enchanted permanent.
		`,
		},
	}
}
