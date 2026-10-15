package h

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Horse is the card definition for Horse.
//
// Type: Token Creature — Horse
//
// Oracle text:
//   Doctors you control have horsemanship. (They can't be blocked except by creatures with horsemanship.)

// HorseToken79029843d4d3483fb2bdc5f7fcaa2ee1 is the card definition for Horse.
var HorseToken79029843d4d3483fb2bdc5f7fcaa2ee1 = newHorseToken79029843d4d3483fb2bdc5f7fcaa2ee1()

func newHorseToken79029843d4d3483fb2bdc5f7fcaa2ee1() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name:      "Horse",
			Colors:    []color.Color{color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Horse},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 2}),
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "Doctors you control have horsemanship. (They can't be blocked except by creatures with horsemanship.)",
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Duration:    game.DurationPermanent,
							Group:       game.ObjectControlledGroup(game.SourcePermanentReference(), game.Selection{SubtypesAny: []types.Sub{types.Sub("Doctor")}}),
							Layer:       game.LayerAbility,
							AddKeywords: []game.Keyword{game.Horsemanship},
						},
					},
				},
			},
			OracleText: `
			Doctors you control have horsemanship. (They can't be blocked except by creatures with horsemanship.)
		`,
		},
	}
}
