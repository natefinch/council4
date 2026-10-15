package t

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// TransguildCourier is the card definition for Transguild Courier.
//
// Type: Artifact Creature — Golem
// Cost: {4}
//
// Oracle text:
//
//	Transguild Courier is all colors.
var TransguildCourier = newTransguildCourier

func newTransguildCourier() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Blue, color.Black, color.Red, color.Green),
		CardFace: game.CardFace{
			Name: "Transguild Courier",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
			}),
			Colors:    []color.Color{color.Black, color.Green, color.Red, color.Blue, color.White},
			Types:     []types.Card{types.Artifact, types.Creature},
			Subtypes:  []types.Sub{types.Golem},
			Power:     opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 3}),
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "Transguild Courier is all colors.",
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Duration:       game.DurationPermanent,
							AffectedSource: true,
							Layer:          game.LayerColor,
							SetColors:      []color.Color{color.White, color.Blue, color.Black, color.Red, color.Green},
						},
					},
				},
			},
			OracleText: `
			Transguild Courier is all colors.
		`,
		},
	}
}
