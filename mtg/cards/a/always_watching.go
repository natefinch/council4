package a

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// AlwaysWatching is the card definition for Always Watching.
//
// Type: Enchantment
// Cost: {1}{W}{W}
//
// Oracle text:
//
//	Nontoken creatures you control get +1/+1 and have vigilance.
var AlwaysWatching = newAlwaysWatching

func newAlwaysWatching() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Always Watching",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.W,
				cost.W,
			}),
			Colors: []color.Color{color.White},
			Types:  []types.Card{types.Enchantment},
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "Nontoken creatures you control get +1/+1 and have vigilance.",
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Duration:       game.DurationPermanent,
							Group:          game.ObjectControlledGroup(game.SourcePermanentReference(), game.Selection{RequiredTypes: []types.Card{types.Creature}, NonToken: true}),
							Layer:          game.LayerPowerToughnessModify,
							PowerDelta:     1,
							ToughnessDelta: 1,
						},
						game.ContinuousEffect{
							Duration:    game.DurationPermanent,
							Group:       game.ObjectControlledGroup(game.SourcePermanentReference(), game.Selection{RequiredTypes: []types.Card{types.Creature}, NonToken: true}),
							Layer:       game.LayerAbility,
							AddKeywords: []game.Keyword{game.Vigilance},
						},
					},
				},
			},
			OracleText: `
			Nontoken creatures you control get +1/+1 and have vigilance.
		`,
		},
	}
}
