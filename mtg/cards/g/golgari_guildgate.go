package g

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
)

// GolgariGuildgate is the card definition for Golgari Guildgate.
//
// Type: Land — Gate
//
// Oracle text:
//
//	This land enters tapped.
//	{T}: Add {B} or {G}.
var GolgariGuildgate = newGolgariGuildgate

func newGolgariGuildgate() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black, color.Green),
		CardFace: game.CardFace{
			Name:     "Golgari Guildgate",
			Types:    []types.Card{types.Land},
			Subtypes: []types.Sub{types.Gate},
			ManaAbilities: []game.ManaAbility{
				game.TapManaChoiceAbility(mana.B, mana.G),
			},
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntersTappedReplacement("This land enters tapped."),
			},
			OracleText: `
			This land enters tapped.
			{T}: Add {B} or {G}.
		`,
		},
	}
}
