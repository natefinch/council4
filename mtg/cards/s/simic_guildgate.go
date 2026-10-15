package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
)

// SimicGuildgate is the card definition for Simic Guildgate.
//
// Type: Land — Gate
//
// Oracle text:
//
//	This land enters tapped.
//	{T}: Add {G} or {U}.
var SimicGuildgate = newSimicGuildgate

func newSimicGuildgate() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue, color.Green),
		CardFace: game.CardFace{
			Name:     "Simic Guildgate",
			Types:    []types.Card{types.Land},
			Subtypes: []types.Sub{types.Gate},
			ManaAbilities: []game.ManaAbility{
				game.TapManaChoiceAbility(mana.G, mana.U),
			},
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntersTappedReplacement("This land enters tapped."),
			},
			OracleText: `
			This land enters tapped.
			{T}: Add {G} or {U}.
		`,
		},
	}
}
