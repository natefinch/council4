package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// SejiriMerfolk is the card definition for Sejiri Merfolk.
//
// Type: Creature — Merfolk Soldier
// Cost: {1}{U}
//
// Oracle text:
//
//	As long as you control a Plains, this creature has first strike and lifelink. (Damage dealt by a creature with lifelink also causes its controller to gain that much life.)
var SejiriMerfolk = newSejiriMerfolk

func newSejiriMerfolk() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Sejiri Merfolk",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.U,
			}),
			Colors:    []color.Color{color.Blue},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Merfolk, types.Soldier},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 1}),
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "As long as you control a Plains, this creature has first strike and lifelink. (Damage dealt by a creature with lifelink also causes its controller to gain that much life.)",
					Condition: opt.Val(game.Condition{
						Text: "As long as you control a Plains",
						ControlsMatching: opt.Val(game.SelectionCount{
							Selection: game.Selection{SubtypesAny: []types.Sub{types.Plains}},
						}),
						SourceCounterKind: counter.PlusOnePlusOne,
					}),
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Duration:       game.DurationPermanent,
							AffectedSource: true,
							Layer:          game.LayerAbility,
							AddKeywords:    []game.Keyword{game.FirstStrike, game.Lifelink},
						},
					},
				},
			},
			OracleText: `
			As long as you control a Plains, this creature has first strike and lifelink. (Damage dealt by a creature with lifelink also causes its controller to gain that much life.)
		`,
		},
	}
}
