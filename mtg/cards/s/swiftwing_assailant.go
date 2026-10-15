package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// SwiftwingAssailant is the card definition for Swiftwing Assailant.
//
// Type: Creature — Bird Warrior
// Cost: {3}{W}
//
// Oracle text:
//
//	Flying
//	Start your engines! (If you have no speed, it starts at 1. It increases once on each of your turns when an opponent loses life. Max speed is 4.)
//	Max speed — This creature gets +0/+1 and has vigilance.
var SwiftwingAssailant = newSwiftwingAssailant

func newSwiftwingAssailant() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Swiftwing Assailant",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.W,
			}),
			Colors:    []color.Color{color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Bird, types.Warrior},
			Power:     opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 3}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
				game.StaticAbility{
					Text: "Max speed — This creature gets +0/+1 and has vigilance.",
					Condition: opt.Val(game.Condition{
						SourceCounterKind:     counter.PlusOnePlusOne,
						ControllerHasMaxSpeed: true,
					}),
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Duration:       game.DurationPermanent,
							AffectedSource: true,
							Layer:          game.LayerPowerToughnessModify,
							ToughnessDelta: 1,
						},
						game.ContinuousEffect{
							Duration:       game.DurationPermanent,
							AffectedSource: true,
							Layer:          game.LayerAbility,
							AddKeywords:    []game.Keyword{game.Vigilance},
						},
					},
				},
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.StartEnginesTriggeredBody,
			},
			OracleText: `
			Flying
			Start your engines! (If you have no speed, it starts at 1. It increases once on each of your turns when an opponent loses life. Max speed is 4.)
			Max speed — This creature gets +0/+1 and has vigilance.
		`,
		},
	}
}
