package c

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// CutthroatIlDal is the card definition for Cutthroat il-Dal.
//
// Type: Creature — Human Rogue
// Cost: {3}{B}
//
// Oracle text:
//
//	Hellbent — This creature has shadow as long as you have no cards in hand. (It can block or be blocked by only creatures with shadow.)
var CutthroatIlDal = newCutthroatIlDal

func newCutthroatIlDal() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Cutthroat il-Dal",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.B,
			}),
			Colors:    []color.Color{color.Black},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Human, types.Rogue},
			Power:     opt.Val(game.PT{Value: 4}),
			Toughness: opt.Val(game.PT{Value: 1}),
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "Hellbent — This creature has shadow as long as you have no cards in hand. (It can block or be blocked by only creatures with shadow.)",
					Condition: opt.Val(game.Condition{
						Text:                "as long as you have no cards in hand",
						ControllerHandEmpty: true,
						SourceCounterKind:   counter.PlusOnePlusOne,
					}),
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Duration:       game.DurationPermanent,
							AffectedSource: true,
							Layer:          game.LayerAbility,
							AddKeywords:    []game.Keyword{game.Shadow},
						},
					},
				},
			},
			OracleText: `
			Hellbent — This creature has shadow as long as you have no cards in hand. (It can block or be blocked by only creatures with shadow.)
		`,
		},
	}
}
