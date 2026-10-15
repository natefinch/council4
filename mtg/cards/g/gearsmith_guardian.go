package g

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// GearsmithGuardian is the card definition for Gearsmith Guardian.
//
// Type: Artifact Creature — Construct
// Cost: {5}
//
// Oracle text:
//
//	This creature gets +2/+0 as long as you control a blue creature.
var GearsmithGuardian = newGearsmithGuardian

func newGearsmithGuardian() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Gearsmith Guardian",
			ManaCost: opt.Val(cost.Mana{
				cost.O(5),
			}),
			Types:     []types.Card{types.Artifact, types.Creature},
			Subtypes:  []types.Sub{types.Construct},
			Power:     opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 5}),
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "This creature gets +2/+0 as long as you control a blue creature.",
					Condition: opt.Val(game.Condition{
						Text: "as long as you control a blue creature",
						ControlsMatching: opt.Val(game.SelectionCount{
							Selection: game.Selection{RequiredTypes: []types.Card{types.Creature}, ColorsAny: []color.Color{color.Blue}},
						}),
						SourceCounterKind: counter.PlusOnePlusOne,
					}),
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Duration:       game.DurationPermanent,
							AffectedSource: true,
							Layer:          game.LayerPowerToughnessModify,
							PowerDelta:     2,
						},
					},
				},
			},
			OracleText: `
			This creature gets +2/+0 as long as you control a blue creature.
		`,
		},
	}
}
