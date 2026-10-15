package g

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// GravitationalShift is the card definition for Gravitational Shift.
//
// Type: Enchantment
// Cost: {3}{U}{U}
//
// Oracle text:
//
//	Creatures with flying get +2/+0.
//	Creatures without flying get -2/-0.
var GravitationalShift = newGravitationalShift

func newGravitationalShift() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Gravitational Shift",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.U,
				cost.U,
			}),
			Colors: []color.Color{color.Blue},
			Types:  []types.Card{types.Enchantment},
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "Creatures with flying get +2/+0.",
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Duration:   game.DurationPermanent,
							Group:      game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}, Keyword: game.Flying}),
							Layer:      game.LayerPowerToughnessModify,
							PowerDelta: 2,
						},
					},
				},
				game.StaticAbility{
					Text: "Creatures without flying get -2/-0.",
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Duration:   game.DurationPermanent,
							Group:      game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}, ExcludedKeyword: game.Flying}),
							Layer:      game.LayerPowerToughnessModify,
							PowerDelta: -2,
						},
					},
				},
			},
			OracleText: `
			Creatures with flying get +2/+0.
			Creatures without flying get -2/-0.
		`,
		},
	}
}
