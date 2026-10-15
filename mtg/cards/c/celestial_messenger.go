package c

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// CelestialMessenger is the card definition for Celestial Messenger.
//
// Type: Creature — Bird Spirit
// Cost: {2}{U}{U}
//
// Oracle text:
//
//	Flash (You may cast this card any time you could cast an instant.)
//	Flying
//	This creature gets +1/+1 as long as you control a Yanling planeswalker.
var CelestialMessenger = newCelestialMessenger

func newCelestialMessenger() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Celestial Messenger",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.U,
				cost.U,
			}),
			Colors:    []color.Color{color.Blue},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Bird, types.Spirit},
			Power:     opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 2}),
			StaticAbilities: []game.StaticAbility{
				game.FlashStaticBody,
				game.FlyingStaticBody,
				game.StaticAbility{
					Text: "This creature gets +1/+1 as long as you control a Yanling planeswalker.",
					Condition: opt.Val(game.Condition{
						Text: "as long as you control a Yanling planeswalker",
						ControlsMatching: opt.Val(game.SelectionCount{
							Selection: game.Selection{RequiredTypes: []types.Card{types.Planeswalker}, SubtypesAny: []types.Sub{types.Yanling}},
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
					},
				},
			},
			OracleText: `
			Flash (You may cast this card any time you could cast an instant.)
			Flying
			This creature gets +1/+1 as long as you control a Yanling planeswalker.
		`,
		},
	}
}
