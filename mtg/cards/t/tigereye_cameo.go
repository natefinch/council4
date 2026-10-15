package t

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// TigereyeCameo is the card definition for Tigereye Cameo.
//
// Type: Artifact
// Cost: {3}
//
// Oracle text:
//
//	{T}: Add {G} or {W}.
var TigereyeCameo = newTigereyeCameo

func newTigereyeCameo() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Green),
		CardFace: game.CardFace{
			Name: "Tigereye Cameo",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
			}),
			Types: []types.Card{types.Artifact},
			ManaAbilities: []game.ManaAbility{
				game.TapManaChoiceAbility(mana.G, mana.W),
			},
			OracleText: `
			{T}: Add {G} or {W}.
		`,
		},
	}
}
