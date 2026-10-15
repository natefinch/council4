package o

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// ObeliskOfNaya is the card definition for Obelisk of Naya.
//
// Type: Artifact
// Cost: {3}
//
// Oracle text:
//
//	{T}: Add {R}, {G}, or {W}.
var ObeliskOfNaya = newObeliskOfNaya

func newObeliskOfNaya() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Red, color.Green),
		CardFace: game.CardFace{
			Name: "Obelisk of Naya",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
			}),
			Types: []types.Card{types.Artifact},
			ManaAbilities: []game.ManaAbility{
				game.TapManaChoiceAbility(mana.R, mana.G, mana.W),
			},
			OracleText: `
			{T}: Add {R}, {G}, or {W}.
		`,
		},
	}
}
