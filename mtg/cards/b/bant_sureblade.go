package b

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// BantSureblade is the card definition for Bant Sureblade.
//
// Type: Creature — Human Soldier
// Cost: {G/U}{W}
//
// Oracle text:
//
//	As long as you control another multicolored permanent, this creature gets +1/+1 and has first strike.
var BantSureblade = newBantSureblade

func newBantSureblade() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Blue, color.Green),
		CardFace: game.CardFace{
			Name: "Bant Sureblade",
			ManaCost: opt.Val(cost.Mana{
				cost.HybridMana(mana.G, mana.U),
				cost.W,
			}),
			Colors:    []color.Color{color.Green, color.Blue, color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Human, types.Soldier},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 1}),
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "As long as you control another multicolored permanent, this creature gets +1/+1 and has first strike.",
					Condition: opt.Val(game.Condition{
						Text: "As long as you control another multicolored permanent",
						ControlsMatching: opt.Val(game.SelectionCount{
							Selection: game.Selection{Multicolored: true, ExcludeSource: true},
						}),
						SourceCounterKind: counter.PlusOnePlusOne,
					}),
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Duration:       game.DurationPermanent,
							AffectedSource: true,
							Layer:          game.LayerPowerToughnessModify,
							PowerDelta:     1,
							ToughnessDelta: 1,
						},
						game.ContinuousEffect{
							Duration:       game.DurationPermanent,
							AffectedSource: true,
							Layer:          game.LayerAbility,
							AddKeywords:    []game.Keyword{game.FirstStrike},
						},
					},
				},
			},
			OracleText: `
			As long as you control another multicolored permanent, this creature gets +1/+1 and has first strike.
		`,
		},
	}
}
