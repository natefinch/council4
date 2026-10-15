package z

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
)

// ZoeticCavern is the card definition for Zoetic Cavern.
//
// Type: Land
//
// Oracle text:
//
//	{T}: Add {C}.
//	Morph {2} (You may cast this card face down as a 2/2 creature for {3}. Turn it face up any time for its morph cost.)
var ZoeticCavern = newZoeticCavern

func newZoeticCavern() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:  "Zoetic Cavern",
			Types: []types.Card{types.Land},
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "Morph {2}",
					KeywordAbilities: []game.KeywordAbility{
						game.MorphKeyword{
							Cost: cost.Mana{
								cost.O(2),
							},
						},
					},
				},
			},
			ManaAbilities: []game.ManaAbility{
				game.TapManaAbility(mana.C),
			},
			OracleText: `
			{T}: Add {C}.
			Morph {2} (You may cast this card face down as a 2/2 creature for {3}. Turn it face up any time for its morph cost.)
		`,
		},
	}
}
