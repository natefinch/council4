package h

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
)

// Heartwood is the card definition for Heartwood.
//
// Type: Token Artifact — Heartwood
//
// Oracle text:
//   {T}: Add {R} or {G}.

// HeartwoodTokened1d8118b182413b99ba52844d35a35b is the card definition for Heartwood.
var HeartwoodTokened1d8118b182413b99ba52844d35a35b = newHeartwoodTokened1d8118b182413b99ba52844d35a35b()

func newHeartwoodTokened1d8118b182413b99ba52844d35a35b() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red, color.Green),
		CardFace: game.CardFace{
			Name:     "Heartwood",
			Colors:   []color.Color{color.Green, color.Red},
			Types:    []types.Card{types.Artifact},
			Subtypes: []types.Sub{types.Sub("Heartwood")},
			ManaAbilities: []game.ManaAbility{
				game.TapManaChoiceAbility(mana.R, mana.G),
			},
			OracleText: `
			{T}: Add {R} or {G}.
		`,
		},
	}
}
