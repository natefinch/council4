package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// SolGrail is the card definition for Sol Grail.
//
// Type: Artifact
// Cost: {3}
//
// Oracle text:
//
//	As this artifact enters, choose a color.
//	{T}: Add one mana of the chosen color.
var SolGrail = newSolGrail

func newSolGrail() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Sol Grail",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
			}),
			Types: []types.Card{types.Artifact},
			ManaAbilities: []game.ManaAbility{
				game.TapChosenColorManaAbility("{T}: Add one mana of the chosen color."),
			},
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntryColorChoiceReplacement("As this artifact enters, choose a color."),
			},
			OracleText: `
			As this artifact enters, choose a color.
			{T}: Add one mana of the chosen color.
		`,
		},
	}
}
