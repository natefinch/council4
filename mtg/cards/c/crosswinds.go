package c

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Crosswinds is the card definition for Crosswinds.
//
// Type: Enchantment
// Cost: {1}{G}
//
// Oracle text:
//
//	Creatures with flying get -2/-0.
var Crosswinds = newCrosswinds

func newCrosswinds() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Crosswinds",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.G,
			}),
			Colors: []color.Color{color.Green},
			Types:  []types.Card{types.Enchantment},
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "Creatures with flying get -2/-0.",
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Duration:   game.DurationPermanent,
							Group:      game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}, Keyword: game.Flying}),
							Layer:      game.LayerPowerToughnessModify,
							PowerDelta: -2,
						},
					},
				},
			},
			OracleText: `
			Creatures with flying get -2/-0.
		`,
		},
	}
}
