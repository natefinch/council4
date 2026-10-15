package g

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// GrimStrider is the card definition for Grim Strider.
//
// Type: Creature — Horror
// Cost: {3}{B}
//
// Oracle text:
//
//	This creature gets -1/-1 for each card in your hand.
var GrimStrider = newGrimStrider

func newGrimStrider() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Grim Strider",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.B,
			}),
			Colors:    []color.Color{color.Black},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Horror},
			Power:     opt.Val(game.PT{Value: 6}),
			Toughness: opt.Val(game.PT{Value: 6}),
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "This creature gets -1/-1 for each card in your hand.",
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Duration:       game.DurationPermanent,
							AffectedSource: true,
							Layer:          game.LayerPowerToughnessModify,
							PowerDeltaDynamic: opt.Val(game.DynamicAmount{
								Kind:        game.DynamicAmountCountCardsInZone,
								Multiplier:  -1,
								CounterKind: counter.PlusOnePlusOne,
								Player:      func() *game.PlayerReference { ref := game.ControllerReference(); return &ref }(),
								CardZone:    zone.Hand,
								Selection:   &game.Selection{},
							}),
							ToughnessDeltaDynamic: opt.Val(game.DynamicAmount{
								Kind:        game.DynamicAmountCountCardsInZone,
								Multiplier:  -1,
								CounterKind: counter.PlusOnePlusOne,
								Player:      func() *game.PlayerReference { ref := game.ControllerReference(); return &ref }(),
								CardZone:    zone.Hand,
								Selection:   &game.Selection{},
							}),
						},
					},
				},
			},
			OracleText: `
			This creature gets -1/-1 for each card in your hand.
		`,
		},
	}
}
