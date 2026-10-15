package p

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// PrimalRage is the card definition for Primal Rage.
//
// Type: Enchantment
// Cost: {1}{G}
//
// Oracle text:
//
//	Creatures you control have trample. (A creature with trample can deal excess combat damage to the player or planeswalker it's attacking.)
var PrimalRage = newPrimalRage

func newPrimalRage() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Primal Rage",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.G,
			}),
			Colors: []color.Color{color.Green},
			Types:  []types.Card{types.Enchantment},
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "Creatures you control have trample. (A creature with trample can deal excess combat damage to the player or planeswalker it's attacking.)",
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Duration:    game.DurationPermanent,
							Group:       game.ObjectControlledGroup(game.SourcePermanentReference(), game.Selection{RequiredTypes: []types.Card{types.Creature}}),
							Layer:       game.LayerAbility,
							AddKeywords: []game.Keyword{game.Trample},
						},
					},
				},
			},
			OracleText: `
			Creatures you control have trample. (A creature with trample can deal excess combat damage to the player or planeswalker it's attacking.)
		`,
		},
	}
}
