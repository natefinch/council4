package h

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// HardyVeteran is the card definition for Hardy Veteran.
//
// Type: Creature — Human Warrior
// Cost: {1}{G}
//
// Oracle text:
//
//	During your turn, this creature gets +0/+2.
var HardyVeteran = newHardyVeteran

func newHardyVeteran() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Hardy Veteran",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.G,
			}),
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Human, types.Warrior},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 2}),
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "During your turn, this creature gets +0/+2.",
					Condition: opt.Val(game.Condition{
						Text:                 "During your turn",
						SourceCounterKind:    counter.PlusOnePlusOne,
						SourceControllerTurn: true,
					}),
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Duration:       game.DurationPermanent,
							AffectedSource: true,
							Layer:          game.LayerPowerToughnessModify,
							ToughnessDelta: 2,
						},
					},
				},
			},
			OracleText: `
			During your turn, this creature gets +0/+2.
		`,
		},
	}
}
