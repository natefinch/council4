package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
)

// SelesnyaGuildgate is the card definition for Selesnya Guildgate.
//
// Type: Land — Gate
//
// Oracle text:
//
//	This land enters tapped.
//	{T}: Add {G} or {W}.
var SelesnyaGuildgate = newSelesnyaGuildgate

func newSelesnyaGuildgate() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Green),
		CardFace: game.CardFace{
			Name:     "Selesnya Guildgate",
			Types:    []types.Card{types.Land},
			Subtypes: []types.Sub{types.Gate},
			ManaAbilities: []game.ManaAbility{
				game.TapManaChoiceAbility(mana.G, mana.W),
			},
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntersTappedReplacement("This land enters tapped."),
			},
			OracleText: `
			This land enters tapped.
			{T}: Add {G} or {W}.
		`,
		},
	}
}
