package v

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
)

// ValgavothSLair is the card definition for Valgavoth's Lair.
//
// Type: Enchantment Land
//
// Oracle text:
//
//	Hexproof
//	This land enters tapped. As it enters, choose a color.
//	{T}: Add one mana of the chosen color.
var ValgavothSLair = newValgavothSLair

func newValgavothSLair() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:  "Valgavoth's Lair",
			Types: []types.Card{types.Enchantment, types.Land},
			StaticAbilities: []game.StaticAbility{
				game.HexproofStaticBody,
			},
			ManaAbilities: []game.ManaAbility{
				game.TapChosenColorManaAbility("{T}: Add one mana of the chosen color."),
			},
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntersTappedColorChoiceReplacement("This land enters tapped. As it enters, choose a color."),
			},
			OracleText: `
			Hexproof
			This land enters tapped. As it enters, choose a color.
			{T}: Add one mana of the chosen color.
		`,
		},
	}
}
