package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// ScornEffigy is the card definition for Scorn Effigy.
//
// Type: Artifact Creature — Scarecrow
// Cost: {3}
//
// Oracle text:
//
//	Foretell {0} (During your turn, you may pay {2} and exile this card from your hand face down. Cast it on a later turn for its foretell cost.)
var ScornEffigy = newScornEffigy

func newScornEffigy() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Scorn Effigy",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
			}),
			Types:     []types.Card{types.Artifact, types.Creature},
			Subtypes:  []types.Sub{types.Scarecrow},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 3}),
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "Foretell {0}",
					KeywordAbilities: []game.KeywordAbility{
						game.ForetellKeyword{
							Cost: cost.Mana{
								cost.O(0),
							},
						},
					},
				},
			},
			OracleText: `
			Foretell {0} (During your turn, you may pay {2} and exile this card from your hand face down. Cast it on a later turn for its foretell cost.)
		`,
		},
	}
}
