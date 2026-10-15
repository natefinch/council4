package i

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
)

// InsomniaCrownCity is the card definition for Insomnia, Crown City.
//
// Type: Land — Town
//
// Oracle text:
//
//	This land enters tapped.
//	{T}: Add {W} or {B}.
var InsomniaCrownCity = newInsomniaCrownCity

func newInsomniaCrownCity() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Black),
		CardFace: game.CardFace{
			Name:     "Insomnia, Crown City",
			Types:    []types.Card{types.Land},
			Subtypes: []types.Sub{types.Town},
			ManaAbilities: []game.ManaAbility{
				game.TapManaChoiceAbility(mana.W, mana.B),
			},
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntersTappedReplacement("This land enters tapped."),
			},
			OracleText: `
			This land enters tapped.
			{T}: Add {W} or {B}.
		`,
		},
	}
}
