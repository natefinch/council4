package c

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// CrashOfRhinoBeetles is the card definition for Crash of Rhino Beetles.
//
// Type: Creature — Insect
// Cost: {4}{G}
//
// Oracle text:
//
//	Trample
//	This creature gets +10/+10 as long as you control ten or more lands.
var CrashOfRhinoBeetles = newCrashOfRhinoBeetles

func newCrashOfRhinoBeetles() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Crash of Rhino Beetles",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
				cost.G,
			}),
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Insect},
			Power:     opt.Val(game.PT{Value: 5}),
			Toughness: opt.Val(game.PT{Value: 5}),
			StaticAbilities: []game.StaticAbility{
				game.TrampleStaticBody,
				game.StaticAbility{
					Text: "This creature gets +10/+10 as long as you control ten or more lands.",
					Condition: opt.Val(game.Condition{
						Text: "as long as you control ten or more lands",
						ControlsMatching: opt.Val(game.SelectionCount{
							Selection: game.Selection{RequiredTypes: []types.Card{types.Land}},
							MinCount:  10,
						}),
						SourceCounterKind: counter.PlusOnePlusOne,
					}),
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Duration:       game.DurationPermanent,
							AffectedSource: true,
							Layer:          game.LayerPowerToughnessModify,
							PowerDelta:     10,
							ToughnessDelta: 10,
						},
					},
				},
			},
			OracleText: `
			Trample
			This creature gets +10/+10 as long as you control ten or more lands.
		`,
		},
	}
}
