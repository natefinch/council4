package r

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// RadiantEssence is the card definition for Radiant Essence.
//
// Type: Creature — Spirit
// Cost: {1}{G}{W}
//
// Oracle text:
//
//	This creature gets +1/+2 as long as an opponent controls a black permanent.
var RadiantEssence = newRadiantEssence

func newRadiantEssence() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Green),
		CardFace: game.CardFace{
			Name: "Radiant Essence",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.G,
				cost.W,
			}),
			Colors:    []color.Color{color.Green, color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Spirit},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 3}),
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "This creature gets +1/+2 as long as an opponent controls a black permanent.",
					Condition: opt.Val(game.Condition{
						Text: "as long as an opponent controls a black permanent",
						AnyOpponentControls: opt.Val(game.SelectionCount{
							Selection: game.Selection{ColorsAny: []color.Color{color.Black}},
						}),
						SourceCounterKind: counter.PlusOnePlusOne,
					}),
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Duration:       game.DurationPermanent,
							AffectedSource: true,
							Layer:          game.LayerPowerToughnessModify,
							PowerDelta:     1,
							ToughnessDelta: 2,
						},
					},
				},
			},
			OracleText: `
			This creature gets +1/+2 as long as an opponent controls a black permanent.
		`,
		},
	}
}
