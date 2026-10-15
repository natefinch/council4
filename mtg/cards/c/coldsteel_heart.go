package c

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// ColdsteelHeart is the card definition for Coldsteel Heart.
//
// Type: Snow Artifact
// Cost: {2}
//
// Oracle text:
//
//	This artifact enters tapped.
//	As this artifact enters, choose a color.
//	{T}: Add one mana of the chosen color.
var ColdsteelHeart = newColdsteelHeart

func newColdsteelHeart() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Coldsteel Heart",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
			}),
			Supertypes: []types.Super{types.Snow},
			Types:      []types.Card{types.Artifact},
			ManaAbilities: []game.ManaAbility{
				game.TapChosenColorManaAbility("{T}: Add one mana of the chosen color."),
			},
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntersTappedReplacement("This artifact enters tapped."),
				game.EntryColorChoiceReplacement("As this artifact enters, choose a color."),
			},
			OracleText: `
			This artifact enters tapped.
			As this artifact enters, choose a color.
			{T}: Add one mana of the chosen color.
		`,
		},
	}
}
