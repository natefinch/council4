package f

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// FelhidePetrifier is the card definition for Felhide Petrifier.
//
// Type: Creature — Minotaur Warrior
// Cost: {2}{B}
//
// Oracle text:
//
//	Deathtouch
//	Other Minotaur creatures you control have deathtouch.
var FelhidePetrifier = newFelhidePetrifier

func newFelhidePetrifier() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Felhide Petrifier",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.B,
			}),
			Colors:    []color.Color{color.Black},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Minotaur, types.Warrior},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 3}),
			StaticAbilities: []game.StaticAbility{
				game.DeathtouchStaticBody,
				game.StaticAbility{
					Text: "Other Minotaur creatures you control have deathtouch.",
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Duration:    game.DurationPermanent,
							Group:       game.ObjectControlledGroupExcluding(game.SourcePermanentReference(), game.Selection{SubtypesAny: []types.Sub{types.Sub("Minotaur")}}, game.SourcePermanentReference()),
							Layer:       game.LayerAbility,
							AddKeywords: []game.Keyword{game.Deathtouch},
						},
					},
				},
			},
			OracleText: `
			Deathtouch
			Other Minotaur creatures you control have deathtouch.
		`,
		},
	}
}
