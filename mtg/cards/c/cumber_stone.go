package c

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// CumberStone is the card definition for Cumber Stone.
//
// Type: Artifact
// Cost: {3}{U}
//
// Oracle text:
//
//	Creatures your opponents control get -1/-0.
var CumberStone = newCumberStone

func newCumberStone() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Cumber Stone",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.U,
			}),
			Colors: []color.Color{color.Blue},
			Types:  []types.Card{types.Artifact},
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "Creatures your opponents control get -1/-0.",
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Duration:   game.DurationPermanent,
							Group:      game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}, Controller: game.ControllerOpponent}),
							Layer:      game.LayerPowerToughnessModify,
							PowerDelta: -1,
						},
					},
				},
			},
			OracleText: `
			Creatures your opponents control get -1/-0.
		`,
		},
	}
}
