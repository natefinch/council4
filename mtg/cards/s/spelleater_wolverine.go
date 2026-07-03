package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// SpelleaterWolverine is the card definition for Spelleater Wolverine.
//
// Type: Creature — Wolverine
// Cost: {2}{R}
//
// Oracle text:
//
//	This creature has double strike as long as there are three or more instant and/or sorcery cards in your graveyard.
var SpelleaterWolverine = newSpelleaterWolverine()

func newSpelleaterWolverine() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Spelleater Wolverine",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.R,
			}),
			Colors:    []color.Color{color.Red},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Wolverine},
			Power:     opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 2}),
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Condition: opt.Val(game.Condition{
						ControllerGraveyardInstantOrSorceryCountAtLeast: 3,
					}),
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Layer:          game.LayerAbility,
							AffectedSource: true,
							AddKeywords: []game.Keyword{
								game.DoubleStrike,
							},
						},
					},
				},
			},
			OracleText: `
			This creature has double strike as long as there are three or more instant and/or sorcery cards in your graveyard.
		`,
		},
	}
}
