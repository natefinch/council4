package w

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// WitnessOfTheAges is the card definition for Witness of the Ages.
//
// Type: Artifact Creature — Golem
// Cost: {6}
//
// Oracle text:
//
//	Morph {5} (You may cast this card face down as a 2/2 creature for {3}. Turn it face up any time for its morph cost.)
var WitnessOfTheAges = newWitnessOfTheAges

func newWitnessOfTheAges() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Witness of the Ages",
			ManaCost: opt.Val(cost.Mana{
				cost.O(6),
			}),
			Types:     []types.Card{types.Artifact, types.Creature},
			Subtypes:  []types.Sub{types.Golem},
			Power:     opt.Val(game.PT{Value: 4}),
			Toughness: opt.Val(game.PT{Value: 4}),
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "Morph {5}",
					KeywordAbilities: []game.KeywordAbility{
						game.MorphKeyword{
							Cost: cost.Mana{
								cost.O(5),
							},
						},
					},
				},
			},
			OracleText: `
			Morph {5} (You may cast this card face down as a 2/2 creature for {3}. Turn it face up any time for its morph cost.)
		`,
		},
	}
}
