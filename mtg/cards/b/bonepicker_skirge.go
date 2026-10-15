package b

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// BonepickerSkirge is the card definition for Bonepicker Skirge.
//
// Type: Creature — Phyrexian Imp
// Cost: {2}{B}
//
// Oracle text:
//
//	Flying
//	Corrupted — As long as an opponent has three or more poison counters, this creature has deathtouch and lifelink.
var BonepickerSkirge = newBonepickerSkirge

func newBonepickerSkirge() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Bonepicker Skirge",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.B,
			}),
			Colors:    []color.Color{color.Black},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Phyrexian, types.Imp},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 2}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
				game.StaticAbility{
					Text: "Corrupted — As long as an opponent has three or more poison counters, this creature has deathtouch and lifelink.",
					Condition: opt.Val(game.Condition{
						Text:                     "As long as an opponent has three or more poison counters",
						AnyOpponentPoisonAtLeast: 3,
						SourceCounterKind:        counter.PlusOnePlusOne,
					}),
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Duration:       game.DurationPermanent,
							AffectedSource: true,
							Layer:          game.LayerAbility,
							AddKeywords:    []game.Keyword{game.Deathtouch, game.Lifelink},
						},
					},
				},
			},
			OracleText: `
			Flying
			Corrupted — As long as an opponent has three or more poison counters, this creature has deathtouch and lifelink.
		`,
		},
	}
}
