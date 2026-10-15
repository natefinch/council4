package o

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// ObeliskOfBant is the card definition for Obelisk of Bant.
//
// Type: Artifact
// Cost: {3}
//
// Oracle text:
//
//	{T}: Add {G}, {W}, or {U}.
var ObeliskOfBant = newObeliskOfBant

func newObeliskOfBant() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Blue, color.Green),
		CardFace: game.CardFace{
			Name: "Obelisk of Bant",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
			}),
			Types: []types.Card{types.Artifact},
			ManaAbilities: []game.ManaAbility{
				game.TapManaChoiceAbility(mana.G, mana.W, mana.U),
			},
			OracleText: `
			{T}: Add {G}, {W}, or {U}.
		`,
		},
	}
}
