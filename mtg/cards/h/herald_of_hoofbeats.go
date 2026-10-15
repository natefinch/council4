package h

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// HeraldOfHoofbeats is the card definition for Herald of Hoofbeats.
//
// Type: Creature — Human Knight
// Cost: {3}{U}
//
// Oracle text:
//
//	Horsemanship (This creature can't be blocked except by creatures with horsemanship.)
//	Other Knights you control have horsemanship.
var HeraldOfHoofbeats = newHeraldOfHoofbeats

func newHeraldOfHoofbeats() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Herald of Hoofbeats",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.U,
			}),
			Colors:    []color.Color{color.Blue},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Human, types.Knight},
			Power:     opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 3}),
			StaticAbilities: []game.StaticAbility{
				game.HorsemanshipStaticBody,
				game.StaticAbility{
					Text: "Other Knights you control have horsemanship.",
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Duration:    game.DurationPermanent,
							Group:       game.ObjectControlledGroupExcluding(game.SourcePermanentReference(), game.Selection{SubtypesAny: []types.Sub{types.Sub("Knight")}}, game.SourcePermanentReference()),
							Layer:       game.LayerAbility,
							AddKeywords: []game.Keyword{game.Horsemanship},
						},
					},
				},
			},
			OracleText: `
			Horsemanship (This creature can't be blocked except by creatures with horsemanship.)
			Other Knights you control have horsemanship.
		`,
		},
	}
}
