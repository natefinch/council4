package d

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
)

// DakmorSalvage is the card definition for Dakmor Salvage.
//
// Type: Land
//
// Oracle text:
//
//	This land enters tapped.
//	{T}: Add {B}.
//	Dredge 2 (If you would draw a card, you may mill two cards instead. If you do, return this card from your graveyard to your hand.)
var DakmorSalvage = newDakmorSalvage

func newDakmorSalvage() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name:  "Dakmor Salvage",
			Types: []types.Card{types.Land},
			StaticAbilities: []game.StaticAbility{
				game.DredgeStaticAbility(2),
			},
			ManaAbilities: []game.ManaAbility{
				game.TapManaAbility(mana.B),
			},
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntersTappedReplacement("This land enters tapped."),
			},
			OracleText: `
			This land enters tapped.
			{T}: Add {B}.
			Dredge 2 (If you would draw a card, you may mill two cards instead. If you do, return this card from your graveyard to your hand.)
		`,
		},
	}
}
